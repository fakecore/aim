package vendors

import (
	"fmt"

	"github.com/fakecore/aim/internal/errors"
)

// VendorConfig represents a vendor configuration from YAML
type VendorConfig struct {
	Endpoints map[string]EndpointConfig `yaml:"endpoints,omitempty"`
}

// EndpointConfig represents an endpoint configuration from YAML
type EndpointConfig struct {
	URL          string `yaml:"url"`
	DefaultModel string `yaml:"default_model,omitempty"`
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

	// Convert EndpointConfig to Endpoint
	endpoints := make(map[string]Endpoint)
	for name, ep := range v.Endpoints {
		endpoints[name] = Endpoint{
			URL:          ep.URL,
			DefaultModel: ep.DefaultModel,
		}
	}

	return &Vendor{
		Endpoints: endpoints,
	}, nil
}

// GetEndpoint gets the endpoint configuration for a specific protocol/endpoint name
func (v *Vendor) GetEndpoint(endpointName string) (*Endpoint, error) {
	ep, ok := v.Endpoints[endpointName]
	if !ok {
		return nil, &errors.Error{
			Code:     "AIM-VEN-002",
			Category: "VEN",
			Message:  fmt.Sprintf("Endpoint '%s' not supported", endpointName),
			Suggestions: []string{
				"Check available endpoints in your vendor configuration",
				"Common endpoints: openai, anthropic",
			},
		}
	}
	return &ep, nil
}

// HasEndpoint checks if a vendor supports a specific endpoint
func (v *Vendor) HasEndpoint(endpointName string) bool {
	_, ok := v.Endpoints[endpointName]
	return ok
}

// ListEndpoints returns all available endpoint names for this vendor
func (v *Vendor) ListEndpoints() []string {
	names := make([]string, 0, len(v.Endpoints))
	for name := range v.Endpoints {
		names = append(names, name)
	}
	return names
}
