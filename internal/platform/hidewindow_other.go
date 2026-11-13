//go:build !windows

package platform

import "os/exec"

// HideWindow — заглушка для non-Windows. На Unix-системах subprocess
// не создают консольных окон, поэтому операция ничего не делает.
func HideWindow(cmd *exec.Cmd) {
}
