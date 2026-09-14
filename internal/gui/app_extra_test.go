package gui

import (
	"testing"
	"time"

	"network-scanner/internal/snmpcollector"
)

// --- app.go tests ---

func TestPartialSNMPKeysFromReport_WithFailures(t *testing.T) {
	report := &snmpcollector.CollectReport{
		Failures: []snmpcollector.DeviceFailure{
			{IP: "192.168.1.1", Kind: snmpcollector.FailureQuery, Message: "query_error"},
		},
	}
	keys := partialSNMPKeysFromReport(report)
	if keys == nil {
		t.Error("expected non-nil keys for FailureQuery")
	}
}

func TestPartialSNMPKeysFromReport_EmptyFailures(t *testing.T) {
	report := &snmpcollector.CollectReport{
		Failures: []snmpcollector.DeviceFailure{},
	}
	keys := partialSNMPKeysFromReport(report)
	// Empty Failures → nil
	if keys != nil {
		t.Errorf("expected nil for empty failures, got %v", keys)
	}
}

func TestPartialSNMPKeysFromReport_MixedReport(t *testing.T) {
	report := &snmpcollector.CollectReport{
		TotalSNMPTargets: 5,
		Connected:        3,
		Partial:          1,
		Failed:           1,
		Failures: []snmpcollector.DeviceFailure{
			{IP: "10.0.0.1", Kind: snmpcollector.FailureQuery, Message: "query_error"},
		},
	}
	keys := partialSNMPKeysFromReport(report)
	if keys == nil {
		t.Error("expected non-nil keys for FailureQuery")
	}
}

func TestFormatDurationMMSS_Zero(t *testing.T) {
	result := formatDurationMMSS(0)
	if result != "00:00" {
		t.Errorf("expected '00:00', got %q", result)
	}
}

func TestFormatDurationMMSS_Small(t *testing.T) {
	result := formatDurationMMSS(30 * time.Second)
	if result != "00:30" {
		t.Errorf("expected '00:30', got %q", result)
	}
}

func TestFormatDurationMMSS_ExactlyOneMinute(t *testing.T) {
	result := formatDurationMMSS(1 * time.Minute)
	if result != "01:00" {
		t.Errorf("expected '01:00', got %q", result)
	}
}

func TestFormatDurationMMSS_FiveMinutes(t *testing.T) {
	result := formatDurationMMSS(5 * time.Minute)
	if result != "05:00" {
		t.Errorf("expected '05:00', got %q", result)
	}
}

func TestFormatDurationMMSS_OneHour(t *testing.T) {
	result := formatDurationMMSS(1 * time.Hour)
	if result != "60:00" {
		t.Errorf("expected '60:00', got %q", result)
	}
}

func TestFormatDurationMMSS_OneHourThirty(t *testing.T) {
	result := formatDurationMMSS(1*time.Hour + 30*time.Minute)
	if result != "90:00" {
		t.Errorf("expected '90:00', got %q", result)
	}
}

func TestFormatDurationMMSS_SmallFraction(t *testing.T) {
	// formatDurationMMSS округляет до ближайшей минуты
	result := formatDurationMMSS(59*time.Second + 999*time.Millisecond)
	// 59.999s округляется до 60s = 01:00
	if result != "01:00" {
		t.Errorf("expected '01:00' (rounded), got %q", result)
	}
}

func TestFormatDurationMMSS_LargeValue(t *testing.T) {
	result := formatDurationMMSS(24 * time.Hour)
	if result != "1440:00" {
		t.Errorf("expected '1440:00', got %q", result)
	}
}

func TestFormatDurationMMSS_Negative(t *testing.T) {
	result := formatDurationMMSS(-1 * time.Minute)
	if result == "" {
		t.Error("expected non-empty result for negative duration")
	}
}
