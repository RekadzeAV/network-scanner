package batch

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestBatchProcessor_ProcessBatch(t *testing.T) {
	processor := NewBatchProcessor(2, 5, 10*time.Second)

	tasks := make([]Task, 10)
	for i := 0; i < 10; i++ {
		tasks[i] = Task{
			ID:      string(rune('A' + i)),
			Payload: i,
		}
	}

	results, err := processor.ProcessBatch(context.Background(), tasks, func(ctx context.Context, task Task) (interface{}, error) {
		time.Sleep(50 * time.Millisecond)
		return task.Payload.(int) * 2, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 10 {
		t.Fatalf("expected 10 results, got %d", len(results))
	}

	for i, r := range results {
		expected := i * 2
		if r.Output.(int) != expected {
			t.Fatalf("task %s: expected %d, got %d", r.TaskID, expected, r.Output)
		}
	}
}

func TestBatchProcessor_ContextCancel(t *testing.T) {
	processor := NewBatchProcessor(2, 5, 5*time.Second)

	tasks := make([]Task, 5)
	for i := 0; i < 5; i++ {
		tasks[i] = Task{
			ID:      string(rune('A' + i)),
			Payload: i,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Отменяем сразу

	// Результат: ошибки могут быть или не быть, главное что процесс завершится
	_, _ = processor.ProcessBatch(ctx, tasks, func(ctx context.Context, task Task) (interface{}, error) {
		time.Sleep(100 * time.Millisecond)
		return task.Payload.(int), nil
	})
	// Тест проверяет что нет panic и deadlock
	t.Log("Context cancellation test passed (no panic)")
}

func TestBatchProcessor_TaskError(t *testing.T) {
	processor := NewBatchProcessor(2, 5, 10*time.Second)

	tasks := make([]Task, 3)
	for i := 0; i < 3; i++ {
		tasks[i] = Task{
			ID:      string(rune('A' + i)),
			Payload: i,
		}
	}

	results, err := processor.ProcessBatch(context.Background(), tasks, func(ctx context.Context, task Task) (interface{}, error) {
		if task.ID == "B" {
			return nil, fmt.Errorf("task B failed")
		}
		return task.Payload.(int), nil
	})

	// Ошибка должна быть, но результаты должны быть собраны
	if err == nil {
		t.Fatal("expected error for task B")
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}

func TestSNMPBatchProcessor(t *testing.T) {
	processor := NewSNMPBatchProcessor()

	requests := []SNMPRequest{
		{Host: "192.168.1.1", OID: "1.3.6.1.2.1.1.5.0", Community: "public"},
		{Host: "192.168.1.2", OID: "1.3.6.1.2.1.1.5.0", Community: "public"},
	}

	responses := processor.ProcessSNMPBatch(context.Background(), requests)

	if len(responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(responses))
	}

	for i, r := range responses {
		if r.Value != "stub" {
			t.Fatalf("response %d: expected stub value, got %s", i, r.Value)
		}
	}
}
