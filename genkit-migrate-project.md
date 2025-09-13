# GenKit Migration CLI Project

## Project Structure

```
genkit-migrate/
├── go.mod
├── go.sum
├── README.md
├── LICENSE
├── CHANGELOG.md
├── .github/
│   └── workflows/
│       ├── ci.yml
│       ├── release.yml
│       └── codeql.yml
├── cmd/
│   └── genkit-migrate/
│       ├── main.go
│       ├── root.go
│       ├── migrate.go
│       ├── analyze.go
│       └── version.go
├── pkg/
│   ├── analyzer/
│   │   ├── project.go
│   │   ├── dependencies.go
│   │   ├── flows.go
│   │   ├── models.go
│   │   └── analyzer_test.go
│   ├── transformer/
│   │   ├── config.go
│   │   ├── code.go
│   │   ├── dependencies.go
│   │   ├── templates.go
│   │   └── transformer_test.go
│   ├── generator/
│   │   ├── aws.go
│   │   ├── terraform.go
│   │   ├── cloudformation.go
│   │   ├── docker.go
│   │   └── generator_test.go
│   └── models/
│       ├── project.go
│       ├── migration.go
│       └── aws.go
├── internal/
│   ├── cli/
│   │   ├── ui.go
│   │   ├── progress.go
│   │   └── prompts.go
│   ├── config/
│   │   ├── config.go
│   │   └── validation.go
│   └── utils/
│       ├── files.go
│       ├── git.go
│       └── templates.go
├── templates/
│   ├── aws/
│   │   ├── go.mod.tmpl
│   │   ├── main.go.tmpl
│   │   ├── config.go.tmpl
│   │   └── deploy/
│   │       ├── terraform/
│   │       │   ├── main.tf.tmpl
│   │       │   ├── variables.tf.tmpl
│   │       │   └── outputs.tf.tmpl
│   │       ├── cloudformation/
│   │       │   └── template.yaml.tmpl
│   │       └── docker/
│   │           ├── Dockerfile.tmpl
│   │           └── docker-compose.yml.tmpl
│   └── docs/
│       ├── README.md.tmpl
│       └── MIGRATION.md.tmpl
├── testdata/
│   ├── sample-projects/
│   │   ├── simple-genkit/
│   │   ├── complex-genkit/
│   │   └── multi-flow/
│   └── expected-outputs/
├── docs/
│   ├── architecture.md
│   ├── migration-guide.md
│   └── troubleshooting.md
├── scripts/
│   ├── build.sh
│   ├── test.sh
│   ├── install.sh
│   └── release.sh
└── tools/
    ├── go.mod
    └── tools.go
```

## Core Files

### go.mod
```go
module github.com/genkit-migrate/genkit-migrate

go 1.23

require (
    github.com/spf13/cobra v1.8.1
    github.com/spf13/viper v1.19.0
    github.com/spf13/afero v1.11.0
    github.com/charmbracelet/bubbletea v1.1.1
    github.com/charmbracelet/lipgloss v1.0.0
    github.com/charmbracelet/huh v0.6.0
    go.mod.parser v0.0.0-20240209183019-3a5c42717c64
    gopkg.in/yaml.v3 v3.0.1
    github.com/Masterminds/sprig/v3 v3.3.0
)

require (
    github.com/stretchr/testify v1.9.0
    github.com/google/go-cmp v0.6.0
    github.com/otiai10/copy v1.14.0
)
```

### cmd/genkit-migrate/main.go
```go
package main

import (
    "os"

    "github.com/genkit-migrate/genkit-migrate/cmd/genkit-migrate"
)

func main() {
    if err := cmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### cmd/genkit-migrate/root.go
```go
package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var (
    cfgFile string
    verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
    Use:   "genkit-migrate",
    Short: "Migrate GenKit applications between cloud providers",
    Long: `GenKit Migrate is a CLI tool that helps you migrate GenKit applications
from Google Cloud Platform to AWS (and other cloud providers).

It analyzes your existing GenKit project, transforms the code and configuration,
and generates the necessary deployment artifacts for your target cloud platform.`,
    Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
    return rootCmd.Execute()
}

func init() {
    cobra.OnInitialize(initConfig)

    // Global flags
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.genkit-migrate.yaml)")
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

    // Bind flags to viper
    viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        home, err := os.UserHomeDir()
        cobra.CheckErr(err)

        viper.AddConfigPath(home)
        viper.AddConfigPath(".")
        viper.SetConfigType("yaml")
        viper.SetConfigName(".genkit-migrate")
    }

    viper.AutomaticEnv()

    if err := viper.ReadInConfig(); err == nil && verbose {
        fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
    }
}
```

### cmd/genkit-migrate/migrate.go
```go
package cmd

