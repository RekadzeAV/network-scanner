package nettools

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// RunWhois вызывает внешнюю утилиту whois, если она есть в PATH.
func RunWhois(ctx context.Context, query string, timeout time.Duration) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", fmt.Errorf("пустой запрос")
	}
	if timeout <= 0 {
		timeout = 45 * time.Second
	}
	path, err := exec.LookPath("whois")
	if err != nil {
		return "", fmt.Errorf("утилита whois не найдена в PATH: установите whois или используйте DNS")
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, path, query)
	out, err := cmd.CombinedOutput()
	s := strings.TrimSpace(string(out))
	if err != nil && s == "" {
		return "", fmt.Errorf("whois: %w", err)
	}
	if runtime.GOOS == "windows" {
		return s + "\n\n(Windows: для полного whois может потребоваться установка клиента, например Sysinternals или пакет whois)", nil
	}
	return s, nil
}
