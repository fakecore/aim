package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/fakecore/aim/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `Manage AIM v1.0 configuration files and settings.`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration",
	Long:  `Initialize AIM configuration file with v1.0 defaults.`,
	RunE:  runConfigInit,
}

var forceFlag bool

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current configuration.`,
	RunE:  runConfigShow,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set configuration value",
	Long:  `Set a configuration value in global settings.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigSet,
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get configuration value",
	Long:  `Get a configuration value from settings.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigGet,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Long:  `List all configuration values from settings.`,
	RunE:  runConfigList,
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit configuration file",
	Long:  `Open the configuration file in your default editor for batch editing.`,
	RunE:  runConfigEdit,
}

func init() {
	// Add subcommands
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configEditCmd)

	// Add flags for init command
	configInitCmd.Flags().BoolVar(&forceFlag, "force", false, "Force overwrite existing configuration")
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	// Get force flag
	force, _ := cmd.Flags().GetBool("force")

	// Create a new loader for initialization
	loader := config.NewLoader()

	// Check if global config already exists
	if _, err := os.Stat(loader.GetGlobalPath()); err == nil && !force {
		return fmt.Errorf("global config already exists at %s. Use --force to overwrite", loader.GetGlobalPath())
	}

	// If force flag is set, remove existing config
	if force {
		if err := os.Remove(loader.GetGlobalPath()); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove existing config: %w", err)
		}
	}

	// Initialize new config
	if err := loader.InitGlobal(); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	fmt.Println("✓ Initialized global configuration")
	fmt.Printf("  File: %s\n", loader.GetGlobalPath())
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Add API keys: aim keys add <key-name> --provider <provider> --key <api-key>")
	fmt.Println("  2. Set defaults: aim config set default-key <key-name>")
	fmt.Println("  3. Run tools: aim run <tool> --key <key-name>")

	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	// Get global configuration manager
	cm := config.GetConfigManager()
	cfg := cm.GetConfig()

	fmt.Printf("\nConfiguration (v%s):\n", cfg.Version)

	// Show settings
	fmt.Println("\nSettings:")
	if cfg.Settings.DefaultTool != "" {
		fmt.Printf("  Default Tool: %s\n", cfg.Settings.DefaultTool)
	}
	if cfg.Settings.DefaultProvider != "" {
		fmt.Printf("  Default Provider: %s\n", cfg.Settings.DefaultProvider)
	}
	if cfg.Settings.DefaultKey != "" {
		fmt.Printf("  Default Key: %s\n", cfg.Settings.DefaultKey)
	}
	if cfg.Settings.Timeout > 0 {
		fmt.Printf("  Timeout: %dms\n", cfg.Settings.Timeout)
	}
	if cfg.Settings.Language != "" {
		fmt.Printf("  Language: %s\n", cfg.Settings.Language)
	}

	// Show keys
	if len(cfg.Keys) > 0 {
		fmt.Println("\nKeys:")
		for name, key := range cfg.Keys {
			fmt.Printf("  %s:\n", name)
			fmt.Printf("    Provider: %s\n", key.Provider)
			if key.Description != "" {
				fmt.Printf("    Description: %s\n", key.Description)
			}
			fmt.Printf("    Key: %s\n", maskKey(key.Key))
		}
	}

	// Show global providers
	if len(cfg.Providers) > 0 {
		fmt.Println("\nGlobal Providers:")
		for name, provider := range cfg.Providers {
			fmt.Printf("  %s:\n", name)
			if provider.BaseURL != "" {
				fmt.Printf("    Base URL: %s\n", provider.BaseURL)
			}
			if provider.Model != "" {
				fmt.Printf("    Model: %s\n", provider.Model)
			}
			if provider.Timeout > 0 {
				fmt.Printf("    Timeout: %dms\n", provider.Timeout)
			}
		}
	}

	// Show tools
	if len(cfg.Tools) > 0 {
		fmt.Println("\nTools:")
		for name, tool := range cfg.Tools {
			fmt.Printf("  %s:\n", name)
			fmt.Printf("    Command: %s\n", tool.Command)
			if tool.Enabled {
				fmt.Printf("    Enabled: true\n")
			}
			if len(tool.Profiles) > 0 {
				fmt.Printf("    Profiles: %s\n", config.GetProfileList(tool.Profiles))
				// Show detailed profile information
				for profileName, profile := range tool.Profiles {
					fmt.Printf("      %s:\n", profileName)
					fmt.Printf("        Provider: %s\n", profile.Provider)
					if profile.BaseURL != "" {
						fmt.Printf("        Base URL: %s\n", profile.BaseURL)
					}
					if profile.Model != "" {
						fmt.Printf("        Model: %s\n", profile.Model)
					}
					if profile.Timeout > 0 {
						fmt.Printf("        Timeout: %dms\n", profile.Timeout)
					}
					if len(profile.Env) > 0 {
						fmt.Printf("        Environment Variables: %d\n", len(profile.Env))
					}
				}
			}
			if tool.Defaults != nil {
				fmt.Printf("    Defaults:\n")
				if tool.Defaults.Timeout > 0 {
					fmt.Printf("      Timeout: %dms\n", tool.Defaults.Timeout)
				}
				if len(tool.Defaults.Env) > 0 {
					fmt.Printf("      Environment Variables: %d\n", len(tool.Defaults.Env))
				}
			}
		}
	}

	// Show aliases - DISABLED: Alias functionality temporarily disabled
	/*
		if len(cfg.Aliases) > 0 {
			fmt.Println("\nAliases:")
			for alias, target := range cfg.Aliases {
				fmt.Printf("  %s -> %s\n", alias, target)
			}
		}
	*/

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	// Get global configuration manager
	cm := config.GetConfigManager()

	// Update configuration
	err := cm.UpdateConfig(func(cfg *config.Config) {
		// Set the appropriate setting
		switch key {
		case "default-tool":
			cfg.Settings.DefaultTool = value
		case "default-provider":
			cfg.Settings.DefaultProvider = value
		case "default-key":
			cfg.Settings.DefaultKey = value
		case "timeout":
			if timeout, err := strconv.Atoi(value); err == nil {
				cfg.Settings.Timeout = timeout
			}
			// If invalid, keep the existing value (silently ignore)
		case "language":
			cfg.Settings.Language = value
		default:
			return
		}
	})

	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	// Force save to disk immediately
	if err := cm.ForceSave(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✓ Set %s = %s\n", key, value)
	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	// Get global configuration manager
	cm := config.GetConfigManager()
	cfg := cm.GetConfig()

	// Get the appropriate setting
	var value string
	switch key {
	case "default-tool":
		value = cfg.Settings.DefaultTool
	case "default-provider":
		value = cfg.Settings.DefaultProvider
	case "default-key":
		value = cfg.Settings.DefaultKey
	case "timeout":
		if cfg.Settings.Timeout > 0 {
			value = fmt.Sprintf("%d", cfg.Settings.Timeout)
		}
	case "language":
		value = cfg.Settings.Language
	default:
		return fmt.Errorf("unknown setting key: %s", key)
	}

	if value == "" {
		fmt.Printf("%s: (not set)\n", key)
	} else {
		fmt.Printf("%s: %s\n", key, value)
	}

	return nil
}

func runConfigList(cmd *cobra.Command, args []string) error {
	// Get global configuration manager
	cm := config.GetConfigManager()
	cfg := cm.GetConfig()

	fmt.Println("\nConfiguration Settings:")

	settings := []struct {
		key   string
		value string
	}{
		{"default-tool", cfg.Settings.DefaultTool},
		{"default-provider", cfg.Settings.DefaultProvider},
		{"default-key", cfg.Settings.DefaultKey},
		{"timeout", fmt.Sprintf("%d", cfg.Settings.Timeout)},
		{"language", cfg.Settings.Language},
	}

	for _, setting := range settings {
		if setting.value == "" {
			fmt.Printf("  %s: (not set)\n", setting.key)
		} else {
			fmt.Printf("  %s: %s\n", setting.key, setting.value)
		}
	}

	return nil
}

func runConfigEdit(cmd *cobra.Command, args []string) error {
	// Get global configuration manager
	cm := config.GetConfigManager()
	loader := config.NewLoader()
	configPath := loader.GetGlobalPath()

	// Check if config exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("configuration file not found. Run 'aim config init' first")
	}

	// Get editor from environment or use default
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		// Try common editors
		for _, e := range []string{"vim", "vi", "nano", "emacs"} {
			if path, err := exec.LookPath(e); err == nil {
				editor = path
				break
			}
		}
	}
	if editor == "" {
		return fmt.Errorf("no editor found. Please set EDITOR or VISUAL environment variable")
	}

	fmt.Printf("Opening configuration file in %s...\n", editor)
	fmt.Printf("  File: %s\n\n", configPath)

	// Open editor
	editCmd := exec.Command(editor, configPath)
	editCmd.Stdin = os.Stdin
	editCmd.Stdout = os.Stdout
	editCmd.Stderr = os.Stderr

	if err := editCmd.Run(); err != nil {
		return fmt.Errorf("failed to run editor: %w", err)
	}

	// Reload and validate configuration
	newCfg, err := loader.Load()
	if err != nil {
		fmt.Printf("⚠️  Warning: Configuration may have errors: %v\n", err)
		fmt.Println("Please check and fix the configuration file.")
		return err
	}

	// Update the config manager with the new configuration
	err = cm.UpdateConfig(func(cfg *config.Config) {
		*cfg = *newCfg
	})
	if err != nil {
		return fmt.Errorf("failed to update configuration: %w", err)
	}

	fmt.Println("\n✓ Configuration updated successfully")
	return nil
}

