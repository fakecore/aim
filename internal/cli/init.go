package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fakecore/aim/internal/config"
	"github.com/fakecore/aim/internal/errors"
	"github.com/fakecore/aim/internal/vendors"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize AIM configuration",
	Long:  `Create a new configuration file with default structure.`,
	RunE:  initialize,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func initialize(cmd *cobra.Command, args []string) error {
	configPath := config.ConfigPath()

	// Check if already exists
	if _, err := os.Stat(configPath); err == nil {
		return &errors.Error{
			Code:     "AIM-CFG-001",
			Category: "CFG",
			Message:  fmt.Sprintf("Configuration already exists at %s", configPath),
		}
	}

	// Create config directory
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Generate config with all built-in vendors
	defaultConfig := generateDefaultConfig()

	if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("AIM configuration initialized at %s\n", configPath)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Set your API keys as environment variables")
	fmt.Println("2. Uncomment and configure accounts in the config file")
	fmt.Println("3. Validate: aim config validate")
	fmt.Println("4. Run a tool: aim run cc -a <account>")

	return nil
}

// generateDefaultConfig generates the default configuration with all built-in vendors
func generateDefaultConfig() string {
	var sb strings.Builder

	sb.WriteString(`version: "2"

# =============================================================================
# VENDORS
# =============================================================================
# All vendors are explicitly defined here. You can modify these or add your own.
# Each vendor specifies which protocols it supports and the corresponding URLs.

vendors:
`)

	// Add all built-in vendors
	for name, vendor := range vendors.BuiltinVendors {
		sb.WriteString(fmt.Sprintf("  %s:\n", name))
		sb.WriteString("    protocols:\n")
		for proto, url := range vendor.Protocols {
			sb.WriteString(fmt.Sprintf("      %s: %s\n", proto, url))
		}
		if len(vendor.DefaultModels) > 0 {
			sb.WriteString("    default_models:\n")
			for proto, model := range vendor.DefaultModels {
				sb.WriteString(fmt.Sprintf("      %s: %s\n", proto, model))
			}
		}
	}

	sb.WriteString(`
# =============================================================================
# ACCOUNTS
# =============================================================================
# Define your API accounts here. Each account references a vendor above.
# The key can be:
#   - Plain text: sk-abc123
#   - Environment variable: ${DEEPSEEK_API_KEY}
#   - Base64 encoded: base64:YWJjMTIz

accounts:
  # Example: DeepSeek account using the deepseek vendor
  # deepseek:
  #   key: ${DEEPSEEK_API_KEY}
  #   vendor: deepseek

  # Example: GLM account with explicit vendor
  # glm-work:
  #   key: ${GLM_WORK_KEY}
  #   vendor: glm

  # Example: Multiple accounts for same vendor
  # glm-personal:
  #   key: ${GLM_PERSONAL_KEY}
  #   vendor: glm

# =============================================================================
# SETTINGS
# =============================================================================
settings:
  # Default account to use when -a flag is not specified
  # default_account: deepseek

  # Command timeout (e.g., 5m, 30s, 1h)
  command_timeout: 5m

  # UI language: auto (detect from system), en, zh
  # language: auto

  # Log level: debug, info, warn, error, silent
  # log_level: warn
`)

	return sb.String()
}
