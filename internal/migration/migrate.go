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
			v2.Settings.DefaultAccount = name
		}
	}

	// Convert providers to vendors
	// All vendors must be explicitly defined in v2
	for name, provider := range v1.Providers {
		if builtin, isBuiltin := vendors.BuiltinVendors[name]; isBuiltin {
			// Use builtin vendor definition as base, apply v1 overrides
			fullURL := provider.BaseURL + provider.APIPath
			builtinURL := builtin.Protocols["openai"]

			v2.Vendors[name] = vendors.VendorConfig{
				Protocols:     make(map[string]string),
				DefaultModels: make(map[string]string),
			}

			// Copy all protocols from builtin
			for proto, url := range builtin.Protocols {
				v2.Vendors[name].Protocols[proto] = url
			}
			// Copy all default models from builtin
			for proto, model := range builtin.DefaultModels {
				v2.Vendors[name].DefaultModels[proto] = model
			}

			// Apply v1 URL override if different
			if fullURL != builtinURL {
				v2.Vendors[name].Protocols["openai"] = fullURL
			}
		} else {
			// Custom provider - create explicit vendor definition
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
