package cli

import (
	"fmt"
	"os"

	"github.com/fakecore/aim/internal/config"
	"github.com/fakecore/aim/internal/tools"
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

	var issues []ValidationIssue

	// Validate version
	if cfg.Version != "2" {
		issues = append(issues, ValidationIssue{
			Level:   "error",
			Field:   "version",
			Message: fmt.Sprintf("Unsupported version '%s', expected '2'", cfg.Version),
		})
	}

	// Validate vendors section exists
	if len(cfg.Vendors) == 0 {
		issues = append(issues, ValidationIssue{
			Level:   "error",
			Field:   "vendors",
			Message: "No vendors defined. Run 'aim init' to regenerate config with built-in vendors.",
		})
	}

	// Validate each vendor has required fields
	for name, v := range cfg.Vendors {
		if len(v.Protocols) == 0 {
			issues = append(issues, ValidationIssue{
				Level:   "error",
				Field:   fmt.Sprintf("vendors.%s", name),
				Message: "Vendor has no protocols defined",
			})
		}
		for proto, url := range v.Protocols {
			if url == "" {
				issues = append(issues, ValidationIssue{
					Level:   "error",
					Field:   fmt.Sprintf("vendors.%s.protocols.%s", name, proto),
					Message: "Protocol URL cannot be empty",
				})
			}
		}
	}

	// Validate accounts section exists
	if len(cfg.Accounts) == 0 {
		issues = append(issues, ValidationIssue{
			Level:   "warning",
			Field:   "accounts",
			Message: "No accounts defined",
		})
	}

	// Validate each account
	for name, acc := range cfg.Accounts {
		prefix := fmt.Sprintf("accounts.%s", name)

		// Check vendor is specified
		if acc.Vendor == "" {
			issues = append(issues, ValidationIssue{
				Level:   "error",
				Field:   fmt.Sprintf("%s.vendor", prefix),
				Message: "Vendor must be explicitly specified (e.g., vendor: deepseek)",
			})
		} else {
			// Check vendor exists in config
			_, err := vendors.Resolve(acc.Vendor, cfg.Vendors)
			if err != nil {
				issues = append(issues, ValidationIssue{
					Level:   "error",
					Field:   fmt.Sprintf("%s.vendor", prefix),
					Message: fmt.Sprintf("Vendor '%s' not defined in configuration", acc.Vendor),
				})
			}
		}

		// Check key can be resolved
		if acc.Key == "" {
			issues = append(issues, ValidationIssue{
				Level:   "error",
				Field:   fmt.Sprintf("%s.key", prefix),
				Message: "API key is required",
			})
		} else {
			_, err := config.ResolveKey(acc.Key)
			if err != nil {
				issues = append(issues, ValidationIssue{
					Level:   "error",
					Field:   fmt.Sprintf("%s.key", prefix),
					Message: err.Error(),
				})
			}
		}
	}

	// Validate default_account if specified
	if cfg.Settings.DefaultAccount != "" {
		if _, ok := cfg.Accounts[cfg.Settings.DefaultAccount]; !ok {
			issues = append(issues, ValidationIssue{
				Level:   "error",
				Field:   "settings.default_account",
				Message: fmt.Sprintf("Default account '%s' not found in accounts", cfg.Settings.DefaultAccount),
			})
		}
	}

	// Validate tools can work with configured vendors
	for toolName, tool := range tools.BuiltinTools {
		for vendorName, vendor := range cfg.Vendors {
			if _, ok := vendor.Protocols[tool.Protocol]; !ok {
				// This is just informational, not an error
				// Some vendors may not support all tools
				_ = toolName
				_ = vendorName
			}
		}
	}

	// Print results
	if len(issues) == 0 {
		fmt.Println("✓ Configuration is valid")
		fmt.Printf("  Vendors: %d\n", len(cfg.Vendors))
		fmt.Printf("  Accounts: %d\n", len(cfg.Accounts))
		return nil
	}

	// Count by level
	var errors, warnings int
	for _, issue := range issues {
		if issue.Level == "error" {
			errors++
		} else {
			warnings++
		}
	}

	fmt.Printf("Configuration issues found: %d errors, %d warnings\n\n", errors, warnings)

	for _, issue := range issues {
		symbol := "✗"
		if issue.Level == "warning" {
			symbol = "⚠"
		}
		// Human readable format: no error codes in user interface
		fmt.Printf("%s %s: %s\n", symbol, issue.Field, issue.Message)
	}

	if errors > 0 {
		fmt.Println("\nSuggestions:")
		fmt.Println("  → Run 'aim init' to regenerate a valid configuration")
		fmt.Println("  → Edit the config file: aim config edit")
		os.Exit(1)
	}

	return nil
}

// ValidationIssue represents a configuration validation issue
type ValidationIssue struct {
	Level   string
	Field   string
	Message string
}
