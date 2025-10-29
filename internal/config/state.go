package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// State represents the current runtime state for v2.0
type State struct {
	Version string       `yaml:"version"`
	Current CurrentState `yaml:"current"`
	Tools   ToolStates   `yaml:"tools,omitempty"`
}

// CurrentState represents the current active configuration for v2.0
type CurrentState struct {
	Tool        string    `yaml:"tool"`
	Key         string    `yaml:"key"`
	Provider    string    `yaml:"provider"`
	LastUpdated time.Time `yaml:"last_updated"`
}

// ToolStates maps tool names to their individual states
type ToolStates map[string]*ToolState

// ToolState represents state for a specific tool in v2.0
type ToolState struct {
	Key         string    `yaml:"key"`
	Provider    string    `yaml:"provider"`
	LastUsed    time.Time `yaml:"last_used,omitempty"`
	LastUpdated time.Time `yaml:"last_updated"`
}

// StateManager manages state persistence for v2.0
type StateManager struct {
	statePath string
}

// NewStateManager creates a new state manager
func NewStateManager() *StateManager {
	homeDir, _ := os.UserHomeDir()
	statePath := filepath.Join(homeDir, ".aim", "config", "state.yaml")

	// Check if AIM_HOME is set (for testing environments)
	if aimHome := os.Getenv("AIM_HOME"); aimHome != "" {
		statePath = filepath.Join(aimHome, "config", "state.yaml")
	}

	return &StateManager{
		statePath: statePath,
	}
}

// NewStateManagerWithPath creates a state manager with custom path
func NewStateManagerWithPath(path string) *StateManager {
	return &StateManager{
		statePath: path,
	}
}

// Load loads the current state
func (sm *StateManager) Load() (*State, error) {
	data, err := os.ReadFile(sm.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default state if file doesn't exist
			return sm.defaultState(), nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state State
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	return &state, nil
}

// Save saves the current state
func (sm *StateManager) Save(state *State) error {
	// Ensure directory exists
	dir := filepath.Dir(sm.statePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write file
	if err := os.WriteFile(sm.statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// defaultState returns the default state for v2.0
func (sm *StateManager) defaultState() *State {
	return &State{
		Version: "2.0",
		Current: CurrentState{
			Tool:        "claude-code",
			Key:         "",
			Provider:    "deepseek",
			LastUpdated: time.Now(),
		},
		Tools: make(ToolStates),
	}
}

// SetCurrent sets the current active configuration for v2.0
func (s *State) SetCurrent(tool, key, provider string) {
	s.Current = CurrentState{
		Tool:        tool,
		Key:         key,
		Provider:    provider,
		LastUpdated: time.Now(),
	}

	// Also update tool-specific state
	s.SetToolState(tool, key, provider)
}

// SetToolState sets state for a specific tool in v2.0
func (s *State) SetToolState(tool, key, provider string) {
	if s.Tools == nil {
		s.Tools = make(ToolStates)
	}

	s.Tools[tool] = &ToolState{
		Key:         key,
		Provider:    provider,
		LastUpdated: time.Now(),
	}
}

// GetCurrentKey returns the current key for a tool
func (s *State) GetCurrentKey(tool string) string {
	// Check tool-specific state first
	if s.Tools != nil {
		if toolState, ok := s.Tools[tool]; ok {
			return toolState.Key
		}
	}

	// Fall back to global current state if tool matches
	if s.Current.Tool == tool {
		return s.Current.Key
	}

	// Return empty if no state found
	return ""
}

// GetCurrentProvider returns the current provider for a tool
func (s *State) GetCurrentProvider(tool string) string {
	// Check tool-specific state first
	if s.Tools != nil {
		if toolState, ok := s.Tools[tool]; ok {
			return toolState.Provider
		}
	}

	// Fall back to global current state if tool matches
	if s.Current.Tool == tool {
		return s.Current.Provider
	}

	// Return empty if no state found
	return ""
}

// UpdateLastUsed updates the last used timestamp for a tool
func (s *State) UpdateLastUsed(tool string) {
	if s.Tools == nil {
		s.Tools = make(ToolStates)
	}

	if toolState, ok := s.Tools[tool]; ok {
		toolState.LastUsed = time.Now()
	} else {
		// Create new state if it doesn't exist
		s.Tools[tool] = &ToolState{
			LastUsed:    time.Now(),
			LastUpdated: time.Now(),
		}
	}
}

// GetCurrentModel returns the current model for a tool (for compatibility)
// In v2.0, model is determined by key+provider+tool combination
func (s *State) GetCurrentModel(tool string) string {
	// This method is kept for compatibility but should not be used in v2.0
	// Model resolution is now handled by the resolver
	return ""
}
