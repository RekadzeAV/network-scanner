package gui

import (
	"context"
	"errors"
	"testing"
	"time"
)

func waitOperationStatus(t *testing.T, m *OperationsManager, id string, expected OperationStatus, timeout time.Duration) Operation {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		op, ok := m.Get(id)
		if ok && op.Status == expected {
			return op
		}
		time.Sleep(5 * time.Millisecond)
	}
	op, _ := m.Get(id)
	t.Fatalf("operation %s did not reach status %s (current=%s)", id, expected, op.Status)
	return Operation{}
}

func TestOperationsManager_RunSuccess(t *testing.T) {
	m := NewOperationsManager()
	id := m.Run(OperationTypeTool, "tool ping", func(ctx context.Context) error {
		return nil
	})
	if id == "" {
		t.Fatalf("expected non-empty operation id")
	}
	op := waitOperationStatus(t, m, id, OperationSuccess, 500*time.Millisecond)
	if op.Error != "" {
		t.Fatalf("unexpected error: %q", op.Error)
	}
	if op.Duration < 0 {
		t.Fatalf("expected non-negative duration")
	}
}

func TestOperationsManager_RunFailedAndRetry(t *testing.T) {
	m := NewOperationsManager()
	attempt := 0
	id := m.Run(OperationTypeTool, "tool dns", func(ctx context.Context) error {
		attempt++
		if attempt == 1 {
			return errors.New("first error")
		}
		return nil
	})
	op := waitOperationStatus(t, m, id, OperationFailed, 500*time.Millisecond)
	if op.Error == "" {
		t.Fatalf("expected failure error text")
	}
	newID, ok := m.Retry(id)
	if !ok || newID == "" {
		t.Fatalf("retry should succeed")
	}
	if newID == id {
		t.Fatalf("retry should create new operation id")
	}
	waitOperationStatus(t, m, newID, OperationSuccess, 500*time.Millisecond)
}

func TestOperationsManager_Cancel(t *testing.T) {
	m := NewOperationsManager()
	id := m.Run(OperationTypeScan, "scan", func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
			return nil
		}
	})
	time.Sleep(25 * time.Millisecond)
	if !m.Cancel(id) {
		t.Fatalf("expected cancel to succeed")
	}
	waitOperationStatus(t, m, id, OperationCanceled, 700*time.Millisecond)
}

func TestOperationsManager_SubscribeReceivesUpdates(t *testing.T) {
	m := NewOperationsManager()
	events := m.Subscribe(8)
	id := m.Run(OperationTypeInternal, "event probe", func(ctx context.Context) error {
		return nil
	})
	timeout := time.After(500 * time.Millisecond)
	seenQueued := false
	seenFinal := false
	for !(seenQueued && seenFinal) {
		select {
		case ev := <-events:
			if ev.Operation.ID != id {
				continue
			}
			if ev.Operation.Status == OperationQueued {
				seenQueued = true
			}
			if ev.Operation.Status == OperationSuccess {
				seenFinal = true
			}
		case <-timeout:
			t.Fatalf("did not observe expected events queued=%v final=%v", seenQueued, seenFinal)
		}
	}
}
