package gui

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type OperationType string

const (
	OperationTypeScan     OperationType = "scan"
	OperationTypeTopology OperationType = "topology"
	OperationTypeTool     OperationType = "tool"
	OperationTypeExport   OperationType = "export"
	OperationTypeInternal OperationType = "internal"
)

type OperationStatus string

const (
	OperationQueued   OperationStatus = "queued"
	OperationRunning  OperationStatus = "running"
	OperationSuccess  OperationStatus = "success"
	OperationFailed   OperationStatus = "failed"
	OperationCanceled OperationStatus = "canceled"
)

type Operation struct {
	ID         string
	Type       OperationType
	Title      string
	Status     OperationStatus
	StartedAt  time.Time
	FinishedAt time.Time
	Duration   time.Duration
	Error      string
	CanRetry   bool
	CanCancel  bool
}

type OperationTask func(context.Context) error

type OperationEvent struct {
	Operation Operation
}

type managedOp struct {
	Operation
	cancel context.CancelFunc
	task   OperationTask
}

type OperationsManager struct {
	mu          sync.RWMutex
	ops         map[string]*managedOp
	subscribers []chan OperationEvent
	sequence    uint64
}

func NewOperationsManager() *OperationsManager {
	return &OperationsManager{
		ops: make(map[string]*managedOp),
	}
}

func (m *OperationsManager) Run(opType OperationType, title string, task OperationTask) string {
	if task == nil {
		return ""
	}
	id := m.nextID()
	ctx, cancel := context.WithCancel(context.Background())
	op := &managedOp{
		Operation: Operation{
			ID:        id,
			Type:      opType,
			Title:     title,
			Status:    OperationQueued,
			CanRetry:  false,
			CanCancel: true,
		},
		cancel: cancel,
		task:   task,
	}

	m.mu.Lock()
	m.ops[id] = op
	m.mu.Unlock()
	m.notify(op.Operation)

	go m.execute(ctx, op)
	return id
}

func (m *OperationsManager) Retry(id string) (string, bool) {
	m.mu.RLock()
	op, ok := m.ops[id]
	m.mu.RUnlock()
	if !ok || op == nil || op.task == nil {
		return "", false
	}
	if op.Status != OperationFailed && op.Status != OperationCanceled {
		return "", false
	}
	return m.Run(op.Type, op.Title, op.task), true
}

func (m *OperationsManager) Cancel(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	op, ok := m.ops[id]
	if !ok || op == nil || op.cancel == nil {
		return false
	}
	if op.Status == OperationSuccess || op.Status == OperationFailed || op.Status == OperationCanceled {
		return false
	}
	op.cancel()
	return true
}

func (m *OperationsManager) Get(id string) (Operation, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	op, ok := m.ops[id]
	if !ok || op == nil {
		return Operation{}, false
	}
	return op.Operation, true
}

func (m *OperationsManager) List() []Operation {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Operation, 0, len(m.ops))
	for _, op := range m.ops {
		out = append(out, op.Operation)
	}
	return out
}

func (m *OperationsManager) Subscribe(buffer int) <-chan OperationEvent {
	if buffer < 1 {
		buffer = 1
	}
	ch := make(chan OperationEvent, buffer)
	m.mu.Lock()
	m.subscribers = append(m.subscribers, ch)
	m.mu.Unlock()
	return ch
}

func (m *OperationsManager) execute(ctx context.Context, op *managedOp) {
	m.setRunning(op.ID)
	err := op.task(ctx)
	m.finish(op.ID, err, ctx.Err())
}

func (m *OperationsManager) setRunning(id string) {
	m.mu.Lock()
	op, ok := m.ops[id]
	if !ok || op == nil {
		m.mu.Unlock()
		return
	}
	op.Status = OperationRunning
	op.StartedAt = time.Now()
	op.CanCancel = true
	op.CanRetry = false
	snapshot := op.Operation
	m.mu.Unlock()
	m.notify(snapshot)
}

func (m *OperationsManager) finish(id string, err error, ctxErr error) {
	m.mu.Lock()
	op, ok := m.ops[id]
	if !ok || op == nil {
		m.mu.Unlock()
		return
	}
	op.FinishedAt = time.Now()
	if !op.StartedAt.IsZero() {
		op.Duration = op.FinishedAt.Sub(op.StartedAt)
	}
	op.CanCancel = false
	op.CanRetry = true
	switch {
	case ctxErr == context.Canceled:
		op.Status = OperationCanceled
		op.Error = ""
	case err != nil:
		op.Status = OperationFailed
		op.Error = err.Error()
	default:
		op.Status = OperationSuccess
		op.Error = ""
		op.CanRetry = false
	}
	snapshot := op.Operation
	m.mu.Unlock()
	m.notify(snapshot)
}

func (m *OperationsManager) nextID() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sequence++
	return fmt.Sprintf("op-%d", m.sequence)
}

func (m *OperationsManager) notify(op Operation) {
	m.mu.RLock()
	subs := append([]chan OperationEvent(nil), m.subscribers...)
	m.mu.RUnlock()
	event := OperationEvent{Operation: op}
	for _, ch := range subs {
		select {
		case ch <- event:
		default:
		}
	}
}
