package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kumahq/kuma/app/kuma-cni-install/cmd"
)

func main() {
	// Create context that cancels on termination signal
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func(sigChan chan os.Signal, cancel context.CancelFunc) {
		sig := <-sigChan
		log.Printf("Exit signal received: %s", sig)
		cancel()
	}(sigChan, cancel)

	rootCmd := cmd.GetCommand()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
