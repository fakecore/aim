package setup

import "errors"

// Setup related error definitions
var (
	ErrMissingKeyName    = errors.New("key name is required")
	ErrMissingToolName   = errors.New("tool name is required")
	ErrInvalidFormat     = errors.New("invalid output format")
	ErrUnsupportedFormat = errors.New("unsupported output format")
	ErrToolNotSupported  = errors.New("tool not supported")
	ErrKeyNotFound       = errors.New("key not found")
	ErrConfigNotFound    = errors.New("tool config file not found")
	ErrBackupFailed      = errors.New("failed to backup existing config")
	ErrInstallFailed     = errors.New("failed to install configuration")
	ErrInstallerNotFound = errors.New("installer not found")
	ErrFormatterNotFound = errors.New("formatter not found")
	ErrNoDefaultTool     = errors.New("no default tool configured")
)

// List of supported tools
var SupportedTools = []string{"cc", "claude-code", "codex"}

// Supported environment variable types
var SupportedEnvTypes = []string{"zsh", "bash", "fish", "json"}

// Supported command formats
var SupportedCommandFormats = []string{"raw", "shell", "json", "simple"}

// Version information (imported from main package)
const Version = "2.0.0"
