package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

type TestSetup struct {
	T       *testing.T
	TmpDir  string
	Config  string
	Env     map[string]string
}

func NewTestSetup(t *testing.T, config string) *TestSetup {
	tmpDir := t.TempDir()

	// Write config
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	return &TestSetup{
		T:      t,
		TmpDir: tmpDir,
		Config: config,
		Env:    make(map[string]string),
	}
}

func (s *TestSetup) SetEnv(key, value string) {
	s.Env[key] = value
}

func (s *TestSetup) Run(args ...string) *Result {
	// Build aim binary from cmd/aim
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(s.TmpDir, "aim"), "./cmd/aim")
	buildCmd.Dir = "/Users/dylan/code/aim"
	if out, err := buildCmd.CombinedOutput(); err != nil {
		s.T.Fatalf("Failed to build aim: %v\n%s", err, out)
	}

	// Run aim command
	cmd := exec.Command(filepath.Join(s.TmpDir, "aim"), args...)
	cmd.Env = os.Environ()

	// Set config path
	cmd.Env = append(cmd.Env, "AIM_CONFIG="+filepath.Join(s.TmpDir, "config.yaml"))

	// Add custom env vars
	for k, v := range s.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	out, err := cmd.CombinedOutput()

	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	}

	return &Result{
		ExitCode: exitCode,
		Stdout:   string(out),
		Stderr:   string(out),
	}
}

type Result struct {
	ExitCode int
	Stdout   string
	Stderr   string
}
