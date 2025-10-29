package cmd

import (
	"fmt"

	"github.com/fakecore/aim/internal/config"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test [key-name]",
	Short: "Test key configuration",
	Long:  `Test connectivity and configuration for keys using the v2.0 simplified configuration.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runTest,
}

func init() {
	testCmd.Flags().Bool("all", false, "Test all configured keys")
	testCmd.Flags().Bool("verbose", false, "Verbose output")
}

func runTest(cmd *cobra.Command, args []string) error {
	testAll, _ := cmd.Flags().GetBool("all")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Get global configuration manager
	cm := config.GetConfigManager()
	cfg := cm.GetConfig()

	// Create resolver
	resolver := config.NewResolver(cfg)

	var keysToTest []string

	if testAll {
		// Test all configured keys
		for keyName := range cfg.Keys {
			keysToTest = append(keysToTest, keyName)
		}
	} else if len(args) > 0 {
		// Test specific key
		keyName := args[0]
		if _, exists := cfg.GetKey(keyName); !exists {
			return fmt.Errorf("key '%s' not found", keyName)
		}
		keysToTest = append(keysToTest, keyName)
	} else {
		// Test default key
		defaultKey := cfg.Settings.DefaultKey
		if defaultKey == "" {
			return fmt.Errorf("no default key configured. Use 'aim config set default-key <key-name>' or specify a key to test")
		}
		keysToTest = append(keysToTest, defaultKey)
	}

	if len(keysToTest) == 0 {
		return fmt.Errorf("no keys to test")
	}

	fmt.Printf("Testing %d key(s)...\n\n", len(keysToTest))

	// Test each key
	successCount := 0
	for _, keyName := range keysToTest {
		if err := testSingleKey(resolver, keyName, verbose); err != nil {
			fmt.Printf("❌ %s: %v\n", keyName, err)
		} else {
			fmt.Printf("✅ %s: OK\n", keyName)
			successCount++
		}
	}

	fmt.Printf("\nTest Results: %d/%d passed\n", successCount, len(keysToTest))
	if successCount < len(keysToTest) {
		return fmt.Errorf("some tests failed")
	}

	return nil
}

// testSingleKey tests a single key configuration
func testSingleKey(resolver *config.Resolver, keyName string, verbose bool) error {
	// Validate key
	if err := resolver.ValidateKey(keyName); err != nil {
		return fmt.Errorf("invalid key: %w", err)
	}

	// Get key configuration
	cfg := resolver.GetConfig()
	key, _ := cfg.GetKey(keyName)

	if verbose {
		fmt.Printf("Testing key '%s' (provider: %s)...\n", keyName, key.Provider)
	}

	// Test by trying to resolve with a known tool
	// We'll use claude-code as the test tool since it's commonly available
	runtime, err := resolver.Resolve("claude-code", keyName, "")
	if err != nil {
		return fmt.Errorf("failed to resolve configuration: %w", err)
	}

	if verbose {
		fmt.Printf("  Resolved configuration:\n")
		fmt.Printf("    Tool: %s\n", runtime.Tool)
		fmt.Printf("    Provider: %s\n", runtime.Provider)
		fmt.Printf("    Base URL: %s\n", runtime.BaseURL)
		fmt.Printf("    Model: %s\n", runtime.Model)
		fmt.Printf("    Timeout: %v\n", runtime.Timeout)
	}

	// Basic validation - check if we have the required fields
	if runtime.APIKey == "" {
		return fmt.Errorf("API key is empty")
	}

	if runtime.BaseURL == "" {
		return fmt.Errorf("base URL is not configured")
	}

	if runtime.Model == "" {
		return fmt.Errorf("model is not configured")
	}

	// For now, we'll just validate the configuration structure
	// In a real implementation, you might want to make a test API call
	// But since we're in alpha stage, basic validation is sufficient

	if verbose {
		fmt.Printf("  Configuration validation: PASSED\n")
	}

	return nil
}
