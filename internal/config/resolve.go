package config

import (
	"github.com/fakecore/aim/internal/errors"
	"github.com/fakecore/aim/internal/vendors"
)

// ResolveAccount resolves an account with all dependencies
func (c *Config) ResolveAccount(name string, tool string, toolProtocol string) (*ResolvedAccount, error) {
	// Find account
	acc, ok := c.Accounts[name]
	if !ok {
		// Build available list
		available := make([]string, 0, len(c.Accounts))
		for n := range c.Accounts {
			available = append(available, n)
		}
		_ = available // TODO: include in error details
		return nil, errors.Wrap(errors.ErrAccountNotFound, name)
	}

	// Resolve key
	key, err := ResolveKey(acc.Key)
	if err != nil {
		return nil, err
	}

	// Resolve vendor
	vendor, err := vendors.Resolve(acc.Vendor, c.Vendors)
	if err != nil {
		return nil, err
	}

	// Get protocol URL
	protocolURL, err := vendor.GetProtocolURL(toolProtocol)
	if err != nil {
		return nil, errors.Wrap(errors.ErrProtocolNotSupported, acc.Vendor, toolProtocol)
	}

	return &ResolvedAccount{
		Name:        name,
		Key:         key,
		Vendor:      acc.Vendor,
		Protocol:    toolProtocol,
		ProtocolURL: protocolURL,
	}, nil
}

// GetDefaultAccount returns the default account name
func (c *Config) GetDefaultAccount() (string, error) {
	if c.Options.DefaultAccount != "" {
		return c.Options.DefaultAccount, nil
	}

	// If only one account, use it
	if len(c.Accounts) == 1 {
		for name := range c.Accounts {
			return name, nil
		}
	}

	return "", errors.Wrap(errors.ErrKeyNotSet, "default")
}
