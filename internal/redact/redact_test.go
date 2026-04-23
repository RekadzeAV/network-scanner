package redact

import (
	"strings"
	"testing"
)

func TestSanitizeText_MasksSecrets(t *testing.T) {
	input := "password=abc123 token:xyz --password qwerty -p hidden apiKey=key123"
	got := SanitizeText(input)
	for _, secret := range []string{"abc123", "xyz", "qwerty", "hidden", "key123"} {
		if strings.Contains(got, secret) {
			t.Fatalf("secret leaked in output: %s", got)
		}
	}
	if !strings.Contains(got, "***") {
		t.Fatalf("expected masking marker: %s", got)
	}
}
