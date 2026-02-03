package migration

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/fakecore/aim/internal/config"
	"github.com/fakecore/aim/internal/vendors"
	"gopkg.in/yaml.v3"
)

// LoadV1 loads a v1 config file
func LoadV1(path string) (*V1Config, error) {
	var cfg V1Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Migrate converts v1 config to v2
// Note: v1 used "providers" and "keys", v2 uses "vendors", "keys", and "accounts"
func Migrate(v1 *V1Config) *config.Config {
	v2 := &config.Config{
		Version:  "2",
		Keys:     make(map[string]config.Key),
		Accounts: make(map[string]config.Account),
		Vendors:  make(map[string]vendors.VendorConfig),
	}

	// Convert v1 keys to v2 keys
	// In v1: keys had value and provider
	// In v2: keys have value, vendor, and optional endpoints restriction
	for name, key := range v1.Keys {
		v2.Keys[name] = config.Key{
			Value:  key.Value,
			Vendor: key.Provider,
			// No endpoints restriction by default
		}

		// Create an account with the same name that references the key
		v2.Accounts[name] = config.Account{
			Key: name,
		}

		if key.IsDefault {
			v2.Settings.DefaultAccount = name
		}
	}

	// Convert v1 providers to v2 vendors
	// In v1: providers had base_url and api_path
	// In v2: vendors have endpoints with url and default_model
	for name, provider := range v1.Providers {
		if builtin, isBuiltin := vendors.BuiltinVendors[name]; isBuiltin {
			// Use builtin vendor definition - copy all endpoints
			endpoints := make(map[string]vendors.EndpointConfig)
			for epName, ep := range builtin.Endpoints {
				endpoints[epName] = vendors.EndpointConfig{
					URL:          ep.URL,
					DefaultModel: ep.DefaultModel,
				}
			}

			// Apply v1 URL override if provided
			if provider.BaseURL != "" || provider.APIPath != "" {
				fullURL := provider.BaseURL + provider.APIPath
				// Override the openai endpoint
				if _, ok := endpoints["openai"]; ok {
					endpoints["openai"] = vendors.EndpointConfig{
						URL:          fullURL,
						DefaultModel: builtin.Endpoints["openai"].DefaultModel,
					}
				}
			}

			v2.Vendors[name] = vendors.VendorConfig{
				Endpoints: endpoints,
			}
		} else {
			// Custom provider - create explicit vendor definition
			fullURL := provider.BaseURL + provider.APIPath
			v2.Vendors[name] = vendors.VendorConfig{
				Endpoints: map[string]vendors.EndpointConfig{
					"openai": {
						URL: fullURL,
					},
				},
			}
		}
	}

	return v2
}

// WriteV2 writes v2 config to file
func WriteV2(cfg *config.Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
