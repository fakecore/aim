package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fakecore/aim/internal/config"
	"github.com/fakecore/aim/internal/migration"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate v1 config to v2",
	Long:  `Convert AIM v1 configuration to v2 format.`,
	RunE:  runMigrate,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

func runMigrate(cmd *cobra.Command, args []string) error {
	// Find v1 config
	home, _ := os.UserHomeDir()
	v1Path := filepath.Join(home, ".config", "aim", "config.toml")

	if _, err := os.Stat(v1Path); os.IsNotExist(err) {
		return fmt.Errorf("v1 config not found at %s", v1Path)
	}

	// Check if v2 already exists
	v2Path := config.ConfigPath()
	if _, err := os.Stat(v2Path); err == nil {
		return fmt.Errorf("v2 config already exists at %s", v2Path)
	}

	// Load v1
	v1, err := migration.LoadV1(v1Path)
	if err != nil {
		return fmt.Errorf("failed to load v1 config: %w", err)
	}

	// Migrate
	v2 := migration.Migrate(v1)

	// Write v2
	if err := migration.WriteV2(v2, v2Path); err != nil {
		return fmt.Errorf("failed to write v2 config: %w", err)
	}

	fmt.Printf("Migrated configuration from v1 to v2\n")
	fmt.Printf("  From: %s\n", v1Path)
	fmt.Printf("  To:   %s\n", v2Path)
	fmt.Println()
	fmt.Println("Please review the migrated configuration:")
	fmt.Printf("  aim config validate\n")
	fmt.Printf("  aim config show\n")

	return nil
}
