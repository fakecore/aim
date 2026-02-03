package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fakecore/aim/internal/config"
	"github.com/fakecore/aim/internal/errors"
	"github.com/fakecore/aim/internal/tools"
	"github.com/fakecore/aim/internal/vendors"
	"github.com/spf13/cobra"
)

var forceFlag bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize AIM configuration",
	Long:  `Create a new configuration file with default structure.`,
	RunE:  initialize,
}

func init() {
	initCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force re-initialization, backing up existing config")
	rootCmd.AddCommand(initCmd)
}

func initialize(cmd *cobra.Command, args []string) error {
	configPath := config.ConfigPath()

	// Check if already exists
	if _, err := os.Stat(configPath); err == nil {
		if !forceFlag {
			return &errors.Error{
				Code:     "AIM-CFG-001",
				Category: "CFG",
				Message:  fmt.Sprintf("Configuration already exists at %s", configPath),
				Suggestions: []string{
					"Use --force to overwrite with backup",
					"Or edit the existing config file directly",
				},
			}
		}
		// Backup existing config
		backupPath := configPath + ".backup"
		if err := os.Rename(configPath, backupPath); err != nil {
			return fmt.Errorf("failed to backup existing config: %w", err)
		}
		fmt.Printf("Backed up existing config to %s\n", backupPath)
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
	fmt.Println("2. Configure keys in the config file")
	fmt.Println("3. Configure accounts to reference keys")
	fmt.Println("4. Validate: aim config validate")
	fmt.Println("5. Run a tool: aim run cc -a <account>")

	return nil
}

// generateDefaultConfig generates the default configuration with all built-in vendors
func generateDefaultConfig() string {
	var sb strings.Builder

	sb.WriteString(`version: "2"

# =============================================================================
# TOOLS
# =============================================================================
# Tool protocol mappings. All tools must be mapped to a protocol (endpoint type).
# Common protocols: openai, anthropic

tools:
`)

	// Add tools in alphabetical order
	toolNames := make([]string, 0, len(tools.BuiltinToolsConfig))
	for name := range tools.BuiltinToolsConfig {
		toolNames = append(toolNames, name)
	}
	for i := 0; i < len(toolNames); i++ {
		for j := i + 1; j < len(toolNames); j++ {
			if toolNames[i] > toolNames[j] {
				toolNames[i], toolNames[j] = toolNames[j], toolNames[i]
			}
		}
	}

	for _, name := range toolNames {
		toolCfg := tools.BuiltinToolsConfig[name]
		sb.WriteString(fmt.Sprintf("  %s:\n", name))
		sb.WriteString(fmt.Sprintf("    protocol: %s\n", toolCfg.Protocol))
	}

	sb.WriteString(`
# =============================================================================
# VENDORS
# =============================================================================
# All vendors are explicitly defined here. Each vendor has multiple endpoints.
# Endpoint names correspond to protocol types (openai, anthropic, etc.)

vendors:
`)

	// Add all built-in vendors (alphabetically)
	vendorNames := make([]string, 0, len(vendors.BuiltinVendors))
	for name := range vendors.BuiltinVendors {
		vendorNames = append(vendorNames, name)
	}
	for i := 0; i < len(vendorNames); i++ {
		for j := i + 1; j < len(vendorNames); j++ {
			if vendorNames[i] > vendorNames[j] {
				vendorNames[i], vendorNames[j] = vendorNames[j], vendorNames[i]
			}
		}
	}

	for _, name := range vendorNames {
		vendor := vendors.BuiltinVendors[name]
		sb.WriteString(fmt.Sprintf("  %s:\n", name))
		sb.WriteString("    endpoints:\n")

		// Sort endpoints
		endpointNames := make([]string, 0, len(vendor.Endpoints))
		for epName := range vendor.Endpoints {
			endpointNames = append(endpointNames, epName)
		}
		for i := 0; i < len(endpointNames); i++ {
			for j := i + 1; j < len(endpointNames); j++ {
				if endpointNames[i] > endpointNames[j] {
					endpointNames[i], endpointNames[j] = endpointNames[j], endpointNames[i]
				}
			}
		}

		for _, epName := range endpointNames {
			ep := vendor.Endpoints[epName]
			sb.WriteString(fmt.Sprintf("      %s:\n", epName))
			sb.WriteString(fmt.Sprintf("        url: %s\n", ep.URL))
			if ep.DefaultModel != "" {
				sb.WriteString(fmt.Sprintf("        default_model: %s\n", ep.DefaultModel))
			}
		}
	}

	sb.WriteString(`
# =============================================================================
# KEYS
# =============================================================================
# Define your API keys here. Keys are referenced by accounts.
# Key formats:
#   - Plain text: sk-abc123
#   - Environment variable: ${API_KEY}
#   - Base64 encoded: base64:YWJjMTIz

keys:
  # Example: DeepSeek key
  # deepseek-main:
  #   value: ${DEEPSEEK_API_KEY}
  #   vendor: deepseek
  #   # Optional: restrict to specific endpoints
  #   # endpoints: [openai]

  # Example: GLM key
  # glm-main:
  #   value: ${GLM_API_KEY}
  #   vendor: glm

# =============================================================================
# ACCOUNTS
# =============================================================================
# Accounts reference keys and can override endpoint/model settings.
# This allows multiple usage profiles with different settings.

accounts:
  # Example: DeepSeek account (uses deepseek-main key)
  # deepseek:
  #   key: deepseek-main
  #   # Optional: override endpoint
  #   # endpoint: openai
  #   # Optional: override model
  #   # model: deepseek-chat

  # Example: GLM account with different model
  # glm:
  #   key: glm-main
  #   model: glm-4.7

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
