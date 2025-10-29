package cmd

import (
	"context"
	"fmt"

	"github.com/fakecore/aim/internal/setup"
	"github.com/spf13/cobra"
)

// setupCmd setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup environment, install configuration, or generate commands",
	Long: `Setup provides four main functions:
1. env - Export environment variables for tools
2. export - Export environment variables for tools (alias for env)
3. install - Install configuration to tool config files
4. command - Generate execution commands

Examples:
	 # Export environment variables
	 eval $(aim setup env cc --key glm-test)
	 aim setup env codex --key glm-test --type fish
	 eval $(aim setup export cc --key glm-test)
	 aim setup export codex --key glm-test --type fish

	 # Install configuration
	 aim setup install cc --key glm-test
	 aim setup install codex --key glm-test --dry-run

	 # Generate commands
	 aim setup command cc --key glm-test
	 aim setup command codex --key glm-test --format json`,
}

func init() {
	// Add subcommands
	setupCmd.AddCommand(setupEnvCmd)
	setupCmd.AddCommand(setupInstallCmd)
	setupCmd.AddCommand(setupCommandCmd)
	setupCmd.AddCommand(setupExportCmd)
	setupCmd.AddCommand(setupRestoreCmd)

	// Initialize flags
	initSetupEnvFlags()
	initSetupInstallFlags()
	initSetupCommandFlags()
	initSetupExportFlags()
	initSetupRestoreFlags()
}

// setupEnvCmd env subcommand
var setupEnvCmd = &cobra.Command{
	Use:   "env <tool>",
	Short: "Export environment variables",
	Long: `Export environment variables for a specific tool and key configuration.

Examples:
  # Export to zsh (default)
  eval $(aim setup env cc --key glm-test)

  # Export to fish
  aim setup env codex --key glm-test --type fish | source

  # Export to JSON
  aim setup env cc --key glm-test --type json`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// No arguments provided, show help instead of error
			cmd.Help()
			return nil
		}
		if len(args) != 1 {
			return fmt.Errorf("accepts 1 arg(s), received %d", len(args))
		}
		return nil
	},
	RunE: runSetupEnv,
}

// setupInstallCmd install subcommand
var setupInstallCmd = &cobra.Command{
	Use:   "install <tool>",
	Short: "Install configuration to tool",
	Long: `Install configuration to tool's config file.

Examples:
  # Install to Claude Code
  aim setup install cc --key glm-test

  # Install to Codex with dry run
  aim setup install codex --key glm-test --dry-run

  # Force overwrite existing config
  aim setup install cc --key glm-test --force`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("accepts 1 arg(s), received %d", len(args))
		}
		return runSetupInstall(cmd, args)
	},
}

// setupCommandCmd command subcommand
var setupCommandCmd = &cobra.Command{
	Use:   "command <tool>",
	Short: "Generate execution command",
	Long: `Generate execution command for a specific tool and key configuration.

Examples:
	 # Generate raw command (with environment variables)
	 aim setup command cc --key glm-test

	 # Generate shell-escaped command (with environment variables)
	 aim setup command codex --key glm-test --format shell

	 # Generate JSON output (with environment variables)
	 aim setup command cc --key glm-test --format json

	 # Generate simple command (without environment variables)
	 aim setup command cc --key glm-test --format simple`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("accepts 1 arg(s), received %d", len(args))
		}
		return runSetupCommand(cmd, args)
	},
}

// setupExportCmd export subcommand
var setupExportCmd = &cobra.Command{
	Use:   "export <tool>",
	Short: "Export environment variables (alias for env)",
	Long: `Export environment variables for a specific tool and key configuration.
This is an alias for the 'env' subcommand.

Examples:
  # Export to zsh (default)
  eval $(aim setup export cc --key glm-test)

  # Export to fish
  aim setup export codex --key glm-test --type fish | source

  # Export to JSON
  aim setup export cc --key glm-test --type json`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("accepts 1 arg(s), received %d", len(args))
		}
		return runSetupEnv(cmd, args)
	},
}

// setupRestoreCmd restore subcommand
var setupRestoreCmd = &cobra.Command{
	Use:   "restore <tool>",
	Short: "Restore configuration from backup",
	Long: `Restore configuration from backup for a specific tool.

Examples:
  # Restore Claude Code configuration from latest backup
  aim setup restore claude-code

  # Restore Codex configuration from specific backup
  aim setup restore codex --backup-path ~/.codex/config.toml.bak.20251029_1504`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("accepts 1 arg(s), received %d", len(args))
		}
		return runSetupRestore(cmd, args)
	},
}

