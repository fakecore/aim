package setup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// ClaudeCodeInstaller Claude Code installer
type ClaudeCodeInstaller struct{}

func NewClaudeCodeInstaller() *ClaudeCodeInstaller {
	return &ClaudeCodeInstaller{}
}

func (i *ClaudeCodeInstaller) Install(req *InstallRequest) error {
	// Get config path
	configPath, err := i.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if existing config is managed by AIM
	managedByAIM, existingConfig, err := i.checkManagedByAIM(configPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to check existing config: %w", err)
	}

	if managedByAIM {
		// If already managed by AIM, overwrite directly
		return i.installAIMManaged(req, configPath)
	} else if existingConfig != nil {
		// If there's a non-AIM managed config, copy and modify
		return i.installNonAIMManaged(req, configPath, existingConfig)
	} else {
		// No existing config, install directly
		return i.installNew(req, configPath)
	}
}

func (i *ClaudeCodeInstaller) Backup(req *InstallRequest) error {
	configPath, err := i.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file doesn't exist, no need to backup
		return nil
	}

	// Generate backup path
	backupPath := req.BackupPath
	if backupPath == "" {
		backupPath = configPath + ".bak." + time.Now().Format("20060102_1504")
	}

	// Copy file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}

func (i *ClaudeCodeInstaller) GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(home, ".claude", "settings.json"), nil
}

func (i *ClaudeCodeInstaller) ValidateConfig(path string) error {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist: %s", path)
	}

	// Try to parse JSON
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

func (i *ClaudeCodeInstaller) ConvertConfig(req *InstallRequest) (interface{}, error) {
	runtime := req.Runtime

	// Build Claude Code configuration
	claudeConfig := map[string]interface{}{
		"env": map[string]interface{}{
			"ANTHROPIC_AUTH_TOKEN":                     runtime.APIKey,
			"ANTHROPIC_BASE_URL":                       "https://open.bigmodel.cn/api/anthropic",
			"API_TIMEOUT_MS":                           "3000000",
			"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": 1,
		},
	}

	// If runtime.BaseURL is not empty and not the default value, use it
	if runtime.BaseURL != "" && runtime.BaseURL != "https://open.bigmodel.cn/api/anthropic" {
		// Convert AIM's base_url to anthropic format
		// Assume AIM's base_url is in format like "https://open.bigmodel.cn/api/paas/v4"
		// Need to convert to "https://open.bigmodel.cn/api/anthropic"
		baseURL := runtime.BaseURL
		// Simple conversion logic, replace path part with /api/anthropic
		if strings.Contains(baseURL, "open.bigmodel.cn") {
			baseURL = "https://open.bigmodel.cn/api/anthropic"
		}
		claudeConfig["env"].(map[string]interface{})["ANTHROPIC_BASE_URL"] = baseURL
	}

	// Add managed_by_aim field
	claudeConfig["managed_by_aim"] = map[string]interface{}{
		"version":    Version,
		"tool":       runtime.Tool,
		"key":        req.KeyName,
		"managed_at": time.Now().Format(time.RFC3339),
		"backup":     true,
	}

	return claudeConfig, nil
}

// CodexInstaller Codex installer
type CodexInstaller struct{}

func NewCodexInstaller() *CodexInstaller {
	return &CodexInstaller{}
}

func (i *CodexInstaller) Install(req *InstallRequest) error {
	// Get config path
	configPath, err := i.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if existing config is managed by AIM
	managedByAIM, existingConfig, err := i.checkManagedByAIM(configPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to check existing config: %w", err)
	}

	if managedByAIM {
		// If already managed by AIM, overwrite directly
		return i.installAIMManaged(req, configPath)
	} else if existingConfig != nil {
		// If there's a non-AIM managed config, copy and modify
		return i.installNonAIMManaged(req, configPath, existingConfig)
	} else {
		// No existing config, install directly
		return i.installNew(req, configPath)
	}
}

func (i *CodexInstaller) Backup(req *InstallRequest) error {
	configPath, err := i.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file doesn't exist, no need to backup
		return nil
	}

	// Generate backup path
	backupPath := req.BackupPath
	if backupPath == "" {
		backupPath = configPath + ".bak." + time.Now().Format("20060102_1504")
	}

	// Copy file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}

func (i *CodexInstaller) GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(home, ".codex", "config.toml"), nil
}

func (i *CodexInstaller) ValidateConfig(path string) error {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist: %s", path)
	}

	// Try to parse TOML
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config map[string]interface{}
	if err := toml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

func (i *CodexInstaller) ConvertConfig(req *InstallRequest) (interface{}, error) {
	runtime := req.Runtime

	// Build Codex configuration, conforming to official standards
	codexConfig := map[string]interface{}{
		"model":           runtime.Model,
		"approval_policy": "auto", // Default auto-approval policy
		"sandbox_mode":    false,  // Default not using sandbox mode
		"model_providers": map[string]interface{}{
			runtime.Provider: map[string]interface{}{
				"api_key":  runtime.APIKey,
				"base_url": runtime.BaseURL,
			},
		},
		"settings": map[string]interface{}{
			"temperature": 0.7,
			"max_tokens":  4096,
		},
	}

	// Add managed_by_aim field, maintaining backward compatibility
	codexConfig["managed_by_aim"] = map[string]interface{}{
		"version":    Version,
		"tool":       runtime.Tool,
		"key":        req.KeyName,
		"managed_at": time.Now().Format(time.RFC3339),
		"backup":     true,
	}

	return codexConfig, nil
}

