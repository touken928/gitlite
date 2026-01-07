package config

import (
	"os"
	"strconv"
)

// Config holds application configuration loaded from environment variables
type Config struct {
	Port     string // SSH server port
	DataPath string // Base directory for data storage
}

// Get retrieves an environment variable value, returning a default if not set
func Get(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

// GetInt retrieves an environment variable as an integer, returning a default if not set or invalid
func GetInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultValue
}

// Load creates a Config instance with values from environment variables
func Load() *Config {
	return &Config{
		Port:     Get("GITLITE_PORT", "2222"),
		DataPath: Get("GITLITE_DATA", "data"),
	}
}
