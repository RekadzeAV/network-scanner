package banner

import (
	"net"
	"strings"
	"testing"
	"time"
)

// === Integration: parseHTTPResponse ===

func TestIntegrationParseHTTPResponse_StatusOnly(t *testing.T) {
	// Создаём mock-conn с HTTP response
	data := "HTTP/1.1 200 OK\r\n\r\n"
	conn := &mockConn{data: data}
	result := parseHTTPResponse(conn)
	if result != "HTTP/1.1 200 OK" {
		t.Errorf("expected 'HTTP/1.1 200 OK', got %q", result)
	}
}

func TestIntegrationParseHTTPResponse_WithServer(t *testing.T) {
	data := "HTTP/1.1 200 OK\r\nServer: nginx/1.25.0\r\n\r\n"
	conn := &mockConn{data: data}
	result := parseHTTPResponse(conn)
	expected := "HTTP/1.1 200 OK | Server=nginx/1.25.0"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestIntegrationParseHTTPResponse_WithServerAndPowered(t *testing.T) {
	data := "HTTP/1.1 200 OK\r\nServer: Apache\r\nX-Powered-By: PHP/8.0\r\n\r\n"
	conn := &mockConn{data: data}
	result := parseHTTPResponse(conn)
	expected := "HTTP/1.1 200 OK | Server=Apache | X-Powered-By=PHP/8.0"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestIntegrationParseHTTPResponse_EmptyResponse(t *testing.T) {
	conn := &mockConn{data: ""}
	result := parseHTTPResponse(conn)
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestIntegrationParseHTTPResponse_CaseInsensitiveServer(t *testing.T) {
	data := "HTTP/1.1 200 OK\r\nserver: nginx\r\n\r\n"
	conn := &mockConn{data: data}
	result := parseHTTPResponse(conn)
	expected := "HTTP/1.1 200 OK | Server=nginx"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestIntegrationParseHTTPResponse_CaseInsensitivePowered(t *testing.T) {
	data := "HTTP/1.1 200 OK\r\nx-powered-by: ASP.NET\r\n\r\n"
	conn := &mockConn{data: data}
	result := parseHTTPResponse(conn)
	expected := "HTTP/1.1 200 OK | X-Powered-By=ASP.NET"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestIntegrationParseHTTPResponse_MultipleHeaders(t *testing.T) {
	data := "HTTP/1.1 301 Moved\r\nServer: haproxy\r\nX-Powered-By: Express\r\nContent-Type: text/html\r\n\r\n"
	conn := &mockConn{data: data}
	result := parseHTTPResponse(conn)
	expected := "HTTP/1.1 301 Moved | Server=haproxy | X-Powered-By=Express"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// === Integration: sanitizeBanner Edge Cases ===

func TestIntegrationSanitizeBanner_TabChars(t *testing.T) {
	result := sanitizeBanner([]byte("Hello\tWorld"))
	if result != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", result)
	}
}

func TestIntegrationSanitizeBanner_MixedWhitespace(t *testing.T) {
	result := sanitizeBanner([]byte("A\nB\tC\rD"))
	if result != "A B C D" {
		t.Errorf("expected 'A B C D', got %q", result)
	}
}

func TestIntegrationSanitizeBanner_AllWhitespace(t *testing.T) {
	result := sanitizeBanner([]byte("\n\t\r\n"))
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestIntegrationSanitizeBanner_Boundary31(t *testing.T) {
	result := sanitizeBanner([]byte{31, 32, 33})
	// 31 — non-printable, 32 — space, 33 — !
	// TrimSpace удалит пробел, останется "!"
	if result != "!" {
		t.Errorf("expected '!', got %q", result)
	}
}

func TestIntegrationSanitizeBanner_Boundary126(t *testing.T) {
	result := sanitizeBanner([]byte{126, 127})
	// 126 — ~, 127 — non-printable
	if result != "~" {
		t.Errorf("expected '~', got %q", result)
	}
}

func TestIntegrationSanitizeBanner_NumericOnly(t *testing.T) {
	result := sanitizeBanner([]byte("12345"))
	if result != "12345" {
		t.Errorf("expected '12345', got %q", result)
	}
}

// === Integration: normalizeByPort Edge Cases ===

func TestIntegrationNormalizeByPort_WhitespaceOnly(t *testing.T) {
	result := normalizeByPort(22, "   \t\n  ")
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestIntegrationNormalizeByPort_UpperCaseSSH(t *testing.T) {
	result := normalizeByPort(22, "SSH-2.0-OpenSSH_9.3")
	if result != "SSH-2.0-OpenSSH_9.3" {
		t.Errorf("expected 'SSH-2.0-OpenSSH_9.3', got %q", result)
	}
}

func TestIntegrationNormalizeByPort_LowerCaseSSH(t *testing.T) {
	result := normalizeByPort(22, "ssh-2.0-openssh_9.3")
	// toUpper сделает "SSH-2.0-OPENSPP_9.3" — должно совпасть
	if result != "ssh-2.0-openssh_9.3" {
		t.Errorf("expected 'ssh-2.0-openssh_9.3', got %q", result)
	}
}

func TestIntegrationNormalizeByPort_FTP220WithSpace(t *testing.T) {
	result := normalizeByPort(21, "220  Welcome")
	if result != "FTP 220  Welcome" {
		t.Errorf("expected 'FTP 220  Welcome', got %q", result)
	}
}

func TestIntegrationNormalizeByPort_SMTP220(t *testing.T) {
	result := normalizeByPort(25, "220 mail.example.com")
	if result != "SMTP 220 mail.example.com" {
		t.Errorf("expected 'SMTP 220 mail.example.com', got %q", result)
	}
}

func TestIntegrationNormalizeByPort_SMTPUpperCase(t *testing.T) {
	result := normalizeByPort(587, "smtp greeting")
	// toUpper сделает "SMTP GREETING" — должен совпасть префикс
	if result != "SMTP smtp greeting" {
		t.Errorf("expected 'SMTP smtp greeting', got %q", result)
	}
}

func TestIntegrationNormalizeByPort_POP3PlusOK(t *testing.T) {
	result := normalizeByPort(110, "+OK Dovecot ready.")
	if result != "POP3 +OK Dovecot ready." {
		t.Errorf("expected 'POP3 +OK Dovecot ready.', got %q", result)
	}
}

func TestIntegrationNormalizeByPort_IMAPOK(t *testing.T) {
	result := normalizeByPort(143, "* OK IMAP4rev1")
	if result != "IMAP * OK IMAP4rev1" {
		t.Errorf("expected 'IMAP * OK IMAP4rev1', got %q", result)
	}
}

func TestIntegrationNormalizeByPort_IMAPContainsIMAP(t *testing.T) {
	result := normalizeByPort(143, "IMAP server ready")
	// Contains(strings.ToUpper(r), "IMAP") — должно совпасть
	if result != "IMAP IMAP server ready" {
		t.Errorf("expected 'IMAP IMAP server ready', got %q", result)
	}
}

func TestIntegrationNormalizeByPort_RandomPortNoNormalization(t *testing.T) {
	result := normalizeByPort(9999, "custom banner")
	if result != "custom banner" {
		t.Errorf("expected 'custom banner', got %q", result)
	}
}

func TestIntegrationNormalizeByPort_RandomPortEmpty(t *testing.T) {
	result := normalizeByPort(9999, "")
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

// === Integration: ExtractVersionHint Edge Cases ===

func TestIntegrationExtractVersionHint_EmptyString(t *testing.T) {
	result := ExtractVersionHint(80, "")
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestIntegrationExtractVersionHint_WhitespaceOnly(t *testing.T) {
	result := ExtractVersionHint(80, "   \t\n  ")
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestIntegrationExtractVersionHint_NoResponseChinese(t *testing.T) {
	result := ExtractVersionHint(80, "нет ответа")
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestIntegrationExtractVersionHint_SSHUpperCase(t *testing.T) {
	result := ExtractVersionHint(22, "SSH-2.0-OpenSSH_9.3")
	if result != "SSH-2.0-OpenSSH_9.3" {
		t.Errorf("expected 'SSH-2.0-OpenSSH_9.3', got %q", result)
	}
}

func TestIntegrationExtractVersionHint_SSHLowerCase(t *testing.T) {
	result := ExtractVersionHint(22, "ssh-2.0-openssh_9.3")
	// toUpper сделает "SSH-2.0-OPENSPP_9.3" — должно совпасть
	if result != "ssh-2.0-openssh_9.3" {
		t.Errorf("expected 'ssh-2.0-openssh_9.3', got %q", result)
	}
}

func TestIntegrationExtractVersionHint_FTPWithCode(t *testing.T) {
	result := ExtractVersionHint(21, "FTP 220 FileZilla Server 1.8.0")
	want := "FileZilla Server 1.8.0"
	if result != want {
		t.Errorf("expected %q, got %q", want, result)
	}
}

func TestIntegrationExtractVersionHint_FTPNoCode(t *testing.T) {
	result := ExtractVersionHint(21, "FTP plain banner")
	want := "plain banner"
	if result != want {
		t.Errorf("expected %q, got %q", want, result)
	}
}

func TestIntegrationExtractVersionHint_SMTPWithCode(t *testing.T) {
	result := ExtractVersionHint(25, "SMTP 220 smtp.example.com")
	want := "smtp.example.com"
	if result != want {
		t.Errorf("expected %q, got %q", want, result)
	}
}

func TestIntegrationExtractVersionHint_POP3PlusOK(t *testing.T) {
	result := ExtractVersionHint(110, "POP3 +OK Dovecot")
	want := "Dovecot"
	if result != want {
		t.Errorf("expected %q, got %q", want, result)
	}
}

func TestIntegrationExtractVersionHint_IMAPOK(t *testing.T) {
	result := ExtractVersionHint(143, "IMAP * OK ready")
	want := "* OK ready"
	if result != want {
		t.Errorf("expected %q, got %q", want, result)
	}
}

func TestIntegrationExtractVersionHint_HTTPStatusOnly(t *testing.T) {
	result := ExtractVersionHint(80, "HTTP/1.1 200 OK")
	if result != "HTTP/1.1 200 OK" {
		t.Errorf("expected 'HTTP/1.1 200 OK', got %q", result)
	}
}

func TestIntegrationExtractVersionHint_HTTPServerOnly(t *testing.T) {
	result := ExtractVersionHint(80, "Server=nginx")
	if result != "nginx" {
		t.Errorf("expected 'nginx', got %q", result)
	}
}

func TestIntegrationExtractVersionHint_HTTPServerLowercase(t *testing.T) {
	result := ExtractVersionHint(80, "server=nginx/1.25")
	if result != "nginx/1.25" {
		t.Errorf("expected 'nginx/1.25', got %q", result)
	}
}

func TestIntegrationExtractVersionHint_HTTPNoParts(t *testing.T) {
	result := ExtractVersionHint(80, "custom banner")
	if result != "custom banner" {
		t.Errorf("expected 'custom banner', got %q", result)
	}
}

func TestIntegrationExtractVersionHint_HTTP443(t *testing.T) {
	result := ExtractVersionHint(443, "HTTP/1.1 200 OK | Server=Apache")
	if result != "HTTP/1.1 200 OK (Apache)" {
		t.Errorf("expected 'HTTP/1.1 200 OK (Apache)', got %q", result)
	}
}

func TestIntegrationExtractVersionHint_HTTP8080(t *testing.T) {
	result := ExtractVersionHint(8080, "HTTP/1.1 301 Moved | Server=tomcat")
	if result != "HTTP/1.1 301 Moved (tomcat)" {
		t.Errorf("expected 'HTTP/1.1 301 Moved (tomcat)', got %q", result)
	}
}

func TestIntegrationExtractVersionHint_HTTP8443(t *testing.T) {
	result := ExtractVersionHint(8443, "HTTP/1.1 200 OK | Server=haproxy")
	if result != "HTTP/1.1 200 OK (haproxy)" {
		t.Errorf("expected 'HTTP/1.1 200 OK (haproxy)', got %q", result)
	}
}

func TestIntegrationExtractVersionHint_LongBanner121Chars(t *testing.T) {
	long := strings.Repeat("A", 121)
	result := ExtractVersionHint(80, long)
	if len(result) != 120 {
		t.Errorf("expected 120 chars, got %d", len(result))
	}
	if result[117:] != "..." {
		t.Errorf("expected '...' suffix, got %q", result[117:])
	}
}

func TestIntegrationExtractVersionHint_LongBanner200Chars(t *testing.T) {
	long := strings.Repeat("B", 200)
	result := ExtractVersionHint(80, long)
	if len(result) != 120 {
		t.Errorf("expected 120 chars, got %d", len(result))
	}
	if result[117:] != "..." {
		t.Errorf("expected '...' suffix, got %q", result[117:])
	}
}

func TestIntegrationExtractVersionHint_LongBanner120Chars(t *testing.T) {
	long := strings.Repeat("C", 120)
	result := ExtractVersionHint(80, long)
	if len(result) != 120 {
		t.Errorf("expected 120 chars, got %d", len(result))
	}
	if strings.HasSuffix(result, "...") {
		t.Error("expected no '...' suffix for exactly 120 chars")
	}
}

func TestIntegrationExtractVersionHint_Ports9999(t *testing.T) {
	result := ExtractVersionHint(9999, "CustomService v1.0")
	if result != "CustomService v1.0" {
		t.Errorf("expected 'CustomService v1.0', got %q", result)
	}
}

// === Integration: trimMailLikePrefix Edge Cases ===

func TestIntegrationTrimMailLikePrefix_JustPlusOK(t *testing.T) {
	result := trimMailLikePrefix("+OK")
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestIntegrationTrimMailLikePrefix_220WithDash(t *testing.T) {
	result := trimMailLikePrefix("220 -continuation")
	if result != "continuation" {
		t.Errorf("expected 'continuation', got %q", result)
	}
}

func TestIntegrationTrimMailLikePrefix_220WithDot(t *testing.T) {
	result := trimMailLikePrefix("220 .continuation")
	if result != "continuation" {
		t.Errorf("expected 'continuation', got %q", result)
	}
}

func TestIntegrationTrimMailLikePrefix_250OK(t *testing.T) {
	result := trimMailLikePrefix("250 OK")
	if result != "OK" {
		t.Errorf("expected 'OK', got %q", result)
	}
}

func TestIntegrationTrimMailLikePrefix_550Error(t *testing.T) {
	result := trimMailLikePrefix("550 User not found")
	if result != "User not found" {
		t.Errorf("expected 'User not found', got %q", result)
	}
}

func TestIntegrationTrimMailLikePrefix_WithSpaces(t *testing.T) {
	result := trimMailLikePrefix("  +OK  server  ")
	if result != "server" {
		t.Errorf("expected 'server', got %q", result)
	}
}

func TestIntegrationTrimMailLikePrefix_ShortString(t *testing.T) {
	result := trimMailLikePrefix("AB")
	if result != "AB" {
		t.Errorf("expected 'AB', got %q", result)
	}
}

func TestIntegrationTrimMailLikePrefix_OnlyDigits(t *testing.T) {
	result := trimMailLikePrefix("220")
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestIntegrationTrimMailLikePrefix_EmptyString(t *testing.T) {
	result := trimMailLikePrefix("")
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestIntegrationTrimMailLikePrefix_PlainText(t *testing.T) {
	result := trimMailLikePrefix("Dovecot ready.")
	if result != "Dovecot ready." {
		t.Errorf("expected 'Dovecot ready.', got %q", result)
	}
}

// === Integration: isDigit Comprehensive ===

func TestIntegrationIsDigit_Comprehensive(t *testing.T) {
	// All digits
	for i := byte('0'); i <= '9'; i++ {
		if !isDigit(i) {
			t.Errorf("isDigit('%c') should be true", i)
		}
	}

	// Common non-digits
	nonDigits := []byte{'a', 'A', ' ', '\n', '\t', '+', '-', '.', '_', '!', '@', '#'}
	for _, c := range nonDigits {
		if isDigit(c) {
			t.Errorf("isDigit('%c') should be false", c)
		}
	}
}

// === Integration: isPlainHTTPPort Comprehensive ===

func TestIntegrationIsPlainHTTPPort_Comprehensive(t *testing.T) {
	if !isPlainHTTPPort(80) {
		t.Error("isPlainHTTPPort(80) should be true")
	}
	if !isPlainHTTPPort(8080) {
		t.Error("isPlainHTTPPort(8080) should be true")
	}
	if isPlainHTTPPort(443) {
		t.Error("isPlainHTTPPort(443) should be false")
	}
	if isPlainHTTPPort(8081) {
		t.Error("isPlainHTTPPort(8081) should be false")
	}
	if isPlainHTTPPort(22) {
		t.Error("isPlainHTTPPort(22) should be false")
	}
	if isPlainHTTPPort(0) {
		t.Error("isPlainHTTPPort(0) should be false")
	}
}

// === Integration: isTLSHTTPPort Comprehensive ===

func TestIntegrationIsTLSHTTPPort_Comprehensive(t *testing.T) {
	if !isTLSHTTPPort(443) {
		t.Error("isTLSHTTPPort(443) should be true")
	}
	if !isTLSHTTPPort(8443) {
		t.Error("isTLSHTTPPort(8443) should be true")
	}
	if isTLSHTTPPort(80) {
		t.Error("isTLSHTTPPort(80) should be false")
	}
	if isTLSHTTPPort(8444) {
		t.Error("isTLSHTTPPort(8444) should be false")
	}
	if isTLSHTTPPort(22) {
		t.Error("isTLSHTTPPort(22) should be false")
	}
	if isTLSHTTPPort(0) {
		t.Error("isTLSHTTPPort(0) should be false")
	}
}

// === Integration: Full Pipeline (logic only) ===

func TestIntegrationFullBannerPipeline_Logic(t *testing.T) {
	// SSH: normalizeByPort → ExtractVersionHint
	sshBanner := normalizeByPort(22, "SSH-2.0-OpenSSH_9.3")
	sshHint := ExtractVersionHint(22, sshBanner)
	if sshHint != "SSH-2.0-OpenSSH_9.3" {
		t.Errorf("SSH pipeline: expected 'SSH-2.0-OpenSSH_9.3', got %q", sshHint)
	}

	// FTP: normalizeByPort → ExtractVersionHint
	ftpBanner := normalizeByPort(21, "220 FileZilla")
	ftpHint := ExtractVersionHint(21, ftpBanner)
	want := "FileZilla"
	if ftpHint != want {
		t.Errorf("FTP pipeline: expected %q, got %q", want, ftpHint)
	}

	// SMTP: normalizeByPort → ExtractVersionHint
	smtpBanner := normalizeByPort(25, "220 mail.example.com")
	smtpHint := ExtractVersionHint(25, smtpBanner)
	want = "mail.example.com"
	if smtpHint != want {
		t.Errorf("SMTP pipeline: expected %q, got %q", want, smtpHint)
	}

	// HTTP: ExtractVersionHint directly
	httpHint := ExtractVersionHint(80, "HTTP/1.1 200 OK | Server=nginx")
	want = "HTTP/1.1 200 OK (nginx)"
	if httpHint != want {
		t.Errorf("HTTP pipeline: expected %q, got %q", want, httpHint)
	}
}

// === Integration: sanitizeBanner Special Cases ===

func TestIntegrationSanitizeBanner_AllPrintableRange(t *testing.T) {
	// Тестируем весь printable диапазон (32-126)
	var buf []byte
	for i := 32; i < 127; i++ {
		buf = append(buf, byte(i))
	}
	result := sanitizeBanner(buf)
	// TrimSpace удалит начальные/конечные пробелы (32 — это пробел)
	// Поэтому длина может быть меньше 95
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestIntegrationSanitizeBanner_NonPrintableRange(t *testing.T) {
	// Все символы < 32 и > 126 должны быть удалены
	var buf []byte
	for i := 0; i < 256; i++ {
		buf = append(buf, byte(i))
	}
	result := sanitizeBanner(buf)
	// Должны остаться только printable символы (32-126)
	for _, c := range result {
		if c < 32 || c >= 127 {
			t.Errorf("unexpected non-printable char in result: %d", c)
		}
	}
}

// mockConn — mock net.Conn для тестирования parseHTTPResponse
type mockConn struct {
	data   string
	offset int
}

func (m *mockConn) Read(p []byte) (n int, err error) {
	if m.offset >= len(m.data) {
		return 0, net.ErrClosed
	}
	n = copy(p, m.data[m.offset:])
	m.offset += n
	return n, nil
}

func (m *mockConn) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockConn) Close() error                       { return nil }
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }
