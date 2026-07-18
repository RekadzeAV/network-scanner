package banner

import (
	"testing"
)

// ============================================================================
// sanitizeBanner — ветки: printable, whitespace, mixed, empty, all non-printable
// ============================================================================

func TestSanitizeBanner_Printable(t *testing.T) {
	got := sanitizeBanner([]byte("Hello World"))
	if got != "Hello World" {
		t.Fatalf("sanitizeBanner(\"Hello World\") = %q, want %q", got, "Hello World")
	}
}

func TestSanitizeBanner_Whitespace(t *testing.T) {
	got := sanitizeBanner([]byte("a\nb\tc\rd"))
	if got != "a b c d" {
		t.Fatalf("sanitizeBanner with whitespace = %q, want %q", got, "a b c d")
	}
}

func TestSanitizeBanner_Mixed(t *testing.T) {
	got := sanitizeBanner([]byte("HTTP/1.1 200 OK\r\nServer: nginx\n"))
	if got != "HTTP/1.1 200 OK  Server: nginx" {
		t.Fatalf("sanitizeBanner mixed = %q, want %q", got, "HTTP/1.1 200 OK  Server: nginx")
	}
}

func TestSanitizeBanner_Empty(t *testing.T) {
	got := sanitizeBanner([]byte(""))
	if got != "" {
		t.Fatalf("sanitizeBanner(empty) = %q, want %q", got, "")
	}
}

func TestSanitizeBanner_AllNonPrintable(t *testing.T) {
	got := sanitizeBanner([]byte{0, 1, 2, 3, 4, 5})
	if got != "" {
		t.Fatalf("sanitizeBanner(non-printable) = %q, want empty", got)
	}
}

func TestSanitizeBanner_BoundaryPrintable(t *testing.T) {
	got := sanitizeBanner([]byte{31, 32, 65, 90, 126, 127})
	// 31 — non-printable, 32 — space (trimmed by TrimSpace), 65-90 — A-Z, 126 — ~, 127 — non-printable
	want := "AZ~"
	if got != want {
		t.Fatalf("sanitizeBanner(boundary) = %q, want %q", got, want)
	}
}

// ============================================================================
// normalizeByPort — ветки: SSH(22), FTP(21), SMTP(25/587), POP3(110), IMAP(143)
// ============================================================================

func TestNormalizeByPort_SSH(t *testing.T) {
	got := normalizeByPort(22, "SSH-2.0-OpenSSH_9.3")
	if got != "SSH-2.0-OpenSSH_9.3" {
		t.Fatalf("normalizeByPort(22, SSH) = %q, want %q", got, "SSH-2.0-OpenSSH_9.3")
	}
}

func TestNormalizeByPort_SSH_NoPrefix(t *testing.T) {
	// SSH порт без SSH- префикса должен вернуть как есть
	got := normalizeByPort(22, "Some random banner")
	if got != "Some random banner" {
		t.Fatalf("normalizeByPort(22, no prefix) = %q, want %q", got, "Some random banner")
	}
}

func TestNormalizeByPort_FTP(t *testing.T) {
	got := normalizeByPort(21, "220 FileZilla Server")
	if got != "FTP 220 FileZilla Server" {
		t.Fatalf("normalizeByPort(21, FTP) = %q, want %q", got, "FTP 220 FileZilla Server")
	}
}

func TestNormalizeByPort_FTP_No220(t *testing.T) {
	// FTP порт без 220 префикса
	got := normalizeByPort(21, "Some other banner")
	if got != "Some other banner" {
		t.Fatalf("normalizeByPort(21, no 220) = %q, want %q", got, "Some other banner")
	}
}

func TestNormalizeByPort_SMTP_220(t *testing.T) {
	got := normalizeByPort(25, "220 smtp.example.com")
	if got != "SMTP 220 smtp.example.com" {
		t.Fatalf("normalizeByPort(25, SMTP 220) = %q, want %q", got, "SMTP 220 smtp.example.com")
	}
}

