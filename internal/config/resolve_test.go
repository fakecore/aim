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
				Endpoints: map[string]vendors.EndpointConfig{
					"openai": {
						URL:          "https://api.deepseek.com/v1",
						DefaultModel: "deepseek-chat",
					},
					"anthropic": {
						URL:          "https://api.deepseek.com/anthropic",
						DefaultModel: "deepseek-chat",
					},
				},
			},
		},
		Keys: map[string]Key{
			"deepseek-key": {
				Value:  "sk-test",
				Vendor: "deepseek",
			},
		},
		Accounts: map[string]Account{
			"deepseek": {Key: "deepseek-key"},
		},
	}

	resolved, err := cfg.ResolveAccount("deepseek", "claude-code", "anthropic")
	require.NoError(t, err)
	assert.Equal(t, "deepseek", resolved.Name)
	assert.Equal(t, "sk-test", resolved.Key)
	assert.Equal(t, "anthropic", resolved.Endpoint)
	assert.Equal(t, "https://api.deepseek.com/anthropic", resolved.EndpointURL)
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

func TestResolveAccount_KeyNotDefined(t *testing.T) {
	cfg := &Config{
		Version: "2",
		Accounts: map[string]Account{
			"test": {Key: "undefined-key"},
		},
	}

	_, err := cfg.ResolveAccount("test", "claude-code", "anthropic")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestResolveAccount_VendorNotDefined(t *testing.T) {
	cfg := &Config{
		Version: "2",
		Keys: map[string]Key{
			"test-key": {
				Value:  "sk-test",
				Vendor: "undefined-vendor",
			},
		},
		Accounts: map[string]Account{
			"test": {Key: "test-key"},
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
				Endpoints: map[string]vendors.EndpointConfig{
					"anthropic": {
						URL: "https://api.deepseek.com/anthropic",
					},
				},
			},
		},
		Keys: map[string]Key{
			"test-key": {
				Value:  "${TEST_KEY}",
				Vendor: "deepseek",
			},
		},
		Accounts: map[string]Account{
			"test": {Key: "test-key"},
		},
	}

	resolved, err := cfg.ResolveAccount("test", "claude-code", "anthropic")
	require.NoError(t, err)
	assert.Equal(t, "sk-from-env", resolved.Key)
}

func TestResolveAccount_AccountEndpointOverride(t *testing.T) {
	cfg := &Config{
		Version: "2",
		Vendors: map[string]vendors.VendorConfig{
			"deepseek": {
				Endpoints: map[string]vendors.EndpointConfig{
					"openai": {
						URL:          "https://api.deepseek.com/v1",
						DefaultModel: "deepseek-chat",
					},
					"anthropic": {
						URL:          "https://api.deepseek.com/anthropic",
						DefaultModel: "deepseek-chat",
					},
				},
			},
		},
		Keys: map[string]Key{
			"deepseek-key": {
				Value:  "sk-test",
				Vendor: "deepseek",
			},
		},
		Accounts: map[string]Account{
			"deepseek": {
				Key: "deepseek-key",
				Endpoints: map[string]string{
					"anthropic": "openai", // Override anthropic to use openai endpoint
				},
			},
		},
	}

	resolved, err := cfg.ResolveAccount("deepseek", "claude-code", "anthropic")
	require.NoError(t, err)
	assert.Equal(t, "openai", resolved.Endpoint)
	assert.Equal(t, "https://api.deepseek.com/v1", resolved.EndpointURL)
}

