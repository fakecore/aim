package vendors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve_FromConfig(t *testing.T) {
	// All vendors must be defined in config file
	vendors := map[string]VendorConfig{
		"deepseek": {
			Protocols: map[string]string{
				"openai":    "https://api.deepseek.com/v1",
				"anthropic": "https://api.deepseek.com/anthropic",
			},
			DefaultModels: map[string]string{
				"openai":    "deepseek-chat",
				"anthropic": "deepseek-chat",
			},
		},
	}

	v, err := Resolve("deepseek", vendors)
	require.NoError(t, err)
	assert.Equal(t, "https://api.deepseek.com/v1", v.Protocols["openai"])
	assert.Equal(t, "https://api.deepseek.com/anthropic", v.Protocols["anthropic"])
	assert.Equal(t, "deepseek-chat", v.DefaultModels["openai"])
}

func TestResolve_CustomVendor(t *testing.T) {
	customVendors := map[string]VendorConfig{
		"custom": {
			Protocols: map[string]string{
				"openai": "https://custom.com/v1",
			},
			DefaultModels: map[string]string{
				"openai": "custom-model",
			},
		},
	}

	v, err := Resolve("custom", customVendors)
	require.NoError(t, err)
	assert.Equal(t, "https://custom.com/v1", v.Protocols["openai"])
	assert.Equal(t, "custom-model", v.DefaultModels["openai"])
}

func TestResolve_NotFound(t *testing.T) {
	// Vendor not defined in config - should error
	vendors := map[string]VendorConfig{
		"other": {
			Protocols: map[string]string{
				"openai": "https://other.com/v1",
			},
		},
	}

	_, err := Resolve("nonexistent", vendors)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not defined")
}

func TestResolve_EmptyVendors(t *testing.T) {
	// Empty vendors map
	_, err := Resolve("deepseek", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not defined")
}

func TestGetProtocolURL(t *testing.T) {
	v := &Vendor{
		Protocols: map[string]string{
			"openai": "https://api.example.com",
		},
	}

	url, err := v.GetProtocolURL("openai")
	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com", url)

	_, err = v.GetProtocolURL("anthropic")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

func TestGetDefaultModel(t *testing.T) {
	v := &Vendor{
		DefaultModels: map[string]string{
			"openai":    "gpt-4",
			"anthropic": "claude-3",
		},
	}

	assert.Equal(t, "gpt-4", v.GetDefaultModel("openai"))
	assert.Equal(t, "claude-3", v.GetDefaultModel("anthropic"))
	assert.Equal(t, "", v.GetDefaultModel("unknown"))
}

func TestGetDefaultModel_Nil(t *testing.T) {
	v := &Vendor{
		Protocols: map[string]string{
			"openai": "https://api.example.com",
		},
	}

	assert.Equal(t, "", v.GetDefaultModel("openai"))
}
