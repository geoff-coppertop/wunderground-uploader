package transmogrifier

import (
	"testing"
)

func TestFormatFieldSuccess(t *testing.T) {
	// _, _, err := formatField("winddirection", 65.0)
	_, _, err := formatField("windspeedaverage", 0.5)

	if err != nil {
		t.Errorf("%w", err)
	}
}

func TestFormatFieldTransmogrify(t *testing.T) {
	_, _, err := formatField("temperature", 20.5)

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
