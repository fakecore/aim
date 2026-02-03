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
			Endpoints: map[string]EndpointConfig{
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
	}

	v, err := Resolve("deepseek", vendors)
	require.NoError(t, err)
	assert.Equal(t, "https://api.deepseek.com/v1", v.Endpoints["openai"].URL)
	assert.Equal(t, "deepseek-chat", v.Endpoints["openai"].DefaultModel)
}

func TestResolve_CustomVendor(t *testing.T) {
	customVendors := map[string]VendorConfig{
		"custom": {
			Endpoints: map[string]EndpointConfig{
				"openai": {
					URL:          "https://custom.com/v1",
					DefaultModel: "custom-model",
				},
			},
		},
	}

	v, err := Resolve("custom", customVendors)
	require.NoError(t, err)
	assert.Equal(t, "https://custom.com/v1", v.Endpoints["openai"].URL)
	assert.Equal(t, "custom-model", v.Endpoints["openai"].DefaultModel)
}

func TestResolve_NotFound(t *testing.T) {
	// Vendor not defined in config - should error
	vendors := map[string]VendorConfig{
		"other": {
			Endpoints: map[string]EndpointConfig{
				"openai": {
					URL: "https://other.com/v1",
				},
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

func TestGetEndpoint(t *testing.T) {
	v := &Vendor{
		Endpoints: map[string]Endpoint{
			"openai": {
				URL:          "https://api.example.com",
				DefaultModel: "gpt-4",
			},
		},
	}

	ep, err := v.GetEndpoint("openai")
	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com", ep.URL)
	assert.Equal(t, "gpt-4", ep.DefaultModel)

	_, err = v.GetEndpoint("anthropic")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

func TestHasEndpoint(t *testing.T) {
	v := &Vendor{
		Endpoints: map[string]Endpoint{
			"openai":    {URL: "https://api.example.com/openai"},
			"anthropic": {URL: "https://api.example.com/anthropic"},
		},
	}

	assert.True(t, v.HasEndpoint("openai"))
	assert.True(t, v.HasEndpoint("anthropic"))
	assert.False(t, v.HasEndpoint("unknown"))
}

func TestListEndpoints(t *testing.T) {
	v := &Vendor{
		Endpoints: map[string]Endpoint{
			"openai":    {URL: "https://api.example.com/openai"},
			"anthropic": {URL: "https://api.example.com/anthropic"},
		},
	}

	endpoints := v.ListEndpoints()
	assert.Len(t, endpoints, 2)
	assert.Contains(t, endpoints, "openai")
	assert.Contains(t, endpoints, "anthropic")
}
