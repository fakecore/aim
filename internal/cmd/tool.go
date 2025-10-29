package cmd

import (
	"fmt"

	"github.com/fakecore/aim/internal/config"
	"github.com/spf13/cobra"
)

var toolCmd = &cobra.Command{
	Use:   "tool",
	Short: "Manage tools",
	Long:  `Manage AI CLI tool configurations using v1.0 inheritance configuration.`,
}

var toolListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tools",
	Long:  `List all configured tools and their providers.`,
	RunE:  runToolList,
}

var toolAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a tool configuration",
	Long:  `Add a new tool configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runToolAdd,
}

var toolRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a tool configuration",
	Long:  `Remove a tool configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runToolRemove,
}

func init() {
	// Flags for add command
	toolAddCmd.Flags().String("command", "", "Command to execute (required)")
	toolAddCmd.Flags().String("provider", "", "Default provider for this tool")
	toolAddCmd.Flags().String("base-url", "", "Default base URL for this tool")
	toolAddCmd.Flags().String("model", "", "Default model for this tool")
	toolAddCmd.MarkFlagRequired("command")

	// Add subcommands
	toolCmd.AddCommand(toolListCmd)
	toolCmd.AddCommand(toolAddCmd)
	toolCmd.AddCommand(toolRemoveCmd)
}

func runToolList(cmd *cobra.Command, args []string) error {
	// Get global configuration manager
	cm := config.GetConfigManager()
	cfg := cm.GetConfig()

	fmt.Println("\nConfigured Tools:")

	if len(cfg.Tools) == 0 {
		fmt.Println("  (none)")
	} else {
		for name, tool := range cfg.Tools {
			fmt.Printf("  • %s\n", name)
			fmt.Printf("    Command: %s\n", tool.Command)
			if tool.Enabled {
				fmt.Printf("    Status: enabled\n")
			} else {
				fmt.Printf("    Status: disabled\n")
			}
			if len(tool.Profiles) > 0 {
				fmt.Printf("    Profiles: %s\n", config.GetProfileList(tool.Profiles))
			}
			if len(tool.FieldMapping) > 0 {
				fmt.Printf("    Field Mappings: %d mappings\n", len(tool.FieldMapping))
			}
			if tool.Defaults != nil {
				fmt.Printf("    Defaults: timeout=%dms", tool.Defaults.Timeout)
				if len(tool.Defaults.Env) > 0 {
					fmt.Printf(", env=%d variables", len(tool.Defaults.Env))
				}
				fmt.Println()
			}
			fmt.Println()
		}
	}

	fmt.Println("Tip: Use 'aim tool add <name>' to add new tools")
	fmt.Println("     Use 'aim tool remove <name>' to remove tools")

	return nil
}

func runToolAdd(cmd *cobra.Command, args []string) error {
	toolName := args[0]
	command, _ := cmd.Flags().GetString("command")
	provider, _ := cmd.Flags().GetString("provider")
	baseURL, _ := cmd.Flags().GetString("base-url")
	model, _ := cmd.Flags().GetString("model")

	// Get global configuration manager
	cm := config.GetConfigManager()
	cfg := cm.GetConfig()

	// Check if tool already exists
	if _, exists := cfg.GetTool(toolName); exists {
		return fmt.Errorf("tool '%s' already exists", toolName)
	}

	// Update configuration
	err := cm.UpdateConfig(func(cfg *config.Config) {
		// Create tool configuration
		tool := &config.ToolConfig{
			Command:      command,
			Enabled:      true,
			Profiles:     make(map[string]*config.ToolProfile),
			FieldMapping: make(map[string]string),
		}

		// Add default profile if specified
		if provider != "" {
			profile := &config.ToolProfile{
				Provider: provider,
			}
			if baseURL != "" {
				profile.BaseURL = baseURL
			}
			if model != "" {
				profile.Model = model
			}
			tool.Profiles[provider] = profile
		}

		// Add to configuration
		if cfg.Tools == nil {
			cfg.Tools = make(map[string]*config.ToolConfig)
		}
		cfg.Tools[toolName] = tool
	})

	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	// Force save to disk immediately
	if err := cm.ForceSave(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✓ Added tool '%s'\n", toolName)
	fmt.Printf("  Command: %s\n", command)
	if provider != "" {
		fmt.Printf("  Default Provider: %s\n", provider)
	}

	return nil
}

func runToolRemove(cmd *cobra.Command, args []string) error {
	toolName := args[0]

	// Get global configuration manager
	cm := config.GetConfigManager()
	cfg := cm.GetConfig()

	// Check if tool exists
	if _, exists := cfg.GetTool(toolName); !exists {
		return fmt.Errorf("tool '%s' not found", toolName)
	}

	// Update configuration
	err := cm.UpdateConfig(func(cfg *config.Config) {
		delete(cfg.Tools, toolName)
	})

	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	// Force save to disk immediately
	if err := cm.ForceSave(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✓ Removed tool '%s'\n", toolName)

	return nil
}
