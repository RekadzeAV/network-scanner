package commands

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestWithRecovery(t *testing.T) {
	reg := NewRegistry(WithRecovery())
	err := reg.Register(Command{
		Name: "boom",
		Risk: RiskRead,
		Handler: func(ctx context.Context, req Request) (Response, error) {
			panic("exploded")
		},
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	resp, err := reg.Dispatch(context.Background(), NewRequest("boom", SourceAPI))
	if err == nil {
		t.Fatal("panic should be converted to error")
	}
	if !strings.Contains(err.Error(), "exploded") {
		t.Errorf("error = %v, want to contain panic value", err)
	}
	if resp.OK {
		t.Error("response should not be OK after panic")
	}
}

func TestWithRecoveryWithoutPanic(t *testing.T) {
	reg := NewRegistry(WithRecovery())
	if err := reg.Register(Command{Name: "ok", Risk: RiskRead, Handler: nopHandler}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	resp, err := reg.Dispatch(context.Background(), NewRequest("ok", SourceAPI))
	if err != nil || !resp.OK {
		t.Errorf("resp = %+v, err = %v", resp, err)
	}
}

func TestWithTimeout(t *testing.T) {
	var sawDeadline bool
	reg := NewRegistry(WithTimeout(50 * time.Millisecond))
	err := reg.Register(Command{
		Name: "slow",
		Risk: RiskRead,
		Handler: func(ctx context.Context, req Request) (Response, error) {
			_, sawDeadline = ctx.Deadline()
			<-ctx.Done()
			return NewFail("cancelled"), ctx.Err()
		},
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	_, err = reg.Dispatch(context.Background(), NewRequest("slow", SourceCLI))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error = %v, want DeadlineExceeded", err)
	}
	if !sawDeadline {
		t.Error("handler should receive context with deadline")
	}
}

func TestWithTimeoutDisabled(t *testing.T) {
	var hadDeadline bool
	reg := NewRegistry(WithTimeout(0))
	if err := reg.Register(Command{
		Name: "fast",
		Risk: RiskRead,
		Handler: func(ctx context.Context, req Request) (Response, error) {
			_, hadDeadline = ctx.Deadline()
			return NewOK("", nil), nil
		},
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if _, err := reg.Dispatch(context.Background(), NewRequest("fast", SourceCLI)); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if hadDeadline {
		t.Error("timeout <= 0 should not add deadline")
	}
}

func TestWithAudit(t *testing.T) {
	audit := NewMemoryAudit(0)
	reg := NewRegistry(WithAudit(audit))

	if err := reg.Register(
		Command{Name: "ok", Risk: RiskRead, Handler: nopHandler},
		Command{Name: "bad", Risk: RiskRead, Handler: func(ctx context.Context, req Request) (Response, error) {
			return NewFail("нет доступа"), errors.New("нет доступа")
		}},
		Command{Name: "soft", Risk: RiskRead, Handler: func(ctx context.Context, req Request) (Response, error) {
			return NewFail("пустой результат"), nil
		}},
	); err != nil {
		t.Fatalf("Register: %v", err)
	}

	_, _ = reg.Dispatch(context.Background(), NewRequest("ok", SourceGUI))
	_, _ = reg.Dispatch(context.Background(), NewRequest("bad", SourceCLI))
	_, _ = reg.Dispatch(context.Background(), NewRequest("soft", SourceAPI))

	entries := audit.Entries()
	if len(entries) != 3 {
		t.Fatalf("audit entries = %d, want 3", len(entries))
	}

	if !entries[0].OK || entries[0].Command != "ok" || entries[0].Source != SourceGUI {
		t.Errorf("entry[0] = %+v", entries[0])
	}
	if entries[1].OK || entries[1].Err == "" {
		t.Errorf("entry[1] = %+v, want failure with error text", entries[1])
	}
	if entries[2].OK {
		t.Errorf("entry[2] = %+v, want OK=false for failed response", entries[2])
	}
	for i, entry := range entries {
		if entry.Duration < 0 {
			t.Errorf("entry[%d].Duration = %v", i, entry.Duration)
		}
	}
}

func TestWithAuditRecordsRejectedCommands(t *testing.T) {
	audit := NewMemoryAudit(0)
	reg := NewRegistry(WithAudit(audit))

	// Команда не найдена — middleware не вызывается, аудита нет.
	if _, err := reg.Dispatch(context.Background(), NewRequest("ghost", SourceAPI)); !errors.Is(err, ErrCommandNotFound) {
		t.Fatalf("err = %v", err)
	}
	if audit.Len() != 0 {
		t.Errorf("unknown command should not be audited, entries = %d", audit.Len())
	}
}

func TestWithAuditNilSink(t *testing.T) {
	reg := NewRegistry(WithAudit(nil))
	if err := reg.Register(Command{Name: "x", Risk: RiskRead, Handler: nopHandler}); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if _, err := reg.Dispatch(context.Background(), NewRequest("x", SourceTest)); err != nil {
		t.Errorf("nil sink should be tolerated: %v", err)
	}
}

func TestAuditFuncSink(t *testing.T) {
	var got []AuditEntry
	var mu sync.Mutex
	sink := AuditFunc(func(entry AuditEntry) {
		mu.Lock()
		defer mu.Unlock()
		got = append(got, entry)
	})

	reg := NewRegistry(WithAudit(sink))
	if err := reg.Register(Command{Name: "x", Risk: RiskRead, Handler: nopHandler}); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if _, err := reg.Dispatch(context.Background(), NewRequest("x", SourceAPI)); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 1 || got[0].Command != "x" {
		t.Errorf("sink got %v", got)
	}
}

func TestMemoryAudit(t *testing.T) {
	audit := NewMemoryAudit(2)

	audit.Emit(AuditEntry{Command: "a"})
	audit.Emit(AuditEntry{Command: "b"})
	audit.Emit(AuditEntry{Command: "c"})

	if audit.Len() != 2 {
		t.Fatalf("Len = %d, want 2 (limit)", audit.Len())
	}
	entries := audit.Entries()
	if entries[0].Command != "b" || entries[1].Command != "c" {
		t.Errorf("entries = %v, want [b c]", entries)
	}

	entries[0].Command = "mutated"
	again := audit.Entries()
	if again[0].Command == "mutated" {
		t.Error("Entries must return a copy")
	}

	last, ok := audit.Last()
	if !ok || last.Command != "c" {
		t.Errorf("Last = %+v, %v", last, ok)
	}

	audit.Reset()
	if audit.Len() != 0 {
		t.Error("Reset should clear entries")
	}
	if _, ok := audit.Last(); ok {
		t.Error("Last on empty should return false")
	}
}

func TestAuditEntryString(t *testing.T) {
	ok := AuditEntry{Command: "scan", Source: SourceCLI, OK: true, Started: time.Unix(0, 0), Duration: time.Millisecond}
	if got := ok.String(); !strings.Contains(got, "scan") || !strings.Contains(got, "ok") {
		t.Errorf("String = %q", got)
	}

	fail := AuditEntry{Command: "scan", Source: SourceAPI, Err: "boom", Started: time.Unix(0, 0)}
	got := fail.String()
	if !strings.Contains(got, "fail") || !strings.Contains(got, "boom") {
		t.Errorf("String = %q", got)
	}
}
