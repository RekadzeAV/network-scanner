package nettools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func runCmd(ctx context.Context, args []string, timeout time.Duration) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("нет команды")
	}
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, args[0], args[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	s := out.String()
	if err != nil && strings.TrimSpace(s) == "" {
		return "", fmt.Errorf("%v: %s", err, args[0])
	}
	return strings.TrimSpace(s), nil
}
