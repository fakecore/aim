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
func Migrate(v1 *V1Config) *config.Config {
	v2 := &config.Config{
		Version:  "2",
		Accounts: make(map[string]config.Account),
		Vendors:  make(map[string]vendors.VendorConfig),
	}

	// Convert keys to accounts
	for name, key := range v1.Keys {
		v2.Accounts[name] = config.Account{
			Key:    key.Value,
			Vendor: key.Provider,
		}
		if key.IsDefault {
			v2.Options.DefaultAccount = name
		}
	}

	// Convert providers to vendors
	// If provider name matches a builtin, use builtin reference
	// Otherwise create custom vendor with openai protocol
	for name, provider := range v1.Providers {
		if _, isBuiltin := vendors.BuiltinVendors[name]; isBuiltin {
			// Use builtin vendor, but allow base override if different URL
			fullURL := provider.BaseURL + provider.APIPath
			builtinURL := vendors.BuiltinVendors[name].Protocols["openai"]
			if fullURL != builtinURL {
				// Custom URL, create override with base
				v2.Vendors[name] = vendors.VendorConfig{
					Base: name,
					Protocols: map[string]string{
						"openai": fullURL,
					},
				}
			}
			// If URL matches builtin, don't add to vendors (use builtin directly)
		} else {
			// Custom provider
			v2.Vendors[name] = vendors.VendorConfig{
				Protocols: map[string]string{
					"openai": provider.BaseURL + provider.APIPath,
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
