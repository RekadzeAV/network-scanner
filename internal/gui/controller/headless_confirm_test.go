package controller

import (
	"testing"
)

// TestConfirmDangerousAction_nilWindow проверяет, что при nil window
// действие не выполняется.
func TestConfirmDangerousAction_nilWindow(t *testing.T) {
	called := false
	confirmDangerousAction(nil, "Title", "Message", func() {
		called = true
	})
	if called {
		t.Fatal("onConfirm should not be called when window is nil")
	}
}

// TestConfirmDangerousAction_nilCallback проверяет, что при nil onConfirm
// функция безопасно завершается.
func TestConfirmDangerousAction_nilCallback(t *testing.T) {
	confirmDangerousAction(nil, "Title", "Message", nil) // не должен паниковать
}

// TestRebootConfirmMessage проверяет формат сообщения подтверждения перезагрузки.
func TestRebootConfirmMessage(t *testing.T) {
	msg := rebootConfirmMessage("192.168.1.1")
	if msg == "" {
		t.Fatal("message should not be empty")
	}
	if !contains(msg, "192.168.1.1") {
		t.Fatalf("message should contain target: %q", msg)
	}
	if !contains(msg, "Перезагрузить устройство") {
		t.Fatalf("message should contain action text: %q", msg)
	}
	if !contains(msg, "необратимо") {
		t.Fatalf("message should contain warning about irreversibility: %q", msg)
	}
}

// TestRebootConfirmMessage_emptyTarget проверяет сообщение с пустым target.
func TestRebootConfirmMessage_emptyTarget(t *testing.T) {
	msg := rebootConfirmMessage("")
	if msg == "" {
		t.Fatal("message should not be empty even with empty target")
	}
	if !contains(msg, "Перезагрузить устройство") {
		t.Fatalf("message should contain action text: %q", msg)
	}
}

// contains — хелпер для проверки подстроки.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
