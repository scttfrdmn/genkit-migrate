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

type Analyzer struct {
	config *Config
}

type Config struct {
	SourceProvider string
	TargetProvider string
	Verbose        bool
}

func New(config *Config) *Analyzer {
	return &Analyzer{config: config}
}

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

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") ||
			strings.Contains(path, "vendor/") ||
			strings.Contains(path, ".git/") {
			return nil
		}

		sourceFile, err := a.parseGoFile(path, projectPath)
		if err != nil {
			if a.config.Verbose {
				fmt.Printf("Warning: failed to parse %s: %v\n", path, err)
			}
			return nil
		}

		if sourceFile != nil {
			relPath, _ := filepath.Rel(projectPath, path)
			project.Files[relPath] = sourceFile

			project.Flows = append(project.Flows, sourceFile.Flows...)
			project.Models = append(project.Models, sourceFile.Models...)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk project directory: %w", err)
	}

	err = a.analyzeDependencies(project)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze dependencies: %w", err)
	}

	err = a.analyzeConfiguration(project)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze configuration: %w", err)
	}

	return project, nil
}

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

	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		sourceFile.Imports = append(sourceFile.Imports, importPath)

		if strings.Contains(importPath, "genkit") ||
			strings.Contains(importPath, "firebase/genkit") ||
			strings.Contains(importPath, "genkit/go/plugins") {
			sourceFile.HasGenKit = true
		}
	}

	if !sourceFile.HasGenKit {
		return nil, nil
	}

	ast.Inspect(node, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			if flow := a.extractFlow(node, fset); flow != nil {
				sourceFile.Flows = append(sourceFile.Flows, flow)
			}

			if model := a.extractModel(node, fset); model != nil {
				sourceFile.Models = append(sourceFile.Models, model)
			}
		}
		return true
	})

	return sourceFile, nil
}

func (a *Analyzer) extractFlow(call *ast.CallExpr, fset *token.FileSet) *models.Flow {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if sel.Sel.Name == "DefineFlow" && len(call.Args) >= 2 {
			if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
				flowName := strings.Trim(lit.Value, `"`)

				return &models.Flow{
					Name:     flowName,
					Position: fset.Position(call.Pos()),
				}
			}
		}
	}
	return nil
}

func (a *Analyzer) extractModel(call *ast.CallExpr, fset *token.FileSet) *models.Model {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if sel.Sel.Name == "Model" && len(call.Args) >= 1 {
			if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
				modelName := strings.Trim(lit.Value, `"`)

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

func (a *Analyzer) detectModelProvider(modelName string) string {
	switch {
	case strings.Contains(modelName, "googleai/") || strings.Contains(modelName, "vertexai/"):
		return "gcp"
	case strings.Contains(modelName, "openai/") || strings.Contains(modelName, "gpt-"):
		return "openai"
	case strings.Contains(modelName, "anthropic/") || strings.Contains(modelName, "claude-"):
		return "anthropic"
	case strings.Contains(modelName, "ollama/"):
		return "ollama"
	case strings.Contains(modelName, "bedrock/") || strings.Contains(modelName, "amazon.") || strings.Contains(modelName, "anthropic."):
		return "aws"
	default:
		return "unknown"
	}
}

func (a *Analyzer) analyzeDependencies(project *models.Project) error {
	goModPath := filepath.Join(project.Path, "go.mod")

	content, err := os.ReadFile(goModPath)
	if err != nil {
		return fmt.Errorf("failed to read go.mod: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "require ") ||
			(strings.Contains(line, " v") && !strings.HasPrefix(line, "module") && !strings.HasPrefix(line, "go ")) {

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

func (a *Analyzer) analyzeConfiguration(project *models.Project) error {
	configFiles := []string{"config.yaml", "config.json", ".env", "app.yaml"}

	for _, filename := range configFiles {
		configPath := filepath.Join(project.Path, filename)
		if _, err := os.Stat(configPath); err == nil {
			project.Configuration[filename] = configPath
		}
	}

	return nil
}
