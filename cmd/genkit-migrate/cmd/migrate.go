package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/genkit-migrate/genkit-migrate/internal/cli"
	"github.com/genkit-migrate/genkit-migrate/pkg/analyzer"
	"github.com/genkit-migrate/genkit-migrate/pkg/generator"
	"github.com/genkit-migrate/genkit-migrate/pkg/transformer"
	"github.com/spf13/cobra"
)

var (
	sourcePath   string
	targetPath   string
	fromProvider string
	toProvider   string
	dryRun       bool
	interactive  bool
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

	if err := migrateCmd.MarkFlagRequired("source"); err != nil {
		// This should never fail with a valid flag name
		panic(fmt.Sprintf("failed to mark source flag as required: %v", err))
	}
}

func runMigrate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	ui := cli.NewUI(interactive, verbose)

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

	ui.StartProgress("Transforming project...")

	transformer := transformer.New(&transformer.Config{
		SourceProvider: fromProvider,
		TargetProvider: toProvider,
		TargetPath:     targetAbs,
		DryRun:         dryRun,
	})

	migration, err := transformer.TransformProject(ctx, project)
	if err != nil {
		ui.StopProgress()
		return fmt.Errorf("transformation failed: %w", err)
	}

	ui.StopProgress()
	ui.Success("Project transformation complete")

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
