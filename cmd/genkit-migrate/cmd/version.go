package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	commit    = "dev"
	buildTime = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of genkit-migrate",
	Long:  `Print the version number of genkit-migrate`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("genkit-migrate %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Built: %s\n", buildTime)
		fmt.Printf("Go version: %s\n", runtime.Version())
		fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("Copyright 2025 Scott Friedman\n")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
