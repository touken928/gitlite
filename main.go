package main

import (
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/touken928/gitlite/internal/config"
	"github.com/touken928/gitlite/internal/logging"
	"github.com/touken928/gitlite/internal/server"
)

func main() {
	// Initialize logger
	logging.Init()
	defer logging.Sync()

	// Load configuration
	cfg := config.Load()

	srv, err := server.New(cfg)
	if err != nil {
		logging.Get().Fatal("Failed to create server", zap.Error(err))
	}

	go func() {
		if err := srv.Start(); err != nil {
			logging.Get().Fatal("Failed to start server", zap.Error(err))
		}
	}()

	logging.Get().Info("Git server started, listening on port", zap.String("port", cfg.Port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logging.Get().Info("Shutting down server...")
	srv.Stop()
}