func TestNormalizeByPort_SMTP_prefix(t *testing.T) {
	got := normalizeByPort(587, "SMTP greeting")
	if got != "SMTP SMTP greeting" {
		t.Fatalf("normalizeByPort(587, SMTP prefix) = %q, want %q", got, "SMTP SMTP greeting")
	}
}

func TestNormalizeByPort_SMTP_NoMatch(t *testing.T) {
	got := normalizeByPort(25, "Some random banner")
	if got != "Some random banner" {
		t.Fatalf("normalizeByPort(25, no match) = %q, want %q", got, "Some random banner")
	}
}

func TestNormalizeByPort_POP3(t *testing.T) {
	got := normalizeByPort(110, "+OK Dovecot ready.")
	if got != "POP3 +OK Dovecot ready." {
		t.Fatalf("normalizeByPort(110, POP3) = %q, want %q", got, "POP3 +OK Dovecot ready.")
	}
}

func TestNormalizeByPort_POP3_NoMatch(t *testing.T) {
	got := normalizeByPort(110, "Some other banner")
	if got != "Some other banner" {
		t.Fatalf("normalizeByPort(110, no match) = %q, want %q", got, "Some other banner")
	}
}

func TestNormalizeByPort_IMAP(t *testing.T) {
	got := normalizeByPort(143, "* OK IMAP server")
	if got != "IMAP * OK IMAP server" {
		t.Fatalf("normalizeByPort(143, IMAP) = %q, want %q", got, "IMAP * OK IMAP server")
	}
}

func TestNormalizeByPort_IMAP_ContainsIMAP(t *testing.T) {
	got := normalizeByPort(143, "Some IMAP banner")
	if got != "IMAP Some IMAP banner" {
		t.Fatalf("normalizeByPort(143, contains IMAP) = %q, want %q", got, "IMAP Some IMAP banner")
	}
}

func TestNormalizeByPort_IMAP_NoMatch(t *testing.T) {
	got := normalizeByPort(143, "Some other banner")
	if got != "Some other banner" {
		t.Fatalf("normalizeByPort(143, no match) = %q, want %q", got, "Some other banner")
	}
}

func TestNormalizeByPort_Empty(t *testing.T) {
	got := normalizeByPort(22, "")
	if got != "" {
		t.Fatalf("normalizeByPort(empty) = %q, want empty", got)
	}
}

func TestNormalizeByPort_RandomPort(t *testing.T) {
	got := normalizeByPort(9999, "custom banner")
	if got != "custom banner" {
		t.Fatalf("normalizeByPort(9999) = %q, want %q", got, "custom banner")
	}
}

func TestNormalizeByPort_WhitespaceOnly(t *testing.T) {
	got := normalizeByPort(22, "   ")
	if got != "" {
		t.Fatalf("normalizeByPort(whitespace) = %q, want empty", got)
	}
}

// ============================================================================
// ExtractVersionHint — ветки: HTTP status+server, status only, server only, long
// ============================================================================

func TestExtractVersionHint_HTTP_StatusAndServer(t *testing.T) {
	got := ExtractVersionHint(80, "HTTP/1.1 200 OK | Server=nginx/1.25.0")
	want := "HTTP/1.1 200 OK (nginx/1.25.0)"
	if got != want {
		t.Fatalf("HTTP status+server = %q, want %q", got, want)
	}
}

func TestExtractVersionHint_HTTP_StatusOnly(t *testing.T) {
	got := ExtractVersionHint(80, "HTTP/1.1 200 OK")
	want := "HTTP/1.1 200 OK"
	if got != want {
		t.Fatalf("HTTP status only = %q, want %q", got, want)
	}
}

func TestExtractVersionHint_HTTP_ServerOnly(t *testing.T) {
	got := ExtractVersionHint(80, "Server=Apache/2.4.52")
	want := "Apache/2.4.52"
	if got != want {
		t.Fatalf("HTTP server only = %q, want %q", got, want)
	}
}

func TestExtractVersionHint_HTTP_ServerLowercase(t *testing.T) {
	got := ExtractVersionHint(80, "server=nginx")
	want := "nginx"
	if got != want {
		t.Fatalf("HTTP server lowercase = %q, want %q", got, want)
	}
}

