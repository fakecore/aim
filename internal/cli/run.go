package cli

import (
	"fmt"
	"os"
	"syscall"

	"github.com/fakecore/aim/internal/config"
	"github.com/fakecore/aim/internal/errors"
	"github.com/fakecore/aim/internal/tools"
	"github.com/spf13/cobra"
)

var (
	accountName string
	dryRun      bool
	native      bool
	model       string
)

var runCmd = &cobra.Command{
	Use:   "run <tool> [-- <args>...]",
	Short: "Run an AI tool with the specified account",
	Long:  `Run an AI tool (claude-code, codex, etc.) with environment configured for the specified account.

Use -- to pass arguments to the tool:
  aim run cc -- -h           # Show claude-code help
  aim run codex -- file.txt  # Run codex on a file`,
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: false,
	RunE:               run,
}

func init() {
	runCmd.Flags().StringVarP(&accountName, "account", "a", "", "Account to use")
	runCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be executed without running")
	runCmd.Flags().BoolVar(&native, "native", false, "Run tool without env injection")
	runCmd.Flags().StringVarP(&model, "model", "m", "", "Model to use (e.g., claude-3-opus-20240229, gpt-4o)")
}

func run(cmd *cobra.Command, args []string) error {
	toolName := args[0]

	// Load config first (needed for tool protocol resolution)
	cfg, cfgErr := config.Load(config.ConfigPath())

	// Resolve tool with config (to get protocol from config)
	tool, err := tools.ResolveWithConfig(toolName, cfg)
	if err != nil {
		return errors.Wrap(errors.ErrToolNotFound, toolName)
	}

	// Get remaining args (after --)
	toolArgs := args[1:]

	// If native mode, skip account resolution
	if native {
		if dryRun {
			printDryRunNative(tool, toolArgs)
			return nil
		}
		return execute(tool, nil, toolArgs, true)
	}

	// Try to resolve account
	var resolved *config.ResolvedAccount
	if cfgErr == nil {
		// Determine account
		accName := accountName
		if accName == "" {
			accName, _ = cfg.GetDefaultAccount()
		}
		if accName != "" {
			resolved, err = cfg.ResolveAccount(accName, tool.Name, tool.Protocol)
			if err != nil && accountName != "" {
				// User explicitly specified an account, so report the error
				return err
			}
		}
	} else if accountName != "" {
		// User explicitly specified an account but config failed to load
		return cfgErr
	}

	// Dry run mode
	if dryRun {
		if resolved != nil {
			printDryRun(tool, resolved, toolArgs)
		} else {
			printDryRunNative(tool, toolArgs)
		}
		return nil
	}

	// Execute - fallback to native if no account resolved
	if resolved == nil {
		// Print colored warning
		fmt.Fprintf(os.Stderr, "\033[33m⚡ No account configured, running in native mode\033[0m\n")
	}
	return execute(tool, resolved, toolArgs, resolved == nil)
}

func execute(tool *tools.Tool, acc *config.ResolvedAccount, args []string, native bool) error {
	// Build env vars
	env := os.Environ()

	if !native {
		// Inject API key (all tools need this)
		if apiKeyVar, ok := tool.EnvVars["api_key"]; ok {
			env = append(env, fmt.Sprintf("%s=%s", apiKeyVar, acc.Key))
		}

		// Inject base URL (if tool supports it)
		if baseURLVar, ok := tool.EnvVars["base_url"]; ok {
			env = append(env, fmt.Sprintf("%s=%s", baseURLVar, acc.EndpointURL))
		}

		// Inject model
		// Priority: 1. User specified (-m flag), 2. Vendor default model (for openai), 3. Skip
		// For anthropic protocol, most vendors auto-map model names, so we don't inject default
		modelToUse := model
		if modelToUse == "" && acc.Endpoint == "openai" && acc.Model != "" {
			modelToUse = acc.Model
		}
		if modelToUse != "" {
			if modelVar, ok := tool.EnvVars["model"]; ok {
				env = append(env, fmt.Sprintf("%s=%s", modelVar, modelToUse))
			}
		}
	}

	// Find executable path
	binPath, err := findExecutable(tool.Command)
	if err != nil {
		return errors.Wrap(errors.ErrCommandNotFound, tool.Command)
	}

	// Build argv (first element is the command name)
	argv := append([]string{tool.Command}, args...)

	// Replace current process with the tool
	return syscall.Exec(binPath, argv, env)
}

func findExecutable(name string) (string, error) {
	// If it's an absolute path, use it directly
	if len(name) > 0 && name[0] == '/' {
		return name, nil
	}

	// Search in PATH
	path := os.Getenv("PATH")
	for _, dir := range splitPath(path) {
		full := dir + "/" + name
		if info, err := os.Stat(full); err == nil && !info.IsDir() {
			return full, nil
		}
	}
	return "", fmt.Errorf("executable not found: %s", name)
}

func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	return splitPathSeparator(path)
}

func splitPathSeparator(path string) []string {
	var result []string
	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] == ':' {
			if i > start {
				result = append(result, path[start:i])
			}
			start = i + 1
		}
	}
	if start < len(path) {
		result = append(result, path[start:])
	}
	return result
}

func printDryRunNative(tool *tools.Tool, args []string) {
	fmt.Printf("Tool: %s (command: %s)\n", tool.Name, tool.Command)
	fmt.Println("Mode: native (no env injection)")
	fmt.Println()
	fmt.Printf("Command: %s %v\n", tool.Command, args)
}

func printDryRun(tool *tools.Tool, acc *config.ResolvedAccount, args []string) {
	fmt.Printf("Tool: %s (command: %s)\n", tool.Name, tool.Command)
	fmt.Printf("Account: %s\n", acc.Name)
	fmt.Printf("Key: %s...\n", acc.Key[:min(len(acc.Key), 8)])
	fmt.Printf("Endpoint: %s\n", acc.Endpoint)
	fmt.Printf("URL: %s\n", acc.EndpointURL)
	// Show model info
	// For anthropic protocol, model is auto-mapped by vendor endpoint (not injected)
	// For openai protocol, model is injected if specified by user or vendor has default
	modelToShow := model
	if modelToShow == "" && acc.Endpoint == "openai" && acc.Model != "" {
		modelToShow = acc.Model
	}
	if modelToShow != "" {
		if modelVar, ok := tool.EnvVars["model"]; ok {
			if model != "" {
				fmt.Printf("Model: %s (user specified via %s)\n", modelToShow, modelVar)
			} else {
				fmt.Printf("Model: %s (vendor default via %s)\n", modelToShow, modelVar)
			}
		} else {
			fmt.Printf("Model: %s (ignored - %s doesn't support model env var)\n", modelToShow, tool.Name)
		}
	}
	if modelToShow == "" && acc.Endpoint == "anthropic" {
		fmt.Printf("Model: auto-mapped by %s endpoint\n", acc.Vendor)
	}
	fmt.Println()
	fmt.Println("Environment:")
	// Show all configured env vars
	if apiKeyVar, ok := tool.EnvVars["api_key"]; ok {
		fmt.Printf("  %s=%s...\n", apiKeyVar, acc.Key[:min(len(acc.Key), 8)])
	}
	if baseURLVar, ok := tool.EnvVars["base_url"]; ok {
		fmt.Printf("  %s=%s\n", baseURLVar, acc.EndpointURL)
	}
	if modelToShow != "" {
		if modelVar, ok := tool.EnvVars["model"]; ok {
			fmt.Printf("  %s=%s\n", modelVar, modelToShow)
		}
	}
	fmt.Println()
	fmt.Printf("Command: %s %v\n", tool.Command, args)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
