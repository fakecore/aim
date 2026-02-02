package cli

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/fakecore/aim/internal/errors"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "aim",
	Short: "AI Model Manager - Manage AI tools and providers",
	Long:  `AIM is a unified CLI tool for managing multiple AI CLI tools and model providers.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Check if it's our custom error type
		if aimErr, ok := err.(*errors.Error); ok {
			log.Error(aimErr)
			os.Exit(aimErr.ExitCode())
		}
		log.Error(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(runCmd)
}
