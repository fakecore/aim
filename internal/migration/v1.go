package migration

// V1Config represents the v1 configuration format
type V1Config struct {
	Version   string                `toml:"version"`
	Settings  V1Settings            `toml:"settings"`
	Keys      map[string]V1Key      `toml:"keys"`
	Providers map[string]V1Provider `toml:"providers"`
	Tools     map[string]V1Tool     `toml:"tools"`
}

type V1Settings struct {
	DefaultProvider string `toml:"default_provider"`
}

type V1Key struct {
	Value     string `toml:"value"`
	Provider  string `toml:"provider"`
	IsDefault bool   `toml:"is_default"`
}

type V1Provider struct {
	BaseURL string            `toml:"base_url"`
	APIPath string            `toml:"api_path"`
	Headers map[string]string `toml:"headers"`
}

type V1Tool struct {
	Name     string `toml:"name"`
	Protocol string `toml:"protocol"`
}