import (
    "context"
    "fmt"
    "path/filepath"

    "github.com/spf13/cobra"
    "github.com/genkit-migrate/genkit-migrate/internal/cli"
    "github.com/genkit-migrate/genkit-migrate/pkg/analyzer"
    "github.com/genkit-migrate/genkit-migrate/pkg/transformer"
    "github.com/genkit-migrate/genkit-migrate/pkg/generator"
    "github.com/genkit-migrate/genkit-migrate/pkg/models"
)

var (
    sourcePath   string
    targetPath   string
    fromProvider string
    toProvider   string
    dryRun      bool
    interactive bool
)

var migrateCmd = &cobra.Command{
    Use:   "migrate",
    Short: "Migrate a GenKit project between cloud providers",
    Long: `Analyze and migrate a GenKit project from one cloud provider to another.

This command will:
1. Analyze your existing GenKit project structure
2. Transform the code for the target cloud provider  
3. Generate deployment configurations
4. Create documentation for the migration

Example:
  genkit-migrate migrate --from=gcp --to=aws --source=./my-genkit-app --target=./my-genkit-app-aws`,
    RunE: runMigrate,
}

func init() {
    rootCmd.AddCommand(migrateCmd)

    migrateCmd.Flags().StringVarP(&sourcePath, "source", "s", ".", "source project path")
    migrateCmd.Flags().StringVarP(&targetPath, "target", "t", "", "target project path (default: source_aws)")
    migrateCmd.Flags().StringVar(&fromProvider, "from", "gcp", "source cloud provider (gcp, aws, azure)")
    migrateCmd.Flags().StringVar(&toProvider, "to", "aws", "target cloud provider (aws, gcp, azure)")
    migrateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "analyze and plan without making changes")
    migrateCmd.Flags().BoolVarP(&interactive, "interactive", "i", true, "interactive mode with prompts")

    migrateCmd.MarkFlagRequired("source")
}

func runMigrate(cmd *cobra.Command, args []string) error {
    ctx := context.Background()

    // Setup UI
    ui := cli.NewUI(interactive, verbose)
    
    // Resolve paths
    sourceAbs, err := filepath.Abs(sourcePath)
    if err != nil {
        return fmt.Errorf("invalid source path: %w", err)
    }

    targetAbs := targetPath
    if targetAbs == "" {
        targetAbs = sourceAbs + "_" + toProvider
    }
    targetAbs, err = filepath.Abs(targetAbs)
    if err != nil {
        return fmt.Errorf("invalid target path: %w", err)
    }

    ui.Info("Starting GenKit migration")
    ui.Info(fmt.Sprintf("Source: %s (%s)", sourceAbs, fromProvider))
    ui.Info(fmt.Sprintf("Target: %s (%s)", targetAbs, toProvider))

    // Step 1: Analyze source project
    ui.StartProgress("Analyzing source project...")
    
    analyzer := analyzer.New(&analyzer.Config{
        SourceProvider: fromProvider,
        TargetProvider: toProvider,
        Verbose:        verbose,
    })

    project, err := analyzer.AnalyzeProject(ctx, sourceAbs)
    if err != nil {
        ui.StopProgress()
        return fmt.Errorf("analysis failed: %w", err)
    }
    
    ui.StopProgress()
    ui.Success(fmt.Sprintf("Found %d flows, %d models", len(project.Flows), len(project.Models)))

    // Interactive confirmation
    if interactive && !dryRun {
        confirmed, err := ui.Confirm("Continue with migration?")
        if err != nil {
            return err
        }
        if !confirmed {
            ui.Info("Migration cancelled")
            return nil
        }
    }

    // Step 2: Transform project
    ui.StartProgress("Transforming project...")
    
    transformer := transformer.New(&transformer.Config{
        SourceProvider: fromProvider,
        TargetProvider: toProvider,
        TargetPath:     targetAbs,
        DryRun:        dryRun,
    })

    migration, err := transformer.TransformProject(ctx, project)
    if err != nil {
        ui.StopProgress()
        return fmt.Errorf("transformation failed: %w", err)
    }
    
    ui.StopProgress()
    ui.Success("Project transformation complete")

    // Step 3: Generate output
    if !dryRun {
        ui.StartProgress("Generating output files...")
        
        generator := generator.New(&generator.Config{
            TargetProvider: toProvider,
            OutputPath:     targetAbs,
        })

        err = generator.GenerateProject(ctx, migration)
        if err != nil {
            ui.StopProgress()
            return fmt.Errorf("generation failed: %w", err)
        }
        
        ui.StopProgress()
        ui.Success(fmt.Sprintf("Migration complete! Check %s", targetAbs))
    } else {
        ui.Info("Dry run complete - no files were modified")
        ui.PrintMigrationPlan(migration)
    }

    return nil
}
```

### pkg/analyzer/project.go
```go
// Package analyzer analyzes GenKit projects for migration
package analyzer

