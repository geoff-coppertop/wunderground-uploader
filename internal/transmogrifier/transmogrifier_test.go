package transmogrifier

import (
	"testing"
)

func TestFormatFieldSuccess(t *testing.T) {
	_, _, err := formatField("wspd", 0.5)

	if err != nil {
		t.Errorf("%w", err)
	}
}

func TestFormatFieldTransmogrify(t *testing.T) {
	_, _, err := formatField("temp", 20.5)

	if err != nil {
		t.Errorf("Unexpected error")
	}
}

func TestFormatInvalidField(t *testing.T) {
	_, _, err := formatField("banana", 65.0)

	if err == nil {
		t.Errorf("expected error")
	}
}

func TestFormatInvalidFieldValue(t *testing.T) {
	_, _, err := formatField("winddir", 65)

	if err == nil {
		t.Errorf("expected error")
	}
}
