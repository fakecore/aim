package setup

// ToolInstaller tool installer interface
type ToolInstaller interface {
	Install(req *InstallRequest) error
	Backup(req *InstallRequest) error
	GetConfigPath() (string, error)
	ValidateConfig(path string) error
	ConvertConfig(req *InstallRequest) (interface{}, error)
}

// OutputFormatter output formatter base interface
type OutputFormatter interface {
	GetName() string
	GetContentType() string
}

// EnvFormatter environment variable formatter interface
type EnvFormatter interface {
	OutputFormatter
	FormatEnv(result *SetupResult) string
}

// CommandFormatter command formatter interface
type CommandFormatter interface {
	OutputFormatter
	FormatCommand(result *SetupResult) string
}
