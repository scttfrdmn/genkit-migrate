package generator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/genkit-migrate/genkit-migrate/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go/token"
)

func TestGenerateProject(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "genkit-generator-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a temporary source directory with test files
	sourceDir, err := os.MkdirTemp("", "genkit-source-*")
	require.NoError(t, err)
	defer os.RemoveAll(sourceDir)

	// Create test source file
	err = os.WriteFile(filepath.Join(sourceDir, "main.go"), []byte("package main\n"), 0644)
	require.NoError(t, err)

	generator := New(&Config{
		TargetProvider: "aws",
		OutputPath:     tempDir,
	})

	migration := &models.Migration{
		Project: &models.Project{
			Path:           sourceDir,
			SourceProvider: "gcp",
			TargetProvider: "aws",
			Files: map[string]*models.SourceFile{
				"main.go": {
					Path:        filepath.Join(sourceDir, "main.go"),
					PackageName: "main",
					HasGenKit:   true,
				},
			},
			Flows: []*models.Flow{
				{Name: "summarize", Position: token.Position{}},
			},
			Models: []*models.Model{
				{Name: "googleai/gemini-1.5-pro", Provider: "gcp", Position: token.Position{}},
			},
		},
		Changes: []*models.Change{
			{Type: "dependency", Description: "Updated dependencies for aws", File: "go.mod"},
		},
		NewFiles: map[string]string{
			"go.mod":                       "module test-app\n\ngo 1.23",
			"config.yaml":                  "region: us-east-1",
			"terraform/main.tf":            "# Terraform configuration",
			"Dockerfile":                   "FROM golang:1.23",
			".github/workflows/deploy.yml": "name: Deploy",
		},
		DeleteFiles: []string{},
		Commands:    []string{},
	}

	err = generator.GenerateProject(context.Background(), migration)
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(tempDir, "go.mod"))
	assert.FileExists(t, filepath.Join(tempDir, "config.yaml"))
	assert.FileExists(t, filepath.Join(tempDir, "terraform", "main.tf"))
	assert.FileExists(t, filepath.Join(tempDir, "Dockerfile"))
	assert.FileExists(t, filepath.Join(tempDir, ".github", "workflows", "deploy.yml"))
	assert.FileExists(t, filepath.Join(tempDir, "MIGRATION.md"))

	content, err := os.ReadFile(filepath.Join(tempDir, "go.mod"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "module test-app")

	migrationDoc, err := os.ReadFile(filepath.Join(tempDir, "MIGRATION.md"))
	require.NoError(t, err)
	assert.Contains(t, string(migrationDoc), "GenKit Migration to aws")
	assert.Contains(t, string(migrationDoc), "Changes Applied**: 1")
}

func TestGenerateReadme(t *testing.T) {
	generator := New(&Config{
		TargetProvider: "aws",
	})

	migration := &models.Migration{
		Project: &models.Project{
			SourceProvider: "gcp",
			TargetProvider: "aws",
			Flows: []*models.Flow{
				{Name: "summarize"},
				{Name: "translate"},
			},
			Models: []*models.Model{
				{Name: "googleai/gemini-1.5-pro", Provider: "gcp"},
			},
		},
		Changes: []*models.Change{
			{Type: "dependency", Description: "Updated dependencies", File: "go.mod"},
			{Type: "model", Description: "Mapped model", File: "main.go"},
		},
	}

	readme := generator.generateReadme(migration)

	assert.Contains(t, readme, "GenKit Migration to aws")
	assert.Contains(t, readme, "Source Provider**: gcp")
	assert.Contains(t, readme, "Target Provider**: aws")
	assert.Contains(t, readme, "Flows Found**: 2")
	assert.Contains(t, readme, "Models Found**: 1")
	assert.Contains(t, readme, "Changes Applied**: 2")
	assert.Contains(t, readme, "AWS Deployment")
	assert.Contains(t, readme, "terraform init")
	assert.Contains(t, readme, "Model Mappings Applied")
}
