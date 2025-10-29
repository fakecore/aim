package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version information (set by build flags)
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"

	// Global flags
	cfgFile string
	verbose bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "aim",
	Short: "AI CLI tool manager",
	Long: `AIM (AI Interface Manager) is a unified manager for AI CLI tools.

It provides seamless switching between different AI models and providers,
with support for tools like Claude Code, OpenAI Codex, and more.

Features:
  - Multi-tool support (Claude Code, Codex, Cursor, etc.)
  - Multi-provider support (Official APIs, custom)
  - Hierarchical configuration (global + project-level)
  - Environment variable management
  - Configuration validation and testing
  - Cross-platform support

Examples:
  # Switch to a model
  aim use deepseek --global

  # Export environment variables
  eval $(aim env deepseek)

  # Test provider configuration
  aim test deepseek

  # Initialize configuration
  aim config init`,
	SilenceUsage:  true,
	SilenceErrors: false,
}

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/aim/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(useCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(providerCmd)
	rootCmd.AddCommand(toolCmd)
	rootCmd.AddCommand(keysCmd)
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if verbose {
		fmt.Fprintln(os.Stderr, "Verbose mode enabled")
	}

	// Configuration is now initialized in main.go
	// No need for individual command loading
}
