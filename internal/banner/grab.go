package banner

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"net/textproto"
	"strings"
	"time"
)

// GrabTCP читает первые байты с открытого TCP-порта (баннер).
func GrabTCP(host string, port int, readTimeout time.Duration) (string, error) {
	if readTimeout <= 0 {
		readTimeout = 2 * time.Second
	}
	// Для HTTP-портов инициируем запрос HEAD, иначе часть сервисов молчит.
	if isTLSHTTPPort(port) {
		if s, err := grabTLSHTTP(host, port, readTimeout); err == nil && s != "" {
			return s, nil
		}
	} else if isPlainHTTPPort(port) {
		if s, err := grabPlainHTTP(host, port, readTimeout); err == nil && s != "" {
			return s, nil
		}
	}

	raw, err := readFirstTCPBytes(host, port, readTimeout)
	if err != nil {
		return "", err
	}
	normalized := normalizeByPort(port, raw)
	if normalized != "" {
		return normalized, nil
	}
	return raw, nil
}

func sanitizeBanner(b []byte) string {
	var sb strings.Builder
	for _, c := range b {
		if c >= 32 && c < 127 {
			sb.WriteByte(byte(c))
		} else if c == '\n' || c == '\r' || c == '\t' {
			sb.WriteByte(' ')
		}
	}
	return strings.TrimSpace(sb.String())
}

func readFirstTCPBytes(host string, port int, readTimeout time.Duration) (string, error) {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, readTimeout)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(readTimeout))
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if n > 0 {
		return sanitizeBanner(buf[:n]), nil
	}
	if err != nil {
		return "", err
	}
	return "", nil
}

func grabPlainHTTP(host string, port int, timeout time.Duration) (string, error) {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))
	req := "HEAD / HTTP/1.0\r\nHost: " + host + "\r\nUser-Agent: network-scanner\r\nConnection: close\r\n\r\n"
	if _, err = conn.Write([]byte(req)); err != nil {
		return "", err
	}
	return parseHTTPResponse(conn), nil
}

func grabTLSHTTP(host string, port int, timeout time.Duration) (string, error) {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	dialer := &net.Dialer{Timeout: timeout}
	cfg := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: true, // Для баннер-граббинга не валидируем сертификат.
		MinVersion:         tls.VersionTLS10,
	}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, cfg)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))
	req := "HEAD / HTTP/1.0\r\nHost: " + host + "\r\nUser-Agent: network-scanner\r\nConnection: close\r\n\r\n"
	if _, err = conn.Write([]byte(req)); err != nil {
		return "", err
	}
	return parseHTTPResponse(conn), nil
}

func parseHTTPResponse(conn net.Conn) string {
	r := textproto.NewReader(bufio.NewReader(conn))
	status, err := r.ReadLine()
	if err != nil {
		return ""
	}
	server := ""
	powered := ""
	for {
		line, err := r.ReadLine()
		if err != nil || strings.TrimSpace(line) == "" {
			break
		}
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "server:") {
			server = strings.TrimSpace(line[len("server:"):])
		}
		if strings.HasPrefix(lower, "x-powered-by:") {
			powered = strings.TrimSpace(line[len("x-powered-by:"):])
		}
	}
	parts := []string{status}
	if server != "" {
		parts = append(parts, "Server="+server)
	}
	if powered != "" {
		parts = append(parts, "X-Powered-By="+powered)
	}
	return sanitizeBanner([]byte(strings.Join(parts, " | ")))
}

func normalizeByPort(port int, raw string) string {
	r := strings.TrimSpace(raw)
	if r == "" {
		return ""
	}
	switch port {
	case 22:
		if strings.HasPrefix(strings.ToUpper(r), "SSH-") {
			return r
		}
	case 21:
		if strings.HasPrefix(r, "220") {
			return "FTP " + r
		}
	case 25, 587:
		if strings.HasPrefix(r, "220") || strings.HasPrefix(strings.ToUpper(r), "SMTP") {
			return "SMTP " + r
		}
	case 110:
		if strings.HasPrefix(r, "+OK") {
			return "POP3 " + r
		}
	case 143:
		if strings.HasPrefix(r, "* OK") || strings.Contains(strings.ToUpper(r), "IMAP") {
			return "IMAP " + r
		}
	}
	return r
}

// ExtractVersionHint извлекает краткую версию/сигнатуру службы из баннера.
func ExtractVersionHint(port int, banner string) string {
	b := strings.TrimSpace(banner)
	if b == "" || b == "нет ответа" {
		return ""
	}
	switch port {
	case 22:
		if strings.HasPrefix(strings.ToUpper(b), "SSH-") {
			return b
		}
	case 21:
		if strings.HasPrefix(b, "FTP ") {
			return trimMailLikePrefix(strings.TrimSpace(strings.TrimPrefix(b, "FTP ")))
		}
	case 25, 587:
		if strings.HasPrefix(b, "SMTP ") {
			return trimMailLikePrefix(strings.TrimSpace(strings.TrimPrefix(b, "SMTP ")))
		}
	case 110:
		if strings.HasPrefix(b, "POP3 ") {
			return trimMailLikePrefix(strings.TrimSpace(strings.TrimPrefix(b, "POP3 ")))
		}
	case 143:
		if strings.HasPrefix(b, "IMAP ") {
			return trimMailLikePrefix(strings.TrimSpace(strings.TrimPrefix(b, "IMAP ")))
		}
	case 80, 443, 8080, 8443:
		parts := strings.Split(b, "|")
		status := ""
		server := ""
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if strings.HasPrefix(strings.ToUpper(p), "HTTP/") {
				status = p
				continue
			}
			if strings.HasPrefix(strings.ToLower(p), "server=") {
				server = strings.TrimSpace(strings.TrimPrefix(p, "Server="))
				server = strings.TrimSpace(strings.TrimPrefix(server, "server="))
			}
		}
		if status != "" && server != "" {
			return status + " (" + server + ")"
		}
		if status != "" {
			return status
		}
		if server != "" {
			return server
		}
	}
	if len(b) > 120 {
		return b[:117] + "..."
	}
	return b
}

func trimMailLikePrefix(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "+OK")
	// Срезаем типовые коды ответов: 220/250/5xx и т.п.
	if len(s) >= 3 && isDigit(s[0]) && isDigit(s[1]) && isDigit(s[2]) {
		s = strings.TrimSpace(s[3:])
		if strings.HasPrefix(s, "-") || strings.HasPrefix(s, ".") {
			s = strings.TrimSpace(s[1:])
		}
	}
	return strings.TrimSpace(s)
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isPlainHTTPPort(port int) bool {
	switch port {
	case 80, 8080:
		return true
	default:
		return false
	}
}

func isTLSHTTPPort(port int) bool {
	switch port {
	case 443, 8443:
		return true
	default:
		return false
	}
}
