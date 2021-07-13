package config

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestGetConfigNoEnv(t *testing.T) {
	_, err := GetConfig()

	if err == nil {
		t.Errorf("expected error")
	}
}

func TestGetConfigInvalidLevel(t *testing.T) {
	os.Setenv("DEBUG", "banana")

	_, err := GetConfig()

	if err == nil {
		t.Errorf("expected error")
	}
}
func TestGetConfigValid(t *testing.T) {
	os.Setenv("DEBUG", "info")

	cfg, err := GetConfig()

	if err != nil {
		t.Errorf("Unexpected error, got %v", err)
	}

	if cfg.Debug != logrus.InfoLevel {
		t.Errorf("Expected info debug level, got %v", cfg.Debug)
	}
}
