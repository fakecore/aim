package config

import (
	"fmt"

	"github.com/fakecore/aim/internal/vendors"
	"gopkg.in/yaml.v3"
)

// Config represents the full AIM configuration
type Config struct {
	Version  string                       `yaml:"version"`
	Tools    map[string]ToolConfig        `yaml:"tools,omitempty"`
	Vendors  map[string]vendors.VendorConfig `yaml:"vendors,omitempty"`
	Keys     map[string]Key               `yaml:"keys,omitempty"`
	Accounts map[string]Account           `yaml:"accounts,omitempty"`
	Settings Settings                     `yaml:"settings,omitempty"`
}

// ToolConfig represents a tool's protocol configuration
type ToolConfig struct {
	Protocol string `yaml:"protocol"`
}

// Key represents a key configuration
type Key struct {
	Value     string            `yaml:"value"`
	Vendor    string            `yaml:"vendor"`
	Endpoints map[string]string `yaml:"endpoints,omitempty"` // Optional: protocol-specific endpoint selection (protocol -> endpoint)
}

// Account represents an account configuration
// Account references a Key and can override endpoint/model
type Account struct {
	Key      string            `yaml:"key"`                  // Reference to a key name
	Endpoints map[string]string `yaml:"endpoints,omitempty"` // Optional: protocol-specific endpoint overrides (protocol -> endpoint)
	Model    string            `yaml:"model,omitempty"`      // Optional: override model
}

// UnmarshalYAML implements custom YAML unmarshaling for Account
// to support both shorthand (string key reference) and full object formats
func (a *Account) UnmarshalYAML(node *yaml.Node) error {
	// Try string format first (shorthand: "account: key-name")
	if node.Kind == yaml.ScalarNode {
		a.Key = node.Value
		return nil
	}

	// Try object format (full: "account: {key: ..., endpoint: ...}")
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

// Settings represents global settings
type Settings struct {
	DefaultAccount string `yaml:"default_account,omitempty"`
	CommandTimeout string `yaml:"command_timeout,omitempty"`
	Language       string `yaml:"language,omitempty"`
	LogLevel       string `yaml:"log_level,omitempty"`
}

// ResolvedAccount represents a fully resolved account with all dependencies
type ResolvedAccount struct {
	Name        string
	Key         string
	Vendor      string
	Endpoint    string
	EndpointURL string
	Model       string
}
