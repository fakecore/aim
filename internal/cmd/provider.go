package cmd

import (
	"fmt"
	"strings"

	"github.com/fakecore/aim/internal/config"
	"github.com/fakecore/aim/internal/provider"
	"github.com/spf13/cobra"
)

var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Manage providers",
	Long:  `Manage AI provider configurations using v1.0 inheritance configuration.`,
}

var providerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all providers",
	Long:  `List all available providers (builtin and configured).`,
	RunE:  runProviderList,
}

var providerAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a global provider configuration",
	Long:  `Add a global provider configuration that can be used by all tools.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runProviderAdd,
}

var providerRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a global provider configuration",
	Long:  `Remove a global AI provider configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runProviderRemove,
}

var providerInfoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show provider information",
	Long:  `Show detailed information about a provider.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runProviderInfo,
}

func init() {
	// Flags for add command
	providerAddCmd.Flags().String("base-url", "", "Base URL for the provider")
	providerAddCmd.Flags().String("model", "", "Default model for the provider")
	providerAddCmd.Flags().Int("timeout", 0, "Timeout in milliseconds")

	// Add subcommands
	providerCmd.AddCommand(providerListCmd)
	providerCmd.AddCommand(providerAddCmd)
	providerCmd.AddCommand(providerRemoveCmd)
	providerCmd.AddCommand(providerInfoCmd)
}

func runProviderList(cmd *cobra.Command, args []string) error {
	// Get global configuration manager
	cm := config.GetConfigManager()
	cfg := cm.GetConfig()

	fmt.Println("\nBuiltin Providers:")
	builtinProviders := provider.GetBuiltinProviders()
	for _, providerInfo := range builtinProviders {
		fmt.Printf("  • %-20s (%s)\n", providerInfo.Name, providerInfo.DisplayName)
		if providerInfo.Description != "" {
			fmt.Printf("    Description: %s\n", providerInfo.Description)
		}
		if len(providerInfo.Endpoints) > 0 {
			fmt.Printf("    Endpoints: ")
			endpointNames := make([]string, len(providerInfo.Endpoints))
			for i, ep := range providerInfo.Endpoints {
				endpointNames[i] = ep.Name
			}
			fmt.Printf("%s\n", strings.Join(endpointNames, ", "))
		}
		fmt.Println()
	}

	if len(cfg.Providers) > 0 {
		fmt.Println("\nGlobal Provider Configurations:")
		for name, provider := range cfg.Providers {
			fmt.Printf("  • %-20s (global)\n", name)
			if provider != nil && provider.BaseURL != "" {
				fmt.Printf("    Base URL: %s\n", provider.BaseURL)
			}
			if provider != nil && provider.Model != "" {
				fmt.Printf("    Model: %s\n", provider.Model)
			}
			if provider != nil && provider.Timeout > 0 {
				fmt.Printf("    Timeout: %dms\n", provider.Timeout)
			}
			fmt.Println()
		}
	} else {
		fmt.Println("\nGlobal Provider Configurations:")
		fmt.Println("  (none)")
		fmt.Println("\nTip: Use 'aim provider add <name>' to add global provider configurations")
	}

	return nil
}

func runProviderAdd(cmd *cobra.Command, args []string) error {
	providerName := args[0]

	// Get global configuration manager
	cm := config.GetConfigManager()
	cfg := cm.GetConfig()

	// Check if provider already exists
	if _, exists := cfg.GetProvider(providerName); exists {
		return fmt.Errorf("provider '%s' already exists", providerName)
	}

	baseURL, _ := cmd.Flags().GetString("base-url")
	model, _ := cmd.Flags().GetString("model")
	timeout, _ := cmd.Flags().GetInt("timeout")

	// Update configuration
	err := cm.UpdateConfig(func(cfg *config.Config) {
		// Create provider configuration
		provider := &config.Provider{}

		if baseURL != "" {
			provider.BaseURL = baseURL
		}
		if model != "" {
			provider.Model = model
		}
		if timeout > 0 {
			provider.Timeout = timeout
		}

		// Add to configuration
		if cfg.Providers == nil {
			cfg.Providers = make(map[string]*config.Provider)
		}
		cfg.Providers[providerName] = provider
	})

	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	// Force save to disk immediately
	if err := cm.ForceSave(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✓ Added global provider configuration '%s'\n", providerName)
	if baseURL != "" {
		fmt.Printf("  Base URL: %s\n", baseURL)
	}
	if model != "" {
		fmt.Printf("  Model: %s\n", model)
	}
	if timeout > 0 {
		fmt.Printf("  Timeout: %dms\n", timeout)
	}

	return nil
}

func runProviderRemove(cmd *cobra.Command, args []string) error {
	providerName := args[0]

	// Get global configuration manager
	cm := config.GetConfigManager()
	cfg := cm.GetConfig()

	// Check if provider exists
	if _, exists := cfg.GetProvider(providerName); !exists {
		return fmt.Errorf("provider '%s' not found", providerName)
	}

	// Update configuration
	err := cm.UpdateConfig(func(cfg *config.Config) {
		delete(cfg.Providers, providerName)
	})

	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	// Force save to disk immediately
	if err := cm.ForceSave(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✓ Removed global provider configuration '%s'\n", providerName)

	return nil
}

func runProviderInfo(cmd *cobra.Command, args []string) error {
	providerName := args[0]

	// Get global configuration manager
	cm := config.GetConfigManager()
	cfg := cm.GetConfig()

	fmt.Printf("\nProvider: %s\n", providerName)

	// Check if it's a builtin provider
	builtinInfo, isBuiltin := provider.GetBuiltinProvider(providerName)
	if isBuiltin {
		// Use the new FormatProviderInfo function
		info, err := provider.FormatProviderInfo(providerName)
		if err != nil {
			return fmt.Errorf("failed to format provider info: %w", err)
		}
		fmt.Println(info)

		// Also show global config if exists
		if builtinInfo.Website != "" {
			fmt.Printf("  Website: %s\n", builtinInfo.Website)
		}
	}

	// Check if there's a global configuration
	globalProvider, hasGlobal := cfg.GetProvider(providerName)
	if hasGlobal {
		fmt.Printf("  Type: global configuration\n")
		if globalProvider.BaseURL != "" {
			fmt.Printf("  Base URL: %s\n", globalProvider.BaseURL)
		}
		if globalProvider.Model != "" {
			fmt.Printf("  Model: %s\n", globalProvider.Model)
		}
		if globalProvider.Timeout > 0 {
			fmt.Printf("  Timeout: %dms\n", globalProvider.Timeout)
		}
	}

	if !isBuiltin && !hasGlobal {
		fmt.Printf("  Type: not found\n")
		fmt.Printf("\nAvailable providers:\n%s\n", provider.FormatProviderListWithDetails())
		return fmt.Errorf("provider '%s' not found", providerName)
	}

	// Show which tools use this provider
	resolver := config.NewResolver(cfg)
	providers := resolver.ListProviders()
	if providers[providerName] {
		fmt.Printf("\n  Used by tools:\n")
		for toolName, tool := range cfg.Tools {
			for profileName, profile := range tool.Profiles {
				if profile.Provider == providerName {
					fmt.Printf("    • %s (profile: %s)\n", toolName, profileName)
					break
				}
			}
		}
	}

	return nil
}
