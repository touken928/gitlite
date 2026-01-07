package main

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"go.uber.org/zap"

	"github.com/touken928/gitlite/internal/logger"
	"github.com/touken928/gitlite/internal/server"
)

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultValue
}

func main() {
	// Initialize logger
	logger.Init()
	defer logger.Sync()

	// Read config from environment variables
	port := getEnv("GITLITE_PORT", "2222")
	dataPath := getEnv("GITLITE_DATA", "data")

	srv, err := server.New(port, dataPath)
	if err != nil {
		logger.Get().Fatal("Failed to create server", zap.Error(err))
	}

	go func() {
		if err := srv.Start(); err != nil {
			logger.Get().Fatal("Failed to start server", zap.Error(err))
		}
	}()

	logger.Get().Info("Git server started, listening on port", zap.String("port", port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Get().Info("Shutting down server...")
	srv.Stop()
}
