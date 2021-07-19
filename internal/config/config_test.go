package config

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func SetValidTestConfig() {
	os.Setenv("DEBUG", "info")

	os.Setenv("SERVER_URL", "mqtt://example.com:1883")
	os.Setenv("TOPIC", "weather")
	os.Setenv("KA_TIME", "10")
	os.Setenv("CRD_TIME", "100")

	os.Setenv("STATION_ID", "ID")
	os.Setenv("STATION_KEY", "KEY")
}

func TestGetConfigNoEnv(t *testing.T) {
	_, err := GetConfig()

	if err == nil {
		t.Errorf("expected error")
	}
}
func TestGetConfigValid(t *testing.T) {
	SetValidTestConfig()

	cfg, err := GetConfig()

	if err != nil {
		t.Errorf("Unexpected error, got %v", err)
	}

	if cfg.Debug != logrus.InfoLevel {
		t.Errorf("Expected info debug level, got %v", cfg.Debug)
	}
}

func TestGetConfigInvalidValues(t *testing.T) {
	var tests = []struct {
		key   string
		value string
	}{
		{"TOPIC", ""},
		{"DEBUG", ""},
		{"DEBUG", "banana"},
		{"SERVER_URL", ""},
		{"KA_TIME", ""},
		{"KA_TIME", "a"},
		{"CRD_TIME", ""},
	}

	for _, test := range tests {
		SetValidTestConfig()
		os.Setenv(test.key, test.value)

		if _, err := GetConfig(); err == nil {
			t.Error("Test Failed: {} inputted, expected error.", test)
		}
	}
}
