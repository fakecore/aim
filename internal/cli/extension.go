package cli

import (
	"fmt"
	"os"

	"github.com/fakecore/aim/internal/extension"
	"github.com/spf13/cobra"
)

var extensionCmd = &cobra.Command{
	Use:   "extension",
	Short: "Manage vendor extensions",
	Long:  `Add, list, and update vendor extensions for custom providers.`,
}

var extensionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed extensions",
	RunE:  extensionList,
}

func init() {
	rootCmd.AddCommand(extensionCmd)
	extensionCmd.AddCommand(extensionListCmd)
}

func extensionList(cmd *cobra.Command, args []string) error {
	dir := extension.DefaultDir()
	if envDir := os.Getenv("AIM_EXTENSIONS"); envDir != "" {
		dir = envDir
	}

	exts, err := extension.LoadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to load extensions: %w", err)
	}

	if len(exts) == 0 {
		fmt.Println("No custom extensions installed.")
		fmt.Println("Builtin vendors: deepseek, glm, kimi, qwen")
		return nil
	}

	fmt.Println("Installed extensions:")
	fmt.Println()
	for name, ext := range exts {
		fmt.Printf("  %s", name)
		if ext.Version != "" {
			fmt.Printf(" (%s)", ext.Version)
		}
		fmt.Println()
		if ext.Description != "" {
			fmt.Printf("    %s\n", ext.Description)
		}
		fmt.Printf("    Protocols: %d\n", len(ext.Protocols))
	}

	return nil
}
