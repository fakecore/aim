package setup

import (
	"fmt"
	"strings"
	"time"

	"github.com/fakecore/aim/internal/config"
)

// SetupRequest setup request
type SetupRequest struct {
	ToolName string `json:"tool_name"`
	KeyName  string `json:"key_name"`
	Type     string `json:"type,omitempty"`     // for env subcommand
	Format   string `json:"format,omitempty"`   // for command subcommand
	DryRun   bool   `json:"dry_run,omitempty"`  // for install subcommand
	Force    bool   `json:"force,omitempty"`    // for install subcommand
	WithEnv  bool   `json:"with_env,omitempty"` // for command subcommand - include environment variables
}

// SetupResult setup result
type SetupResult struct {
	Request   *SetupRequest         `json:"request"`
	Runtime   *config.RuntimeConfig `json:"runtime"`
	EnvVars   map[string]string     `json:"env_vars"`
	Command   string                `json:"command,omitempty"`
	Generated time.Time             `json:"generated"`
	Metadata  *SetupMetadata        `json:"metadata"`
}

// SetupMetadata setup metadata
type SetupMetadata struct {
	Duration   time.Duration `json:"duration"`
	Source     string        `json:"source"`
	Version    string        `json:"version"`
	BackupPath string        `json:"backup_path,omitempty"`
	ConfigPath string        `json:"config_path,omitempty"`
}

// InstallRequest install request
type InstallRequest struct {
	*SetupRequest
	BackupPath string                `json:"backup_path,omitempty"`
	Runtime    *config.RuntimeConfig `json:"runtime,omitempty"`
}

// NewSetupRequest creates a new setup request
func NewSetupRequest(toolName, keyName string) *SetupRequest {
	return &SetupRequest{
		ToolName: toolName,
		KeyName:  keyName,
	}
}

// NewInstallRequest creates a new install request
func NewInstallRequest(toolName, keyName string) *InstallRequest {
	return &InstallRequest{
		SetupRequest: NewSetupRequest(toolName, keyName),
	}
}

// NewSetupResult creates a new setup result
func NewSetupResult(req *SetupRequest) *SetupResult {
	return &SetupResult{
		Request:   req,
		EnvVars:   make(map[string]string),
		Generated: time.Now(),
		Metadata: &SetupMetadata{
			Version: Version,
		},
	}
}

// Validate validates the setup request
func (r *SetupRequest) Validate() error {
	if r.ToolName == "" {
		return ErrMissingToolName
	}
	if r.KeyName == "" {
		return ErrMissingKeyName
	}

	// Validate tool name
	if !isSupportedTool(r.ToolName) {
		return fmt.Errorf("%w: %s", ErrToolNotSupported, r.ToolName)
	}

	// Validate type (for env subcommand)
	if r.Type != "" && !isSupportedEnvType(r.Type) {
		return fmt.Errorf("%w: %s", ErrInvalidFormat, r.Type)
	}

	// Validate format (for command subcommand)
	if r.Format != "" && !isSupportedCommandFormat(r.Format) {
		return fmt.Errorf("%w: %s", ErrInvalidFormat, r.Format)
	}

	return nil
}

// Normalize normalizes the setup request
func (r *SetupRequest) Normalize() {
	if r.Type == "" {
		r.Type = "zsh" // default environment variable type
	}
	if r.Format == "" {
		r.Format = "raw" // default command format
	}
	r.Type = strings.ToLower(r.Type)
	r.Format = strings.ToLower(r.Format)
}

// GetCanonicalToolName gets the canonical name of the tool
func (r *SetupRequest) GetCanonicalToolName() string {
	switch r.ToolName {
	case "cc":
		return "claude-code"
	case "claude-code":
		return "claude-code"
	case "codex":
		return "codex"
	default:
		return r.ToolName
	}
}

// isSupportedTool checks if the tool is supported
func isSupportedTool(toolName string) bool {
	for _, tool := range SupportedTools {
		if tool == toolName {
			return true
		}
	}
	return false
}

// isSupportedEnvType checks if the environment variable type is supported
func isSupportedEnvType(envType string) bool {
	for _, t := range SupportedEnvTypes {
		if t == envType {
			return true
		}
	}
	return false
}

// isSupportedCommandFormat checks if the command format is supported
func isSupportedCommandFormat(format string) bool {
	for _, f := range SupportedCommandFormats {
		if f == format {
			return true
		}
	}
	return false
}