// checkManagedByAIM checks if the configuration is managed by AIM
func (i *ClaudeCodeInstaller) checkManagedByAIM(configPath string) (bool, map[string]interface{}, error) {
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return false, nil, nil
	}

	// Read existing configuration
	data, err := os.ReadFile(configPath)
	if err != nil {
		return false, nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var existingConfig map[string]interface{}
	if err := json.Unmarshal(data, &existingConfig); err != nil {
		return false, nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Check if there's a managed_by_aim field
	if managedByAIM, exists := existingConfig["managed_by_aim"]; exists {
		if managedMap, ok := managedByAIM.(map[string]interface{}); ok {
			// Check if there's backup information
			if hasBackup, ok := managedMap["backup"].(bool); ok && hasBackup {
				return true, existingConfig, nil
			}
		}
	}

	return false, existingConfig, nil
}

// installAIMManaged installs to existing AIM-managed configuration
func (i *ClaudeCodeInstaller) installAIMManaged(req *InstallRequest, configPath string) error {
	// Convert configuration
	claudeConfig, err := i.ConvertConfig(req)
	if err != nil {
		return fmt.Errorf("failed to convert config: %w", err)
	}

	// Serialize configuration
	configData, err := json.MarshalIndent(claudeConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Directly overwrite the config file
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// installNonAIMManaged installs to non-AIM managed configuration
func (i *ClaudeCodeInstaller) installNonAIMManaged(req *InstallRequest, configPath string, existingConfig map[string]interface{}) error {
	// Create new configuration based on existing config
	newConfig := make(map[string]interface{})

	// Copy all fields from existing configuration
	for k, v := range existingConfig {
		newConfig[k] = v
	}

	// Convert AIM configuration
	claudeConfig, err := i.ConvertConfig(req)
	if err != nil {
		return fmt.Errorf("failed to convert config: %w", err)
	}

	// Merge configurations, prioritizing fields from AIM config
	if claudeMap, ok := claudeConfig.(map[string]interface{}); ok {
		for k, v := range claudeMap {
			newConfig[k] = v
		}
	}

	// Serialize configuration
	configData, err := json.MarshalIndent(newConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write configuration file
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// installNew installs new configuration
func (i *ClaudeCodeInstaller) installNew(req *InstallRequest, configPath string) error {
	// Convert configuration
	claudeConfig, err := i.ConvertConfig(req)
	if err != nil {
		return fmt.Errorf("failed to convert config: %w", err)
	}

	// Serialize configuration
	configData, err := json.MarshalIndent(claudeConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write configuration file
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// checkManagedByAIM checks if the configuration is managed by AIM
func (i *CodexInstaller) checkManagedByAIM(configPath string) (bool, map[string]interface{}, error) {
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return false, nil, nil
	}

	// Read existing configuration
	data, err := os.ReadFile(configPath)
	if err != nil {
		return false, nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse TOML
	var existingConfig map[string]interface{}
	if err := toml.Unmarshal(data, &existingConfig); err != nil {
		return false, nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Check if there's a managed_by_aim field
	if managedByAIM, exists := existingConfig["managed_by_aim"]; exists {
		if managedMap, ok := managedByAIM.(map[string]interface{}); ok {
			// Check if there's backup information
			if hasBackup, ok := managedMap["backup"].(bool); ok && hasBackup {
				return true, existingConfig, nil
			}
		}
	}

	return false, existingConfig, nil
}

// installAIMManaged installs to existing AIM-managed configuration
func (i *CodexInstaller) installAIMManaged(req *InstallRequest, configPath string) error {
	// Convert configuration
	codexConfig, err := i.ConvertConfig(req)
	if err != nil {
		return fmt.Errorf("failed to convert config: %w", err)
	}

	// Serialize configuration to TOML format
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(codexConfig); err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	configData := buf.Bytes()

	// Directly overwrite the config file
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// installNonAIMManaged installs to non-AIM managed configuration
func (i *CodexInstaller) installNonAIMManaged(req *InstallRequest, configPath string, existingConfig map[string]interface{}) error {
	// Create new configuration based on existing config
	newConfig := make(map[string]interface{})

	// Copy all fields from existing configuration
	for k, v := range existingConfig {
		newConfig[k] = v
	}

	// Convert AIM configuration
	codexConfig, err := i.ConvertConfig(req)
	if err != nil {
		return fmt.Errorf("failed to convert config: %w", err)
	}

	// Merge configurations, prioritizing fields from AIM config
	if codexMap, ok := codexConfig.(map[string]interface{}); ok {
		for k, v := range codexMap {
			newConfig[k] = v
		}
	}

	// Serialize configuration to TOML format
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(newConfig); err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	configData := buf.Bytes()

	// Write configuration file
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// installNew installs new configuration
func (i *CodexInstaller) installNew(req *InstallRequest, configPath string) error {
	// Convert configuration
	codexConfig, err := i.ConvertConfig(req)
	if err != nil {
		return fmt.Errorf("failed to convert config: %w", err)
	}

	// Serialize configuration to TOML format
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(codexConfig); err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	configData := buf.Bytes()

	// Write configuration file
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
