package setup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fakecore/aim/internal/config"
	"github.com/fakecore/aim/internal/constants"
	"github.com/fakecore/aim/internal/tool"
)

// SetupManager setup manager
type SetupManager struct {
	configManager   *config.ConfigManager
	preparerManager *tool.EnvironmentPreparerManager
	installers      map[string]ToolInstaller
	formatters      map[string]OutputFormatter
	logger          Logger
}

// Logger log interface
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// DefaultLogger default log implementation
type DefaultLogger struct{}

func (l *DefaultLogger) Debugf(format string, args ...interface{}) {}
func (l *DefaultLogger) Infof(format string, args ...interface{})  { fmt.Printf(format+"\n", args...) }
func (l *DefaultLogger) Errorf(format string, args ...interface{}) { fmt.Printf(format+"\n", args...) }

// NewSetupManager creates a setup manager
func NewSetupManager(logger Logger) *SetupManager {
	if logger == nil {
		logger = &DefaultLogger{}
	}

	manager := &SetupManager{
		configManager:   config.GetConfigManager(),
		preparerManager: tool.NewEnvironmentPreparerManager(),
		installers:      make(map[string]ToolInstaller),
		formatters:      make(map[string]OutputFormatter),
		logger:          logger,
	}

	// Register default installers
	manager.registerDefaultInstallers()

	// Register default formatters
	manager.registerDefaultFormatters()

	return manager
}

// registerDefaultInstallers registers default installers
func (sm *SetupManager) registerDefaultInstallers() {
	sm.installers["claude-code"] = NewClaudeCodeInstaller()
	sm.installers["cc"] = NewClaudeCodeInstaller() // cc is an alias for claude-code
	sm.installers["codex"] = NewCodexInstaller()
}

// registerDefaultFormatters registers default formatters
func (sm *SetupManager) registerDefaultFormatters() {
	// Environment variable formatters
	sm.formatters["zsh"] = NewZshFormatter()
	sm.formatters["bash"] = NewBashFormatter()
	sm.formatters["fish"] = NewFishFormatter()
	sm.formatters["json"] = NewJSONFormatter(false)

	// Command formatters
	sm.formatters["raw"] = NewRawCommandFormatter()
	sm.formatters["shell"] = NewShellCommandFormatter()
	sm.formatters["json-command"] = NewJSONCommandFormatter()
	sm.formatters["simple"] = NewSimpleCommandFormatter()
}

// ExportEnv exports environment variables
func (sm *SetupManager) ExportEnv(ctx context.Context, req *SetupRequest) (*SetupResult, error) {
	startTime := time.Now()

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}
	req.Normalize()

	// Create result object
	result := NewSetupResult(req)

	// Get configuration
	cfg := sm.configManager.GetConfig()
	result.Metadata.Source = sm.configManager.GetConfigPath()

	// Create resolver
	resolver := config.NewResolver(cfg)

	// Validate key
	if err := resolver.ValidateKey(req.KeyName); err != nil {
		return nil, sm.enrichKeyError(err, cfg)
	}

	// Get canonical tool name
	canonicalToolName := req.GetCanonicalToolName()

	// Validate tool
	if err := resolver.ValidateTool(canonicalToolName); err != nil {
		return nil, sm.enrichToolError(err, cfg)
	}

	// Resolve runtime configuration
	runtime, err := resolver.Resolve(canonicalToolName, req.KeyName, "")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve configuration: %w", err)
	}
	result.Runtime = runtime

	// Prepare tool-specific environment variables
	_, toolEnvVars, err := sm.preparerManager.PrepareEnvironment(runtime)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare tool environment: %w", err)
	}

	// Merge environment variables
	result.EnvVars = sm.mergeEnvVars(runtime.EnvVars, toolEnvVars)

	// Update metadata
	result.Metadata.Duration = time.Since(startTime)

	sm.logger.Debugf("Environment export completed in %v", result.Metadata.Duration)

	return result, nil
}

