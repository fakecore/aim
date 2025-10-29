package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fakecore/aim/internal/cmd"
	"github.com/fakecore/aim/internal/config"
)

func main() {
	// Initialize configuration at startup
	if err := config.GetConfigManager().Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Set up exit handlers for graceful shutdown
	setupExitHandlers()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// setupExitHandlers sets up signal handlers for graceful shutdown
func setupExitHandlers() {
	// Handle normal exit
	defer func() {
		if err := config.GetConfigManager().Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to save configuration: %v\n", err)
		}
	}()

	// Handle interrupt signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		// Save configuration on interrupt
		if err := config.GetConfigManager().Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to save configuration: %v\n", err)
		}
		os.Exit(0)
	}()
}
