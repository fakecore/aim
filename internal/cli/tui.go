package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fakecore/aim/internal/config"
	"github.com/fakecore/aim/internal/tui"
	"github.com/fakecore/aim/internal/vendors"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Open TUI config editor",
	Long:  `Launch the interactive Terminal UI for managing AIM configuration.`,
	RunE:  runTUI,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

func runTUI(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(config.ConfigPath())
	if err != nil {
		cfg = &config.Config{
			Version:  "2",
			Accounts: make(map[string]config.Account),
			Vendors:  make(map[string]vendors.VendorConfig),
		}
	}

	model := tui.New(cfg)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}