// GenerateCommand generates execution command
func (sm *SetupManager) GenerateCommand(ctx context.Context, req *SetupRequest) (*SetupResult, error) {
	startTime := time.Now()

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}
	req.Normalize()

	// Create result object
	result := NewSetupResult(req)

	// Get configuration
	cfg := sm.configManager.GetConfig()
	result.Metadata.Source = sm.configManager.GetConfigPath()

	// Create resolver
	resolver := config.NewResolver(cfg)

	// Validate key
	if err := resolver.ValidateKey(req.KeyName); err != nil {
		return nil, sm.enrichKeyError(err, cfg)
	}

	// Get canonical tool name
	canonicalToolName := req.GetCanonicalToolName()

	// Validate tool
	if err := resolver.ValidateTool(canonicalToolName); err != nil {
		return nil, sm.enrichToolError(err, cfg)
	}

	// Resolve runtime configuration
	runtime, err := resolver.Resolve(canonicalToolName, req.KeyName, "")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve configuration: %w", err)
	}
	result.Runtime = runtime

	// Prepare tool-specific environment variables and arguments
	toolArgs, toolEnvVars, err := sm.preparerManager.PrepareEnvironment(runtime)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare tool environment: %w", err)
	}

	// Merge environment variables
	result.EnvVars = sm.mergeEnvVars(runtime.EnvVars, toolEnvVars)

	// Generate command (includes environment variables by default)
	result.Command = sm.buildCommandWithEnv(canonicalToolName, runtime, toolArgs, result.EnvVars)

	// Update metadata
	result.Metadata.Duration = time.Since(startTime)

	sm.logger.Debugf("Command generation completed in %v", result.Metadata.Duration)

	return result, nil
}

// InstallConfig installs configuration
func (sm *SetupManager) InstallConfig(ctx context.Context, req *InstallRequest) (*SetupResult, error) {
	startTime := time.Now()

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}
	req.Normalize()

	// Create result object
	result := NewSetupResult(req.SetupRequest)

	// Get configuration
	cfg := sm.configManager.GetConfig()
	result.Metadata.Source = sm.configManager.GetConfigPath()

	// Create resolver
	resolver := config.NewResolver(cfg)

	// Validate key
	if err := resolver.ValidateKey(req.KeyName); err != nil {
		return nil, sm.enrichKeyError(err, cfg)
	}

	// Get canonical tool name
	canonicalToolName := req.GetCanonicalToolName()

	// Validate tool
	if err := resolver.ValidateTool(canonicalToolName); err != nil {
		return nil, sm.enrichToolError(err, cfg)
	}

	// Resolve runtime configuration
	runtime, err := resolver.Resolve(canonicalToolName, req.KeyName, "")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve configuration: %w", err)
	}
	result.Runtime = runtime

	// Get installer
	installer, exists := sm.installers[canonicalToolName]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrInstallerNotFound, canonicalToolName)
	}

	// Update Runtime field of install request
	req.Runtime = runtime

	// Get configuration file path
	configPath, err := installer.GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}
	result.Metadata.ConfigPath = configPath

	// Backup existing configuration
	if !req.DryRun {
		backupPath, err := sm.backupConfig(installer, req)
		if err != nil {
			return nil, fmt.Errorf("failed to backup config: %w", err)
		}
		result.Metadata.BackupPath = backupPath
	}

	// Install configuration
	if !req.DryRun {
		if err := installer.Install(req); err != nil {
			return nil, fmt.Errorf("failed to install config: %w", err)
		}
	}

	// Prepare tool-specific environment variables
	_, toolEnvVars, err := sm.preparerManager.PrepareEnvironment(runtime)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare tool environment: %w", err)
	}

	// Merge environment variables
	result.EnvVars = sm.mergeEnvVars(runtime.EnvVars, toolEnvVars)

	// Update metadata
	result.Metadata.Duration = time.Since(startTime)

	sm.logger.Debugf("Config installation completed in %v", result.Metadata.Duration)

	return result, nil
}

