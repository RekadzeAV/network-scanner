package commands

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// WithRecovery — превращает панику обработчика в ошибку, не роняя процесс.
func WithRecovery() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req Request) (resp Response, err error) {
			defer func() {
				if r := recover(); r != nil {
					resp = Response{}
					err = fmt.Errorf("command %q panicked: %v", req.Command, r)
				}
			}()
			return next(ctx, req)
		}
	}
}

// WithTimeout — ограничивает время выполнения команды.
// Отмена контекста не прерывает обработчики, которые его не проверяют,
// поэтому таймаут носит характер соглашения для корректных handlers.
func WithTimeout(timeout time.Duration) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req Request) (Response, error) {
			if timeout <= 0 {
				return next(ctx, req)
			}
			if ctx == nil {
				ctx = context.Background()
			}
			runCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			return next(runCtx, req)
		}
	}
}

// AuditEntry — запись о выполненной команде.
type AuditEntry struct {
	Command  string
	Source   Source
	OK       bool
	Err      string
	Started  time.Time
	Duration time.Duration
}

// String — человекочитаемое представление записи.
func (e AuditEntry) String() string {
	status := "ok"
	if !e.OK {
		status = "fail"
		if e.Err != "" {
			status = "fail: " + e.Err
		}
	}
	return fmt.Sprintf("%s %s [%s] %s (%s)", e.Started.Format(time.RFC3339), e.Command, e.Source, status, e.Duration)
}

// AuditSink — приёмник записей аудита (журнал, БД, память).
type AuditSink interface {
	Emit(AuditEntry)
}

// AuditFunc — адаптация функции под AuditSink.
type AuditFunc func(AuditEntry)

// Emit — реализует AuditSink.
func (f AuditFunc) Emit(entry AuditEntry) { f(entry) }

// MemoryAudit — потокобезопасный буфер записей в памяти (тесты, отладка).
type MemoryAudit struct {
	mu      sync.Mutex
	entries []AuditEntry
	limit   int
}

// NewMemoryAudit — буфер с ограничением на число записей (<=0 — без ограничения).
func NewMemoryAudit(limit int) *MemoryAudit {
	return &MemoryAudit{limit: limit}
}

// Emit — реализует AuditSink.
func (m *MemoryAudit) Emit(entry AuditEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = append(m.entries, entry)
	if m.limit > 0 && len(m.entries) > m.limit {
		m.entries = m.entries[len(m.entries)-m.limit:]
	}
}

// Entries — копия всех записей.
func (m *MemoryAudit) Entries() []AuditEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]AuditEntry{}, m.entries...)
}

// Last — последняя запись (false если буфер пуст).
func (m *MemoryAudit) Last() (AuditEntry, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.entries) == 0 {
		return AuditEntry{}, false
	}
	return m.entries[len(m.entries)-1], true
}

// Len — число записей.
func (m *MemoryAudit) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.entries)
}

// Reset — очищает буфер.
func (m *MemoryAudit) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = nil
}

// WithAudit — пишет запись аудита по итогам каждой команды,
// включая команды, завершившиеся ошибкой.
func WithAudit(sink AuditSink) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req Request) (Response, error) {
			started := time.Now()
			resp, err := next(ctx, req)
			if sink != nil {
				entry := AuditEntry{
					Command:  req.Command,
					Source:   req.Source,
					OK:       err == nil && resp.OK,
					Started:  started,
					Duration: time.Since(started),
				}
				if err != nil {
					entry.Err = err.Error()
				}
				sink.Emit(entry)
			}
			return resp, err
		}
	}
}
