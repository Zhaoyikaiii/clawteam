// Command im-core is the IM Core service entry point
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Zhaoyikaiii/clawteam/internal/config"
	"github.com/Zhaoyikaiii/clawteam/internal/im"
)

func main() {
	// Load configuration
	cfg := config.MustLoad()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Starting IM Core Service")
	logger.Info("API address", "addr", cfg.GetAPIAddr())

	// Initialize components
	eventBus := im.NewInMemoryEventBus()
	presence := im.NewPresenceManager()
	repo := im.NewInMemoryMessageRepository()
	msgService := im.NewMessageService(repo, eventBus)
	api := im.NewAPI(msgService, presence, eventBus, logger)

	// Create context for graceful shutdown
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- api.Serve(cfg.GetAPIAddr())
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		logger.Info("Shutting down gracefully...")
		cancel()
	case err := <-serverErr:
		if err != nil {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}

	logger.Info("IM Core Service stopped")
}
