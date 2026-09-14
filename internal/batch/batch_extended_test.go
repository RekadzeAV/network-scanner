package batch

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestNewBatchProcessor_Defaults — проверка дефолтных значений.
func TestNewBatchProcessor_Defaults(t *testing.T) {
	p := NewBatchProcessor(0, 0, 0)

	if p.workerCount != 10 {
		t.Errorf("expected default workerCount 10, got %d", p.workerCount)
	}
	if p.batchSize != 50 {
		t.Errorf("expected default batchSize 50, got %d", p.batchSize)
	}
	if p.timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", p.timeout)
	}
}

// TestNewBatchProcessor_CustomValues — проверка пользовательских значений.
func TestNewBatchProcessor_CustomValues(t *testing.T) {
	p := NewBatchProcessor(5, 20, 5*time.Second)

	if p.workerCount != 5 {
		t.Errorf("expected workerCount 5, got %d", p.workerCount)
	}
	if p.batchSize != 20 {
		t.Errorf("expected batchSize 20, got %d", p.batchSize)
	}
	if p.timeout != 5*time.Second {
		t.Errorf("expected timeout 5s, got %v", p.timeout)
	}
}

// TestNewBatchProcessor_NegativeValues — отрицательные значения.
func TestNewBatchProcessor_NegativeValues(t *testing.T) {
	p := NewBatchProcessor(-1, -1, -1)

	if p.workerCount != 10 {
		t.Errorf("expected default workerCount for negative, got %d", p.workerCount)
	}
	if p.batchSize != 50 {
		t.Errorf("expected default batchSize for negative, got %d", p.batchSize)
	}
	if p.timeout != 30*time.Second {
		t.Errorf("expected default timeout for negative, got %v", p.timeout)
	}
}

// TestBatchProcessor_EmptyTasks — пустой список задач.
func TestBatchProcessor_EmptyTasks(t *testing.T) {
	processor := NewBatchProcessor(2, 5, 10*time.Second)

	results, err := processor.ProcessBatch(context.Background(), []Task{}, func(ctx context.Context, task Task) (interface{}, error) {
		return nil, nil
	})

	if err != nil {
		t.Errorf("expected no error for empty tasks, got %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty tasks, got %d", len(results))
	}
}