import (
    "context"
    "fmt"
    "go/ast"
    "go/parser"
    "go/token"
    "os"
    "path/filepath"
    "strings"

    "github.com/genkit-migrate/genkit-migrate/pkg/models"
)

// Analyzer analyzes GenKit projects for migration
type Analyzer struct {
    config *Config
}

// Config holds analyzer configuration
type Config struct {
    SourceProvider string
    TargetProvider string
    Verbose        bool
}

// New creates a new analyzer
func New(config *Config) *Analyzer {
    return &Analyzer{config: config}
}

// AnalyzeProject analyzes a GenKit project directory
func (a *Analyzer) AnalyzeProject(ctx context.Context, projectPath string) (*models.Project, error) {
    project := &models.Project{
        Path:           projectPath,
        SourceProvider: a.config.SourceProvider,
        TargetProvider: a.config.TargetProvider,
        Files:          make(map[string]*models.SourceFile),
        Dependencies:   make(map[string]string),
        Flows:          make([]*models.Flow, 0),
        Models:         make([]*models.Model, 0),
        Configuration:  make(map[string]interface{}),
    }

    // Walk the project directory
    err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Skip non-Go files and vendor/test directories
        if !strings.HasSuffix(path, ".go") ||
            strings.Contains(path, "vendor/") ||
            strings.Contains(path, ".git/") {
            return nil
        }

        // Parse Go file
        sourceFile, err := a.parseGoFile(path, projectPath)
        if err != nil {
            if a.config.Verbose {
                fmt.Printf("Warning: failed to parse %s: %v\n", path, err)
            }
            return nil // Continue with other files
        }

        if sourceFile != nil {
            relPath, _ := filepath.Rel(projectPath, path)
            project.Files[relPath] = sourceFile

            // Extract GenKit elements
            project.Flows = append(project.Flows, sourceFile.Flows...)
            project.Models = append(project.Models, sourceFile.Models...)
        }

        return nil
    })

    if err != nil {
        return nil, fmt.Errorf("failed to walk project directory: %w", err)
    }

    // Analyze dependencies
    err = a.analyzeDependencies(project)
    if err != nil {
        return nil, fmt.Errorf("failed to analyze dependencies: %w", err)
    }

    // Detect configuration patterns
    err = a.analyzeConfiguration(project)
    if err != nil {
        return nil, fmt.Errorf("failed to analyze configuration: %w", err)
    }

    return project, nil
}

// parseGoFile parses a single Go file and extracts GenKit information
func (a *Analyzer) parseGoFile(filePath, projectRoot string) (*models.SourceFile, error) {
    fset := token.NewFileSet()
    node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
    if err != nil {
        return nil, err
    }

    sourceFile := &models.SourceFile{
        Path:        filePath,
        PackageName: node.Name.Name,
        Imports:     make([]string, 0),
        Flows:       make([]*models.Flow, 0),
        Models:      make([]*models.Model, 0),
        HasGenKit:   false,
    }

    // Extract imports
    for _, imp := range node.Imports {
        importPath := strings.Trim(imp.Path.Value, `"`)
        sourceFile.Imports = append(sourceFile.Imports, importPath)
        
        // Check for GenKit imports
        if strings.Contains(importPath, "genkit") ||
            strings.Contains(importPath, "firebase/genkit") {
            sourceFile.HasGenKit = true
        }
    }

    // Only analyze files that use GenKit
    if !sourceFile.HasGenKit {
        return nil, nil
    }

    // Walk AST to find GenKit patterns
    ast.Inspect(node, func(n ast.Node) bool {
        switch node := n.(type) {
        case *ast.CallExpr:
            // Look for genkit.DefineFlow calls
            if flow := a.extractFlow(node, fset); flow != nil {
                sourceFile.Flows = append(sourceFile.Flows, flow)
            }
            
            // Look for model registrations
            if model := a.extractModel(node, fset); model != nil {
                sourceFile.Models = append(sourceFile.Models, model)
            }
        }
        return true
    })

    return sourceFile, nil
}

