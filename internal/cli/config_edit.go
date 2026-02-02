package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fakecore/aim/internal/config"
	"github.com/spf13/cobra"
)

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit configuration file",
	Long:  `Open the configuration file in your default editor ($EDITOR).`,
	RunE:  configEdit,
}

func init() {
	configCmd.AddCommand(configEditCmd)
}

func configEdit(cmd *cobra.Command, args []string) error {
	configPath := config.ConfigPath()

	// Get editor from environment
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		// Try common editors
		for _, e := range []string{"vim", "nano", "vi"} {
			if _, err := exec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
	}
	if editor == "" {
		return fmt.Errorf("no editor found. Set $EDITOR environment variable")
	}

	// Ensure config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		defaultConfig := `version: "2"
accounts:
  # Add your accounts here
  # deepseek: ${DEEPSEEK_API_KEY}
`
		if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
	}

	// Open editor
	editorCmd := exec.Command(editor, configPath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	return editorCmd.Run()
}
