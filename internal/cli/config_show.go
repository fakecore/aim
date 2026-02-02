package cli

import (
	"fmt"

	"github.com/fakecore/aim/internal/config"
	"github.com/fakecore/aim/internal/errors"
	"github.com/fakecore/aim/internal/vendors"
	"github.com/spf13/cobra"
)

var (
	showAccount string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage AIM configuration",
	Long:  `View and edit AIM configuration files and settings.`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show configuration for an account",
	Long:  `Display resolved configuration including account, vendor, and protocol information.`,
	RunE:  configShow,
}

func init() {
	configShowCmd.Flags().StringVarP(&showAccount, "account", "a", "", "Account to show")
	configCmd.AddCommand(configShowCmd)
}

func configShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(config.ConfigPath())
	if err != nil {
		return err
	}

	account := showAccount
	if account == "" {
		account, err = cfg.GetDefaultAccount()
		if err != nil {
			return err
		}
	}

	acc, ok := cfg.Accounts[account]
	if !ok {
		return errors.Wrap(errors.ErrAccountNotFound, account)
	}

	key, err := config.ResolveKey(acc.Key)
	if err != nil {
		return err
	}

	vendor, err := vendors.Resolve(acc.Vendor, cfg.Vendors)
	if err != nil {
		return err
	}

	fmt.Printf("Account: %s\n", account)
	fmt.Printf("Vendor: %s\n", acc.Vendor)
	fmt.Printf("Key: %s...\n", truncate(key, 8))
	fmt.Println()
	fmt.Println("Protocols:")
	for proto, url := range vendor.Protocols {
		fmt.Printf("  %s: %s\n", proto, url)
	}

	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
