package gui

import (
	"context"
	"testing"
	"time"
)

// --- operations.go extended tests ---

func TestOperationType_Constants(t *testing.T) {
	if OperationTypeScan != "scan" {
		t.Errorf("expected 'scan', got %q", OperationTypeScan)
	}
	if OperationTypeTopology != "topology" {
		t.Errorf("expected 'topology', got %q", OperationTypeTopology)
	}
	if OperationTypeTool != "tool" {
		t.Errorf("expected 'tool', got %q", OperationTypeTool)
	}
	if OperationTypeExport != "export" {
		t.Errorf("expected 'export', got %q", OperationTypeExport)
	}
	if OperationTypeInternal != "internal" {
		t.Errorf("expected 'internal', got %q", OperationTypeInternal)
	}
}

func TestOperationStatus_Constants(t *testing.T) {
	if OperationQueued != "queued" {
		t.Errorf("expected 'queued', got %q", OperationQueued)
	}
	if OperationRunning != "running" {
		t.Errorf("expected 'running', got %q", OperationRunning)
	}
	if OperationSuccess != "success" {
		t.Errorf("expected 'success', got %q", OperationSuccess)
	}
	if OperationFailed != "failed" {
		t.Errorf("expected 'failed', got %q", OperationFailed)
	}
	if OperationCanceled != "canceled" {
		t.Errorf("expected 'canceled', got %q", OperationCanceled)
	}
}

func TestOperationsManager_RunNilTask(t *testing.T) {
	m := NewOperationsManager()
	id := m.Run(OperationTypeTool, "nil task", nil)
	if id != "" {
		t.Errorf("expected empty ID for nil task, got %q", id)
	}
}

func TestOperationsManager_GetNonExistent(t *testing.T) {
	m := NewOperationsManager()
	_, ok := m.Get("nonexistent")
	if ok {
		t.Error("expected ok=false for non-existent ID")
	}
}

func TestOperationsManager_ListEmpty(t *testing.T) {
	m := NewOperationsManager()
	list := m.List()
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d items", len(list))
	}
}

func TestOperationsManager_ListWithOps(t *testing.T) {
	m := NewOperationsManager()
	m.Run(OperationTypeTool, "task1", func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})
	m.Run(OperationTypeTool, "task2", func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})
	time.Sleep(50 * time.Millisecond)
	list := m.List()
	if len(list) < 2 {
		t.Errorf("expected at least 2 items, got %d", len(list))
	}
}

func TestOperationsManager_CancelNonExistent(t *testing.T) {
	m := NewOperationsManager()
	if m.Cancel("nonexistent") {
		t.Error("expected false for cancelling non-existent op")
	}
}

func TestOperationsManager_RetryNonExistent(t *testing.T) {
	m := NewOperationsManager()
	_, ok := m.Retry("nonexistent")
	if ok {
		t.Error("expected false for retrying non-existent op")
	}
}

func TestOperationsManager_RetryNotFailed(t *testing.T) {
	m := NewOperationsManager()
	id := m.Run(OperationTypeTool, "task", func(ctx context.Context) error {
		time.Sleep(200 * time.Millisecond)
		return nil
	})
	time.Sleep(50 * time.Millisecond)
	// Операция ещё running — retry не должен работать
	_, ok := m.Retry(id)
	if ok {
		t.Error("expected false for retrying running op")
	}
}

func TestOperationsManager_SubscribeZeroBuffer(t *testing.T) {
	m := NewOperationsManager()
	ch := m.Subscribe(0)
	if ch == nil {
		t.Fatal("expected non-nil channel")
	}
}

func TestOperationsManager_SubscribeMultiple(t *testing.T) {
	m := NewOperationsManager()
	ch1 := m.Subscribe(5)
	ch2 := m.Subscribe(5)
	if ch1 == nil || ch2 == nil {
		t.Fatal("expected non-nil channels")
	}
}

func TestOperationsManager_NextID_Sequence(t *testing.T) {
	m := NewOperationsManager()
	id1 := m.nextID()
	id2 := m.nextID()
	if id1 == id2 {
		t.Error("expected different IDs")
	}
}

