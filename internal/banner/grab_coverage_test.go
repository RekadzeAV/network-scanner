package banner

import (
	"net"
	"strings"
	"testing"
	"time"
)

// ============================================================================
// M1.4: Тесты для непокрытых функций banner (0.0% → 100%)
// ============================================================================

// TestGrabPlainHTTP_ConnectionRefused — ветка: соединение отклонено
func TestGrabPlainHTTP_ConnectionRefused(t *testing.T) {
	// Используем порт, который точно не слушается
	banner, err := grabPlainHTTP("192.0.2.1", 59999, 100*time.Millisecond)
	if err == nil {
		t.Error("expected error for refused connection")
	}
	if banner != "" {
		t.Errorf("expected empty banner, got: %q", banner)
	}
}

// TestGrabPlainHTTP_Timeout — ветка: таймаут соединения
func TestGrabPlainHTTP_Timeout(t *testing.T) {
	// Используем зарезервированную IP-область для таймаута
	banner, err := grabPlainHTTP("192.0.2.1", 59998, 50*time.Millisecond)
	if err == nil {
		t.Error("expected error or timeout")
	}
	// Banner может быть пустым или частично заполненным
	_ = banner
}

// TestGrabTLSHTTP_ConnectionRefused — ветка: TLS соединение отклонено
func TestGrabTLSHTTP_ConnectionRefused(t *testing.T) {
	banner, err := grabTLSHTTP("192.0.2.1", 59997, 100*time.Millisecond)
	if err == nil {
		t.Error("expected error for refused TLS connection")
	}
	if banner != "" {
		t.Errorf("expected empty banner, got: %q", banner)
	}
}

// TestGrabTLSHTTP_Timeout — ветка: TLS таймаут
func TestGrabTLSHTTP_Timeout(t *testing.T) {
	banner, err := grabTLSHTTP("192.0.2.1", 59996, 50*time.Millisecond)
	if err == nil {
		t.Error("expected error or timeout")
	}
	_ = banner
}

// ============================================================================
// M1.4: Дополнительные тесты для повышения покрытия
// ============================================================================

// TestGrabWithMockServer — тест с мокированным сервером
func TestGrabWithMockServer(t *testing.T) {
	// Создаём простой HTTP сервер для тестирования
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("cannot create listener: %v", err)
	}
	defer listener.Close()

	// Отвечаем HTTP баннером
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Читаем запрос
		buf := make([]byte, 1024)
		conn.Read(buf)

		// Отправляем ответ
		response := "HTTP/1.0 200 OK\r\nServer: TestServer/1.0\r\nX-Powered-By: Test\r\n\r\n"
		conn.Write([]byte(response))
	}()

	// Даем серверу время на запуск
	time.Sleep(50 * time.Millisecond)

	// Тестируем grabPlainHTTP
	port := listener.Addr().(*net.TCPAddr).Port
	banner, err := grabPlainHTTP("127.0.0.1", port, 1*time.Second)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if banner == "" {
		t.Error("expected non-empty banner from mock server")
	}
}

// TestExtractVersionHint_Edges — граничные случаи текущего port-scoped API.
//
// API принимает порт и баннер: известный порт с правильным префиксом даёт
// извлечённую версию, иначе баннер возвращается как есть (с усечением >120 байт).
func TestExtractVersionHint_Edges(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		input    string
		expected string
	}{
		{"empty string", 80, "", ""},
		{"unknown port returns banner", 0, "Apache", "Apache"},
		{"ssh prefix", 22, "SSH-2.0-OpenSSH_8.9", "SSH-2.0-OpenSSH_8.9"},
		{"ssh wrong prefix falls back to banner", 22, "SomeOther", "SomeOther"},
		{"http status and server", 80, "HTTP/1.1|200 OK|Server=Apache/2.4.41", "HTTP/1.1 (Apache/2.4.41)"},
		{"http status only", 443, "HTTP/2|404 Not Found", "HTTP/2"},
		{"placeholder no answer", 80, "нет ответа", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractVersionHint(tt.port, tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestExtractVersionHint_TruncatesLongBanner — ветка: баннер длиннее 120 байт.
func TestExtractVersionHint_TruncatesLongBanner(t *testing.T) {
	long := strings.Repeat("x", 130)
	got := ExtractVersionHint(0, long)

	if len(got) != 120 {
		t.Fatalf("expected truncated length 120, got %d (%q)", len(got), got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("expected ellipsis suffix, got %q", got)
	}
}

// TestNormalizeByPort_PortMapping — нормализация баннера по порту протокола.
func TestNormalizeByPort_PortMapping(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		raw      string
		expected string
	}{
		{"empty", 80, "", ""},
		{"ssh keeps version", 22, "SSH-2.0-OpenSSH_8.9", "SSH-2.0-OpenSSH_8.9"},
		{"ssh wrong prefix unchanged", 22, "banner", "banner"},
		{"ftp adds prefix", 21, "220---------- Welcome", "FTP 220---------- Welcome"},
		{"smtp adds prefix", 25, "220 mail.example.com", "SMTP 220 mail.example.com"},
		{"submission adds prefix", 587, "SMTP ready", "SMTP SMTP ready"},
		{"pop3 adds prefix", 110, "+OK server ready", "POP3 +OK server ready"},
		{"imap adds prefix", 143, "* OK IMAP4rev1", "IMAP * OK IMAP4rev1"},
		{"http passthrough", 80, "HTTP/1.1|200", "HTTP/1.1|200"},
		{"unknown port passthrough", 9999, "raw", "raw"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeByPort(tt.port, tt.raw)
			if result != tt.expected {
				t.Errorf("port %d: got %q, want %q", tt.port, result, tt.expected)
			}
		})
	}
}

// TestTrimMailLikePrefix — тестирование mail-префиксов
func TestTrimMailLikePrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no prefix", "OpenSSH_8.2", "OpenSSH_8.2"},
		{"with prefix", "SMTP mail server", "mail server"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trimMailLikePrefix(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestIsDigit — тестирование проверки цифр
func TestIsDigit(t *testing.T) {
	if !isDigit('0') || !isDigit('9') {
		t.Error("expected isDigit to return true for 0-9")
	}
	if isDigit('a') || isDigit(' ') {
		t.Error("expected isDigit to return false for non-digits")
	}
}

// TestIsPlainHTTPPort — тестирование HTTP-портов
func TestIsPlainHTTPPort(t *testing.T) {
	httpPorts := []int{80, 8080, 8000, 8888, 8880}
	for _, port := range httpPorts {
		if !isPlainHTTPPort(port) {
			t.Errorf("expected port %d to be HTTP port", port)
		}
	}

	nonHTTPPorts := []int{443, 22, 53, 25}
	for _, port := range nonHTTPPorts {
		if isPlainHTTPPort(port) {
			t.Errorf("expected port %d to NOT be HTTP port", port)
		}
	}
}

// TestIsTLSHTTPPort — тестирование TLS-портов
func TestIsTLSHTTPPort(t *testing.T) {
	tlsPorts := []int{443, 8443, 465, 993, 995}
	for _, port := range tlsPorts {
		if !isTLSHTTPPort(port) {
			t.Errorf("expected port %d to be TLS port", port)
		}
	}

	nonTLSPorts := []int{80, 22, 53, 25}
	for _, port := range nonTLSPorts {
		if isTLSHTTPPort(port) {
			t.Errorf("expected port %d to NOT be TLS port", port)
		}
	}
}
