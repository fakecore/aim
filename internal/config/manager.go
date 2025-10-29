package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// ConfigManager manages global configuration state
type ConfigManager struct {
	loader      *Loader
	config      *Config
	state       *State
	stateMgr    *StateManager
	modified    bool
	mutex       sync.RWMutex
	initialized bool
}

var (
	globalManager *ConfigManager
	once          sync.Once
)

// GetConfigManager returns global configuration manager instance
func GetConfigManager() *ConfigManager {
	once.Do(func() {
		globalManager = &ConfigManager{
			loader:   NewLoader(),
			stateMgr: NewStateManager(),
			modified: false,
		}
	})
	return globalManager
}

// Initialize initializes configuration and state
func (cm *ConfigManager) Initialize() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cm.initialized {
		return nil
	}

	// Check if global config file exists, create if missing
	if cm.isConfigMissing() {
		// Auto-initialize with default configuration
		if initErr := cm.autoInitialize(); initErr != nil {
			return fmt.Errorf("failed to auto-initialize configuration: %w", initErr)
		}
	}

	// Load configuration file
	cfg, err := cm.loader.Load()
	if err != nil {
		return fmt.Errorf(`configuration initialization failed: %w

Possible solutions:
1. Check file permissions: %s
2. Validate YAML syntax in configuration file
3. Check environment variables: AIM_CONFIG_PATH, AIM_HOME
4. Try removing and recreating: rm %s && aim config init

For detailed troubleshooting, see: https://github.com/fakecore/aim/blob/main/README.md#-故障排查`,
			err, cm.loader.globalPath, cm.loader.globalPath)
	}
	cm.config = cfg

	// Load state file (create default if missing)
	state, err := cm.stateMgr.Load()
	if err != nil {
		// Auto-create default state if missing
		if cm.isStateMissing() {
			state = cm.stateMgr.defaultState()
			if saveErr := cm.stateMgr.Save(state); saveErr != nil {
				// Non-fatal: continue with default state in memory
				fmt.Fprintf(os.Stderr, "Warning: Could not save state file: %v\n", saveErr)
			}
		} else {
			return fmt.Errorf(`state initialization failed: %w

Possible solutions:
1. Check directory permissions: %s
2. Check available disk space
3. Remove corrupted state file: rm -rf %s`,
				err, filepath.Dir(cm.stateMgr.statePath), cm.stateMgr.statePath)
		}
	}
	cm.state = state

	cm.initialized = true
	return nil
}

// isConfigMissing checks if the configuration file is missing
func (cm *ConfigManager) isConfigMissing() bool {
	_, err := os.Stat(cm.loader.globalPath)
	return os.IsNotExist(err)
}

// isStateMissing checks if the state file is missing
func (cm *ConfigManager) isStateMissing() bool {
	_, err := os.Stat(cm.stateMgr.statePath)
	return os.IsNotExist(err)
}

// autoInitialize automatically creates default configuration
func (cm *ConfigManager) autoInitialize() error {
	fmt.Fprintf(os.Stderr, "ℹ No configuration found, creating default config at %s\n", cm.loader.globalPath)

	if err := cm.loader.InitGlobalSilent(); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "✓ Configuration initialized successfully\n")
	return nil
}

// GetConfig returns current configuration (read-only access)
func (cm *ConfigManager) GetConfig() *Config {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	if !cm.initialized {
		// This should not happen since we force initialize in main
		panic("configuration not initialized - call Initialize() first")
	}

	return cm.config
}

// GetState returns current state (read-only access)
func (cm *ConfigManager) GetState() *State {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	if !cm.initialized {
		panic("state not initialized - call Initialize() first")
	}

	return cm.state
}

// UpdateConfig updates configuration and marks it as modified
func (cm *ConfigManager) UpdateConfig(updateFunc func(*Config)) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if !cm.initialized {
		return fmt.Errorf("configuration not initialized")
	}

	updateFunc(cm.config)
	cm.modified = true
	return nil
}

// UpdateState updates state and marks it as modified
func (cm *ConfigManager) UpdateState(updateFunc func(*State)) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if !cm.initialized {
		return fmt.Errorf("state not initialized")
	}

	updateFunc(cm.state)
	cm.modified = true
	return nil
}

// ForceSave forces immediate save of configuration and state
func (cm *ConfigManager) ForceSave() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if !cm.initialized {
		return fmt.Errorf("configuration not initialized")
	}

	return cm.forceSaveUnsafe()
}

// SaveIfModified saves configuration and state only if they have been modified
func (cm *ConfigManager) SaveIfModified() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if !cm.modified {
		return nil // No changes, no need to save
	}

	return cm.forceSaveUnsafe()
}

// Save saves configuration and state (always saves)
func (cm *ConfigManager) Save() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if !cm.initialized {
		return fmt.Errorf("configuration not initialized")
	}

	return cm.forceSaveUnsafe()
}

// forceSaveUnsafe saves configuration and state without locking (for internal use)
func (cm *ConfigManager) forceSaveUnsafe() error {
	// Save configuration
	if err := cm.loader.SaveGlobal(cm.config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Save state
	if err := cm.stateMgr.Save(cm.state); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	cm.modified = false
	return nil
}

// GetConfigPath returns the global configuration file path
func (cm *ConfigManager) GetConfigPath() string {
	return cm.loader.GetGlobalPath()
}

// GetLocalConfigPath returns the local configuration file path
func (cm *ConfigManager) GetLocalConfigPath() string {
	return cm.loader.GetLocalPath()
}

// InitGlobal initializes the global configuration file
func (cm *ConfigManager) InitGlobal() error {
	return cm.loader.InitGlobal()
}
