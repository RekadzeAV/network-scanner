package nettools

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"
)

// RunTraceroute вызывает tracert (Windows) или traceroute (Unix).
func RunTraceroute(ctx context.Context, host string, timeout time.Duration) (string, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return "", fmt.Errorf("пустой хост")
	}
	var args []string
	switch runtime.GOOS {
	case "windows":
		args = []string{"tracert", "-d", "-h", "30", host}
	default:
		args = []string{"traceroute", "-m", "30", "-n", host}
	}
	return runCmd(ctx, args, timeout)
}
