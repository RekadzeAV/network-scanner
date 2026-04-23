package nettools

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ToolErrorCode описывает нормализованный тип ошибки инструментов.
type ToolErrorCode string

const (
	ToolErrorNotInstalled    ToolErrorCode = "not_installed"
	ToolErrorPermissionDenied ToolErrorCode = "permission_denied"
	ToolErrorTimeout         ToolErrorCode = "timeout"
	ToolErrorNetwork         ToolErrorCode = "network_error"
	ToolErrorParse           ToolErrorCode = "parse_error"
	ToolErrorUnknown         ToolErrorCode = "unknown"
)

// ToolError содержит нормализованную ошибку для CLI/GUI.
type ToolError struct {
	Code    ToolErrorCode
	Tool    string
	Message string
	Err     error
}

func (e *ToolError) Error() string {
	tool := strings.TrimSpace(e.Tool)
	msg := strings.TrimSpace(e.Message)
	if tool == "" {
		tool = "tool"
	}
	if msg == "" {
		msg = "ошибка выполнения"
	}
	return fmt.Sprintf("%s (%s): %s", tool, e.Code, msg)
}

func (e *ToolError) Unwrap() error {
	return e.Err
}

func newToolError(tool string, code ToolErrorCode, message string, err error) error {
	return &ToolError{
		Code:    code,
		Tool:    strings.TrimSpace(tool),
		Message: strings.TrimSpace(message),
		Err:     err,
	}
}

// HumanizeToolError возвращает короткое пользовательское описание ошибки инструмента.
func HumanizeToolError(err error) string {
	if err == nil {
		return ""
	}
	var te *ToolError
	if !errors.As(err, &te) {
		return err.Error()
	}
	base := strings.TrimSpace(te.Message)
	if base == "" {
		base = "не удалось выполнить инструмент"
	}
	switch te.Code {
	case ToolErrorNotInstalled:
		return fmt.Sprintf("%s. Установите утилиту или проверьте PATH.", base)
	case ToolErrorPermissionDenied:
		return fmt.Sprintf("%s. Запустите с достаточными правами.", base)
	case ToolErrorTimeout:
		return fmt.Sprintf("%s. Увеличьте timeout и повторите.", base)
	case ToolErrorNetwork:
		return fmt.Sprintf("%s. Проверьте сеть, DNS и доступность хоста.", base)
	default:
		return base
	}
}

func runCmd(ctx context.Context, args []string, timeout time.Duration) (string, error) {
	if len(args) == 0 {
		return "", newToolError("", ToolErrorUnknown, "не задана команда", nil)
	}
	tool := strings.TrimSpace(args[0])
	if tool == "" {
		return "", newToolError("", ToolErrorUnknown, "пустое имя команды", nil)
	}
	toolPath, err := exec.LookPath(tool)
	if err != nil {
		return "", newToolError(tool, ToolErrorNotInstalled, "утилита не найдена в PATH", err)
	}
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, toolPath, args[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	s := out.String()
	trimmed := strings.TrimSpace(s)
	if errors.Is(cctx.Err(), context.DeadlineExceeded) {
		return trimmed, newToolError(tool, ToolErrorTimeout, "превышен таймаут выполнения", cctx.Err())
	}
	if err != nil {
		lower := strings.ToLower(trimmed)
		switch {
		case strings.Contains(lower, "permission denied"),
			strings.Contains(lower, "access is denied"),
			strings.Contains(lower, "operation not permitted"):
			return trimmed, newToolError(tool, ToolErrorPermissionDenied, "недостаточно прав для запуска", err)
		case strings.Contains(lower, "network is unreachable"),
			strings.Contains(lower, "name or service not known"),
			strings.Contains(lower, "no route to host"):
			return trimmed, newToolError(tool, ToolErrorNetwork, "сетевая ошибка при выполнении", err)
		default:
			msg := "ошибка выполнения команды"
			if trimmed == "" {
				msg = err.Error()
			}
			return trimmed, newToolError(tool, ToolErrorUnknown, msg, err)
		}
	}
	return trimmed, nil
}
