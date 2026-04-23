package remoteexec

import "network-scanner/internal/redact"

// SanitizeText masks common secret patterns in logs/output.
func SanitizeText(s string) string {
	return redact.SanitizeText(s)
}
