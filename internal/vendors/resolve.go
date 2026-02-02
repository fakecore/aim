package vendors

import (
	"fmt"

	"github.com/fakecore/aim/internal/errors"
)

// VendorConfig represents a vendor configuration from YAML
type VendorConfig struct {
	Builtin   string            `yaml:"builtin,omitempty"`
	Base      string            `yaml:"base,omitempty"`
	Protocols map[string]string `yaml:"protocols,omitempty"`
}

// Resolve resolves a vendor configuration
func Resolve(name string, vendors map[string]VendorConfig) (*Vendor, error) {
	// Check user-defined vendors
	if v, ok := vendors[name]; ok {
		return resolveWithBase(v, vendors)
	}

	// Check builtin vendors
	if v, ok := BuiltinVendors[name]; ok {
		return &v, nil
	}

	return nil, errors.Wrap(errors.ErrVendorNotFound, name)
}

// resolveWithBase resolves a vendor with base inheritance
func resolveWithBase(v VendorConfig, allVendors map[string]VendorConfig) (*Vendor, error) {
	result := &Vendor{
		Protocols: make(map[string]string),
	}

	// If has base, merge from base first
	if v.Base != "" {
		base, err := Resolve(v.Base, allVendors)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve base vendor '%s': %w", v.Base, err)
		}
		for proto, url := range base.Protocols {
			result.Protocols[proto] = url
		}
	}

	// Apply overrides
	for proto, url := range v.Protocols {
		result.Protocols[proto] = url
	}

	return result, nil
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
