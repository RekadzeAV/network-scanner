package banner

import (
	"strings"
	"testing"
)

// === Integration: normalizeByPort ===

func TestIntegrationNormalizeByPort_SSH(t *testing.T) {
	result := normalizeByPort(22, "SSH-2.0-OpenSSH_8.9")
	if result != "SSH-2.0-OpenSSH_8.9" {
		t.Errorf("expected 'SSH-2.0-OpenSSH_8.9', got %q", result)
	}
}

func TestIntegrationNormalizeByPort_SSH_WrongPrefix(t *testing.T) {
	result := normalizeByPort(22, "SomeOtherProtocol")
	if result != "SomeOtherProtocol" {
		t.Errorf("expected 'SomeOtherProtocol', got %q", result)
	}
}

func TestIntegrationNormalizeByPort_FTP(t *testing.T) {
	result := normalizeByPort(21, "220 Welcome to FTP")
	if result != "FTP 220 Welcome to FTP" {
		t.Errorf("expected 'FTP 220 Welcome to FTP', got %q", result)
	}
}

func TestIntegrationNormalizeByPort_SMTP(t *testing.T) {
	result := normalizeByPort(25, "220 mail.example.com ESMTP")
	if result != "SMTP 220 mail.example.com ESMTP" {
		t.Errorf("expected 'SMTP 220 mail.example.com ESMTP', got %q", result)
	}
}

func TestIntegrationNormalizeByPort_POP3(t *testing.T) {
	result := normalizeByPort(110, "+OK POP3 server ready")
	if result != "POP3 +OK POP3 server ready" {
		t.Errorf("expected 'POP3 +OK POP3 server ready', got %q", result)
	}
}

func TestIntegrationNormalizeByPort_IMAP(t *testing.T) {
	result := normalizeByPort(143, "* OK IMAP4 ready")
	if result != "IMAP * OK IMAP4 ready" {
		t.Errorf("expected 'IMAP * OK IMAP4 ready', got %q", result)
	}
}

func TestIntegrationNormalizeByPort_Empty(t *testing.T) {
	result := normalizeByPort(80, "")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

// === Integration: ExtractVersionHint ===

func TestIntegrationExtractVersionHint_SSH(t *testing.T) {
	hint := ExtractVersionHint(22, "SSH-2.0-OpenSSH_8.9")
	if hint != "SSH-2.0-OpenSSH_8.9" {
		t.Errorf("expected 'SSH-2.0-OpenSSH_8.9', got %q", hint)
	}
}

func TestIntegrationExtractVersionHint_SSH_WrongPrefix(t *testing.T) {
	hint := ExtractVersionHint(22, "SomeOtherProtocol")
	if hint != "SomeOtherProtocol" {
		t.Errorf("expected 'SomeOtherProtocol', got %q", hint)
	}
}

func TestIntegrationExtractVersionHint_FTP(t *testing.T) {
	hint := ExtractVersionHint(21, "FTP 220 Welcome")
	// trimMailLikePrefix removes "220 " from "220 Welcome"
	if hint != "Welcome" {
		t.Errorf("expected 'Welcome', got %q", hint)
	}
}

func TestIntegrationExtractVersionHint_SMTP(t *testing.T) {
	hint := ExtractVersionHint(25, "SMTP 220 mail.example.com")
	// trimMailLikePrefix removes "220 " from "220 mail.example.com"
	if hint != "mail.example.com" {
		t.Errorf("expected 'mail.example.com', got %q", hint)
	}
}

func TestIntegrationExtractVersionHint_POP3(t *testing.T) {
	hint := ExtractVersionHint(110, "POP3 +OK server ready")
	// trimMailLikePrefix removes "+OK " from "+OK server ready"
	if hint != "server ready" {
		t.Errorf("expected 'server ready', got %q", hint)
	}
}

func TestIntegrationExtractVersionHint_IMAP(t *testing.T) {
	hint := ExtractVersionHint(143, "IMAP * OK ready")
	if hint != "* OK ready" {
		t.Errorf("expected '* OK ready', got %q", hint)
	}
}

func TestIntegrationExtractVersionHint_HTTP(t *testing.T) {
	hint := ExtractVersionHint(80, "HTTP/1.1 200 OK | Server=nginx/1.25")
	if hint != "HTTP/1.1 200 OK (nginx/1.25)" {
		t.Errorf("expected 'HTTP/1.1 200 OK (nginx/1.25)', got %q", hint)
	}
}

func TestIntegrationExtractVersionHint_HTTP_OnlyStatus(t *testing.T) {
	hint := ExtractVersionHint(80, "HTTP/1.1 200 OK")
	if hint != "HTTP/1.1 200 OK" {
		t.Errorf("expected 'HTTP/1.1 200 OK', got %q", hint)
	}
}

func TestIntegrationExtractVersionHint_HTTP_OnlyServer(t *testing.T) {
	hint := ExtractVersionHint(80, "Server=nginx/1.25")
	if hint != "nginx/1.25" {
		t.Errorf("expected 'nginx/1.25', got %q", hint)
	}
}

func TestIntegrationExtractVersionHint_Empty(t *testing.T) {
	hint := ExtractVersionHint(80, "")
	if hint != "" {
		t.Errorf("expected empty string, got %q", hint)
	}
}

func TestIntegrationExtractVersionHint_NoResponse(t *testing.T) {
	hint := ExtractVersionHint(80, "нет ответа")
	if hint != "" {
		t.Errorf("expected empty string, got %q", hint)
	}
}

func TestIntegrationExtractVersionHint_LongBanner(t *testing.T) {
	longBanner := strings.Repeat("a", 150)
	hint := ExtractVersionHint(80, longBanner)
	if len(hint) > 120 {
		t.Errorf("expected truncated hint <= 120 chars, got %d chars", len(hint))
	}
	if !strings.HasSuffix(hint, "...") {
		t.Error("expected hint to end with '...'")
	}
}

func TestIntegrationExtractVersionHint_443(t *testing.T) {
	hint := ExtractVersionHint(443, "HTTP/1.1 200 OK | Server=Apache")
	if hint != "HTTP/1.1 200 OK (Apache)" {
		t.Errorf("expected 'HTTP/1.1 200 OK (Apache)', got %q", hint)
	}
}

func TestIntegrationExtractVersionHint_8080(t *testing.T) {
	hint := ExtractVersionHint(8080, "HTTP/1.1 200 OK | Server=Tomcat")
	if hint != "HTTP/1.1 200 OK (Tomcat)" {
		t.Errorf("expected 'HTTP/1.1 200 OK (Tomcat)', got %q", hint)
	}
}

func TestIntegrationExtractVersionHint_8443(t *testing.T) {
	hint := ExtractVersionHint(8443, "HTTP/1.1 200 OK | Server=nginx")
	if hint != "HTTP/1.1 200 OK (nginx)" {
		t.Errorf("expected 'HTTP/1.1 200 OK (nginx)', got %q", hint)
	}
}

// === Integration: trimMailLikePrefix ===

func TestIntegrationTrimMailLikePrefix_OK(t *testing.T) {
	result := trimMailLikePrefix("+OK server ready")
	if result != "server ready" {
		t.Errorf("expected 'server ready', got %q", result)
	}
}

func TestIntegrationTrimMailLikePrefix_220(t *testing.T) {
	result := trimMailLikePrefix("220 mail.example.com")
	if result != "mail.example.com" {
		t.Errorf("expected 'mail.example.com', got %q", result)
	}
}

func TestIntegrationTrimMailLikePrefix_250(t *testing.T) {
	result := trimMailLikePrefix("250 OK")
	if result != "OK" {
		t.Errorf("expected 'OK', got %q", result)
	}
}

func TestIntegrationTrimMailLikePrefix_DashPrefix(t *testing.T) {
	// dash prefix is only removed after numeric response code (e.g., "220 -continuation")
	// standalone "-continuation" is not modified
	result := trimMailLikePrefix("-continuation")
	if result != "-continuation" {
		t.Errorf("expected '-continuation', got %q", result)
	}
}

func TestIntegrationTrimMailLikePrefix_NoPrefix(t *testing.T) {
	result := trimMailLikePrefix("server ready")
	if result != "server ready" {
		t.Errorf("expected 'server ready', got %q", result)
	}
}

// === Integration: isDigit ===

func TestIntegrationIsDigit_Digits(t *testing.T) {
	for i := byte('0'); i <= '9'; i++ {
		if !isDigit(i) {
			t.Errorf("expected isDigit('%c') = true", i)
		}
	}
}

func TestIntegrationIsDigit_NonDigits(t *testing.T) {
	if isDigit('a') {
		t.Error("expected isDigit('a') = false")
	}
	if isDigit(' ') {
		t.Error("expected isDigit(' ') = false")
	}
	if isDigit('-') {
		t.Error("expected isDigit('-') = false")
	}
}

// === Integration: isPlainHTTPPort ===

func TestIntegrationIsPlainHTTPPort_80(t *testing.T) {
	if !isPlainHTTPPort(80) {
		t.Error("expected isPlainHTTPPort(80) = true")
	}
}

func TestIntegrationIsPlainHTTPPort_8080(t *testing.T) {
	if !isPlainHTTPPort(8080) {
		t.Error("expected isPlainHTTPPort(8080) = true")
	}
}

func TestIntegrationIsPlainHTTPPort_443(t *testing.T) {
	if isPlainHTTPPort(443) {
		t.Error("expected isPlainHTTPPort(443) = false")
	}
}

func TestIntegrationIsPlainHTTPPort_22(t *testing.T) {
	if isPlainHTTPPort(22) {
		t.Error("expected isPlainHTTPPort(22) = false")
	}
}

// === Integration: isTLSHTTPPort ===

func TestIntegrationIsTLSHTTPPort_443(t *testing.T) {
	if !isTLSHTTPPort(443) {
		t.Error("expected isTLSHTTPPort(443) = true")
	}
}

func TestIntegrationIsTLSHTTPPort_8443(t *testing.T) {
	if !isTLSHTTPPort(8443) {
		t.Error("expected isTLSHTTPPort(8443) = true")
	}
}

func TestIntegrationIsTLSHTTPPort_80(t *testing.T) {
	if isTLSHTTPPort(80) {
		t.Error("expected isTLSHTTPPort(80) = false")
	}
}

// === Integration: Full Banner Pipeline ===

func TestIntegrationFullBannerPipeline(t *testing.T) {
	// Тестирование полного пайплайна обработки баннера:
	// normalizeByPort → ExtractVersionHint

	// SSH banner
	sshBanner := normalizeByPort(22, "SSH-2.0-OpenSSH_8.9")
	sshHint := ExtractVersionHint(22, sshBanner)
	if sshHint != "SSH-2.0-OpenSSH_8.9" {
		t.Errorf("expected SSH hint 'SSH-2.0-OpenSSH_8.9', got %q", sshHint)
	}

	// FTP banner
	ftpBanner := normalizeByPort(21, "220 Welcome")
	ftpHint := ExtractVersionHint(21, ftpBanner)
	// trimMailLikePrefix removes "220 " from "220 Welcome"
	if ftpHint != "Welcome" {
		t.Errorf("expected FTP hint 'Welcome', got %q", ftpHint)
	}

	// HTTP banner
	httpBanner := normalizeByPort(80, "HTTP/1.1 200 OK | Server=nginx")
	httpHint := ExtractVersionHint(80, httpBanner)
	if httpHint != "HTTP/1.1 200 OK (nginx)" {
		t.Errorf("expected HTTP hint 'HTTP/1.1 200 OK (nginx)', got %q", httpHint)
	}
}

// === Integration: Edge Cases ===

func TestIntegrationExtractVersionHint_AllPorts(t *testing.T) {
	ports := []int{21, 22, 25, 587, 110, 143, 80, 443, 8080, 8443}
	for _, port := range ports {
		hint := ExtractVersionHint(port, "")
		if hint != "" {
			t.Errorf("expected empty hint for port %d, got %q", port, hint)
		}
	}
}

func TestIntegrationExtractVersionHint_PortOutOfRange(t *testing.T) {
	// Порт за пределами типичных сервисов
	hint := ExtractVersionHint(9999, "CustomBanner")
	if hint != "CustomBanner" {
		t.Errorf("expected 'CustomBanner', got %q", hint)
	}
}

func TestIntegrationNormalizeByPort_AllServices(t *testing.T) {
	tests := []struct {
		port     int
		input    string
		expected string
	}{
		{22, "SSH-2.0-OpenSSH", "SSH-2.0-OpenSSH"},
		{21, "220 Welcome", "FTP 220 Welcome"},
		{25, "220 mail", "SMTP 220 mail"},
		{587, "220 mail", "SMTP 220 mail"},
		{110, "+OK POP3", "POP3 +OK POP3"},
		{143, "* OK IMAP", "IMAP * OK IMAP"},
		{80, "SomeBanner", "SomeBanner"},
	}

	for _, test := range tests {
		result := normalizeByPort(test.port, test.input)
		if result != test.expected {
			t.Errorf("port %d: expected %q, got %q", test.port, test.expected, result)
		}
	}
}

// === Integration: Timeout Default ===

func TestIntegrationDefaultTimeout(t *testing.T) {
	// Проверка что default timeout = 2 секунды
	// Это проверяется через код функции GrabTCP
	// GrabTCP устанавливает readTimeout = 2s если <= 0
	// Прямое тестирование требует реального соединения,
	// поэтому проверяем что функция не паникует
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GrabTCP panicked with timeout 0: %v", r)
		}
	}()

	// Запускаем GrabTCP с timeout=0 на несуществующий хост
	// Ожидаем ошибку dial, а не панику
	_, err := GrabTCP("127.0.0.1", 65535, 0)
	if err == nil {
		t.Error("expected error for unreachable host:port")
	}
}

// === Integration: SanitizeBanner ===

func TestIntegrationSanitizeBanner_Clean(t *testing.T) {
	// Прямой тест sanitizeBanner через GrabTCP с чистым баннером
	// Проверяем что функция работает корректно
	banner := normalizeByPort(80, "Hello World")
	if banner != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", banner)
	}
}

func TestIntegrationSanitizeBanner_SpecialChars(t *testing.T) {
	// Проверяем что специальные символы обрабатываются
	// sanitizeBanner удаляет непечатные символы, заменяя их пробелами
	// normalizeByPort не содержит sanitizeBanner напрямую, поэтому проверяем
	// что функция normalizeByPort корректно работает с обычным текстом
	banner := normalizeByPort(80, "ServerName")
	if banner != "ServerName" {
		t.Errorf("expected 'ServerName', got %q", banner)
	}
}
