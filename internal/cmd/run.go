package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fakecore/aim/internal/config"
	"github.com/fakecore/aim/internal/tool"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <tool> --key <key-name> [--provider <provider>] [flags] [-- tool-args]",
	Short: "Run a tool with specific configuration",
	Long: `Run a tool with a specific key and provider using the v2.0 simplified configuration.

Examples:
  # Run with default provider from key
  aim run claude-code --key deepseek-work

  # Override provider
  aim run claude-code --key deepseek-work --provider glm

  # Use default key from settings
  aim run claude-code

  # Pass additional arguments to the tool
  aim run claude-code --key deepseek-work -- --help

  # Pass CLI-specific arguments (space-separated in one flag)
  aim run claude-code --key glm-coding --cli-args "--dangerously-skip-permissions"

  # Pass multiple CLI-specific arguments (space-separated)
  aim run claude-code --key glm-coding --cli-args "-arg1 -arg2 -arg3"

  # Or use multiple flags (each will be split on spaces)
  aim run claude-code --key glm-coding --cli-args "-arg1" --cli-args "-arg2"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runRun,
}

func init() {
	// Flags for run command
	runCmd.Flags().String("key", "", "Key name to use (required unless default is set)")
	runCmd.Flags().String("provider", "", "Provider to use (overrides key's default provider)")
	runCmd.Flags().String("model", "", "Model to use (overrides configuration)")
	runCmd.Flags().Int("timeout", 0, "Timeout in milliseconds (overrides configuration)")
	runCmd.Flags().StringSlice("cli-args", []string{}, "Additional arguments to pass to the CLI tool")
	runCmd.Flags().Bool("native", false, "Use tool's native configuration (no env vars)")
}

func runRun(cmd *cobra.Command, args []string) error {
	toolName := args[0]
	keyName, _ := cmd.Flags().GetString("key")
	providerName, _ := cmd.Flags().GetString("provider")
	modelName, _ := cmd.Flags().GetString("model")
	timeout, _ := cmd.Flags().GetInt("timeout")
	additionalArgs, _ := cmd.Flags().GetStringSlice("cli-args")
	nativeMode, _ := cmd.Flags().GetBool("native")

	// Get canonical name (handle aliases like cc -> claude-code)
	canonicalToolName := tool.GetCanonicalName(toolName)

	// Check if tool is supported
	if !tool.IsToolSupported(canonicalToolName) {
		return fmt.Errorf("unsupported tool: %s. Currently supported tools: [codex claude-code (cc)]", toolName)
	}

	// If native mode is enabled, skip all configuration
	if nativeMode {
		return runNative(cmd, args, canonicalToolName)
	}

	// Get global configuration manager
	cm := config.GetConfigManager()
	cfg := cm.GetConfig()

	// Create resolver
	resolver := config.NewResolver(cfg)

	// Validate tool
	if err := resolver.ValidateTool(canonicalToolName); err != nil {
		return fmt.Errorf("invalid tool: %w", err)
	}

	// If no key specified, use default
	if keyName == "" {
		keyName = cfg.Settings.DefaultKey
		if keyName == "" {
			return fmt.Errorf("no key specified. Use --key <key-name> or set default key with 'aim config set default-key <key-name>'")
		}
	}

	// Validate key
	if err := resolver.ValidateKey(keyName); err != nil {
		return fmt.Errorf("invalid key: %w", err)
	}

	// Resolve runtime configuration first to get the provider
	runtime, err := resolver.Resolve(canonicalToolName, keyName, providerName)
	if err != nil {
		return fmt.Errorf("failed to resolve configuration: %w", err)
	}

	// Initialize environment preparer manager
	preparerManager := tool.NewEnvironmentPreparerManager()

	// Validate tool environment configuration with resolved provider
	if err := preparerManager.ValidateEnvironment(canonicalToolName, runtime.Provider); err != nil {
		return fmt.Errorf("tool environment validation failed: %w", err)
	}

	// Apply command line overrides
	if modelName != "" {
		runtime.Model = modelName
		runtime.ModelOverride = true
	}
	if timeout > 0 {
		runtime.Timeout = time.Duration(timeout) * time.Millisecond
	}

	// Rebuild env vars after overrides so model/timeout changes are reflected.
	if err := resolver.UpdateRuntimeEnvVars(runtime); err != nil {
		return fmt.Errorf("failed to update runtime env vars: %w", err)
	}

	// Claude Code: omit model env var only if model is not explicitly set in profile.
	// If profile has model: "-", we don't pass ANTHROPIC_MODEL.
	// If profile has model: "xxx", we pass ANTHROPIC_MODEL="xxx".
	if canonicalToolName == string(tool.ToolTypeClaudeCode) {
		toolConfig, _ := cfg.GetTool(canonicalToolName)
		toolProfile, _ := cfg.GetToolProfile(canonicalToolName, runtime.Profile)

		// Only remove model env vars if:
		// 1. Model is empty (not set), AND
		// 2. Not explicitly overridden via CLI
		if runtime.Model == "" && !runtime.ModelOverride {
			removeModelEnvVars(toolConfig, toolProfile, runtime.Profile, runtime.EnvVars)
		}
	}

	// Prepare tool-specific environment and arguments
	toolConfigArgs, toolEnvVars, err := preparerManager.PrepareEnvironment(runtime)
	if err != nil {
		return fmt.Errorf("failed to prepare tool environment: %w", err)
	}

	// Merge tool-specific environment variables with runtime environment variables
	for key, value := range toolEnvVars {
		runtime.EnvVars[key] = value
	}

	// Get tool command
	toolConfig, _ := cfg.GetTool(canonicalToolName)

	// Find real binary
	realBinary, err := findRealBinary(toolConfig.Command)
	if err != nil {
		return fmt.Errorf("failed to find binary '%s': %w", toolConfig.Command, err)
	}

	// Extract tool arguments after --
	var toolArgs []string
	if len(os.Args) > 0 {
		// Find the -- separator
		for i, arg := range os.Args {
			if arg == "--" && i+1 < len(os.Args) {
				toolArgs = os.Args[i+1:]
				break
			}
		}
	}

	// Prepend tool-specific config to tool args
	if len(toolConfigArgs) > 0 {
		toolArgs = append(toolConfigArgs, toolArgs...)
	}

	// Add additional arguments from --cli-args flag
	// Split each value on spaces to support space-separated arguments
	if len(additionalArgs) > 0 {
		for _, argGroup := range additionalArgs {
			// Split on whitespace and append non-empty parts
			parts := splitWhitespace(argGroup)
			toolArgs = append(toolArgs, parts...)
		}
	}

	// Show what we're running
	if verbose {
		fmt.Fprintf(os.Stderr, "Running: %s (canonical: %s) with key=%s, provider=%s, profile=%s, model=%s\n",
			toolName, canonicalToolName, keyName, runtime.Provider, runtime.Profile, runtime.Model)
	}

	// Execute with environment
	return execWithEnv(realBinary, toolArgs, runtime.EnvVars)
}

func removeModelEnvVars(toolConfig *config.ToolConfig, toolProfile *config.ToolProfile, profileName string, envVars map[string]string) {
	for _, envKey := range modelEnvKeys(toolConfig, toolProfile, profileName) {
		delete(envVars, envKey)
	}
}

func modelEnvKeys(toolConfig *config.ToolConfig, toolProfile *config.ToolProfile, profileName string) []string {
	var keys []string

	addKeys := func(mapping map[string]string) {
		for envKey, fieldPath := range mapping {
			if isModelFieldPath(fieldPath, profileName) {
				keys = append(keys, envKey)
			}
		}
	}

	if toolProfile != nil {
		addKeys(toolProfile.FieldMapping)
	}
	if toolConfig != nil {
		addKeys(toolConfig.FieldMapping)
	}

	return keys
}

func isModelFieldPath(fieldPath, profileName string) bool {
	if fieldPath == "" {
		return false
	}
	if fieldPath == "profiles.{current_profile}.model" {
		return true
	}
	if profileName != "" && fieldPath == fmt.Sprintf("profiles.%s.model", profileName) {
		return true
	}
	return false
}

// findRealBinary finds real binary in PATH (excluding ~/.aim/bin and AIM_HOME/bin)
func findRealBinary(name string) (string, error) {
	// Search in PATH, excluding ~/.aim/bin and AIM_HOME/bin
	pathEnv := os.Getenv("PATH")
	homeDir, _ := os.UserHomeDir()
	aimBin := filepath.Join(homeDir, ".aim", "bin")

	// Also exclude AIM_HOME/bin to avoid infinite loop in test environments
	aimHomeBin := ""
	if aimHome := os.Getenv("AIM_HOME"); aimHome != "" {
		aimHomeBin = filepath.Join(aimHome, "bin")
	}

	for _, dir := range filepath.SplitList(pathEnv) {
		// Skip ~/.aim/bin to avoid infinite loop
		if dir == aimBin {
			continue
		}

		// Skip AIM_HOME/bin in test environments
		if aimHomeBin != "" && dir == aimHomeBin {
			continue
		}

		binPath := filepath.Join(dir, name)
		if _, err := os.Stat(binPath); err == nil {
			// Check if executable
			if isExecutable(binPath) {
				return binPath, nil
			}
		}
	}

	return "", fmt.Errorf("binary '%s' not found in PATH", name)
}

// isExecutable checks if a file is executable
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	mode := info.Mode()
	return !mode.IsDir() && mode&0111 != 0
}

// runNative runs a tool in native mode (no environment variables)
func runNative(cmd *cobra.Command, args []string, canonicalToolName string) error {
	toolName := args[0]

	// Check for conflicting flags
	if cmd.Flags().Changed("key") {
		fmt.Fprintf(os.Stderr, "Warning: --native flag specified, ignoring --key\n")
	}
	if cmd.Flags().Changed("provider") {
		fmt.Fprintf(os.Stderr, "Warning: --native flag specified, ignoring --provider\n")
	}
	if cmd.Flags().Changed("model") {
		fmt.Fprintf(os.Stderr, "Warning: --native flag specified, ignoring --model\n")
	}
	if cmd.Flags().Changed("timeout") {
		fmt.Fprintf(os.Stderr, "Warning: --native flag specified, ignoring --timeout\n")
	}

	// Check if tool is supported
	if !tool.IsToolSupported(canonicalToolName) {
		return fmt.Errorf("unsupported tool: %s. Currently supported tools: [codex claude-code (cc)]", toolName)
	}

	// Get tool command
	cm := config.GetConfigManager()
	cfg := cm.GetConfig()
	toolConfig, ok := cfg.GetTool(canonicalToolName)
	if !ok {
		return fmt.Errorf("tool '%s' not configured", canonicalToolName)
	}

	// Find real binary
	realBinary, err := findRealBinary(toolConfig.Command)
	if err != nil {
		return fmt.Errorf("failed to find binary '%s': %w", toolConfig.Command, err)
	}

	// Extract tool arguments after --
	var toolArgs []string
	if len(os.Args) > 0 {
		// Find the -- separator
		for i, arg := range os.Args {
			if arg == "--" && i+1 < len(os.Args) {
				toolArgs = os.Args[i+1:]
				break
			}
		}
	}

	// Get additional cli-args
	additionalArgs, _ := cmd.Flags().GetStringSlice("cli-args")
	if len(additionalArgs) > 0 {
		for _, argGroup := range additionalArgs {
			// Split on whitespace and append non-empty parts
			parts := splitWhitespace(argGroup)
			toolArgs = append(toolArgs, parts...)
		}
	}

	// Show what we're running
	if verbose {
		fmt.Fprintf(os.Stderr, "Running %s in native mode (no env vars)\n", canonicalToolName)
	}

	// Execute with NO environment variables
	execCmd := exec.Command(realBinary, toolArgs...)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}

// execWithEnv executes a command with environment variables
func execWithEnv(binary string, args []string, envVars map[string]string) error {
	// Debug output
	if verbose {
		fmt.Fprintf(os.Stderr, "Executing binary: %s with args: %v\n", binary, args)
		fmt.Fprintf(os.Stderr, "Environment variables:\n")
		for key, value := range envVars {
			fmt.Fprintf(os.Stderr, "  %s=%s\n", key, value)
		}
	}

	cmd := exec.Command(binary, args...)

	// Copy current environment
	cmd.Env = os.Environ()

	// Add/override with our environment variables
	for key, value := range envVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Set stdio
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute
	return cmd.Run()
}

// splitWhitespace splits a string on whitespace and returns non-empty parts
func splitWhitespace(s string) []string {
	var result []string
	for _, part := range strings.Fields(s) {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
