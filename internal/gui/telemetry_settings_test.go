package gui

import (
	"testing"

	"network-scanner/internal/telemetry"
)

// --- telemetry_settings.go tests ---

func TestNewTelemetrySettingsManager_Nil(t *testing.T) {
	tsm := NewTelemetrySettingsManager(nil)
	if tsm == nil {
		t.Fatal("expected non-nil TelemetrySettingsManager")
	}
	if tsm.telemetry != nil {
		t.Error("expected telemetry to be nil")
	}
}

func TestNewTelemetrySettingsManager_WithTelemetry(t *testing.T) {
	tel := telemetry.NewTelemetry(nil)
	tsm := NewTelemetrySettingsManager(tel)
	if tsm == nil {
		t.Fatal("expected non-nil TelemetrySettingsManager")
	}
	if tsm.telemetry != tel {
		t.Error("expected telemetry to be set")
	}
}

func TestTelemetrySettingsManager_CreateSettingsTab_NilTelemetry(t *testing.T) {
	tsm := &TelemetrySettingsManager{}
	result := tsm.CreateSettingsTab()
	if result == nil {
		t.Fatal("expected non-nil result for nil telemetry")
	}
}

func TestTelemetrySettingsManager_CreateSettingsTab_WithTelemetry(t *testing.T) {
	tel := telemetry.NewTelemetry(nil)
	tsm := NewTelemetrySettingsManager(tel)
	result := tsm.CreateSettingsTab()
	if result == nil {
		t.Fatal("expected non-nil result for valid telemetry")
	}
}

func TestTelemetrySettingsManager_OnChanged(t *testing.T) {
	tsm := &TelemetrySettingsManager{}
	tsm.OnChanged(func(value bool) {
		_ = value
	})
	if tsm.onChanged == nil {
		t.Error("expected onChanged to be set")
	}
}

func TestTelemetrySettingsManager_GetStatus_NilTelemetry(t *testing.T) {
	tsm := &TelemetrySettingsManager{}
	status := tsm.GetStatus()
	if status != "Недоступна" {
		t.Errorf("expected 'Недоступна', got %q", status)
	}
}

func TestTelemetrySettingsManager_GetStatus_WithTelemetry(t *testing.T) {
	tel := telemetry.NewTelemetry(nil)
	tsm := NewTelemetrySettingsManager(tel)
	status := tsm.GetStatus()
	if status == "" {
		t.Error("expected non-empty status")
	}
}

func TestTelemetrySettingsManager_GetStatus_EmptyStats(t *testing.T) {
	tel := telemetry.NewTelemetry(nil)
	tsm := NewTelemetrySettingsManager(tel)
	status := tsm.GetStatus()
	if status == "" {
		t.Error("expected non-empty status for empty stats")
	}
}
