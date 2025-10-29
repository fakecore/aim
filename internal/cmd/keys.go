package cmd

import (
	"fmt"

	"github.com/fakecore/aim/internal/config"
	"github.com/fakecore/aim/internal/provider"
	"github.com/spf13/cobra"
)

var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Manage API keys",
	Long:  `Manage API keys for AI providers using the v1.0 inheritance configuration.`,
}

var keysAddCmd = &cobra.Command{
	Use:   "add <key-name> --provider <provider> --key <api-key>",
	Short: "Add an API key",
	Long: `Add an API key for a provider using the v1.0 configuration format.

Examples:
  # Add a DeepSeek key
  aim keys add deepseek-work --provider deepseek --key sk-xxx-work --description "Work DeepSeek account"

  # Add a GLM key
  aim keys add glm-shared --provider glm --key glm-xxx --description "Shared GLM account"

Available providers:
` + provider.FormatProviderForHelp() + `

Use 'aim provider list' to see all available providers.`,
	Args: cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		providerName, _ := cmd.Flags().GetString("provider")
		apiKey, _ := cmd.Flags().GetString("key")

		if providerName == "" {
			fmt.Println("❌ Error: Missing required flag --provider")
			fmt.Println("\nAvailable providers:")
			fmt.Println(provider.FormatProviderListWithDetails())
			fmt.Println("\nExample usage:")
			fmt.Println("  aim keys add my-key --provider glm --key your-api-key")
			return fmt.Errorf("--provider flag is required")
		}

		// Get current config
		cm := config.GetConfigManager()
		cfg := cm.GetConfig()

		// Check if provider exists in config (global providers or tool providers)
		providerExists := false

		// Check global providers
		if _, exists := cfg.Providers[providerName]; exists {
			providerExists = true
		}

		// Check tool profiles
		if !providerExists {
			for _, tool := range cfg.Tools {
				if tool.Profiles != nil {
					for _, profile := range tool.Profiles {
						if profile.Provider == providerName {
							providerExists = true
							break
						}
					}
				}
			}
		}

		if !providerExists {
			// Provider not in config, check if it's a builtin
			if provider.IsBuiltinProvider(providerName) {
				fmt.Printf("❌ Error: Provider '%s' is not configured\n\n", providerName)
				fmt.Printf("'%s' is a builtin provider but hasn't been added to your configuration yet.\n\n", providerName)
				fmt.Println("To use this provider, it will be automatically available in the default configuration.")
				fmt.Println("However, you may need to reinitialize your config or manually add it.")
				fmt.Println("\nConfigured providers:")
				listConfiguredProviders(cfg)
				return fmt.Errorf("provider '%s' not configured", providerName)
			} else {
				// Check if it's a builtin provider with an endpoint suffix
				baseProvider := providerName
				isEndpointVariant := false
				if _, exists := provider.GetBuiltinProvider(providerName); exists {
					baseProvider = providerName
					isEndpointVariant = false
				} else {
					// Try to find base provider (e.g., "glm" from "glm-coding")
					for name := range provider.GetBuiltinProviders() {
						if len(providerName) > len(name) && providerName[:len(name)] == name {
							baseProvider = name
							isEndpointVariant = true
							break
						}
					}
				}

				if isEndpointVariant {
					fmt.Printf("❌ Error: Provider '%s' is not configured\n\n", providerName)
					fmt.Printf("'%s' appears to be an endpoint variant of '%s', but it's not in your configuration.\n\n", providerName, baseProvider)
					fmt.Println("To add this provider endpoint to your configuration:")
					fmt.Printf("  1. Check available endpoints: aim provider info %s\n", baseProvider)
					fmt.Println("  2. Manually add it to your config.yaml, or")
					fmt.Println("  3. Reinitialize your configuration")
					fmt.Println("\nConfigured providers:")
					listConfiguredProviders(cfg)
					return fmt.Errorf("provider '%s' not configured", providerName)
				} else {
					fmt.Printf("❌ Error: Unknown provider '%s'\n\n", providerName)
					fmt.Println("This provider is not builtin and not configured.")
					fmt.Println("\nConfigured providers:")
					listConfiguredProviders(cfg)
					fmt.Println("\nBuiltin providers:")
					fmt.Println(provider.FormatProviderListWithDetails())
					return fmt.Errorf("unknown provider: %s", providerName)
				}
			}
		}

		if apiKey == "" {
			fmt.Println("❌ Error: Missing required flag --key")
			fmt.Println("\nExample usage:")
			fmt.Println("  aim keys add my-key --provider glm --key your-api-key")
			return fmt.Errorf("--key flag is required")
		}

		return nil
	},
	RunE: runKeysAdd,
}

var keysListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all API keys",
	Long:  `List all configured API keys (masked).`,
	RunE:  runKeysList,
}

var keysRemoveCmd = &cobra.Command{
	Use:   "remove <key-name>",
	Short: "Remove an API key",
	Long: `Remove an API key configuration.

Examples:
  aim keys remove deepseek-work    # Remove the DeepSeek API key
  aim keys remove glm-shared       # Remove the GLM API key`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			fmt.Println("❌ Error: Missing key name")
			fmt.Println("\nUsage: aim keys remove <key-name>")
			fmt.Println("\nAvailable keys:")
			cm := config.GetConfigManager()
			cfg := cm.GetConfig()
			if len(cfg.Keys) == 0 {
				fmt.Println("  (no keys configured)")
				fmt.Println("\nUse 'aim keys add <name> --provider <provider> --key <api-key>' to add a key")
			} else {
				for name := range cfg.Keys {
					fmt.Printf("  • %s\n", name)
				}
			}
			return fmt.Errorf("key name required")
		}
		if len(args) > 1 {
			fmt.Println("❌ Error: Too many arguments")
			fmt.Println("\nUsage: aim keys remove <key-name>")
			return fmt.Errorf("only one key name allowed")
		}
		return nil
	},
	RunE: runKeysRemove,
}

var keysShowCmd = &cobra.Command{
	Use:   "show <key-name>",
	Short: "Show full API key",
	Long: `Show the full (unmasked) API key for a key configuration.

Examples:
  aim keys show deepseek-work    # Show the full DeepSeek API key
  aim keys show glm-shared       # Show the full GLM API key`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			fmt.Println("❌ Error: Missing key name")
			fmt.Println("\nUsage: aim keys show <key-name>")
			fmt.Println("\nAvailable keys:")
			cm := config.GetConfigManager()
			cfg := cm.GetConfig()
			if len(cfg.Keys) == 0 {
				fmt.Println("  (no keys configured)")
				fmt.Println("\nUse 'aim keys add <name> --provider <provider> --key <api-key>' to add a key")
			} else {
				for name := range cfg.Keys {
					fmt.Printf("  • %s\n", name)
				}
			}
			return fmt.Errorf("key name required")
		}
		if len(args) > 1 {
			fmt.Println("❌ Error: Too many arguments")
			fmt.Println("\nUsage: aim keys show <key-name>")
			return fmt.Errorf("only one key name allowed")
		}
		return nil
	},
	RunE: runKeysShow,
}

func init() {
	// Flags for add command
	keysAddCmd.Flags().String("provider", "", "Provider name (required)")
	keysAddCmd.Flags().String("key", "", "API key value (required)")
	keysAddCmd.Flags().String("description", "", "Description of the key")
	keysAddCmd.MarkFlagRequired("provider")
	keysAddCmd.MarkFlagRequired("key")

	// Add subcommands
	keysCmd.AddCommand(keysAddCmd)
	keysCmd.AddCommand(keysListCmd)
	keysCmd.AddCommand(keysRemoveCmd)
	keysCmd.AddCommand(keysShowCmd)
}

