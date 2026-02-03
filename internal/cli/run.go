package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/fakecore/aim/internal/config"
	"github.com/fakecore/aim/internal/errors"
	"github.com/fakecore/aim/internal/tools"
	"github.com/spf13/cobra"
)

var (
	accountName string
	timeout     string
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
	runCmd.Flags().StringVar(&timeout, "timeout", "", "Command timeout (e.g., 5m, 1h)")
	runCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be executed without running")
	runCmd.Flags().BoolVar(&native, "native", false, "Run tool without env injection")
	runCmd.Flags().StringVarP(&model, "model", "m", "", "Model to use (e.g., claude-3-opus-20240229, gpt-4o)")
}

func run(cmd *cobra.Command, args []string) error {
	toolName := args[0]

	// Resolve tool
	tool, err := tools.Resolve(toolName)
	if err != nil {
		return errors.Wrap(errors.ErrToolNotFound, toolName)
	}

	// Load config
	cfg, err := config.Load(config.ConfigPath())
	if err != nil {
		return err
	}

	// Determine account
	if accountName == "" {
		accountName, err = cfg.GetDefaultAccount()
		if err != nil {
			return err
		}
	}

	// Resolve account
	resolved, err := cfg.ResolveAccount(accountName, tool.Name, tool.Protocol)
	if err != nil {
		return err
	}

	// Get timeout
	timeoutDuration := cfg.Options.CommandTimeout
	if timeout != "" {
		timeoutDuration = timeout
	}
	duration, err := time.ParseDuration(timeoutDuration)
	if err != nil {
		return fmt.Errorf("invalid timeout: %w", err)
	}

	// Get remaining args (after --)
	toolArgs := cmd.Flags().Args()

	// Dry run mode
	if dryRun {
		printDryRun(tool, resolved, duration, toolArgs)
		return nil
	}

	// Execute
	return execute(tool, resolved, duration, toolArgs, native)
}

func execute(tool *tools.Tool, acc *config.ResolvedAccount, timeout time.Duration, args []string, native bool) error {
	// Build env vars
	env := os.Environ()

	if !native {
		// Inject API key (all tools need this)
		if apiKeyVar, ok := tool.EnvVars["api_key"]; ok {
			env = append(env, fmt.Sprintf("%s=%s", apiKeyVar, acc.Key))
		}

		// Inject base URL (if tool supports it)
		if baseURLVar, ok := tool.EnvVars["base_url"]; ok {
			env = append(env, fmt.Sprintf("%s=%s", baseURLVar, acc.ProtocolURL))
		}

		// Inject model
		// Priority: 1. User specified (-m flag), 2. Vendor default model (for openai), 3. Skip
		// For anthropic protocol, most vendors auto-map model names, so we don't inject default
		modelToUse := model
		if modelToUse == "" && acc.Protocol == "openai" && acc.Model != "" {
			modelToUse = acc.Model
		}
		if modelToUse != "" {
			if modelVar, ok := tool.EnvVars["model"]; ok {
				env = append(env, fmt.Sprintf("%s=%s", modelVar, modelToUse))
			}
		}
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(ctx, tool.Command, args...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Create new process group for signal forwarding
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Start command
	if err := cmd.Start(); err != nil {
		if os.IsNotExist(err) {
			return errors.Wrap(errors.ErrCommandNotFound, tool.Command)
		}
		return errors.WrapWithCause(errors.ErrToolNotFound, err, tool.Name)
	}

	// Forward signals to process group
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range sigChan {
			syscall.Kill(-cmd.Process.Pid, sig.(syscall.Signal))
		}
	}()

	// Wait for completion
	err := cmd.Wait()
	signal.Stop(sigChan)
	close(sigChan)

	// Check timeout
	if ctx.Err() == context.DeadlineExceeded {
		return errors.Wrap(errors.ErrExecutionTimeout, timeout)
	}

	return err
}

func printDryRun(tool *tools.Tool, acc *config.ResolvedAccount, timeout time.Duration, args []string) {
	fmt.Printf("Tool: %s (command: %s)\n", tool.Name, tool.Command)
	fmt.Printf("Account: %s\n", acc.Name)
	fmt.Printf("Key: %s...\n", acc.Key[:min(len(acc.Key), 8)])
	fmt.Printf("Protocol: %s\n", acc.Protocol)
	fmt.Printf("URL: %s\n", acc.ProtocolURL)
	// Show model info
	// For anthropic protocol, model is auto-mapped by vendor endpoint (not injected)
	// For openai protocol, model is injected if specified by user or vendor has default
	modelToShow := model
	if modelToShow == "" && acc.Protocol == "openai" && acc.Model != "" {
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
	if modelToShow == "" && acc.Protocol == "anthropic" {
		fmt.Printf("Model: auto-mapped by %s endpoint\n", acc.Vendor)
	}
	fmt.Printf("Timeout: %s\n", timeout)
	fmt.Println()
	fmt.Println("Environment:")
	// Show all configured env vars
	if apiKeyVar, ok := tool.EnvVars["api_key"]; ok {
		fmt.Printf("  %s=%s...\n", apiKeyVar, acc.Key[:min(len(acc.Key), 8)])
	}
	if baseURLVar, ok := tool.EnvVars["base_url"]; ok {
		fmt.Printf("  %s=%s\n", baseURLVar, acc.ProtocolURL)
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

