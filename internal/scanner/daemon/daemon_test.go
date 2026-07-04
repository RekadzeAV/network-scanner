package daemon

import (
	"testing"
	"time"

	"network-scanner/internal/scanner"
)

func TestNewRunner(t *testing.T) {
	r := NewRunner()
	if r == nil {
		t.Fatal("expected runner instance")
	}
	if r.Events() == nil {
		t.Fatal("expected events channel")
	}
	if r.CurrentScanner() != nil {
		t.Fatal("scanner should be nil before Start")
	}
	if r.IsRunning() {
		t.Fatal("runner should not be running before Start")
	}
}

func TestStopWithoutStart(t *testing.T) {
	r := NewRunner()
	r.Stop()
}

func TestStartRejectsWhenAlreadyRunning(t *testing.T) {
	r := NewRunner()
	r.running = true
	if err := r.Start(Config{}); err == nil {
		t.Fatal("expected error when starting already running runner")
	}
}

func TestStartEmitsErrorWhenFactoryReturnsNil(t *testing.T) {
	r := NewRunnerWithFactory(func(Config) *scanner.NetworkScanner {
		return nil
	})
	err := r.Start(Config{})
	if err == nil {
		t.Fatal("expected start error")
	}
	select {
	case ev := <-r.Events():
		if ev.Kind != EventError {
			t.Fatalf("expected error event, got %s", ev.Kind)
		}
		if ev.Err == nil {
			t.Fatal("expected event error payload")
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected error event in channel")
	}
}
