package transformer

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/genkit-migrate/genkit-migrate/pkg/models"
)

type Transformer struct {
	config *Config
}

type Config struct {
	SourceProvider string
	TargetProvider string
	TargetPath     string
	DryRun         bool
}

func New(config *Config) *Transformer {
	return &Transformer{config: config}
}

func (t *Transformer) TransformProject(ctx context.Context, project *models.Project) (*models.Migration, error) {
	migration := &models.Migration{
		Project:     project,
		Changes:     make([]*models.Change, 0),
		NewFiles:    make(map[string]string),
		DeleteFiles: make([]string, 0),
		Commands:    make([]string, 0),
	}

	err := t.transformDependencies(migration)
	if err != nil {
		return nil, fmt.Errorf("failed to transform dependencies: %w", err)
	}

	err = t.transformSourceFiles(migration)
	if err != nil {
		return nil, fmt.Errorf("failed to transform source files: %w", err)
	}

	err = t.transformModels(migration)
	if err != nil {
		return nil, fmt.Errorf("failed to transform models: %w", err)
	}

	err = t.transformConfiguration(migration)
	if err != nil {
		return nil, fmt.Errorf("failed to transform configuration: %w", err)
	}

	err = t.generateDeploymentFiles(migration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate deployment files: %w", err)
	}

	return migration, nil
}

func (t *Transformer) transformDependencies(migration *models.Migration) error {
	project := migration.Project

	goModTemplate := `module {{ .ModuleName }}

go {{ .GoVersion }}

require (
	github.com/firebase/genkit/go v1.0.2
{{- if eq .TargetProvider "aws" }}
	github.com/scttfrdmn/genkit-aws v0.1.0
{{- end }}
{{- range .Dependencies }}
	{{ .Name }} {{ .Version }}
{{- end }}
)
`

	tmpl, err := template.New("go.mod").Parse(goModTemplate)
	if err != nil {
		return err
	}

	var content strings.Builder
	err = tmpl.Execute(&content, map[string]interface{}{
		"ModuleName":     t.extractModuleName(project),
		"GoVersion":      "1.23",
		"TargetProvider": t.config.TargetProvider,
		"Dependencies":   t.filterDependencies(project.Dependencies),
	})
	if err != nil {
		return err
	}

	migration.NewFiles["go.mod"] = content.String()

	migration.Changes = append(migration.Changes, &models.Change{
		Type:        "dependency",
		Description: fmt.Sprintf("Updated dependencies for %s", t.config.TargetProvider),
		File:        "go.mod",
	})

	return nil
}

func (t *Transformer) transformSourceFiles(migration *models.Migration) error {
	project := migration.Project

	for filePath, sourceFile := range project.Files {
		if !sourceFile.HasGenKit {
			continue
		}

		newContent, changes, err := t.transformGoFile(sourceFile)
		if err != nil {
			return fmt.Errorf("failed to transform %s: %w", filePath, err)
		}

		if len(changes) > 0 {
			migration.NewFiles[filePath] = newContent
			migration.Changes = append(migration.Changes, changes...)
		}
	}

	return nil
}

func (t *Transformer) transformGoFile(sourceFile *models.SourceFile) (string, []*models.Change, error) {
	changes := make([]*models.Change, 0)

	content := `package ` + sourceFile.PackageName + `

import (
	"context"
	"github.com/firebase/genkit/go/genkit"
`

	if t.config.TargetProvider == "aws" {
		content += `	genkitaws "github.com/scttfrdmn/genkit-aws/pkg/genkit-aws"
	"github.com/scttfrdmn/genkit-aws/pkg/bedrock"
	"github.com/scttfrdmn/genkit-aws/pkg/monitoring"
`
	}

	content += `)

// TODO: Implement actual AST transformation
// This is a simplified example showing the structure
`

	changes = append(changes, &models.Change{
		Type:        "import",
		Description: fmt.Sprintf("Added %s imports", t.config.TargetProvider),
		File:        sourceFile.Path,
	})

	return content, changes, nil
}

func (t *Transformer) transformModels(migration *models.Migration) error {
	modelMappings := t.getModelMappings()

	for _, model := range migration.Project.Models {
		if newModel, exists := modelMappings[model.Name]; exists {
			migration.Changes = append(migration.Changes, &models.Change{
				Type:        "model",
				Description: fmt.Sprintf("Map model %s -> %s", model.Name, newModel),
				File:        model.Position.Filename,
			})
		}
	}

	return nil
}

