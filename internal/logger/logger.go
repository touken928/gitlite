package logger

import "go.uber.org/zap"

var log *zap.Logger

// Init initializes the global logger with production configuration
func Init() {
	var err error
	log, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
}

// InitWithConfig initializes the logger with custom configuration
func InitWithConfig(cfg zap.Config) {
	var err error
	log, err = cfg.Build()
	if err != nil {
		panic(err)
	}
}

// Get returns the global logger instance
func Get() *zap.Logger {
	return log
}

// Sync flushes any buffered log entries
func Sync() {
	if log != nil {
		log.Sync()
	}
}
