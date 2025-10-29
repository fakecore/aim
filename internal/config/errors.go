package config

import "errors"

// Common configuration errors
var (
	// ErrInvalidVersion is returned when configuration version is invalid
	ErrInvalidVersion = errors.New("invalid configuration version")

	// ErrMissingDefaultTool is returned when no default tool is configured
	ErrMissingDefaultTool = errors.New("missing default tool")

	// ErrMissingDefaultProvider is returned when no default provider is configured
	ErrMissingDefaultProvider = errors.New("missing default provider")

	// ErrMissingDefaultKey is returned when no default key is configured
	ErrMissingDefaultKey = errors.New("missing default key")

	// ErrInvalidYAML is returned when YAML parsing fails
	ErrInvalidYAML = errors.New("invalid YAML format")

	// ErrKeyNotFound is returned when a key is not found
	ErrKeyNotFound = errors.New("key not found")

	// ErrToolNotFound is returned when a tool is not found
	ErrToolNotFound = errors.New("tool not found")

	// ErrProviderNotFound is returned when a provider is not found
	ErrProviderNotFound = errors.New("provider not found")

	// ErrInvalidKey is returned when a key is invalid
	ErrInvalidKey = errors.New("invalid key")

	// ErrInvalidTool is returned when a tool is invalid
	ErrInvalidTool = errors.New("invalid tool")
)
