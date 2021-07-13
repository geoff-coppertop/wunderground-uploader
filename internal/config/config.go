package config

import (
	"fmt"
	// "net/url"
	"os"
	// "strconv"
	// "strings"
	// "time"

	log "github.com/sirupsen/logrus"
)

// Retrieve config from environmental variables

// Configuration will be pulled from the environment using the following keys
const (
	envDebug = "DEBUG" // if "true" then the libraries will be instructed to print debug info
)

// config holds the configuration
type Config struct {
	Debug log.Level
}

// GetConfig - Retrieves the configuration from the environment
func GetConfig() (Config, error) {
	var cfg Config
	var err error
	var debug_level string

	if debug_level, err = stringFromEnv(envDebug); err != nil {
		return Config{}, err
	}

	if cfg.Debug, err = log.ParseLevel(debug_level); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// stringFromEnv - Retrieves a string from the environment and ensures it is not blank (ort non-existent)
func stringFromEnv(key string) (string, error) {
	s := os.Getenv(key)
	if len(s) == 0 {
		return "", fmt.Errorf("environmental variable %s must not be blank", key)
	}
	return s, nil
}
