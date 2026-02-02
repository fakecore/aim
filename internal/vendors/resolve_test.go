package vendors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve_Builtin(t *testing.T) {
	v, err := Resolve("deepseek", nil)
	require.NoError(t, err)
	assert.Equal(t, "https://api.deepseek.com/v1", v.Protocols["openai"])
	assert.Equal(t, "https://api.deepseek.com/anthropic", v.Protocols["anthropic"])
}

func TestResolve_Custom(t *testing.T) {
	customVendors := map[string]VendorConfig{
		"custom": {
			Protocols: map[string]string{
				"openai": "https://custom.com/v1",
			},
		},
	}

	v, err := Resolve("custom", customVendors)
	require.NoError(t, err)
	assert.Equal(t, "https://custom.com/v1", v.Protocols["openai"])
}

func TestResolve_WithBase(t *testing.T) {
	customVendors := map[string]VendorConfig{
		"glm-beta": {
			Base: "glm",
			Protocols: map[string]string{
				"anthropic": "https://beta.bigmodel.cn/anthropic",
			},
		},
	}

	v, err := Resolve("glm-beta", customVendors)
	require.NoError(t, err)
	// Inherited from glm
	assert.Equal(t, "https://open.bigmodel.cn/api/paas/v4", v.Protocols["openai"])
	// Overridden
	assert.Equal(t, "https://beta.bigmodel.cn/anthropic", v.Protocols["anthropic"])
}

func TestResolve_NotFound(t *testing.T) {
	_, err := Resolve("nonexistent", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AIM-VEN-001")
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
	assert.Contains(t, err.Error(), "AIM-VEN-002")
}
