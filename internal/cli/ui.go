package cli

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/genkit-migrate/genkit-migrate/pkg/models"
)

type UI struct {
	interactive bool
	verbose     bool
	spinner     bool
}

func NewUI(interactive, verbose bool) *UI {
	return &UI{
		interactive: interactive,
		verbose:     verbose,
	}
}

func (ui *UI) Info(message string) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	fmt.Printf("%s %s\n", style.Render("ℹ"), message)
}

func (ui *UI) Success(message string) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	fmt.Printf("%s %s\n", style.Render("✓"), message)
}

func (ui *UI) Error(message string) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	fmt.Fprintf(os.Stderr, "%s %s\n", style.Render("✗"), message)
}

func (ui *UI) Warning(message string) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	fmt.Printf("%s %s\n", style.Render("⚠"), message)
}

func (ui *UI) StartProgress(message string) {
	fmt.Printf("⏳ %s", message)
	ui.spinner = true
}

func (ui *UI) StopProgress() {
	if ui.spinner {
		fmt.Printf("\r")
		ui.spinner = false
	}
}

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

func (ui *UI) PrintMigrationPlan(migration *models.Migration) {
	ui.Info("Migration Plan:")
	fmt.Printf("\n")

	if len(migration.Changes) > 0 {
		ui.Info("Changes:")
		for _, change := range migration.Changes {
			fmt.Printf("  • %s: %s\n", change.Type, change.Description)
		}
		fmt.Printf("\n")
	}

	if len(migration.NewFiles) > 0 {
		ui.Info("New files to be created:")
		for filePath := range migration.NewFiles {
			fmt.Printf("  • %s\n", filePath)
		}
		fmt.Printf("\n")
	}

	if len(migration.Commands) > 0 {
		ui.Info("Commands to run:")
		for _, cmd := range migration.Commands {
			fmt.Printf("  • %s\n", cmd)
		}
		fmt.Printf("\n")
	}
}

func (ui *UI) PrintAnalysisTable(project *models.Project) {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true)

	ui.Info("Project Analysis Results")
	fmt.Printf("\n")

	fmt.Printf("%s: %s\n", headerStyle.Render("Project Path"), project.Path)
	fmt.Printf("%s: %s\n", headerStyle.Render("Source Provider"), project.SourceProvider)
	fmt.Printf("%s: %d\n", headerStyle.Render("Total Files"), len(project.Files))
	fmt.Printf("%s: %d\n", headerStyle.Render("Dependencies"), len(project.Dependencies))
	fmt.Printf("\n")

	if len(project.Flows) > 0 {
		fmt.Printf("%s:\n", headerStyle.Render("GenKit Flows"))
		for _, flow := range project.Flows {
			fmt.Printf("  • %s (%s:%d)\n", flow.Name, flow.Position.Filename, flow.Position.Line)
		}
		fmt.Printf("\n")
	}

	if len(project.Models) > 0 {
		fmt.Printf("%s:\n", headerStyle.Render("Models"))
		for _, model := range project.Models {
			fmt.Printf("  • %s (%s) - %s:%d\n", model.Name, model.Provider, model.Position.Filename, model.Position.Line)
		}
		fmt.Printf("\n")
	}

	if len(project.Dependencies) > 0 {
		fmt.Printf("%s:\n", headerStyle.Render("Key Dependencies"))
		for dep, version := range project.Dependencies {
			if ui.isRelevantDependency(dep) {
				fmt.Printf("  • %s %s\n", dep, version)
			}
		}
		fmt.Printf("\n")
	}
}

func (ui *UI) isRelevantDependency(dep string) bool {
	relevantPrefixes := []string{
		"github.com/firebase/genkit",
		"github.com/genkit-",
		"github.com/google",
		"github.com/aws",
		"github.com/azure",
	}

	for _, prefix := range relevantPrefixes {
		if len(dep) >= len(prefix) && dep[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