func TestExtractVersionHint_HTTP_NoParts(t *testing.T) {
	got := ExtractVersionHint(80, "custom banner")
	want := "custom banner"
	if got != want {
		t.Fatalf("HTTP no parts = %q, want %q", got, want)
	}
}

func TestExtractVersionHint_HTTP_8443(t *testing.T) {
	got := ExtractVersionHint(8443, "HTTP/1.1 404 Not Found | Server=tomcat")
	want := "HTTP/1.1 404 Not Found (tomcat)"
	if got != want {
		t.Fatalf("HTTP 8443 = %q, want %q", got, want)
	}
}

func TestExtractVersionHint_HTTP_8080(t *testing.T) {
	got := ExtractVersionHint(8080, "HTTP/1.1 301 Moved | Server=haproxy")
	want := "HTTP/1.1 301 Moved (haproxy)"
	if got != want {
		t.Fatalf("HTTP 8080 = %q, want %q", got, want)
	}
}

func TestExtractVersionHint_HTTP_443(t *testing.T) {
	got := ExtractVersionHint(443, "HTTP/1.1 200 OK | Server=nginx")
	want := "HTTP/1.1 200 OK (nginx)"
	if got != want {
		t.Fatalf("HTTP 443 = %q, want %q", got, want)
	}
}

func TestExtractVersionHint_LongBanner(t *testing.T) {
	longBanner := ""
	for i := 0; i < 200; i++ {
		longBanner += "A"
	}
	got := ExtractVersionHint(80, longBanner)
	want := longBanner[:117] + "..."
	if len(got) >= len(longBanner) {
		t.Fatalf("long banner was not truncated: len(got)=%d, len(banner)=%d", len(got), len(longBanner))
	}
	if got != want {
		t.Fatalf("long banner = %q, want %q", got, want)
	}
}

func TestExtractVersionHint_FTP_TrimMailPrefix(t *testing.T) {
	got := ExtractVersionHint(21, "FTP 220 FileZilla Server 1.8.0")
	want := "FileZilla Server 1.8.0"
	if got != want {
		t.Fatalf("FTP trim mail prefix = %q, want %q", got, want)
	}
}

func TestExtractVersionHint_FTP_NoPrefix(t *testing.T) {
	got := ExtractVersionHint(21, "FTP plain banner")
	want := "plain banner"
	if got != want {
		t.Fatalf("FTP no prefix = %q, want %q", got, want)
	}
}

func TestExtractVersionHint_SMTP_TrimMailPrefix(t *testing.T) {
	got := ExtractVersionHint(25, "SMTP 220 smtp.example.com ESMTP Postfix")
	want := "smtp.example.com ESMTP Postfix"
	if got != want {
		t.Fatalf("SMTP trim mail prefix = %q, want %q", got, want)
	}
}

func TestExtractVersionHint_POP3_TrimMailPrefix(t *testing.T) {
	got := ExtractVersionHint(110, "POP3 +OK Dovecot ready.")
	want := "Dovecot ready."
	if got != want {
		t.Fatalf("POP3 trim mail prefix = %q, want %q", got, want)
	}
}

func TestExtractVersionHint_IMAP_TrimMailPrefix(t *testing.T) {
	got := ExtractVersionHint(143, "IMAP * OK IMAP server")
	want := "* OK IMAP server"
	if got != want {
		t.Fatalf("IMAP trim mail prefix = %q, want %q", got, want)
	}
}

// ============================================================================
// trimMailLikePrefix — ветки: +OK, numeric code, dash/dot prefix
// ============================================================================

func TestTrimMailLikePrefix_Plain(t *testing.T) {
	got := trimMailLikePrefix("Dovecot ready.")
	want := "Dovecot ready."
	if got != want {
		t.Fatalf("plain = %q, want %q", got, want)
	}
}

func TestTrimMailLikePrefix_PLUSOK(t *testing.T) {
	got := trimMailLikePrefix("+OK Dovecot")
	want := "Dovecot"
	if got != want {
		t.Fatalf("+OK = %q, want %q", got, want)
	}
}

