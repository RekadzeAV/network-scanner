package features

import (
	"strings"
	"testing"
)

func TestRegisterAndIsEnabled(t *testing.T) {
	m := NewManager()

	f := m.Register("test.flag", "Test flag", true)
	if !f.enabled.Load() {
		t.Error("flag should be enabled by default")
	}
	if !m.IsEnabled("test.flag") {
		t.Error("IsEnabled should return true for enabled flag")
	}

	f2 := m.Register("test.flag2", "Test flag 2", false)
	if f2.enabled.Load() {
		t.Error("flag should be disabled by default")
	}
	if m.IsEnabled("test.flag2") {
		t.Error("IsEnabled should return false for disabled flag")
	}
}

func TestSetEnabled(t *testing.T) {
	m := NewManager()
	m.Register("test.flag", "Test", true)

	m.SetEnabled("test.flag", false)
	if m.IsEnabled("test.flag") {
		t.Error("flag should be disabled after SetEnabled(false)")
	}

	m.SetEnabled("test.flag", true)
	if !m.IsEnabled("test.flag") {
		t.Error("flag should be enabled after SetEnabled(true)")
	}
}

func TestSetEnabledUnknownFlag(t *testing.T) {
	m := NewManager()
	// Не должно паниковать
	m.SetEnabled("unknown.flag", true)
}

func TestToggle(t *testing.T) {
	m := NewManager()
	m.Register("test.flag", "Test", true)

	// Toggle должен переключить с true на false
	result := m.Toggle("test.flag")
	if result {
		t.Error("Toggle should return false (was true)")
	}
	if m.IsEnabled("test.flag") {
		t.Error("flag should be false after toggle")
	}

	// Toggle должен переключить с false на true
	result = m.Toggle("test.flag")
	if !result {
		t.Error("Toggle should return true (was false)")
	}
	if !m.IsEnabled("test.flag") {
		t.Error("flag should be true after toggle")
	}
}

func TestToggleUnknownFlag(t *testing.T) {
	m := NewManager()
	result := m.Toggle("unknown.flag")
	if result {
		t.Error("Toggle on unknown flag should return false")
	}
}

func TestFlags(t *testing.T) {
	m := NewManager()
	m.Register("flag1", "First", true)
	m.Register("flag2", "Second", false)

	flags := m.Flags()
	if len(flags) != 2 {
		t.Errorf("expected 2 flags, got %d", len(flags))
	}

	names := make(map[string]bool)
	for _, f := range flags {
		names[f.name] = true
	}
	if !names["flag1"] || !names["flag2"] {
		t.Error("missing expected flags in Flags()")
	}
}

func TestStatusReport(t *testing.T) {
	m := NewManager()
	m.Register("flag1", "First flag", true)
	m.Register("flag2", "Second flag", false)

	report := m.StatusReport()

	if !strings.Contains(report, "Feature Flags Status:") {
		t.Error("report should contain header")
	}
	if !strings.Contains(report, "[ON] flag1") {
		t.Error("report should show flag1 as ON")
	}
	if !strings.Contains(report, "[OFF] flag2") {
		t.Error("report should show flag2 as OFF")
	}
	if !strings.Contains(report, "First flag") {
		t.Error("report should contain description")
	}
}

func TestDefaultFlags(t *testing.T) {
	// Проверяем, что предопределённые флаги зарегистрированы
	flags := DefaultManager.Flags()
	if len(flags) == 0 {
		t.Error("DefaultManager should have registered flags")
	}

	// Проверяем ключевые флаги D-трека
	expectedFlags := []string{
		"d1.topology.hardening",
		"d1.topology.fallback",
		"d2.export.schema_validation",
		"d3.gui.responsive",
	}

	report := DefaultManager.StatusReport()
	for _, expected := range expectedFlags {
		if !strings.Contains(report, expected) {
			t.Errorf("DefaultManager should contain flag %q", expected)
		}
	}
}

func TestConcurrency(t *testing.T) {
	m := NewManager()
	m.Register("concurrent.flag", "Concurrent test", true)

	done := make(chan bool, 100)
	for i := 0; i < 50; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				m.SetEnabled("concurrent.flag", j%2 == 0)
				_ = m.IsEnabled("concurrent.flag")
			}
			done <- true
		}()
	}
	for i := 0; i < 50; i++ {
		<-done
	}
}
