package batch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gosnmp/gosnmp"
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

// SNMPClient интерфейс для SNMP операций
type SNMPClient interface {
	Get(oid string) (string, error)
	Close() error
}

// gosnmpClient обёртка над gosnmp
type gosnmpClient struct {
	conn *gosnmp.GoSNMP
}

// NewSNMPClient создаёт новый SNMP клиент
func NewSNMPClient(host, community string, timeout time.Duration) (SNMPClient, error) {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	c := &gosnmp.GoSNMP{
		Target:    host,
		Port:      161,
		Community: community,
		Version:   gosnmp.Version2c,
		Timeout:   timeout,
		Retries:   2,
	}
	if err := c.Connect(); err != nil {
		return nil, fmt.Errorf("snmp connect to %s: %w", host, err)
	}
	return &gosnmpClient{conn: c}, nil
}

// Get выполняет SNMP GET запрос
func (g *gosnmpClient) Get(oid string) (string, error) {
	result, err := g.conn.Get([]string{oid})
	if err != nil {
		return "", fmt.Errorf("snmp get %s: %w", oid, err)
	}
	if len(result.Variables) == 0 {
		return "", fmt.Errorf("empty response for %s", oid)
	}
	pdu := result.Variables[0]
	switch v := pdu.Value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// Close закрывает SNMP соединение
func (g *gosnmpClient) Close() error {
	if g.conn != nil && g.conn.Conn != nil {
		return g.conn.Conn.Close()
	}
	return nil
}

// SNMPBatchProcessor специализированный процессор для SNMP
type SNMPBatchProcessor struct {
	BatchProcessor
	timeout time.Duration
}

// NewSNMPBatchProcessor создаёт SNMP BatchProcessor
func NewSNMPBatchProcessor() *SNMPBatchProcessor {
	return &SNMPBatchProcessor{
		BatchProcessor: *NewBatchProcessor(5, 20, 10*time.Second),
		timeout:        2 * time.Second,
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

// ProcessSNMPBatch выполняет батч SNMP запросов с реальными SNMP запросами
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

		// Создаём SNMP клиент
		client, err := NewSNMPClient(req.Host, req.Community, p.timeout)
		if err != nil {
			return SNMPResponse{
				Host:  req.Host,
				OID:   req.OID,
				Error: err,
			}, nil
		}
		defer client.Close()

		// Выполняем SNMP GET запрос
		value, err := client.Get(req.OID)
		if err != nil {
			return SNMPResponse{
				Host:  req.Host,
				OID:   req.OID,
				Error: err,
			}, nil
		}

		return SNMPResponse{
			Host:  req.Host,
			OID:   req.OID,
			Value: value,
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