func TestTrimMailLikePrefix_NumericCode(t *testing.T) {
	got := trimMailLikePrefix("220 smtp.example.com")
	want := "smtp.example.com"
	if got != want {
		t.Fatalf("numeric code = %q, want %q", got, want)
	}
}

func TestTrimMailLikePrefix_DashPrefix(t *testing.T) {
	got := trimMailLikePrefix("220 -welcome")
	want := "welcome"
	if got != want {
		t.Fatalf("dash prefix = %q, want %q", got, want)
	}
}

func TestTrimMailLikePrefix_DotPrefix(t *testing.T) {
	got := trimMailLikePrefix("220 .welcome")
	want := "welcome"
	if got != want {
		t.Fatalf("dot prefix = %q, want %q", got, want)
	}
}

func TestTrimMailLikePrefix_Whitespace(t *testing.T) {
	got := trimMailLikePrefix("   220 smtp.example.com   ")
	want := "smtp.example.com"
	if got != want {
		t.Fatalf("whitespace = %q, want %q", got, want)
	}
}

func TestTrimMailLikePrefix_Short(t *testing.T) {
	got := trimMailLikePrefix("AB")
	want := "AB"
	if got != want {
		t.Fatalf("short = %q, want %q", got, want)
	}
}

func TestTrimMailLikePrefix_Empty(t *testing.T) {
	got := trimMailLikePrefix("")
	want := ""
	if got != want {
		t.Fatalf("empty = %q, want %q", got, want)
	}
}

// ============================================================================
// isDigit — ветки: 0-9, non-digit
// ============================================================================

func TestIsDigit_Digits(t *testing.T) {
	for i := byte('0'); i <= '9'; i++ {
		if !isDigit(i) {
			t.Fatalf("isDigit(%c) = false, want true", i)
		}
	}
}

func TestIsDigit_NonDigits(t *testing.T) {
	if isDigit('a') {
		t.Fatal("isDigit('a') should be false")
	}
	if isDigit(' ') {
		t.Fatal("isDigit(' ') should be false")
	}
	if isDigit('+') {
		t.Fatal("isDigit('+') should be false")
	}
	if isDigit('-') {
		t.Fatal("isDigit('-') should be false")
	}
}

// ============================================================================
// isPlainHTTPPort / isTLSHTTPPort — ветки: все порты
// ============================================================================

func TestIsPlainHTTPPort_80(t *testing.T) {
	if !isPlainHTTPPort(80) {
		t.Fatal("isPlainHTTPPort(80) should be true")
	}
}

func TestIsPlainHTTPPort_8080(t *testing.T) {
	if !isPlainHTTPPort(8080) {
		t.Fatal("isPlainHTTPPort(8080) should be true")
	}
}

func TestIsPlainHTTPPort_Other(t *testing.T) {
	if isPlainHTTPPort(443) {
		t.Fatal("isPlainHTTPPort(443) should be false")
	}
	if isPlainHTTPPort(22) {
		t.Fatal("isPlainHTTPPort(22) should be false")
	}
	if isPlainHTTPPort(8081) {
		t.Fatal("isPlainHTTPPort(8081) should be false")
	}
}

func TestIsTLSHTTPPort_443(t *testing.T) {
	if !isTLSHTTPPort(443) {
		t.Fatal("isTLSHTTPPort(443) should be true")
	}
}

func TestIsTLSHTTPPort_8443(t *testing.T) {
	if !isTLSHTTPPort(8443) {
		t.Fatal("isTLSHTTPPort(8443) should be true")
	}
}

func TestIsTLSHTTPPort_Other(t *testing.T) {
	if isTLSHTTPPort(80) {
		t.Fatal("isTLSHTTPPort(80) should be false")
	}
	if isTLSHTTPPort(22) {
		t.Fatal("isTLSHTTPPort(22) should be false")
	}
	if isTLSHTTPPort(8444) {
		t.Fatal("isTLSHTTPPort(8444) should be false")
	}
}