func runKeysAdd(cmd *cobra.Command, args []string) error {
	keyName := args[0]
	provider, _ := cmd.Flags().GetString("provider")
	apiKey, _ := cmd.Flags().GetString("key")
	description, _ := cmd.Flags().GetString("description")

	// Get global configuration manager
	cm := config.GetConfigManager()
	cfg := cm.GetConfig()

	// Check if key already exists
	if _, exists := cfg.GetKey(keyName); exists {
		return fmt.Errorf("key '%s' already exists", keyName)
	}

	// Update configuration
	err := cm.UpdateConfig(func(cfg *config.Config) {
		if cfg.Keys == nil {
			cfg.Keys = make(map[string]*config.Key)
		}

		cfg.Keys[keyName] = &config.Key{
			Provider:    provider,
			Key:         apiKey,
			Description: description,
		}
	})

	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	// Force save to disk immediately
	if err := cm.ForceSave(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✓ Added key '%s' for provider '%s'\n", keyName, provider)
	if description != "" {
		fmt.Printf("  Description: %s\n", description)
	}
	fmt.Printf("  Key: %s\n", maskKey(apiKey))

	return nil
}

func runKeysList(cmd *cobra.Command, args []string) error {
	// Get global configuration manager
	cm := config.GetConfigManager()
	cfg := cm.GetConfig()

	fmt.Println("\nConfigured API keys:")

	if len(cfg.Keys) == 0 {
		fmt.Println("  (none)")
	} else {
		// Show configured keys
		for name, key := range cfg.Keys {
			maskedKey := maskKey(key.Key)
			providerDisplay := key.Provider

			fmt.Printf("  ✓ %-20s %-15s %s\n", name+":", providerDisplay+":", maskedKey)
			if key.Description != "" {
				fmt.Printf("    %s\n", key.Description)
			}
		}
	}

	fmt.Println("\nTip: Use 'aim keys add <key-name> --provider <provider> --key <api-key>' to add keys")
	fmt.Println("     Use 'aim keys remove <key-name>' to remove keys")
	fmt.Println("     Use 'aim config edit' to open config file for manual editing")

	return nil
}

func runKeysRemove(cmd *cobra.Command, args []string) error {
	keyName := args[0]

	// Get global configuration manager
	cm := config.GetConfigManager()
	cfg := cm.GetConfig()

	// Check if key exists
	if _, exists := cfg.GetKey(keyName); !exists {
		fmt.Printf("❌ Error: Key '%s' not found\n\n", keyName)
		fmt.Println("Available keys:")
		if len(cfg.Keys) == 0 {
			fmt.Println("  (no keys configured)")
		} else {
			for name := range cfg.Keys {
				fmt.Printf("  • %s\n", name)
			}
		}
		fmt.Println("\nUse 'aim keys list' to see all configured keys")
		return nil
	}

	// Show confirmation
	fmt.Printf("Are you sure you want to remove key '%s'? (y/N): ", keyName)
	var confirm string
	fmt.Scanln(&confirm)

	if confirm != "y" && confirm != "Y" {
		fmt.Println("Cancelled")
		return nil
	}

	// Update configuration
	err := cm.UpdateConfig(func(cfg *config.Config) {
		delete(cfg.Keys, keyName)
	})

	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	// Force save to disk immediately
	if err := cm.ForceSave(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✓ Removed key '%s'\n", keyName)

	return nil
}

func runKeysShow(cmd *cobra.Command, args []string) error {
	keyName := args[0]

	// Get global configuration manager
	cm := config.GetConfigManager()
	cfg := cm.GetConfig()

	// Get key
	key, exists := cfg.GetKey(keyName)
	if !exists {
		fmt.Printf("❌ Error: Key '%s' not found\n\n", keyName)
		fmt.Println("Available keys:")
		if len(cfg.Keys) == 0 {
			fmt.Println("  (no keys configured)")
		} else {
			for name := range cfg.Keys {
				fmt.Printf("  • %s\n", name)
			}
		}
		fmt.Println("\nUse 'aim keys list' to see all configured keys")
		fmt.Println("Use 'aim keys add <name> --provider <provider> --key <api-key>' to add a new key")
		return nil
	}

	// Show warning
	fmt.Println("\n⚠️  WARNING: This will display the full API key")

	// Show full key
	fmt.Printf("\nKey Name: %s\n", keyName)
	fmt.Printf("Provider: %s\n", key.Provider)
	if key.Description != "" {
		fmt.Printf("Description: %s\n", key.Description)
	}
	fmt.Printf("API Key: %s\n\n", key.Key)

	return nil
}

// maskKey masks an API key for display
func maskKey(key string) string {
	if key == "" {
		return ""
	}

	if len(key) <= 8 {
		return "****"
	}

	// Show first 4 and last 4 characters
	return key[:4] + "****..." + key[len(key)-4:]
}

// listConfiguredProviders lists all providers configured in the config
func listConfiguredProviders(cfg *config.Config) {
	// Collect all unique provider names from global and tool configs
	providerSet := make(map[string]bool)

	// From global providers
	for name := range cfg.Providers {
		providerSet[name] = true
	}

	// From tool profiles
	for _, tool := range cfg.Tools {
		if tool.Profiles != nil {
			for _, profile := range tool.Profiles {
				if profile.Provider != "" {
					providerSet[profile.Provider] = true
				}
			}
		}
	}

	if len(providerSet) == 0 {
		fmt.Println("  (none configured)")
		return
	}

	// Display providers
	for name := range providerSet {
		// Check if it's a builtin provider
		if builtinInfo, exists := provider.GetBuiltinProvider(name); exists {
			fmt.Printf("  • %-15s - %s\n", name, builtinInfo.Description)
		} else {
			// Try to match endpoint variant
			matched := false
			for baseName, builtinInfo := range provider.GetBuiltinProviders() {
				if len(name) > len(baseName) && name[:len(baseName)] == baseName {
					fmt.Printf("  • %-15s - %s (endpoint variant)\n", name, builtinInfo.Description)
					matched = true
					break
				}
			}
			if !matched {
				fmt.Printf("  • %-15s - (custom provider)\n", name)
			}
		}
	}
}
