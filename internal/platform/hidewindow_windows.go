//go:build windows

package platform

import (
	"os/exec"
	"syscall"
)

// HideWindow предотвращает появление консольного окна (cmd.exe) при запуске
// subprocess через exec.Cmd. Устанавливает SysProcAttr.HideWindow = true,
// что подавляет создание нового console window для дочернего процесса.
//
// Применяется к ping, arp, tracert, whois, dot и другим утилитам,
// запускаемым через exec.Command в сканере и сетевых инструментах.
func HideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
