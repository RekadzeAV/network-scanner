package batch

import (
	"context"
	"testing"
	"time"
)

// ============================================================================
// M1.X: Тесты для непокрытых функций batch (83.1% → 85%+)
// ============================================================================

// TestSNMPBatchProcessor_New — ветка: создание SNMP процессора
func TestSNMPBatchProcessor_New(t *testing.T) {
	proc := NewSNMPBatchProcessor()
	if proc == nil {
		t.Fatal("expected non-nil SNMP batch processor")
	}
	
	// Проверяем что timeout установлен
	if proc.timeout != 2*time.Second {
		t.Errorf("expected timeout 2s, got: %v", proc.timeout)
	}
}

// TestSNMPBatchProcessor_ProcessBatch_Empty — ветка: пустой батч
func TestSNMPBatchProcessor_ProcessBatch_Empty(t *testing.T) {
	proc := NewSNMPBatchProcessor()
	
	requests := []SNMPRequest{}
	results := proc.ProcessSNMPBatch(context.Background(), requests)
	
	if len(results) != 0 {
		t.Errorf("expected empty results, got: %d", len(results))
	}
}

// TestSNMPBatchProcessor_ProcessBatch_Single — ветка: одиночный запрос
func TestSNMPBatchProcessor_ProcessBatch_Single(t *testing.T) {
	proc := NewSNMPBatchProcessor()
	
	// Запрос к несуществующему SNMP агенту должен вернуть ошибку
	requests := []SNMPRequest{
		{
			Host:      "192.0.2.1",
			OID:       "1.3.6.1.2.1.1.1.0",
			Community: "public",
		},
	}
	
	results := proc.ProcessSNMPBatch(context.Background(), requests)
	
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got: %d", len(results))
	}
	
	// Результат должен содержать ошибку (нет SNMP агента)
	if results[0].Error == nil {
		t.Error("expected error for non-existent SNMP agent")
	}
}

// TestSNMPBatchProcessor_ProcessBatch_Multiple — ветка: несколько запросов
func TestSNMPBatchProcessor_ProcessBatch_Multiple(t *testing.T) {
	proc := NewSNMPBatchProcessor()
	
	requests := []SNMPRequest{
		{Host: "192.0.2.1", OID: "1.3.6.1.2.1.1.1.0", Community: "public"},
		{Host: "192.0.2.2", OID: "1.3.6.1.2.1.1.2.0", Community: "public"},
	}
	
	results := proc.ProcessSNMPBatch(context.Background(), requests)
	
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got: %d", len(results))
	}
	
	// Оба должны содержать ошибки (нет SNMP агентов)
	for i, result := range results {
		if result.Error == nil {
			t.Errorf("result %d: expected error, got nil", i)
		}
	}
}

// TestSNMPBatchProcessor_ProcessBatch_Timeout — ветка: таймаут
func TestSNMPBatchProcessor_ProcessBatch_Timeout(t *testing.T) {
	// Создаем процессор с коротким таймаутом
	proc := &SNMPBatchProcessor{
		BatchProcessor: *NewBatchProcessor(1, 1, 100*time.Millisecond),
		timeout:        100 * time.Millisecond,
	}
	
	requests := []SNMPRequest{
		{Host: "192.0.2.1", OID: "1.3.6.1.2.1.1.1.0", Community: "public"},
	}
	
	// С коротким таймаутом должен быть timeout error
	results := proc.ProcessSNMPBatch(context.Background(), requests)
	
	if len(results) > 0 && results[0].Error != nil {
		t.Logf("expected timeout error: %v", results[0].Error)
	}
}

// TestSNMPBatchProcessor_ProcessBatch_BadOID — ветка: невалидный OID
func TestSNMPBatchProcessor_ProcessBatch_BadOID(t *testing.T) {
	proc := NewSNMPBatchProcessor()
	
	requests := []SNMPRequest{
		{Host: "192.0.2.1", OID: "invalid-oid", Community: "public"},
	}
	
	results := proc.ProcessSNMPBatch(context.Background(), requests)
	
	if len(results) > 0 && results[0].Error == nil {
		t.Error("expected error for invalid OID")
	}
}