// RestoreConfig restores configuration
func (sm *SetupManager) RestoreConfig(ctx context.Context, toolName string, backupPath string) (*SetupResult, error) {
	startTime := time.Now()

	// Get canonical tool name
	canonicalToolName := sm.getCanonicalToolName(toolName)

	// Get installer
	installer, exists := sm.installers[canonicalToolName]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrInstallerNotFound, canonicalToolName)
	}

	// Get configuration file path
	configPath, err := installer.GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	// If no backup path is specified, find the latest backup
	if backupPath == "" {
		backupPath, err = sm.findLatestBackup(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to find backup: %w", err)
		}
	}

	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("backup file does not exist: %s", backupPath)
	}

	// Create result object
	result := &SetupResult{
		Metadata: &SetupMetadata{
			Version:    Version,
			BackupPath: backupPath,
			ConfigPath: configPath,
		},
		Generated: time.Now(),
	}

	// Restore configuration
	if err := sm.restoreFromBackup(installer, backupPath, configPath); err != nil {
		return nil, fmt.Errorf("failed to restore from backup: %w", err)
	}

	// Update metadata
	result.Metadata.Duration = time.Since(startTime)

	sm.logger.Debugf("Configuration restore completed in %v", result.Metadata.Duration)

	return result, nil
}

// FormatEnv formats environment variable output
func (sm *SetupManager) FormatEnv(result *SetupResult, formatType string) (string, error) {
	formatter, exists := sm.formatters[formatType]
	if !exists {
		return "", fmt.Errorf("%w: %s", ErrFormatterNotFound, formatType)
	}

	envFormatter, ok := formatter.(EnvFormatter)
	if !ok {
		return "", fmt.Errorf("formatter %s does not support environment variable formatting", formatType)
	}

	return envFormatter.FormatEnv(result), nil
}

// FormatCommand formats command output
func (sm *SetupManager) FormatCommand(result *SetupResult, formatType string) (string, error) {
	formatter, exists := sm.formatters[formatType]
	if !exists {
		return "", fmt.Errorf("%w: %s", ErrFormatterNotFound, formatType)
	}

	commandFormatter, ok := formatter.(CommandFormatter)
	if !ok {
		return "", fmt.Errorf("formatter %s does not support command formatting", formatType)
	}

	return commandFormatter.FormatCommand(result), nil
}

// mergeEnvVars merges environment variables
func (sm *SetupManager) mergeEnvVars(base, toolSpecific map[string]string) map[string]string {
	merged := make(map[string]string)

	// First add base environment variables
	for k, v := range base {
		merged[k] = v
	}

	// Then add tool-specific environment variables (overriding base variables)
	for k, v := range toolSpecific {
		merged[k] = v
	}

	return merged
}

// buildCommand builds execution command
func (sm *SetupManager) buildCommand(toolName string, _ *config.RuntimeConfig, toolArgs []string) string {
	// Get tool command
	cfg := sm.configManager.GetConfig()
	toolConfig, _ := cfg.GetTool(toolName)

	if toolConfig == nil {
		return ""
	}

	// Build command
	var cmd string
	if toolConfig.Command != "" {
		cmd = toolConfig.Command
	} else {
		// Default to using tool name as command
		cmd = toolName
	}

	// Add tool-specific arguments
	if len(toolArgs) > 0 {
		for _, arg := range toolArgs {
			cmd += " " + arg
		}
	}

	return cmd
}

// buildCommandWithEnv builds execution command with environment variables
func (sm *SetupManager) buildCommandWithEnv(toolName string, runtime *config.RuntimeConfig, toolArgs []string, envVars map[string]string) string {
	// Get base command
	baseCmd := sm.buildCommand(toolName, runtime, toolArgs)
	if baseCmd == "" {
		return ""
	}

	// If no environment variables, return base command directly
	if len(envVars) == 0 {
		return baseCmd
	}

	// Build environment variable export section
	var envExports []string
	for key, value := range envVars {
		// Escape special characters
		escapedValue := EscapeShellValue(value)
		// Handle tab and newline characters specifically for command context
		escapedValue = strings.ReplaceAll(escapedValue, `\t`, `\t`)
		escapedValue = strings.ReplaceAll(escapedValue, `\n`, `\n`)

		envExports = append(envExports, fmt.Sprintf("export %s=\"%s\"", key, escapedValue))
	}

	// Combine environment variable exports and command
	return strings.Join(envExports, " && ") + " && " + baseCmd
}

