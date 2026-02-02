package extension

import (
	"fmt"
)

// Extension represents a vendor extension
type Extension struct {
	Name        string              `yaml:"name"`
	Version     string              `yaml:"version"`
	Description string              `yaml:"description,omitempty"`
	Protocols   map[string]Protocol `yaml:"protocols"`
}

// Protocol represents a protocol configuration
type Protocol struct {
	URL          string            `yaml:"url"`
	EnvTemplate  map[string]string `yaml:"env_template,omitempty"`
	Headers      map[string]string `yaml:"headers,omitempty"`
}

// Validate checks if the extension is valid
func (e *Extension) Validate() error {
	if e.Name == "" {
		return fmt.Errorf("extension name is required")
	}
	if len(e.Protocols) == 0 {
		return fmt.Errorf("extension must define at least one protocol")
	}
	for name, proto := range e.Protocols {
		if proto.URL == "" {
			return fmt.Errorf("protocol %s: URL is required", name)
		}
	}
	return nil
}
