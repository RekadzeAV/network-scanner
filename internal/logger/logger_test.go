package logger

import (
	"testing"
)

// These tests exercise the release (non-debug) build of logger,
// where all functions are no-ops.

func TestInit(t *testing.T) {
	err := Init("test-app", "1.0.0")
	if err != nil {
		t.Fatalf("expected nil error in release mode, got %v", err)
	}
}

func TestClose(t *testing.T) {
	// Should not panic
	Close()
}

func TestLog(t *testing.T) {
	// Should not panic
	Log("test message: %s", "hello")
}

func TestLogError_NilError(t *testing.T) {
	// Should not panic with nil error
	LogError(nil, "context")
}

func TestLogError_WithError(t *testing.T) {
	// Should not panic with actual error
	LogError(nil, "test context")
}

func TestLogDebug(t *testing.T) {
	// Should not panic
	LogDebug("debug message: %d", 42)
}

func TestGetLogFileName(t *testing.T) {
	name := GetLogFileName()
	// In release mode, returns empty string
	if name != "" {
		t.Fatalf("expected empty string in release mode, got %q", name)
	}
}