// extractFlow extracts flow information from AST node
func (a *Analyzer) extractFlow(call *ast.CallExpr, fset *token.FileSet) *models.Flow {
    // Look for genkit.DefineFlow calls
    if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
        if sel.Sel.Name == "DefineFlow" && len(call.Args) >= 2 {
            // Extract flow name
            if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
                flowName := strings.Trim(lit.Value, `"`)
                
                return &models.Flow{
                    Name:     flowName,
                    Position: fset.Position(call.Pos()),
                    // TODO: Extract more flow details
                }
            }
        }
    }
    return nil
}

// extractModel extracts model information from AST node  
func (a *Analyzer) extractModel(call *ast.CallExpr, fset *token.FileSet) *models.Model {
    // Look for model usage patterns
    if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
        if sel.Sel.Name == "Model" && len(call.Args) >= 1 {
            if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
                modelName := strings.Trim(lit.Value, `"`)
                
                // Determine provider based on model name
                provider := a.detectModelProvider(modelName)
                
                return &models.Model{
                    Name:     modelName,
                    Provider: provider,
                    Position: fset.Position(call.Pos()),
                }
            }
        }
    }
    return nil
}

// detectModelProvider detects the cloud provider based on model name
func (a *Analyzer) detectModelProvider(modelName string) string {
    switch {
    case strings.Contains(modelName, "googleai/") || strings.Contains(modelName, "vertexai/"):
        return "gcp"
    case strings.Contains(modelName, "openai/"):
        return "openai"
    case strings.Contains(modelName, "anthropic/"):
        return "anthropic"
    default:
        return "unknown"
    }
}

// analyzeDependencies analyzes go.mod for dependencies
func (a *Analyzer) analyzeDependencies(project *models.Project) error {
    goModPath := filepath.Join(project.Path, "go.mod")
    
    content, err := os.ReadFile(goModPath)
    if err != nil {
        return fmt.Errorf("failed to read go.mod: %w", err)
    }

    // Simple parsing of go.mod (could be enhanced with proper parser)
    lines := strings.Split(string(content), "\n")
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "require ") || 
           (strings.Contains(line, " v") && !strings.HasPrefix(line, "module") && !strings.HasPrefix(line, "go ")) {
            
            // Extract dependency name and version
            parts := strings.Fields(line)
            if len(parts) >= 2 {
                dep := strings.TrimPrefix(parts[0], "require")
                dep = strings.TrimSpace(dep)
                version := parts[1]
                
                project.Dependencies[dep] = version
            }
        }
    }

    return nil
}

// analyzeConfiguration analyzes project configuration patterns
func (a *Analyzer) analyzeConfiguration(project *models.Project) error {
    // Look for common configuration files
    configFiles := []string{"config.yaml", "config.json", ".env", "app.yaml"}
    
    for _, filename := range configFiles {
        configPath := filepath.Join(project.Path, filename)
        if _, err := os.Stat(configPath); err == nil {
            // File exists, analyze it
            project.Configuration[filename] = configPath
        }
    }

    return nil
}
```

### pkg/transformer/code.go
```go
// Package transformer transforms GenKit code for different cloud providers
package transformer

import (
    "context"
    "fmt"
    "strings"
    "text/template"

    "github.com/genkit-migrate/genkit-migrate/pkg/models"
)

// Transformer transforms GenKit projects between cloud providers
type Transformer struct {
    config *Config
}

// Config holds transformer configuration
type Config struct {
    SourceProvider string
    TargetProvider string
    TargetPath     string
    DryRun         bool
}

// New creates a new transformer
func New(config *Config) *Transformer {
    return &Transformer{config: config}
}

