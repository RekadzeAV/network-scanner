package batch

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BatchProcessor выполняет задачи батчами с ограничением параллелизма
type BatchProcessor struct {
	workerCount int
	batchSize   int
	timeout     time.Duration
}

// NewBatchProcessor создаёт новый BatchProcessor
func NewBatchProcessor(workerCount, batchSize int, timeout time.Duration) *BatchProcessor {
	if workerCount <= 0 {
		workerCount = 10
	}
	if batchSize <= 0 {
		batchSize = 50
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &BatchProcessor{
		workerCount: workerCount,
		batchSize:   batchSize,
		timeout:     timeout,
	}
}

// Task представляет задачу для выполнения
type Task struct {
	ID      string
	Payload interface{}
}

// Result представляет результат выполнения задачи
type Result struct {
	TaskID string
	Output interface{}
	Error  error
}

// ProcessBatch выполняет батч задач
func (p *BatchProcessor) ProcessBatch(ctx context.Context, tasks []Task, fn func(ctx context.Context, task Task) (interface{}, error)) ([]Result, error) {
	results := make(map[string]Result)
	var mu sync.Mutex

	// Разбиваем на батчи
	for i := 0; i < len(tasks); i += p.batchSize {
		end := i + p.batchSize
		if end > len(tasks) {
			end = len(tasks)
		}
		batch := tasks[i:end]

		// Выполняем батч параллельно
		batchCtx, cancel := context.WithTimeout(ctx, p.timeout)
		errChan := make(chan error, len(batch))

		var wg sync.WaitGroup
		for _, task := range batch {
			wg.Add(1)
			go func(t Task) {
				defer wg.Done()

				output, err := fn(batchCtx, t)
				mu.Lock()
				results[t.ID] = Result{
					TaskID: t.ID,
					Output: output,
					Error:  err,
				}
				mu.Unlock()

				if err != nil {
					errChan <- err
				}
			}(task)
		}

		wg.Wait()
		cancel()

		// Проверяем ошибки батча
		close(errChan)
		for err := range errChan {
			if err != nil {
				return p.resultsToSlice(results, tasks), fmt.Errorf("batch error: %w", err)
			}
		}
	}

	return p.resultsToSlice(results, tasks), nil
}

func (p *BatchProcessor) resultsToSlice(results map[string]Result, tasks []Task) []Result {
	slice := make([]Result, len(tasks))
	for i, task := range tasks {
		slice[i] = results[task.ID]
	}
	return slice
}

// SNMPBatchProcessor специализированный процессор для SNMP
type SNMPBatchProcessor struct {
	BatchProcessor
}

// NewSNMPBatchProcessor создаёт SNMP BatchProcessor
func NewSNMPBatchProcessor() *SNMPBatchProcessor {
	return &SNMPBatchProcessor{
		BatchProcessor: *NewBatchProcessor(5, 20, 10*time.Second),
	}
}

// SNMPRequest представляет SNMP запрос
type SNMPRequest struct {
	Host      string
	OID       string
	Community string
}

// SNMPResponse представляет SNMP ответ
type SNMPResponse struct {
	Host  string
	OID   string
	Value string
	Error error
}

// ProcessSNMPBatch выполняет батч SNMP запросов
func (p *SNMPBatchProcessor) ProcessSNMPBatch(ctx context.Context, requests []SNMPRequest) []SNMPResponse {
	tasks := make([]Task, len(requests))
	for i, req := range requests {
		tasks[i] = Task{
			ID:      fmt.Sprintf("snmp-%d", i),
			Payload: req,
		}
	}

	results, _ := p.ProcessBatch(ctx, tasks, func(ctx context.Context, task Task) (interface{}, error) {
		req := task.Payload.(SNMPRequest)
		// TODO: реальный SNMP запрос
		return SNMPResponse{
			Host:  req.Host,
			OID:   req.OID,
			Value: "stub",
		}, nil
	})

	responses := make([]SNMPResponse, len(results))
	for i, r := range results {
		if r.Output != nil {
			responses[i] = r.Output.(SNMPResponse)
		} else {
			responses[i] = SNMPResponse{
				Host:  requests[i].Host,
				OID:   requests[i].OID,
				Error: r.Error,
			}
		}
	}

	return responses
}
