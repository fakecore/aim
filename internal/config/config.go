package config

import (
	"fmt"

	"github.com/fakecore/aim/internal/vendors"
	"gopkg.in/yaml.v3"
)

// Config represents the full AIM configuration
type Config struct {
	Version  string                      `yaml:"version"`
	Vendors  map[string]vendors.VendorConfig `yaml:"vendors,omitempty"`
	Accounts map[string]Account          `yaml:"accounts"`
	Options  Options                     `yaml:"options,omitempty"`
}

// Account represents an account configuration
type Account struct {
	Key    string `yaml:"key,omitempty"`
	Vendor string `yaml:"vendor,omitempty"`
}

// UnmarshalYAML implements custom YAML unmarshaling for Account
// to support both shorthand (string key) and full object formats
func (a *Account) UnmarshalYAML(node *yaml.Node) error {
	// Try string format first (shorthand: "account: key-value")
	if node.Kind == yaml.ScalarNode {
		a.Key = node.Value
		return nil
	}

	// Try object format (full: "account: {key: ..., vendor: ...}")
	if node.Kind == yaml.MappingNode {
		type rawAccount Account
		var raw rawAccount
		if err := node.Decode(&raw); err != nil {
			return err
		}
		*a = Account(raw)
		return nil
	}

	return fmt.Errorf("account must be a string or object, got %v", node.Kind)
}

// Options represents global options
type Options struct {
	DefaultAccount string `yaml:"default_account,omitempty"`
	CommandTimeout string `yaml:"command_timeout,omitempty"`
}

// ResolvedAccount represents a fully resolved account
type ResolvedAccount struct {
	Name        string
	Key         string
	Vendor      string
	Protocol    string
	ProtocolURL string
	Model       string // Default model for this vendor/protocol
}
