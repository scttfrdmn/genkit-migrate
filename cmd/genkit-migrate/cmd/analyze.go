package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/genkit-migrate/genkit-migrate/internal/cli"
	"github.com/genkit-migrate/genkit-migrate/pkg/analyzer"
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	outputFile   string
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze a GenKit project without migrating",
	Long: `Analyze a GenKit project to understand its structure, dependencies, and cloud provider usage.

This command will:
1. Scan the project for GenKit flows and models
2. Analyze dependencies and imports
3. Detect current cloud provider configuration
4. Output analysis results in the specified format

Example:
  genkit-migrate analyze --source=./my-genkit-app --format=json`,
	RunE: runAnalyze,
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVarP(&sourcePath, "source", "s", ".", "source project path")
	analyzeCmd.Flags().StringVar(&outputFormat, "format", "table", "output format (table, json, yaml)")
	analyzeCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file (default: stdout)")

	analyzeCmd.MarkFlagRequired("source")
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	ui := cli.NewUI(interactive, verbose)

	sourceAbs, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("invalid source path: %w", err)
	}

	ui.Info(fmt.Sprintf("Analyzing GenKit project: %s", sourceAbs))
	ui.StartProgress("Scanning project files...")

	analyzer := analyzer.New(&analyzer.Config{
		SourceProvider: "auto-detect",
		TargetProvider: "",
		Verbose:        verbose,
	})

	project, err := analyzer.AnalyzeProject(ctx, sourceAbs)
	if err != nil {
		ui.StopProgress()
		return fmt.Errorf("analysis failed: %w", err)
	}

	ui.StopProgress()
	ui.Success("Analysis complete")

	switch outputFormat {
	case "json":
		jsonOutput, err := json.MarshalIndent(project, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(jsonOutput))
	case "table":
		ui.PrintAnalysisTable(project)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	return nil
}