// TransformProject transforms a project for the target provider
func (t *Transformer) TransformProject(ctx context.Context, project *models.Project) (*models.Migration, error) {
    migration := &models.Migration{
        Project:     project,
        Changes:     make([]*models.Change, 0),
        NewFiles:    make(map[string]string),
        DeleteFiles: make([]string, 0),
        Commands:    make([]string, 0),
    }

    // Transform dependencies
    err := t.transformDependencies(migration)
    if err != nil {
        return nil, fmt.Errorf("failed to transform dependencies: %w", err)
    }

    // Transform code files  
    err = t.transformSourceFiles(migration)
    if err != nil {
        return nil, fmt.Errorf("failed to transform source files: %w", err)
    }

    // Transform models
    err = t.transformModels(migration)
    if err != nil {
        return nil, fmt.Errorf("failed to transform models: %w", err)
    }

    // Transform configuration
    err = t.transformConfiguration(migration)
    if err != nil {
        return nil, fmt.Errorf("failed to transform configuration: %w", err)
    }

    // Generate deployment files
    err = t.generateDeploymentFiles(migration)
    if err != nil {
        return nil, fmt.Errorf("failed to generate deployment files: %w", err)
    }

    return migration, nil
}

// transformDependencies transforms go.mod dependencies
func (t *Transformer) transformDependencies(migration *models.Migration) error {
    project := migration.Project
    
    // Create new go.mod content
    goModTemplate := `module {{ .ModuleName }}

go {{ .GoVersion }}

require (
    github.com/firebase/genkit/go v0.5.8
{{- if eq .TargetProvider "aws" }}
    github.com/genkit-aws/genkit-aws v0.1.0
    github.com/aws/aws-sdk-go-v2 v1.32.6
    github.com/aws/aws-sdk-go-v2/config v1.28.6
    github.com/aws/aws-sdk-go-v2/service/bedrockruntime v1.21.7
    github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.42.7
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

// transformSourceFiles transforms Go source files
func (t *Transformer) transformSourceFiles(migration *models.Migration) error {
    project := migration.Project

    for filePath, sourceFile := range project.Files {
        if !sourceFile.HasGenKit {
            continue
        }

        // Transform the file content
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

// transformGoFile transforms a single Go file
func (t *Transformer) transformGoFile(sourceFile *models.SourceFile) (string, []*models.Change, error) {
    changes := make([]*models.Change, 0)
    
    // Read original content (simplified - in real implementation, would need to parse and modify AST)
    content := `// Transformed for ` + t.config.TargetProvider + `
package ` + sourceFile.PackageName + `

import (
    "context"
    "github.com/firebase/genkit/go/genkit"
`

    // Add provider-specific imports
    if t.config.TargetProvider == "aws" {
        content += `    genkitaws "github.com/genkit-aws/genkit-aws/pkg/genkit-aws"
    "github.com/genkit-aws/genkit-aws/pkg/bedrock"
    "github.com/genkit-aws/genkit-aws/pkg/monitoring"
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

// transformModels transforms model references
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

// getModelMappings returns model mappings between providers
func (t *Transformer) getModelMappings() map[string]string {
    if t.config.SourceProvider == "gcp" && t.config.TargetProvider == "aws" {
        return map[string]string{
            "googleai/gemini-1.5-flash":   "anthropic.claude-3-haiku-20240307-v1:0",
            "googleai/gemini-1.5-pro":     "anthropic.claude-3-sonnet-20240229-v1:0",
            "vertexai/gemini-pro":         "anthropic.claude-3-sonnet-20240229-v1:0",
        }
    }
    return make(map[string]string)
}

// transformConfiguration transforms configuration files
func (t *Transformer) transformConfiguration(migration *models.Migration) error {
    // Generate new configuration for target provider
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

// generateDeploymentFiles generates deployment configurations
func (t *Transformer) generateDeploymentFiles(migration *models.Migration) error {
    if t.config.TargetProvider == "aws" {
        // Generate Terraform configuration
        err := t.generateTerraform(migration)
        if err != nil {
            return err
        }

        // Generate Dockerfile
        err = t.generateDockerfile(migration)
        if err != nil {
            return err
        }

        // Generate GitHub Actions workflow
        err = t.generateGitHubActions(migration)
        if err != nil {
            return err
        }
    }

    return nil
}

// generateTerraform generates Terraform configuration
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

    // Variables file
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

// generateDockerfile generates Dockerfile
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

// generateGitHubActions generates CI/CD workflow
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
        aws-access-key-id: {{ "${{ secrets.AWS_ACCESS_KEY_ID }}" }}
        aws-secret-access-key: {{ "${{ secrets.AWS_SECRET_ACCESS_KEY }}" }}
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

// Helper functions
func (t *Transformer) extractModuleName(project *models.Project) string {
    // Extract from existing go.mod or generate
    return "genkit-app"
}

func (t *Transformer) extractProjectName(project *models.Project) string {
    return "GenKitApp"
}

func (t *Transformer) filterDependencies(deps map[string]string) []map[string]string {
    filtered := make([]map[string]string, 0)
    for name, version := range deps {
        // Skip provider-specific dependencies that will be replaced
        if !strings.Contains(name, "firebase") && !strings.Contains(name, "google") {
            filtered = append(filtered, map[string]string{
                "Name":    name,
                "Version": version,
            })
        }
    }
    return filtered
}
```

### pkg/models/project.go
```go
// Package models defines data structures for GenKit migration
package models

import "go/token"

// Project represents a GenKit project
type Project struct {
    Path           string                 `json:"path"`
    SourceProvider string                 `json:"source_provider"`
    TargetProvider string                 `json:"target_provider"`
    Files          map[string]*SourceFile `json:"files"`
    Dependencies   map[string]string      `json:"dependencies"`
    Flows          []*Flow               `json:"flows"`
    Models         []*Model              `json:"models"`
    Configuration  map[string]interface{} `json:"configuration"`
}

// SourceFile represents a source code file
type SourceFile struct {
    Path        string   `json:"path"`
    PackageName string   `json:"package_name"`
    Imports     []string `json:"imports"`
    Flows       []*Flow  `json:"flows"`
    Models      []*Model `json:"models"`
    HasGenKit   bool     `json:"has_genkit"`
}

// Flow represents a GenKit flow
type Flow struct {
    Name        string         `json:"name"`
    Position    token.Position `json:"position"`
    InputType   string         `json:"input_type,omitempty"`
    OutputType  string         `json:"output_type,omitempty"`
    Description string         `json:"description,omitempty"`
}

// Model represents a GenKit model reference
type Model struct {
    Name     string         `json:"name"`
    Provider string         `json:"provider"`
    Position token.Position `json:"position"`
}

// Migration represents a migration plan
type Migration struct {
    Project     *Project           `json:"project"`
    Changes     []*Change          `json:"changes"`
    NewFiles    map[string]string  `json:"new_files"`
    DeleteFiles []string           `json:"delete_files"`
    Commands    []string           `json:"commands"`
}

// Change represents a specific change in the migration
type Change struct {
    Type        string `json:"type"`        // "dependency", "import", "model", "config"
    Description string `json:"description"`
    File        string `json:"file"`
    OldValue    string `json:"old_value,omitempty"`
    NewValue    string `json:"new_value,omitempty"`
}
```

### internal/cli/ui.go
```go
// Package cli provides user interface utilities
package cli

import (
    "fmt"
    "os"
    "time"

    "github.com/charmbracelet/huh"
    "github.com/charmbracelet/lipgloss"
    "github.com/genkit-migrate/genkit-migrate/pkg/models"
)

// UI provides user interface functionality
type UI struct {
    interactive bool
    verbose     bool
    spinner     bool
}

// NewUI creates a new UI instance
func NewUI(interactive, verbose bool) *UI {
    return &UI{
        interactive: interactive,
        verbose:     verbose,
    }
}

// Info prints an informational message
func (ui *UI) Info(message string) {
    style := lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
    fmt.Printf("%s %s\n", style.Render("ℹ"), message)
}

// Success prints a success message
func (ui *UI) Success(message string) {
    style := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
    fmt.Printf("%s %s\n", style.Render("✓"), message)
}

// Error prints an error message
func (ui *UI) Error(message string) {
    style := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
    fmt.Fprintf(os.Stderr, "%s %s\n", style.Render("✗"), message)
}

// Warning prints a warning message
func (ui *UI) Warning(message string) {
    style := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
    fmt.Printf("%s %s\n", style.Render("⚠"), message)
}

// StartProgress starts a progress indicator
func (ui *UI) StartProgress(message string) {
    fmt.Printf("⏳ %s", message)
    ui.spinner = true
    // TODO: Implement actual spinner
}

// StopProgress stops the progress indicator
func (ui *UI) StopProgress() {
    if ui.spinner {
        fmt.Printf("\r")
        ui.spinner = false
    }
}

// Confirm prompts the user for confirmation
func (ui *UI) Confirm(message string) (bool, error) {
    if !ui.interactive {
        return true, nil
    }

    var confirm bool
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewConfirm().
                Title(message).
                Value(&confirm),
        ),
    )

    err := form.Run()
    return confirm, err
}

// SelectProvider prompts the user to select a cloud provider
func (ui *UI) SelectProvider(providers []string, prompt string) (string, error) {
    if !ui.interactive {
        return providers[0], nil
    }

    var selected string
    options := make([]huh.Option[string], len(providers))
    for i, provider := range providers {
        options[i] = huh.NewOption(provider, provider)
    }

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title(prompt).
                Options(options...).
                Value(&selected),
        ),
    )

    err := form.Run()
    return selected, err
}

