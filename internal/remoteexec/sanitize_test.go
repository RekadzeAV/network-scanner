package remoteexec

import (
	"strings"
	"testing"
)

func TestSanitizeText_MasksSecretTokens(t *testing.T) {
	input := "password=abc123 token:xyz --password qwerty -p hidden apiKey=key123"
	got := SanitizeText(input)
	if strings.Contains(got, "abc123") || strings.Contains(got, "xyz") || strings.Contains(got, "qwerty") || strings.Contains(got, "hidden") || strings.Contains(got, "key123") {
		t.Fatalf("expected secrets to be masked, got: %s", got)
	}
	if !strings.Contains(got, "***") {
		t.Fatalf("expected masked markers, got: %s", got)
	}
}
