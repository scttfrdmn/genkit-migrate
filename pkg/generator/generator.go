package generator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/genkit-migrate/genkit-migrate/pkg/models"
	"github.com/otiai10/copy"
)

type Generator struct {
	config *Config
}

type Config struct {
	TargetProvider string
	OutputPath     string
}

func New(config *Config) *Generator {
	return &Generator{config: config}
}

func (g *Generator) GenerateProject(ctx context.Context, migration *models.Migration) error {
	err := g.createOutputDirectory()
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	err = g.writeNewFiles(migration)
	if err != nil {
		return fmt.Errorf("failed to write new files: %w", err)
	}

	err = g.copyExistingFiles(migration)
	if err != nil {
		return fmt.Errorf("failed to copy existing files: %w", err)
	}

	err = g.generateDocumentation(migration)
	if err != nil {
		return fmt.Errorf("failed to generate documentation: %w", err)
	}

	return nil
}

func (g *Generator) createOutputDirectory() error {
	return os.MkdirAll(g.config.OutputPath, 0755)
}

func (g *Generator) writeNewFiles(migration *models.Migration) error {
	for filePath, content := range migration.NewFiles {
		fullPath := filepath.Join(g.config.OutputPath, filePath)

		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", fullPath, err)
		}
	}

	return nil
}

func (g *Generator) copyExistingFiles(migration *models.Migration) error {
	project := migration.Project

	for filePath := range project.Files {
		sourcePath := filepath.Join(project.Path, filePath)
		targetPath := filepath.Join(g.config.OutputPath, filePath)

		if _, exists := migration.NewFiles[filePath]; exists {
			continue
		}

		dir := filepath.Dir(targetPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		if err := copy.Copy(sourcePath, targetPath); err != nil {
			return fmt.Errorf("failed to copy file %s: %w", filePath, err)
		}
	}

	return nil
}

func (g *Generator) generateDocumentation(migration *models.Migration) error {
	readmeContent := g.generateReadme(migration)
	readmePath := filepath.Join(g.config.OutputPath, "MIGRATION.md")

	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to write MIGRATION.md: %w", err)
	}

	return nil
}

func (g *Generator) generateReadme(migration *models.Migration) string {
	content := fmt.Sprintf(`# GenKit Migration to %s

This project has been migrated from %s to %s using genkit-migrate.

## Migration Summary

- **Source Provider**: %s
- **Target Provider**: %s
- **Flows Found**: %d
- **Models Found**: %d
- **Changes Applied**: %d

## Changes Made

`,
		g.config.TargetProvider,
		migration.Project.SourceProvider,
		g.config.TargetProvider,
		migration.Project.SourceProvider,
		g.config.TargetProvider,
		len(migration.Project.Flows),
		len(migration.Project.Models),
		len(migration.Changes),
	)

	for _, change := range migration.Changes {
		content += fmt.Sprintf("- **%s**: %s (in %s)\n", change.Type, change.Description, change.File)
	}

	if g.config.TargetProvider == "aws" {
		content += `

## AWS Deployment

### Prerequisites

1. AWS CLI configured with appropriate credentials
2. Terraform installed (>= 1.0)
3. Docker installed (for containerization)

### Deploy with Terraform

` + "```bash" + `
cd terraform
terraform init
terraform plan
terraform apply
` + "```" + `

### Build and Deploy Docker Container

` + "```bash" + `
docker build -t genkit-app .
docker tag genkit-app:latest <your-ecr-repo>:latest
docker push <your-ecr-repo>:latest
` + "```" + `

### Configuration

Update the ` + "`config.yaml`" + ` file with your specific AWS settings:

- AWS region
- Bedrock model preferences
- CloudWatch configuration
- Environment variables

## Next Steps

1. Review all generated files
2. Test the application locally
3. Update any hardcoded values in configuration
4. Deploy to your AWS environment
5. Set up monitoring and logging
6. Test all GenKit flows work correctly

## Model Mappings Applied

The following model mappings were applied during migration:

`

		modelMappings := map[string]string{
			"googleai/gemini-1.5-flash": "anthropic.claude-3-haiku-20240307-v1:0",
			"googleai/gemini-1.5-pro":   "anthropic.claude-3-sonnet-20240229-v1:0",
			"vertexai/gemini-pro":       "anthropic.claude-3-sonnet-20240229-v1:0",
		}

		for oldModel, newModel := range modelMappings {
			content += fmt.Sprintf("- `%s` → `%s`\n", oldModel, newModel)
		}

		content += `
## Support

For issues with the migration tool, please visit:
https://github.com/genkit-migrate/genkit-migrate/issues

For GenKit framework support, please visit:
https://firebase.google.com/docs/genkit
`
	}

	return content
}
