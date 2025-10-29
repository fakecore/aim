package setup

// Temporary placeholder implementations, will be replaced with concrete implementations later
// Note: NewClaudeCodeInstaller and NewCodexInstaller have been implemented in installers.go

// NewZshFormatter creates a Zsh formatter
func NewZshFormatter() EnvFormatter {
	return &ZshFormatter{}
}

// NewBashFormatter creates a Bash formatter
func NewBashFormatter() EnvFormatter {
	return &BashFormatter{}
}

// NewFishFormatter creates a Fish formatter
func NewFishFormatter() EnvFormatter {
	return &FishFormatter{}
}

// NewJSONFormatter creates a JSON formatter
func NewJSONFormatter(pretty bool) EnvFormatter {
	return &JSONFormatter{pretty: pretty}
}

// NewRawCommandFormatter creates a Raw command formatter
func NewRawCommandFormatter() CommandFormatter {
	return &RawCommandFormatter{}
}

// NewShellCommandFormatter creates a Shell command formatter
func NewShellCommandFormatter() CommandFormatter {
	return &ShellCommandFormatter{}
}

// NewJSONCommandFormatter creates a JSON command formatter
func NewJSONCommandFormatter() CommandFormatter {
	return &JSONCommandFormatter{}
}

// NewSimpleCommandFormatter creates a simple command formatter
func NewSimpleCommandFormatter() CommandFormatter {
	return &SimpleCommandFormatter{}
}

// Placeholder implementations - only keeping installer placeholders, formatters have been moved to formatters.go

type claudeCodeInstallerPlaceholder struct{}

func (p *claudeCodeInstallerPlaceholder) Install(req *InstallRequest) error {
	return nil
}

func (p *claudeCodeInstallerPlaceholder) Backup(req *InstallRequest) error {
	return nil
}

func (p *claudeCodeInstallerPlaceholder) GetConfigPath() (string, error) {
	return "", nil
}

func (p *claudeCodeInstallerPlaceholder) ValidateConfig(path string) error {
	return nil
}

func (p *claudeCodeInstallerPlaceholder) ConvertConfig(req *InstallRequest) (interface{}, error) {
	return nil, nil
}

type codexInstallerPlaceholder struct{}

func (p *codexInstallerPlaceholder) Install(req *InstallRequest) error {
	return nil
}

func (p *codexInstallerPlaceholder) Backup(req *InstallRequest) error {
	return nil
}

func (p *codexInstallerPlaceholder) GetConfigPath() (string, error) {
	return "", nil
}

func (p *codexInstallerPlaceholder) ValidateConfig(path string) error {
	return nil
}

func (p *codexInstallerPlaceholder) ConvertConfig(req *InstallRequest) (interface{}, error) {
	return nil, nil
}