func TestOperationsManager_FinishCancelled(t *testing.T) {
	m := NewOperationsManager()
	id := m.Run(OperationTypeTool, "cancellable", func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})
	time.Sleep(50 * time.Millisecond)
	m.Cancel(id)
	time.Sleep(100 * time.Millisecond)
	op, ok := m.Get(id)
	if !ok {
		t.Fatal("expected op to exist")
	}
	if op.Status != OperationCanceled {
		t.Errorf("expected canceled, got %q", op.Status)
	}
}

func TestOperationsManager_FinishFailed(t *testing.T) {
	m := NewOperationsManager()
	id := m.Run(OperationTypeTool, "failing", func(ctx context.Context) error {
		return context.DeadlineExceeded
	})
	time.Sleep(100 * time.Millisecond)
	op, ok := m.Get(id)
	if !ok {
		t.Fatal("expected op to exist")
	}
	if op.Status != OperationFailed {
		t.Errorf("expected failed, got %q", op.Status)
	}
	if op.Error == "" {
		t.Error("expected non-empty error message")
	}
}

func TestOperationsManager_FinishSuccess(t *testing.T) {
	m := NewOperationsManager()
	id := m.Run(OperationTypeTool, "success", func(ctx context.Context) error {
		return nil
	})
	time.Sleep(100 * time.Millisecond)
	op, ok := m.Get(id)
	if !ok {
		t.Fatal("expected op to exist")
	}
	if op.Status != OperationSuccess {
		t.Errorf("expected success, got %q", op.Status)
	}
	if op.CanRetry {
		t.Error("expected CanRetry=false for success")
	}
}

func TestOperationsManager_RetryAfterFailure(t *testing.T) {
	m := NewOperationsManager()
	callCount := 0
	id := m.Run(OperationTypeTool, "retryable", func(ctx context.Context) error {
		callCount++
		if callCount == 1 {
			return context.DeadlineExceeded
		}
		return nil
	})
	time.Sleep(100 * time.Millisecond)
	newID, ok := m.Retry(id)
	if !ok {
		t.Fatal("expected retry to succeed")
	}
	if newID == "" {
		t.Error("expected non-empty new ID")
	}
	if newID == id {
		t.Error("expected different ID for retry")
	}
}

func TestOperationsManager_CancelAlreadyFinished(t *testing.T) {
	m := NewOperationsManager()
	id := m.Run(OperationTypeTool, "quick", func(ctx context.Context) error {
		return nil
	})
	time.Sleep(100 * time.Millisecond)
	// Already finished — cancel should return false
	if m.Cancel(id) {
		t.Error("expected false for cancelling finished op")
	}
}

func TestOperationsManager_OperationDuration(t *testing.T) {
	m := NewOperationsManager()
	id := m.Run(OperationTypeTool, "timed", func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})
	time.Sleep(150 * time.Millisecond)
	op, ok := m.Get(id)
	if !ok {
		t.Fatal("expected op to exist")
	}
	if op.Duration <= 0 {
		t.Error("expected positive duration")
	}
}

func TestOperationsManager_OperationFields(t *testing.T) {
	m := NewOperationsManager()
	id := m.Run(OperationTypeScan, "test scan", func(ctx context.Context) error {
		return nil
	})
	time.Sleep(100 * time.Millisecond)
	op, ok := m.Get(id)
	if !ok {
		t.Fatal("expected op to exist")
	}
	if op.ID != id {
		t.Errorf("expected ID %q, got %q", id, op.ID)
	}
	if op.Type != OperationTypeScan {
		t.Errorf("expected type %q, got %q", OperationTypeScan, op.Type)
	}
	if op.Title != "test scan" {
		t.Errorf("expected title 'test scan', got %q", op.Title)
	}
	if op.StartedAt.IsZero() {
		t.Error("expected non-zero StartedAt")
	}
	if op.FinishedAt.IsZero() {
		t.Error("expected non-zero FinishedAt")
	}
}
