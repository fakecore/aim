package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fakecore/aim/internal/config"
	"github.com/fakecore/aim/internal/errors"
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

	// Write default config
	defaultConfig := `version: "2"

# Accounts define your API keys and associated vendors
accounts:
  # Example: DeepSeek account
  # deepseek: ${DEEPSEEK_API_KEY}

  # Example: GLM account with explicit vendor
  # glm-work:
  #   key: ${GLM_WORK_KEY}
  #   vendor: glm

# Optional: Override or define custom vendors
# vendors:
#   my-company:
#     protocols:
#       openai: https://ai.company.com/v1

# Optional: Global settings
options:
  # default_account: deepseek
  command_timeout: 5m
`

	if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("AIM configuration initialized at %s\n", configPath)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Set your API keys as environment variables")
	fmt.Println("2. Edit the config: aim config edit")
	fmt.Println("3. Validate: aim config validate")
	fmt.Println("4. Run a tool: aim run cc -a <account>")

	return nil
}
