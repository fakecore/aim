package extension

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadDir loads all extensions from a directory
func LoadDir(dir string) (map[string]Extension, error) {
	extensions := make(map[string]Extension)

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return extensions, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".yaml" && filepath.Ext(entry.Name()) != ".yml" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		ext, err := LoadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", entry.Name(), err)
		}

		if err := ext.Validate(); err != nil {
			return nil, fmt.Errorf("invalid extension %s: %w", entry.Name(), err)
		}

		extensions[ext.Name] = *ext
	}

	return extensions, nil
}

// LoadFile loads a single extension from file
func LoadFile(path string) (*Extension, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var ext Extension
	if err := yaml.Unmarshal(data, &ext); err != nil {
		return nil, err
	}

	return &ext, nil
}

// DefaultDir returns the default extensions directory
func DefaultDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "aim", "extensions")
}
