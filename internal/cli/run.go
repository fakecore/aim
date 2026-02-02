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
)

var runCmd = &cobra.Command{
	Use:   "run <tool>",
	Short: "Run an AI tool with the specified account",
	Long:  `Run an AI tool (claude-code, codex, etc.) with environment configured for the specified account.`,
	Args:  cobra.ExactArgs(1),
	RunE:  run,
}

func init() {
	runCmd.Flags().StringVarP(&accountName, "account", "a", "", "Account to use")
	runCmd.Flags().StringVar(&timeout, "timeout", "", "Command timeout (e.g., 5m, 1h)")
	runCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be executed without running")
	runCmd.Flags().BoolVar(&native, "native", false, "Run tool without env injection")
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
		switch tool.Protocol {
		case "anthropic":
			env = append(env, fmt.Sprintf("ANTHROPIC_AUTH_TOKEN=%s", acc.Key))
			env = append(env, fmt.Sprintf("ANTHROPIC_BASE_URL=%s", acc.ProtocolURL))
		case "openai":
			env = append(env, fmt.Sprintf("OPENAI_API_KEY=%s", acc.Key))
			env = append(env, fmt.Sprintf("OPENAI_BASE_URL=%s", acc.ProtocolURL))
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
	fmt.Printf("Timeout: %s\n", timeout)
	fmt.Println()
	fmt.Println("Environment:")
	switch tool.Protocol {
	case "anthropic":
		fmt.Printf("  ANTHROPIC_AUTH_TOKEN=%s...\n", acc.Key[:min(len(acc.Key), 8)])
		fmt.Printf("  ANTHROPIC_BASE_URL=%s\n", acc.ProtocolURL)
	case "openai":
		fmt.Printf("  OPENAI_API_KEY=%s...\n", acc.Key[:min(len(acc.Key), 8)])
		fmt.Printf("  OPENAI_BASE_URL=%s\n", acc.ProtocolURL)
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
