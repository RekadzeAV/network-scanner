//go:build !debug
// +build !debug

package logger

import "testing"

// TestLogStructured_Release — в релизной сборке structured-логгер — заглушка
// (не паникует, ничего не пишет).
func TestLogStructured_Release(t *testing.T) {
	LogStructured("info", "scan started", map[string]interface{}{"network": "192.168.1.0/24"})
	LogStructured("error", "scan failed", nil)
}