func TestResolveAccount_KeyEndpointOverride(t *testing.T) {
	cfg := &Config{
		Version: "2",
		Vendors: map[string]vendors.VendorConfig{
			"deepseek": {
				Endpoints: map[string]vendors.EndpointConfig{
					"openai": {
						URL: "https://api.deepseek.com/v1",
					},
					"anthropic": {
						URL: "https://api.deepseek.com/anthropic",
					},
					"openai-coding": {
						URL: "https://api.deepseek.com/coding/v1",
					},
				},
			},
		},
		Keys: map[string]Key{
			"deepseek-key": {
				Value:  "sk-test",
				Vendor: "deepseek",
				Endpoints: map[string]string{
					"openai": "openai-coding", // Key overrides openai to use coding endpoint
				},
			},
		},
		Accounts: map[string]Account{
			"deepseek": {Key: "deepseek-key"},
		},
	}

	// Key endpoint override should apply
	resolved, err := cfg.ResolveAccount("deepseek", "codex", "openai")
	require.NoError(t, err)
	assert.Equal(t, "openai-coding", resolved.Endpoint)
	assert.Equal(t, "https://api.deepseek.com/coding/v1", resolved.EndpointURL)

	// anthropic protocol should use default (no override in key)
	resolved, err = cfg.ResolveAccount("deepseek", "cc", "anthropic")
	require.NoError(t, err)
	assert.Equal(t, "anthropic", resolved.Endpoint)
}

func TestResolveAccount_AccountOverridesKey(t *testing.T) {
	cfg := &Config{
		Version: "2",
		Vendors: map[string]vendors.VendorConfig{
			"deepseek": {
				Endpoints: map[string]vendors.EndpointConfig{
					"openai": {
						URL: "https://api.deepseek.com/v1",
					},
					"anthropic": {
						URL: "https://api.deepseek.com/anthropic",
					},
					"openai-coding": {
						URL: "https://api.deepseek.com/coding/v1",
					},
				},
			},
		},
		Keys: map[string]Key{
			"deepseek-key": {
				Value:  "sk-test",
				Vendor: "deepseek",
				Endpoints: map[string]string{
					"openai": "openai-coding",
				},
			},
		},
		Accounts: map[string]Account{
			"deepseek": {
				Key: "deepseek-key",
				Endpoints: map[string]string{
					"openai": "openai", // Account overrides key's preference
				},
			},
		},
	}

	// Account endpoint override should take priority over key's
	resolved, err := cfg.ResolveAccount("deepseek", "codex", "openai")
	require.NoError(t, err)
	assert.Equal(t, "openai", resolved.Endpoint)
	assert.Equal(t, "https://api.deepseek.com/v1", resolved.EndpointURL)
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

func TestResolveAccount_ProtocolSpecificEndpoint(t *testing.T) {
	cfg := &Config{
		Version: "2",
		Vendors: map[string]vendors.VendorConfig{
			"glm": {
				Endpoints: map[string]vendors.EndpointConfig{
					"openai": {
						URL:          "https://open.bigmodel.cn/api/paas/v4",
						DefaultModel: "glm-4.7",
					},
					"openai-coding": {
						URL:          "https://open.bigmodel.cn/api/coding/paas/v4",
						DefaultModel: "glm-4-coding",
					},
					"anthropic": {
						URL:          "https://open.bigmodel.cn/api/anthropic",
						DefaultModel: "glm-4.7",
					},
				},
			},
		},
		Keys: map[string]Key{
			"glm-key": {
				Value:  "sk-test",
				Vendor: "glm",
			},
		},
		Accounts: map[string]Account{
			"glm-coding": {
				Key: "glm-key",
				Endpoints: map[string]string{
					"openai":    "openai-coding",
					"anthropic": "anthropic",
				},
			},
		},
	}

	// Test with openai protocol - should use openai-coding
	resolved, err := cfg.ResolveAccount("glm-coding", "codex", "openai")
	require.NoError(t, err)
	assert.Equal(t, "openai-coding", resolved.Endpoint)
	assert.Equal(t, "https://open.bigmodel.cn/api/coding/paas/v4", resolved.EndpointURL)

	// Test with anthropic protocol - should use anthropic
	resolved, err = cfg.ResolveAccount("glm-coding", "cc", "anthropic")
	require.NoError(t, err)
	assert.Equal(t, "anthropic", resolved.Endpoint)
	assert.Equal(t, "https://open.bigmodel.cn/api/anthropic", resolved.EndpointURL)
}
