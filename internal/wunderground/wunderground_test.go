package wunderground

import (
	"testing"
)

func TestBuildRequestSuccess(t *testing.T) {
	data := map[string]string{
		"winddirection": "65.0",
	}

	if _, err := buildRequestString(data, "id", "password"); err != nil {
		t.Errorf("Unexpected error")
	}
}

func TestBuildRequestEmptyData(t *testing.T) {
	data := make(map[string]string)

	if _, err := buildRequestString(data, "id", "password"); err == nil {
		t.Errorf("expected error")
	}
}