func (t *Transformer) getModelMappings() map[string]string {
	if t.config.SourceProvider == "gcp" && t.config.TargetProvider == "aws" {
		return map[string]string{
			"googleai/gemini-1.5-flash": "anthropic.claude-3-haiku-20240307-v1:0",
			"googleai/gemini-1.5-pro":   "anthropic.claude-3-sonnet-20240229-v1:0",
			"googleai/gemini-2.0-flash": "anthropic.claude-3-5-sonnet-20241022-v2:0",
			"vertexai/gemini-pro":       "anthropic.claude-3-sonnet-20240229-v1:0",
			"vertexai/gemini-1.5-pro":   "anthropic.claude-3-sonnet-20240229-v1:0",
			"vertexai/gemini-1.5-flash": "anthropic.claude-3-haiku-20240307-v1:0",
			// Map some models to Amazon Nova for variety
			"googleai/gemini-1.5-flash-8b": "amazon.nova-lite-v1:0",
			"googleai/text-bison":          "amazon.nova-micro-v1:0",
		}
	}
	return make(map[string]string)
}

func (t *Transformer) transformConfiguration(migration *models.Migration) error {
	if t.config.TargetProvider == "aws" {
		configTemplate := `# AWS Configuration for GenKit
region: us-east-1
profile: default

bedrock:
  models:
    - anthropic.claude-3-sonnet-20240229-v1:0
    - amazon.nova-pro-v1:0
  
cloudwatch:
  namespace: "GenKit/{{ .ProjectName }}"
  enabled: true

# Environment variables
environment:
  - GENKIT_ENV=production
  - AWS_REGION=us-east-1
`

		tmpl, err := template.New("config").Parse(configTemplate)
		if err != nil {
			return err
		}

		var content strings.Builder
		err = tmpl.Execute(&content, map[string]interface{}{
			"ProjectName": t.extractProjectName(migration.Project),
		})
		if err != nil {
			return err
		}

		migration.NewFiles["config.yaml"] = content.String()
	}

	return nil
}

func (t *Transformer) generateDeploymentFiles(migration *models.Migration) error {
	if t.config.TargetProvider == "aws" {
		err := t.generateTerraform(migration)
		if err != nil {
			return err
		}

		err = t.generateDockerfile(migration)
		if err != nil {
			return err
		}

		err = t.generateGitHubActions(migration)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Transformer) generateTerraform(migration *models.Migration) error {
	terraformMain := `# Terraform configuration for GenKit on AWS
terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

# Lambda function for GenKit app
resource "aws_lambda_function" "genkit_app" {
  filename         = "genkit-app.zip"
  function_name    = "genkit-app"
  role            = aws_iam_role.lambda_role.arn
  handler         = "main"
  runtime         = "provided.al2"
  
  environment {
    variables = {
      GENKIT_ENV = "production"
      AWS_REGION = var.aws_region
    }
  }
}

# IAM role for Lambda
resource "aws_iam_role" "lambda_role" {
  name = "genkit-lambda-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

# IAM policy for Bedrock access
resource "aws_iam_role_policy" "bedrock_policy" {
  name = "genkit-bedrock-policy"
  role = aws_iam_role.lambda_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "bedrock:InvokeModel",
          "bedrock:InvokeModelWithResponseStream"
        ]
        Resource = "*"
      }
    ]
  })
}
`

	migration.NewFiles["terraform/main.tf"] = terraformMain

	terraformVars := `variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Project name"
  type        = string
  default     = "genkit-app"
}
`

	migration.NewFiles["terraform/variables.tf"] = terraformVars

	return nil
}

func (t *Transformer) generateDockerfile(migration *models.Migration) error {
	dockerfile := `# Multi-stage build for GenKit Go app
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8080
CMD ["./main"]
`

	migration.NewFiles["Dockerfile"] = dockerfile
	return nil
}

func (t *Transformer) generateGitHubActions(migration *models.Migration) error {
	workflow := `name: Deploy to AWS

on:
  push:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v4
      with:
        go-version: '1.23'
    - run: go test -v ./...

  deploy:
    needs: test
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: aws-actions/configure-aws-credentials@v4
      with:
        aws-access-key-id: ${{ "{{ secrets.AWS_ACCESS_KEY_ID }}" }}
        aws-secret-access-key: ${{ "{{ secrets.AWS_SECRET_ACCESS_KEY }}" }}
        aws-region: us-east-1
    
    - name: Deploy with Terraform
      run: |
        cd terraform
        terraform init
        terraform plan
        terraform apply -auto-approve
`

	migration.NewFiles[".github/workflows/deploy.yml"] = workflow
	return nil
}

func (t *Transformer) extractModuleName(project *models.Project) string {
	return "genkit-app"
}

func (t *Transformer) extractProjectName(project *models.Project) string {
	return "GenKitApp"
}

func (t *Transformer) filterDependencies(deps map[string]string) []map[string]string {
	filtered := make([]map[string]string, 0)
	for name, version := range deps {
		if !strings.Contains(name, "firebase") && !strings.Contains(name, "google") {
			filtered = append(filtered, map[string]string{
				"Name":    name,
				"Version": version,
			})
		}
	}
	return filtered
}
