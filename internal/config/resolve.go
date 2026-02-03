package config

import (
	"fmt"

	"github.com/fakecore/aim/internal/errors"
	"github.com/fakecore/aim/internal/vendors"
)

// ResolveAccount resolves an account with all dependencies
// New flow: Account -> Key -> Vendor -> Endpoint
func (c *Config) ResolveAccount(accountName string, toolName string, toolProtocol string) (*ResolvedAccount, error) {
	// 1. Find account
	acc, ok := c.Accounts[accountName]
	if !ok {
		available := make([]string, 0, len(c.Accounts))
		for n := range c.Accounts {
			available = append(available, n)
		}
		err := &errors.Error{
			Code:     "AIM-ACC-001",
			Category: "ACC",
			Message:  fmt.Sprintf("Account '%s' not found", accountName),
			Suggestions: []string{
				"Check your config file for available accounts",
				"Add a new account to your config file",
			},
		}
		if len(available) > 0 {
			err.Details = map[string]interface{}{"available": available}
		}
		return nil, err
	}

	// 2. Resolve key reference
	keyName := acc.Key
	if keyName == "" {
		return nil, &errors.Error{
			Code:     "AIM-ACC-002",
			Category: "ACC",
			Message:  fmt.Sprintf("Account '%s' does not reference a key", accountName),
			Suggestions: []string{
				"Add 'key: <key-name>' to the account configuration",
			},
		}
	}

	key, ok := c.Keys[keyName]
	if !ok {
		available := make([]string, 0, len(c.Keys))
		for n := range c.Keys {
			available = append(available, n)
		}
		err := &errors.Error{
			Code:     "AIM-KEY-001",
			Category: "ACC",
			Message:  fmt.Sprintf("Key '%s' referenced by account '%s' not found", keyName, accountName),
			Suggestions: []string{
				"Add the key to the 'keys' section in your config",
			},
		}
		if len(available) > 0 {
			err.Details = map[string]interface{}{"available": available}
		}
		return nil, err
	}

	// 3. Resolve key value (using ResolveKey from parse.go)
	keyValue, err := ResolveKey(key.Value)
	if err != nil {
		return nil, err
	}

	// 4. Resolve vendor
	vendor, err := vendors.Resolve(key.Vendor, c.Vendors)
	if err != nil {
		return nil, err
	}

	// 5. Determine endpoint to use
	// Priority: Account.Endpoint > Tool.Protocol
	endpointName := acc.Endpoint
	if endpointName == "" {
		endpointName = toolProtocol
	}

	if endpointName == "" {
		return nil, &errors.Error{
			Code:     "AIM-TOO-001",
			Category: "TOO",
			Message:  fmt.Sprintf("Cannot determine endpoint for tool '%s'", toolName),
			Suggestions: []string{
				"Add the tool to the 'tools' section in your config with its protocol",
				"Or specify 'endpoint: <name>' in the account configuration",
			},
		}
	}

	// 6. Check if key is restricted to specific endpoints
	if len(key.Endpoints) > 0 {
		allowed := false
		for _, ep := range key.Endpoints {
			if ep == endpointName {
				allowed = true
				break
			}
		}
		if !allowed {
			err := &errors.Error{
				Code:     "AIM-KEY-002",
				Category: "ACC",
				Message:  fmt.Sprintf("Key '%s' is not allowed to use endpoint '%s'", keyName, endpointName),
				Suggestions: []string{
					"Add the endpoint to the 'endpoints' list in the key configuration",
					"Or remove the 'endpoints' restriction from the key",
				},
			}
			if len(key.Endpoints) > 0 {
				err.Details = map[string]interface{}{"allowed": key.Endpoints}
			}
			return nil, err
		}
	}

	// 7. Get endpoint configuration
	endpoint, err := vendor.GetEndpoint(endpointName)
	if err != nil {
		return nil, err
	}

	// 8. Determine model to use
	// Priority: Account.Model > Endpoint.DefaultModel
	model := acc.Model
	if model == "" {
		model = endpoint.DefaultModel
	}

	return &ResolvedAccount{
		Name:        accountName,
		Key:         keyValue,
		Vendor:      key.Vendor,
		Endpoint:    endpointName,
		EndpointURL: endpoint.URL,
		Model:       model,
	}, nil
}

// GetDefaultAccount returns the default account name
func (c *Config) GetDefaultAccount() (string, error) {
	if c.Settings.DefaultAccount != "" {
		// Verify the account exists
		if _, ok := c.Accounts[c.Settings.DefaultAccount]; ok {
			return c.Settings.DefaultAccount, nil
		}
		return "", &errors.Error{
			Code:     "AIM-ACC-003",
			Category: "ACC",
			Message:  fmt.Sprintf("Default account '%s' not found", c.Settings.DefaultAccount),
			Suggestions: []string{
				"Update 'default_account' in settings to an existing account",
				"Or add the missing account to your config",
			},
		}
	}

	// If only one account, use it
	if len(c.Accounts) == 1 {
		for name := range c.Accounts {
			return name, nil
		}
	}

	return "", &errors.Error{
		Code:     "AIM-ACC-004",
		Category: "ACC",
		Message:  "No default account configured",
		Suggestions: []string{
			"Set 'default_account' in the settings section",
			"Or use -a flag to specify an account",
		},
	}
}
