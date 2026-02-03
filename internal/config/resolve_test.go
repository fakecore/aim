package config

import (
	"os"
	"testing"

	"github.com/fakecore/aim/internal/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveAccount(t *testing.T) {
	cfg := &Config{
		Version: "2",
		Vendors: map[string]vendors.VendorConfig{
			"deepseek": {
				Protocols: map[string]string{
					"openai":    "https://api.deepseek.com/v1",
					"anthropic": "https://api.deepseek.com/anthropic",
				},
				DefaultModels: map[string]string{
					"anthropic": "deepseek-chat",
				},
			},
		},
		Accounts: map[string]Account{
			"deepseek": {Key: "sk-test", Vendor: "deepseek"},
		},
	}

	resolved, err := cfg.ResolveAccount("deepseek", "claude-code", "anthropic")
	require.NoError(t, err)
	assert.Equal(t, "deepseek", resolved.Name)
	assert.Equal(t, "sk-test", resolved.Key)
	assert.Equal(t, "anthropic", resolved.Protocol)
	assert.Equal(t, "https://api.deepseek.com/anthropic", resolved.ProtocolURL)
	assert.Equal(t, "deepseek-chat", resolved.Model)
}

func TestResolveAccount_NotFound(t *testing.T) {
	cfg := &Config{
		Version:  "2",
		Accounts: map[string]Account{},
	}

	_, err := cfg.ResolveAccount("nonexistent", "claude-code", "anthropic")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestResolveAccount_VendorNotDefined(t *testing.T) {
	// Vendor not defined in config
	cfg := &Config{
		Version: "2",
		Accounts: map[string]Account{
			"test": {Key: "sk-test", Vendor: "undefined-vendor"},
		},
	}

	_, err := cfg.ResolveAccount("test", "claude-code", "anthropic")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not defined")
}

func TestResolveAccount_WithEnv(t *testing.T) {
	os.Setenv("TEST_KEY", "sk-from-env")
	defer os.Unsetenv("TEST_KEY")

	cfg := &Config{
		Version: "2",
		Vendors: map[string]vendors.VendorConfig{
			"deepseek": {
				Protocols: map[string]string{
					"anthropic": "https://api.deepseek.com/anthropic",
				},
			},
		},
		Accounts: map[string]Account{
			"test": {Key: "${TEST_KEY}", Vendor: "deepseek"},
		},
	}

	resolved, err := cfg.ResolveAccount("test", "claude-code", "anthropic")
	require.NoError(t, err)
	assert.Equal(t, "sk-from-env", resolved.Key)
}

func TestGetDefaultAccount(t *testing.T) {
	cfg := &Config{
		Settings: Settings{DefaultAccount: "deepseek"},
		Accounts: map[string]Account{
			"deepseek": {},
			"glm":      {},
		},
	}

	name, err := cfg.GetDefaultAccount()
	require.NoError(t, err)
	assert.Equal(t, "deepseek", name)
}

func TestGetDefaultAccount_SingleAccount(t *testing.T) {
	cfg := &Config{
		Accounts: map[string]Account{
			"only": {},
		},
	}

	name, err := cfg.GetDefaultAccount()
	require.NoError(t, err)
	assert.Equal(t, "only", name)
}

func TestGetDefaultAccount_NoDefault(t *testing.T) {
	cfg := &Config{
		Accounts: map[string]Account{
			"acc1": {},
			"acc2": {},
		},
	}

	_, err := cfg.GetDefaultAccount()
	assert.Error(t, err)
}
