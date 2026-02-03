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
		if len(v.Endpoints) == 0 {
			issues = append(issues, ValidationIssue{
				Level:   "error",
				Field:   fmt.Sprintf("vendors.%s", name),
				Message: "Vendor has no endpoints defined",
			})
		}
		for epName, ep := range v.Endpoints {
			if ep.URL == "" {
				issues = append(issues, ValidationIssue{
					Level:   "error",
					Field:   fmt.Sprintf("vendors.%s.endpoints.%s", name, epName),
					Message: "Endpoint URL cannot be empty",
				})
			}
		}
	}

	// Validate keys section exists
	if len(cfg.Keys) == 0 {
		issues = append(issues, ValidationIssue{
			Level:   "error",
			Field:   "keys",
			Message: "No keys defined",
		})
	}

	// Validate each key
	for name, key := range cfg.Keys {
		prefix := fmt.Sprintf("keys.%s", name)

		// Check vendor is specified
		if key.Vendor == "" {
			issues = append(issues, ValidationIssue{
				Level:   "error",
				Field:   fmt.Sprintf("%s.vendor", prefix),
				Message: "Vendor must be explicitly specified (e.g., vendor: deepseek)",
			})
		} else {
			// Check vendor exists in config
			_, err := vendors.Resolve(key.Vendor, cfg.Vendors)
			if err != nil {
				issues = append(issues, ValidationIssue{
					Level:   "error",
					Field:   fmt.Sprintf("%s.vendor", prefix),
					Message: fmt.Sprintf("Vendor '%s' not defined in configuration", key.Vendor),
				})
			}
		}

		// Check key value can be resolved
		if key.Value == "" {
			issues = append(issues, ValidationIssue{
				Level:   "error",
				Field:   fmt.Sprintf("%s.value", prefix),
				Message: "Key value is required",
			})
		} else {
			_, err := config.ResolveKey(key.Value)
			if err != nil {
				issues = append(issues, ValidationIssue{
					Level:   "error",
					Field:   fmt.Sprintf("%s.value", prefix),
					Message: err.Error(),
				})
			}
		}

		// Validate endpoints (if specified)
		for protocol, endpointName := range key.Endpoints {
			vendor, ok := cfg.Vendors[key.Vendor]
			if !ok {
				continue // Vendor error already reported above
			}
			if _, ok := vendor.Endpoints[endpointName]; !ok {
				issues = append(issues, ValidationIssue{
					Level:   "warning",
					Field:   fmt.Sprintf("%s.endpoints.%s", prefix, protocol),
					Message: fmt.Sprintf("Endpoint '%s' not found in vendor '%s'", endpointName, key.Vendor),
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

		// Check key is specified
		if acc.Key == "" {
			issues = append(issues, ValidationIssue{
				Level:   "error",
				Field:   fmt.Sprintf("%s.key", prefix),
				Message: "Account must reference a key (e.g., key: my-key)",
			})
		} else {
			// Check key exists in config
			if _, ok := cfg.Keys[acc.Key]; !ok {
				issues = append(issues, ValidationIssue{
					Level:   "error",
					Field:   fmt.Sprintf("%s.key", prefix),
					Message: fmt.Sprintf("Key '%s' not defined in keys section", acc.Key),
				})
			}
		}

		// Validate endpoint overrides (if specified)
		if len(acc.Endpoints) > 0 && acc.Key != "" {
			if key, ok := cfg.Keys[acc.Key]; ok {
				vendor, ok := cfg.Vendors[key.Vendor]
				if !ok {
					// Vendor error already reported
					continue
				}
				for protocol, endpointName := range acc.Endpoints {
					if _, ok := vendor.Endpoints[endpointName]; !ok {
						issues = append(issues, ValidationIssue{
							Level:   "error",
							Field:   fmt.Sprintf("%s.endpoints.%s", prefix, protocol),
							Message: fmt.Sprintf("Endpoint '%s' not found in vendor '%s'", endpointName, key.Vendor),
						})
					}
				}
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
			if _, ok := vendor.Endpoints[tool.Protocol]; !ok {
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
		fmt.Printf("  Keys: %d\n", len(cfg.Keys))
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