func initSetupEnvFlags() {
	// env subcommand flags
	setupEnvCmd.Flags().String("key", "", "API key name (required)")
	setupEnvCmd.Flags().String("type", "zsh", "Output type: zsh, bash, fish, json")
	setupEnvCmd.MarkFlagRequired("key")
}

func initSetupInstallFlags() {
	// install subcommand flags
	setupInstallCmd.Flags().String("key", "", "API key name (required)")
	setupInstallCmd.Flags().String("backup-path", "", "Custom backup path (optional)")
	setupInstallCmd.Flags().Bool("dry-run", false, "Preview changes without installing")
	setupInstallCmd.Flags().Bool("force", false, "Force overwrite existing config")
	setupInstallCmd.MarkFlagRequired("key")
}

func initSetupCommandFlags() {
	// command subcommand flags
	setupCommandCmd.Flags().String("key", "", "API key name (required)")
	setupCommandCmd.Flags().String("format", "raw", "Output format: raw, shell, json, simple")
	setupCommandCmd.MarkFlagRequired("key")
}

func initSetupExportFlags() {
	// export subcommand flags (same as env)
	setupExportCmd.Flags().String("key", "", "API key name (required)")
	setupExportCmd.Flags().String("type", "zsh", "Output type: zsh, bash, fish, json")
	setupExportCmd.MarkFlagRequired("key")
}

func initSetupRestoreFlags() {
	// restore subcommand flags
	setupRestoreCmd.Flags().String("backup-path", "", "Specific backup file to restore from (optional)")
}

func runSetupEnv(cmd *cobra.Command, args []string) error {
	toolName := args[0]
	keyName, _ := cmd.Flags().GetString("key")
	envType, _ := cmd.Flags().GetString("type")

	// Create setup request
	req := setup.NewSetupRequest(toolName, keyName)
	req.Type = envType

	// Create setup manager
	manager := setup.NewSetupManager(nil)

	// Export environment variables
	ctx := context.Background()
	result, err := manager.ExportEnv(ctx, req)
	if err != nil {
		return fmt.Errorf("export failed: %w", err)
	}

	// Format output
	output, err := manager.FormatEnv(result, req.Type)
	if err != nil {
		return fmt.Errorf("format failed: %w", err)
	}

	// Output to stdout
	fmt.Print(output)

	return nil
}

func runSetupInstall(cmd *cobra.Command, args []string) error {
	toolName := args[0]
	keyName, _ := cmd.Flags().GetString("key")
	backupPath, _ := cmd.Flags().GetString("backup-path")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	force, _ := cmd.Flags().GetBool("force")

	// Create install request
	req := setup.NewInstallRequest(toolName, keyName)
	req.BackupPath = backupPath
	req.DryRun = dryRun
	req.Force = force

	// Create setup manager
	manager := setup.NewSetupManager(nil)

	// Install configuration
	ctx := context.Background()
	result, err := manager.InstallConfig(ctx, req)
	if err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	// Display result
	if dryRun {
		fmt.Println("Dry run completed. No changes were made.")
	} else {
		fmt.Printf("✓ Configuration installed to %s\n", result.Metadata.ConfigPath)
		if result.Metadata.BackupPath != "" {
			fmt.Printf("✓ Backup created at %s\n", result.Metadata.BackupPath)
		}
	}

	return nil
}

func runSetupRestore(cmd *cobra.Command, args []string) error {
	toolName := args[0]
	backupPath, _ := cmd.Flags().GetString("backup-path")

	// Create setup manager
	manager := setup.NewSetupManager(nil)

	// Restore configuration
	ctx := context.Background()
	result, err := manager.RestoreConfig(ctx, toolName, backupPath)
	if err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	// Display result
	fmt.Printf("✓ Configuration restored from %s\n", result.Metadata.BackupPath)
	if result.Metadata.ConfigPath != "" {
		fmt.Printf("✓ Configuration restored to %s\n", result.Metadata.ConfigPath)
	}

	return nil
}

func runSetupCommand(cmd *cobra.Command, args []string) error {
	toolName := args[0]
	keyName, _ := cmd.Flags().GetString("key")
	format, _ := cmd.Flags().GetString("format")

	// Create setup request
	req := setup.NewSetupRequest(toolName, keyName)
	req.Format = format

	// Create setup manager
	manager := setup.NewSetupManager(nil)

	// Generate command
	ctx := context.Background()
	result, err := manager.GenerateCommand(ctx, req)
	if err != nil {
		return fmt.Errorf("command generation failed: %w", err)
	}

	// Format output
	output, err := manager.FormatCommand(result, req.Format)
	if err != nil {
		return fmt.Errorf("format failed: %w", err)
	}

	// Output to stdout
	fmt.Print(output)

	return nil
}
