package vendors

import (
	"fmt"

	"github.com/fakecore/aim/internal/errors"
)

// VendorConfig represents a vendor configuration from YAML
type VendorConfig struct {
	Protocols     map[string]string `yaml:"protocols,omitempty"`
	DefaultModels map[string]string `yaml:"default_models,omitempty"`
}

// Resolve resolves a vendor configuration from the config file
// All vendors must be explicitly defined in the config file
func Resolve(name string, vendors map[string]VendorConfig) (*Vendor, error) {
	v, ok := vendors[name]
	if !ok {
		return nil, &errors.Error{
			Code:       "AIM-VEN-003",
			Category:   "VEN",
			Message:    fmt.Sprintf("Vendor '%s' not defined in configuration", name),
			Suggestions: []string{
				fmt.Sprintf("Add vendor '%s' to your config file's vendors section", name),
				"Run 'aim init' to regenerate config with all built-in vendors",
			},
		}
	}

	return &Vendor{
		Protocols:     v.Protocols,
		DefaultModels: v.DefaultModels,
	}, nil
}

// GetProtocolURL gets the URL for a specific protocol
func (v *Vendor) GetProtocolURL(protocol string) (string, error) {
	url, ok := v.Protocols[protocol]
	if !ok {
		return "", &errors.Error{
			Code:     "AIM-VEN-002",
			Category: "VEN",
			Message:  fmt.Sprintf("Protocol '%s' not supported", protocol),
		}
	}
	return url, nil
}

// GetDefaultModel gets the default model for a specific protocol
func (v *Vendor) GetDefaultModel(protocol string) string {
	if v.DefaultModels == nil {
		return ""
	}
	return v.DefaultModels[protocol]
}