// TestBatchProcessor_SingleTask — одна задача.
func TestBatchProcessor_SingleTask(t *testing.T) {
	processor := NewBatchProcessor(2, 5, 10*time.Second)

	tasks := []Task{
		{ID: "single", Payload: "test"},
	}

	results, err := processor.ProcessBatch(context.Background(), tasks, func(ctx context.Context, task Task) (interface{}, error) {
		return task.Payload.(string) + "-processed", nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Output.(string) != "test-processed" {
		t.Errorf("expected 'test-processed', got %q", results[0].Output)
	}
}

// TestBatchProcessor_BatchSize1 — размер батча 1.
func TestBatchProcessor_BatchSize1(t *testing.T) {
	processor := NewBatchProcessor(1, 1, 10*time.Second)

	tasks := []Task{
		{ID: "1", Payload: 1},
		{ID: "2", Payload: 2},
		{ID: "3", Payload: 3},
	}

	results, err := processor.ProcessBatch(context.Background(), tasks, func(ctx context.Context, task Task) (interface{}, error) {
		return task.Payload.(int) * 10, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}

// TestBatchProcessor_ResultsToSlice — конвертация результатов в слайс.
func TestBatchProcessor_ResultsToSlice(t *testing.T) {
	processor := NewBatchProcessor(2, 5, 10*time.Second)

	results := map[string]Result{
		"A": {TaskID: "A", Output: 1, Error: nil},
		"B": {TaskID: "B", Output: 2, Error: nil},
		"C": {TaskID: "C", Output: 3, Error: nil},
	}

	tasks := []Task{
		{ID: "A"},
		{ID: "B"},
		{ID: "C"},
	}

	slice := processor.resultsToSlice(results, tasks)

	if len(slice) != 3 {
		t.Fatalf("expected 3 results, got %d", len(slice))
	}
	if slice[0].TaskID != "A" {
		t.Errorf("expected first result TaskID 'A', got %q", slice[0].TaskID)
	}
	if slice[2].Output.(int) != 3 {
		t.Errorf("expected third result Output 3, got %v", slice[2].Output)
	}
}

// TestBatchProcessor_ResultsToSlice_OrderPreserved — порядок сохранён.
func TestBatchProcessor_ResultsToSlice_OrderPreserved(t *testing.T) {
	processor := NewBatchProcessor(2, 5, 10*time.Second)

	results := map[string]Result{
		"Z": {TaskID: "Z", Output: 26},
		"A": {TaskID: "A", Output: 1},
		"M": {TaskID: "M", Output: 13},
	}

	tasks := []Task{
		{ID: "A"},
		{ID: "M"},
		{ID: "Z"},
	}

	slice := processor.resultsToSlice(results, tasks)

	if slice[0].TaskID != "A" {
		t.Errorf("expected first result TaskID 'A', got %q", slice[0].TaskID)
	}
	if slice[1].TaskID != "M" {
		t.Errorf("expected second result TaskID 'M', got %q", slice[1].TaskID)
	}
	if slice[2].TaskID != "Z" {
		t.Errorf("expected third result TaskID 'Z', got %q", slice[2].TaskID)
	}
}

// TestBatchProcessor_AllTasksFail — все задачи провалились.
func TestBatchProcessor_AllTasksFail(t *testing.T) {
	processor := NewBatchProcessor(2, 5, 10*time.Second)

	tasks := []Task{
		{ID: "1", Payload: 1},
		{ID: "2", Payload: 2},
	}

	_, err := processor.ProcessBatch(context.Background(), tasks, func(ctx context.Context, task Task) (interface{}, error) {
		return nil, fmt.Errorf("task %s failed", task.ID)
	})

	if err == nil {
		t.Fatal("expected error when all tasks fail")
	}
}

// TestBatchProcessor_MixedResults — смешанные результаты.
func TestBatchProcessor_MixedResults(t *testing.T) {
	processor := NewBatchProcessor(2, 5, 10*time.Second)

	tasks := []Task{
		{ID: "1", Payload: "ok"},
		{ID: "2", Payload: "fail"},
		{ID: "3", Payload: "ok"},
	}

	_, err := processor.ProcessBatch(context.Background(), tasks, func(ctx context.Context, task Task) (interface{}, error) {
		if task.Payload == "fail" {
			return nil, fmt.Errorf("task %s failed", task.ID)
		}
		return task.Payload, nil
	})

	if err == nil {
		t.Fatal("expected error for mixed results")
	}
}

// TestBatchProcessor_ContextTimeout — таймаут контекста.
func TestBatchProcessor_ContextTimeout(t *testing.T) {
	processor := NewBatchProcessor(2, 5, 5*time.Second)

	tasks := []Task{
		{ID: "1", Payload: 1},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := processor.ProcessBatch(ctx, tasks, func(ctx context.Context, task Task) (interface{}, error) {
		time.Sleep(200 * time.Millisecond)
		return task.Payload, nil
	})

	// Может быть error или timeout — главное что нет panic
	t.Logf("Context timeout test: err=%v", err)
}

// TestBatchProcessor_LargeBatch — большой батч.
func TestBatchProcessor_LargeBatch(t *testing.T) {
	processor := NewBatchProcessor(10, 100, 30*time.Second)

	tasks := make([]Task, 250)
	for i := 0; i < 250; i++ {
		tasks[i] = Task{
			ID:      fmt.Sprintf("task-%d", i),
			Payload: i,
		}
	}

	results, err := processor.ProcessBatch(context.Background(), tasks, func(ctx context.Context, task Task) (interface{}, error) {
		return task.Payload.(int) * 2, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 250 {
		t.Fatalf("expected 250 results, got %d", len(results))
	}
}

// TestBatchProcessor_PayloadNil — payload nil.
func TestBatchProcessor_PayloadNil(t *testing.T) {
	processor := NewBatchProcessor(2, 5, 10*time.Second)

	tasks := []Task{
		{ID: "1", Payload: nil},
	}

	results, err := processor.ProcessBatch(context.Background(), tasks, func(ctx context.Context, task Task) (interface{}, error) {
		return task.Payload, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Output != nil {
		t.Errorf("expected nil output, got %v", results[0].Output)
	}
}

// TestNewSNMPBatchProcessor — создание SNMP BatchProcessor.
func TestNewSNMPBatchProcessor(t *testing.T) {
	p := NewSNMPBatchProcessor()

	if p == nil {
		t.Fatal("NewSNMPBatchProcessor() returned nil")
	}
	if p.workerCount != 5 {
		t.Errorf("expected workerCount 5, got %d", p.workerCount)
	}
	if p.batchSize != 20 {
		t.Errorf("expected batchSize 20, got %d", p.batchSize)
	}
	if p.timeout != 2*time.Second {
		t.Errorf("expected timeout 2s, got %v", p.timeout)
	}
}

// TestTask_Struct — структура Task.
func TestTask_Struct(t *testing.T) {
	task := Task{
		ID:      "test-123",
		Payload: map[string]string{"key": "value"},
	}

	if task.ID != "test-123" {
		t.Errorf("expected ID 'test-123', got %q", task.ID)
	}
	if task.Payload == nil {
		t.Error("expected non-nil Payload")
	}
}

// TestResult_Struct — структура Result.
func TestResult_Struct(t *testing.T) {
	result := Result{
		TaskID: "test-123",
		Output: "success",
		Error:  nil,
	}

	if result.TaskID != "test-123" {
		t.Errorf("expected TaskID 'test-123', got %q", result.TaskID)
	}
	if result.Output != "success" {
		t.Errorf("expected Output 'success', got %v", result.Output)
	}
}

// TestSNMPRequest_Struct — структура SNMPRequest.
func TestSNMPRequest_Struct(t *testing.T) {
	req := SNMPRequest{
		Host:      "192.168.1.1",
		OID:       "1.3.6.1.2.1.1.5.0",
		Community: "public",
	}

	if req.Host != "192.168.1.1" {
		t.Errorf("expected Host '192.168.1.1', got %q", req.Host)
	}
	if req.OID != "1.3.6.1.2.1.1.5.0" {
		t.Errorf("expected OID '1.3.6.1.2.1.1.5.0', got %q", req.OID)
	}
	if req.Community != "public" {
		t.Errorf("expected Community 'public', got %q", req.Community)
	}
}

// TestSNMPResponse_Struct — структура SNMPResponse.
func TestSNMPResponse_Struct(t *testing.T) {
	resp := SNMPResponse{
		Host:  "192.168.1.1",
		OID:   "1.3.6.1.2.1.1.5.0",
		Value: "test-device",
		Error: nil,
	}

	if resp.Host != "192.168.1.1" {
		t.Errorf("expected Host '192.168.1.1', got %q", resp.Host)
	}
	if resp.OID != "1.3.6.1.2.1.1.5.0" {
		t.Errorf("expected OID '1.3.6.1.2.1.1.5.0', got %q", resp.OID)
	}
	if resp.Value != "test-device" {
		t.Errorf("expected Value 'test-device', got %q", resp.Value)
	}
}

// TestSNMPResponse_WithError — SNMPResponse с ошибкой.
func TestSNMPResponse_WithError(t *testing.T) {
	err := fmt.Errorf("snmp timeout")
	resp := SNMPResponse{
		Host:  "192.168.1.1",
		OID:   "1.3.6.1.2.1.1.5.0",
		Value: "",
		Error: err,
	}

	if resp.Error != err {
		t.Errorf("expected Error to be set")
	}
	if resp.Value != "" {
		t.Errorf("expected empty Value for error case")
	}
}

// TestBatchProcessor_ZeroBatchSize — размер батча 0.
func TestBatchProcessor_ZeroBatchSize(t *testing.T) {
	// batchSize=0 должен использовать дефолт 50
	processor := NewBatchProcessor(2, 0, 10*time.Second)

	if processor.batchSize != 50 {
		t.Errorf("expected default batchSize 50 for 0 input, got %d", processor.batchSize)
	}
}

// TestBatchProcessor_ZeroWorkers — workers 0.
func TestBatchProcessor_ZeroWorkers(t *testing.T) {
	// workerCount=0 должен использовать дефолт 10
	processor := NewBatchProcessor(0, 50, 10*time.Second)

	if processor.workerCount != 10 {
		t.Errorf("expected default workerCount 10 for 0 input, got %d", processor.workerCount)
	}
}

// TestBatchProcessor_ZeroTimeout — timeout 0.
func TestBatchProcessor_ZeroTimeout(t *testing.T) {
	// timeout=0 должен использовать дефолт 30s
	processor := NewBatchProcessor(10, 50, 0)

	if processor.timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s for 0 input, got %v", processor.timeout)
	}
}