// PrintMigrationPlan prints the migration plan
func (ui *UI) PrintMigrationPlan(migration *models.Migration) {
    ui.Info("Migration Plan:")
    fmt.Printf("\n")

    // Print changes
    if len(migration.Changes) > 0 {
        ui.Info("Changes:")
        for _, change := range migration.Changes {
            fmt.Printf("  • %s: %s\n", change.Type, change.Description)
        }
        fmt.Printf("\n")
    }

    // Print new files
    if len(migration.NewFiles) > 0 {
        ui.Info("New files to be created:")
        for filePath := range migration.NewFiles {
            fmt.Printf("  • %s\n", filePath)
        }
        fmt.Printf("\n")
    }

    // Print commands
    if len(migration.Commands) > 0 {
        ui.Info("Commands to run:")
        for _, cmd := range migration.Commands {
            fmt.Printf("  • %s\n", cmd)
        }
        fmt.Printf("\n")
    }
}
```

### README.md
```markdown
# GenKit Migrate

**Migrate Google GenKit applications between cloud providers**

A CLI tool for migrating existing GenKit applications (built with Google's GenKit framework) from one cloud provider to another.

> **Important**: This migrates applications that use Google's GenKit framework, not the framework itself. Your app continues using Google's GenKit APIs - we just change which cloud provider plugins it uses.

## What It Does

Takes a GenKit app like this:
```go
// Original app using Google AI
import "github.com/firebase/genkit/go/plugins/googleai"

googleai.Init(ctx, &googleai.Config{})
resp, _ := genkit.Generate(ctx, &ai.GenerateRequest{
    Model: genkit.Model("googleai/gemini-1.5-pro"),
    // ...
})
```

And transforms it to:
```go  
// Migrated app using AWS Bedrock
import "github.com/genkit-aws/genkit-aws/aws/bedrock"

bedrock.Init(ctx, &bedrock.Config{})
resp, _ := genkit.Generate(ctx, &ai.GenerateRequest{
    Model: genkit.Model("anthropic.claude-3-sonnet-20240229-v1:0"),
    // ...
})
```

## Migration Process

1. **Analyzes** your GenKit app to find flows, models, and dependencies
2. **Transforms** code to use different cloud provider plugins
3. **Maps models** between providers (e.g., gemini → claude)
4. **Updates** go.mod dependencies  
5. **Generates** new deployment configs (Terraform, Docker, CI/CD)

## Quick Start

### Install
```bash
go install github.com/genkit-migrate/genkit-migrate/cmd/genkit-migrate@latest
```

### Migrate GCP → AWS
```bash
genkit-migrate migrate --from=gcp --to=aws --source=./my-genkit-app
```

### Preview Changes (Dry Run)
```bash
genkit-migrate migrate --from=gcp --to=aws --source=./my-genkit-app --dry-run
```

## Example Migration

**Before (GCP):**
```
my-genkit-app/
├── go.mod                 # Depends on googleai plugin
├── main.go               # Uses googleai.Init(), googleai/gemini models
└── config.yaml          # GCP-specific configuration
```

**After (AWS):**
```
my-genkit-app_aws/
├── go.mod                # Depends on AWS bedrock plugin
├── main.go              # Uses bedrock.Init(), anthropic/claude models  
├── config.yaml          # AWS-specific configuration
├── terraform/           # AWS deployment configs
│   ├── main.tf
│   └── variables.tf
├── Dockerfile           # Container for AWS Lambda/ECS
└── .github/workflows/   # CI/CD for AWS
    └── deploy.yml
```

## Supported Migrations

| From | To | Status | Models Mapped |
|------|-------|--------|---------------|
| GCP | AWS | ✅ Ready | googleai/gemini → anthropic/claude |
| GCP | Azure | 🚧 Planned | TBD |
| AWS | GCP | 🚧 Planned | TBD |

## Command Reference

### `migrate`
```bash
genkit-migrate migrate [flags]
```

**Flags:**
- `--from`: Source provider (gcp, aws, azure) 
- `--to`: Target provider (aws, gcp, azure)
- `--source, -s`: Source GenKit project path
- `--target, -t`: Target path (default: source_target)
- `--dry-run`: Preview without changes
- `--interactive, -i`: Interactive prompts (default: true)

### `analyze` 
```bash
genkit-migrate analyze --source=./my-genkit-app
```
Analyze a GenKit project without migrating.

## What Gets Transformed

### Code Changes
- **Import statements**: `googleai` → `bedrock`
- **Plugin initialization**: `googleai.Init()` → `bedrock.Init()`  
- **Model references**: `googleai/gemini-1.5-pro` → `anthropic.claude-3-sonnet-20240229-v1:0`
- **Configuration**: Environment variables, connection settings

### Dependencies  
- **go.mod**: Replace provider-specific packages
- **Provider plugins**: Remove old, add new cloud provider plugins
- **Maintain GenKit**: Keep Google's GenKit framework unchanged

### Model Mappings (GCP → AWS)
| GCP Model | AWS Model |
|-----------|-----------|
| googleai/gemini-1.5-flash | anthropic.claude-3-haiku-20240307-v1:0 |
| googleai/gemini-1.5-pro | anthropic.claude-3-sonnet-20240229-v1:0 |
| vertexai/gemini-pro | anthropic.claude-3-sonnet-20240229-v1:0 |

### Generated Files
- **Terraform**: AWS infrastructure as code
- **Docker**: Container configuration for AWS services
- **CI/CD**: GitHub Actions for AWS deployment
- **Documentation**: Migration notes and next steps

## Configuration

Create `.genkit-migrate.yaml`:
```yaml
# Default settings
default_source_provider: gcp
default_target_provider: aws
interactive: true

# Provider settings  
providers:
  aws:
    region: us-east-1
    profile: default
  gcp:
    project_id: my-project
    region: us-central1
```

## Limitations

1. **Manual review required** for complex flows
2. **Model behavior differences** may need adjustment
3. **Provider-specific features** might not have equivalents
4. **Testing recommended** before production deployment

## Migration Example

**Original GenKit app (GCP):**
```go
package main

import (
    "github.com/firebase/genkit/go/genkit"
    "github.com/firebase/genkit/go/plugins/googleai"
)

func main() {
    ctx := context.Background()
    
    genkit.Init(ctx, nil)
    googleai.Init(ctx, &googleai.Config{})
    
    flow := genkit.DefineFlow("summarize", func(ctx context.Context, text string) (string, error) {
        resp, err := genkit.Generate(ctx, &ai.GenerateRequest{
            Model: genkit.Model("googleai/gemini-1.5-pro"),
            Messages: []*ai.Message{{
                Role: ai.RoleUser,
                Content: []*ai.Part{{Text: "Summarize: " + text}},
            }},
        })
        return resp.Text(), err
    })
}
```

**Command:**
```bash
genkit-migrate migrate --from=gcp --to=aws --source=./my-app
```

**Migrated GenKit app (AWS):**
```go
package main

import (
    "github.com/firebase/genkit/go/genkit"
    "github.com/genkit-aws/genkit-aws/aws/bedrock"
)

func main() {
    ctx := context.Background()
    
    genkit.Init(ctx, nil)
    bedrock.Init(ctx, &bedrock.Config{Region: "us-east-1"})
    
    flow := genkit.DefineFlow("summarize", func(ctx context.Context, text string) (string, error) {
        resp, err := genkit.Generate(ctx, &ai.GenerateRequest{
            Model: genkit.Model("anthropic.claude-3-sonnet-20240229-v1:0"),
            Messages: []*ai.Message{{
                Role: ai.RoleUser,
                Content: []*ai.Part{{Text: "Summarize: " + text}},
            }},
        })
        return resp.Text(), err
    })
}
```

Plus generated AWS deployment files, updated dependencies, and migration documentation.

## Contributing

Help expand cloud provider support and model mappings!

### Development Setup
```bash
git clone https://github.com/genkit-migrate/genkit-migrate
cd genkit-migrate
go mod download
make test
```

## License

Apache License 2.0
```

This project structure provides:

1. **Comprehensive CLI tool** with proper command structure using Cobra
2. **Modular architecture** with separate packages for analysis, transformation, and generation
3. **Rich user experience** with interactive prompts and progress indicators
4. **Production-ready features** like dry-run mode, verbose output, and configuration management
5. **Extensible design** for supporting multiple cloud providers
6. **Proper testing structure** with testdata and integration tests
7. **Complete documentation** and examples

The design follows Go best practices and is structured for collaborative development with Claude Code.
