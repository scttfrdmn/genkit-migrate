package transformer

import (
	"context"
	"testing"

	"github.com/genkit-migrate/genkit-migrate/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go/token"
)

func TestTransformProject(t *testing.T) {
	transformer := New(&Config{
		SourceProvider: "gcp",
		TargetProvider: "aws",
		TargetPath:     "/tmp/test-output",
		DryRun:         false,
	})

	project := &models.Project{
		Path:           "/tmp/test-source",
		SourceProvider: "gcp",
		TargetProvider: "aws",
		Files: map[string]*models.SourceFile{
			"main.go": {
				Path:        "/tmp/test-source/main.go",
				PackageName: "main",
				HasGenKit:   true,
				Flows: []*models.Flow{
					{Name: "summarize", Position: token.Position{}},
				},
				Models: []*models.Model{
					{Name: "googleai/gemini-1.5-pro", Provider: "gcp", Position: token.Position{}},
				},
			},
		},
		Dependencies: map[string]string{
			"github.com/firebase/genkit/go":                  "v0.5.8",
			"github.com/firebase/genkit/go/plugins/googleai": "v0.5.8",
			"github.com/spf13/cobra":                         "v1.8.1",
		},
		Flows: []*models.Flow{
			{Name: "summarize", Position: token.Position{}},
		},
		Models: []*models.Model{
			{Name: "googleai/gemini-1.5-pro", Provider: "gcp", Position: token.Position{}},
		},
		Configuration: make(map[string]interface{}),
	}

	migration, err := transformer.TransformProject(context.Background(), project)
	require.NoError(t, err)

	assert.NotNil(t, migration)
	assert.Equal(t, project, migration.Project)
	assert.Greater(t, len(migration.Changes), 0)
	assert.Greater(t, len(migration.NewFiles), 0)

	assert.Contains(t, migration.NewFiles, "go.mod")
	assert.Contains(t, migration.NewFiles, "main.go")
	assert.Contains(t, migration.NewFiles, "config.yaml")
	assert.Contains(t, migration.NewFiles, "terraform/main.tf")
	assert.Contains(t, migration.NewFiles, "Dockerfile")
}

func TestGetModelMappings(t *testing.T) {
	transformer := New(&Config{
		SourceProvider: "gcp",
		TargetProvider: "aws",
	})

	mappings := transformer.getModelMappings()

	assert.Contains(t, mappings, "googleai/gemini-1.5-flash")
	assert.Equal(t, "anthropic.claude-3-haiku-20240307-v1:0", mappings["googleai/gemini-1.5-flash"])

	assert.Contains(t, mappings, "googleai/gemini-1.5-pro")
	assert.Equal(t, "anthropic.claude-3-sonnet-20240229-v1:0", mappings["googleai/gemini-1.5-pro"])
}

func TestFilterDependencies(t *testing.T) {
	transformer := New(&Config{})

	deps := map[string]string{
		"github.com/firebase/genkit/go":                  "v0.5.8",
		"github.com/firebase/genkit/go/plugins/googleai": "v0.5.8",
		"github.com/spf13/cobra":                         "v1.8.1",
		"github.com/google/uuid":                         "v1.3.0",
		"gopkg.in/yaml.v3":                               "v3.0.1",
	}

	filtered := transformer.filterDependencies(deps)

	var foundCobra, foundYaml bool
	for _, dep := range filtered {
		if dep["Name"] == "github.com/spf13/cobra" {
			foundCobra = true
		}
		if dep["Name"] == "gopkg.in/yaml.v3" {
			foundYaml = true
		}
	}

	assert.True(t, foundCobra, "Should include non-Google/Firebase dependencies")
	assert.True(t, foundYaml, "Should include non-Google/Firebase dependencies")
}
