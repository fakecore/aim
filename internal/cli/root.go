package cli

import (
	"fmt"
	"os"

	"github.com/fakecore/aim/internal/errors"
	"github.com/fakecore/aim/pkg/clog"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "aim",
	Short:         "AI Model Manager - Manage AI tools and providers",
	Long:          `AIM is a unified CLI tool for managing multiple AI CLI tools and model providers.`,
	SilenceErrors: true, // We handle error printing ourselves
	SilenceUsage:  true, // Don't show usage on error
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Check if it's our custom error type
		if aimErr, ok := err.(*errors.Error); ok {
			// User-facing: human readable message (without error code)
			fmt.Fprintln(os.Stderr, "Error:", aimErr.Error())
			// Log: detailed message with error code
			clog.Errorf(aimErr.LogMessage())
			// Print suggestions if available
			for _, suggestion := range aimErr.Suggestions {
				fmt.Fprintf(os.Stderr, "  → %s\n", suggestion)
			}
			os.Exit(aimErr.ExitCode())
		}
		// Non-AIM errors
		fmt.Fprintln(os.Stderr, "Error:", err)
		clog.Errorf(err.Error())
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(configCmd)
}