// backupConfig backs up configuration
func (sm *SetupManager) backupConfig(installer ToolInstaller, req *InstallRequest) (string, error) {
	backupPath := req.BackupPath
	if backupPath == "" {
		// Generate default backup path
		configPath, err := installer.GetConfigPath()
		if err != nil {
			return "", err
		}
		backupPath = configPath + ".bak." + time.Now().Format("20060102_1504")
	}

	if err := installer.Backup(&InstallRequest{
		SetupRequest: req.SetupRequest,
		BackupPath:   backupPath,
	}); err != nil {
		return "", err
	}

	return backupPath, nil
}

// enrichKeyError enriches key error information
func (sm *SetupManager) enrichKeyError(err error, cfg *config.Config) error {
	if len(cfg.Keys) == 0 {
		return fmt.Errorf("%w\n\nNo keys configured. Use 'aim keys add <name> --provider <provider> --key <api-key>' to add a key.", err)
	}

	var availableKeys []string
	for name := range cfg.Keys {
		availableKeys = append(availableKeys, name)
	}

	return fmt.Errorf("%w\n\nAvailable keys:\n  %s\n\nUse 'aim keys list' to see all keys.",
		err, strings.Join(availableKeys, "\n  "))
}

// enrichToolError enriches tool error information
func (sm *SetupManager) enrichToolError(err error, cfg *config.Config) error {
	var availableTools []string
	for name := range cfg.Tools {
		availableTools = append(availableTools, name)
	}

	return fmt.Errorf("%w\n\nAvailable tools:\n  %s\n\nUse 'aim tool list' to see all tools.",
		err, strings.Join(availableTools, "\n  "))
}

// RegisterInstaller registers an installer
func (sm *SetupManager) RegisterInstaller(toolName string, installer ToolInstaller) {
	sm.installers[toolName] = installer
}

// RegisterFormatter registers a formatter
func (sm *SetupManager) RegisterFormatter(name string, formatter OutputFormatter) {
	sm.formatters[name] = formatter
}

// getCanonicalToolName gets the canonical name of a tool
func (sm *SetupManager) getCanonicalToolName(toolName string) string {
	switch toolName {
	case "cc":
		return "claude-code"
	case "claude-code":
		return "claude-code"
	case "codex":
		return "codex"
	default:
		return toolName
	}
}

// findLatestBackup finds the latest backup file
func (sm *SetupManager) findLatestBackup(configPath string) (string, error) {
	// Get configuration file directory and base name
	configDir := filepath.Dir(configPath)
	configBase := filepath.Base(configPath)

	// Find all backup files
	matches, err := filepath.Glob(filepath.Join(configDir, configBase+".bak.*"))
	if err != nil {
		return "", fmt.Errorf("failed to search for backup files: %w", err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no backup files found")
	}

	// Find the latest backup file
	var latestBackup string
	var latestTime time.Time

	for _, match := range matches {
		// Extract timestamp from filename
		fileName := filepath.Base(match)
		parts := strings.Split(fileName, ".")
		if len(parts) < 3 {
			continue
		}

		timeStr := parts[len(parts)-1] // Get timestamp part
		if parsedTime, err := time.Parse("20060102_1504", timeStr); err == nil {
			if parsedTime.After(latestTime) {
				latestTime = parsedTime
				latestBackup = match
			}
		}
	}

	if latestBackup == "" {
		return "", fmt.Errorf("no valid backup files found")
	}

	return latestBackup, nil
}

// restoreFromBackup restores configuration from backup file
func (sm *SetupManager) restoreFromBackup(installer ToolInstaller, backupPath, configPath string) error {
	// Read backup file
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// Write to configuration file
	if err := os.WriteFile(configPath, data, constants.ConfigFileMode); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
