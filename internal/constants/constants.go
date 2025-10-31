package constants

import "time"

// File permissions
const (
	ConfigFileMode = 0644
	ConfigDirMode  = 0755
)

// Timeout constants
const (
	DefaultTimeout    = 60 * time.Second
	GLMTimeout        = 50 * time.Minute // 3,000,000ms
	GLMCodingTimeout  = 5 * time.Minute  // 300,000ms
)

// Timeout in milliseconds for configuration
const (
	DefaultTimeoutMS = 60000
	GLMTimeoutMS     = 300000
)