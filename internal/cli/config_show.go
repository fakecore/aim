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
	Long:  `Display resolved configuration including account, key, and vendor information.`,
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

	// Get the key referenced by the account
	keyRef, ok := cfg.Keys[acc.Key]
	if !ok {
		return fmt.Errorf("key '%s' referenced by account not found", acc.Key)
	}

	keyValue, err := config.ResolveKey(keyRef.Value)
	if err != nil {
		return err
	}

	vendor, err := vendors.Resolve(keyRef.Vendor, cfg.Vendors)
	if err != nil {
		return err
	}

	fmt.Printf("Account: %s\n", account)
	fmt.Printf("Key: %s\n", acc.Key)
	fmt.Printf("Vendor: %s\n", keyRef.Vendor)
	fmt.Printf("API Key: %s...\n", truncate(keyValue, 8))

	// Show endpoint overrides
	if len(acc.Endpoints) > 0 {
		fmt.Println("Endpoint Overrides (protocol -> endpoint):")
		for protocol, endpoint := range acc.Endpoints {
			fmt.Printf("  %s -> %s\n", protocol, endpoint)
		}
	}
	if len(keyRef.Endpoints) > 0 {
		fmt.Println("Key Endpoint Defaults (protocol -> endpoint):")
		for protocol, endpoint := range keyRef.Endpoints {
			fmt.Printf("  %s -> %s\n", protocol, endpoint)
		}
	}
	if acc.Model != "" {
		fmt.Printf("Model (override): %s\n", acc.Model)
	}
	fmt.Println()
	fmt.Println("Available Endpoints:")
	for epName, ep := range vendor.Endpoints {
		defaultModel := ""
		if ep.DefaultModel != "" {
			defaultModel = fmt.Sprintf(", default model: %s", ep.DefaultModel)
		}
		fmt.Printf("  %s: %s%s\n", epName, ep.URL, defaultModel)
	}

	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