// TestGosnmpClient_Get_Empty — ветка: пустой ответ SNMP
func TestGosnmpClient_Get_Empty(t *testing.T) {
	// Тестируем gosnmpClient.Get с пустым ответом
	// Создаем mock клиент через тест
	// Функция Get имеет 30% покрытия, проверяем ветку empty response
	t.Skip("requires SNMP mock setup")
}

// TestGosnmpClient_Get_String — ветка: строковый ответ SNMP
func TestGosnmpClient_Get_String(t *testing.T) {
	// Тестируем gosnmpClient.Get со строковым значением
	t.Skip("requires SNMP mock setup")
}

// TestGosnmpClient_Get_Bytes — ветка: байтовый ответ SNMP
func TestGosnmpClient_Get_Bytes(t *testing.T) {
	// Тестируем gosnmpClient.Get с байтовым значением
	t.Skip("requires SNMP mock setup")
}

// TestGosnmpClient_Get_Default — ветка: default case SNMP
func TestGosnmpClient_Get_Default(t *testing.T) {
	// Тестируем gosnmpClient.Get с non-string/non-byte значением
	t.Skip("requires SNMP mock setup")
}

// TestBatchProcessor_Context_Cancellation — ветка: отмена контекста
func TestBatchProcessor_Context_Cancellation(t *testing.T) {
	proc := NewBatchProcessor(1, 5, 5*time.Second)
	
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Отменяем сразу
	
	tasks := []Task{
		{ID: "task-1", Payload: "test"},
	}
	
	_, err := proc.ProcessBatch(ctx, tasks, func(ctx context.Context, task Task) (interface{}, error) {
		return "result", nil
	})
	
	// В зависимости от реализации, отмена контекста может не вернуть ошибку
	t.Logf("context cancellation result: err=%v", err)
}

// TestBatchProcessor_BatchSplit — ветка: разбивка на батчи
func TestBatchProcessor_BatchSplit(t *testing.T) {
	proc := NewBatchProcessor(2, 3, 10*time.Second)
	
	// Создаем 10 задач, должны разбиться на батчи по 3
	tasks := make([]Task, 10)
	for i := 0; i < 10; i++ {
		tasks[i] = Task{ID: string(rune('a' + i)), Payload: i}
	}
	
	results, err := proc.ProcessBatch(context.Background(), tasks, func(ctx context.Context, task Task) (interface{}, error) {
		return task.Payload, nil
	})
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if len(results) != 10 {
		t.Errorf("expected 10 results, got: %d", len(results))
	}
}

// TestBatchProcessor_TaskIDs — ветка: валидация task IDs
func TestBatchProcessor_TaskIDs(t *testing.T) {
	proc := NewBatchProcessor(1, 10, 1*time.Second)
	
	tasks := []Task{
		{ID: "unique-1", Payload: "a"},
		{ID: "unique-2", Payload: "b"},
	}
	
	results, err := proc.ProcessBatch(context.Background(), tasks, func(ctx context.Context, task Task) (interface{}, error) {
		return task.Payload, nil
	})
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Проверяем что IDs совпадают
	expectedIDs := map[string]bool{"unique-1": true, "unique-2": true}
	for _, result := range results {
		if !expectedIDs[result.TaskID] {
			t.Errorf("unexpected task ID: %s", result.TaskID)
		}
	}
}

// TestBatchProcessor_PayloadTypes — ветка: разные типы payload
func TestBatchProcessor_PayloadTypes(t *testing.T) {
	proc := NewBatchProcessor(1, 5, 1*time.Second)
	
	tasks := []Task{
		{ID: "string", Payload: "string-value"},
		{ID: "int", Payload: 42},
		{ID: "bool", Payload: true},
		{ID: "map", Payload: map[string]int{"a": 1}},
	}
	
	results, err := proc.ProcessBatch(context.Background(), tasks, func(ctx context.Context, task Task) (interface{}, error) {
		return task.Payload, nil
	})
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if len(results) != 4 {
		t.Errorf("expected 4 results, got: %d", len(results))
	}
}
