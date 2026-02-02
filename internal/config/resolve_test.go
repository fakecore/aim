package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveAccount(t *testing.T) {
	cfg := &Config{
		Version: "2",
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
}

func TestResolveAccount_NotFound(t *testing.T) {
	cfg := &Config{
		Version:  "2",
		Accounts: map[string]Account{},
	}

	_, err := cfg.ResolveAccount("nonexistent", "claude-code", "anthropic")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AIM-ACC-001")
}

func TestResolveAccount_WithEnv(t *testing.T) {
	os.Setenv("TEST_KEY", "sk-from-env")
	defer os.Unsetenv("TEST_KEY")

	cfg := &Config{
		Version: "2",
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
		Options: Options{DefaultAccount: "deepseek"},
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
