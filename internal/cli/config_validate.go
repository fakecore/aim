package cli

import (
	"fmt"

	"github.com/fakecore/aim/internal/config"
	"github.com/fakecore/aim/internal/vendors"
	"github.com/spf13/cobra"
)

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file",
	Long:  `Check configuration for errors and report all issues found.`,
	RunE:  configValidate,
}

func init() {
	configCmd.AddCommand(configValidateCmd)
}

func configValidate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(config.ConfigPath())
	if err != nil {
		return err
	}

	var issues []string
	var firstErr error

	// Validate each account
	for name, acc := range cfg.Accounts {
		// Check key can be resolved
		_, err := config.ResolveKey(acc.Key)
		if err != nil {
			issues = append(issues, fmt.Sprintf("account '%s': %v", name, err))
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		// Check vendor exists
		_, err = vendors.Resolve(acc.Vendor, cfg.Vendors)
		if err != nil {
			issues = append(issues, fmt.Sprintf("account '%s': %v", name, err))
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
	}

	if len(issues) > 0 {
		fmt.Println("Configuration issues found:")
		for _, issue := range issues {
			fmt.Printf("  - %s\n", issue)
		}
		return firstErr
	}

	fmt.Println("Configuration is valid")
	return nil
}
