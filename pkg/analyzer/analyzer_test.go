package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeProject(t *testing.T) {
	testDir := createTestProject(t)
	defer os.RemoveAll(testDir)

	analyzer := New(&Config{
		SourceProvider: "gcp",
		TargetProvider: "aws",
		Verbose:        false,
	})

	project, err := analyzer.AnalyzeProject(context.Background(), testDir)
	require.NoError(t, err)

	assert.Equal(t, testDir, project.Path)
	assert.Equal(t, "gcp", project.SourceProvider)
	assert.Equal(t, "aws", project.TargetProvider)
	assert.Len(t, project.Files, 1)
	assert.Len(t, project.Flows, 1)
	assert.Len(t, project.Models, 1)

	flow := project.Flows[0]
	assert.Equal(t, "summarize", flow.Name)

	model := project.Models[0]
	assert.Equal(t, "googleai/gemini-1.5-pro", model.Name)
	assert.Equal(t, "gcp", model.Provider)
}

func TestDetectModelProvider(t *testing.T) {
	analyzer := New(&Config{})

	tests := []struct {
		modelName string
		expected  string
	}{
		{"googleai/gemini-1.5-pro", "gcp"},
		{"vertexai/gemini-pro", "gcp"},
		{"openai/gpt-4", "openai"},
		{"anthropic/claude-3", "anthropic"},
		{"unknown-model", "unknown"},
	}

	for _, test := range tests {
		result := analyzer.detectModelProvider(test.modelName)
		assert.Equal(t, test.expected, result, "Model: %s", test.modelName)
	}
}

func createTestProject(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "genkit-test-*")
	require.NoError(t, err)

	goModContent := `module test-genkit-app

go 1.23

require (
    github.com/firebase/genkit/go v0.5.8
    github.com/firebase/genkit/go/plugins/googleai v0.5.8
)
`

	mainGoContent := `package main

import (
    "context"
    "github.com/firebase/genkit/go/genkit"
    "github.com/firebase/genkit/go/plugins/googleai"
)

func main() {
    ctx := context.Background()
    genkit.Init(ctx, nil)
    
    genkit.DefineFlow("summarize", func(ctx context.Context, input string) (string, error) {
        resp, err := genkit.Generate(ctx, &ai.GenerateRequest{
            Model: genkit.Model("googleai/gemini-1.5-pro"),
        })
        return resp.Text(), err
    })
}
`

	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(mainGoContent), 0644)
	require.NoError(t, err)

	return tempDir
}
