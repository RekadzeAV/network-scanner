package banner

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// GrabTCP читает первые байты с открытого TCP-порта (баннер).
func GrabTCP(host string, port int, readTimeout time.Duration) (string, error) {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	if readTimeout <= 0 {
		readTimeout = 2 * time.Second
	}
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
