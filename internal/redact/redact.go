package redact

import "regexp"

var sensitivePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(password|passwd|pwd|token|secret|api[-_]?key)\s*[:=]\s*([^\s,;]+)`),
	regexp.MustCompile(`(?i)(--password|--passwd|--token|--secret)\s+([^\s]+)`),
	regexp.MustCompile(`(?i)(-p)\s+([^\s]+)`),
}

// SanitizeText masks common secret patterns in logs/output.
func SanitizeText(s string) string {
	out := s
	for _, re := range sensitivePatterns {
		out = re.ReplaceAllString(out, `$1=***`)
	}
	return out
}
