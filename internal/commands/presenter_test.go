package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"network-scanner/internal/apperror"
)

func TestTextPresenterPresent(t *testing.T) {
	var out bytes.Buffer
	p := NewTextPresenter(&out, &bytes.Buffer{})

	resp := NewOK("готово", []string{"10.0.0.1", "10.0.0.2"}).
		WithWarning("SNMP недоступен").
		WithMeta("duration", "1.2s").
		WithMeta("count", "2")

	if err := p.Present(NewRequest("scan", SourceCLI), resp); err != nil {
		t.Fatalf("Present: %v", err)
	}

	got := out.String()
	for _, want := range []string{"готово", "10.0.0.1", "warning: SNMP недоступен", "duration: 1.2s", "count: 2"} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q:\n%s", want, got)
		}
	}

	// Метаданные выводятся в детерминированном порядке.
	if strings.Index(got, "count: 2") > strings.Index(got, "duration: 1.2s") {
		t.Errorf("meta not sorted:\n%s", got)
	}
}

func TestTextPresenterPresentError(t *testing.T) {
	var errOut bytes.Buffer
	p := NewTextPresenter(io.Discard, &errOut)

	if err := p.PresentError(NewRequest("scan", SourceCLI), errors.New("сбой сети")); err != nil {
		t.Fatalf("PresentError: %v", err)
	}
	if !strings.Contains(errOut.String(), "error: сбой сети") {
		t.Errorf("errOut = %q", errOut.String())
	}

	if err := p.PresentError(NewRequest("scan", SourceCLI), nil); err != nil {
		t.Errorf("nil error should be no-op: %v", err)
	}
}

func TestTextPresenterNilWriters(t *testing.T) {
	p := &TextPresenter{}

	if err := p.Present(NewRequest("x", SourceCLI), NewOK("m", "d").WithWarning("w").WithMeta("k", "v")); err != nil {
		t.Errorf("Present with nil writers: %v", err)
	}
	if err := p.PresentError(NewRequest("x", SourceCLI), errors.New("e")); err != nil {
		t.Errorf("PresentError with nil writers: %v", err)
	}
}

func TestJSONPresenterPresent(t *testing.T) {
	var out bytes.Buffer
	p := NewJSONPresenter(&out)

	resp := NewOK("готово", map[string]int{"hosts": 3}).WithMeta("duration", "1s")
	if err := p.Present(NewRequest("scan", SourceAPI), resp); err != nil {
		t.Fatalf("Present: %v", err)
	}

	var envelope map[string]any
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatalf("invalid JSON %q: %v", out.String(), err)
	}
	if envelope["command"] != "scan" || envelope["ok"] != true || envelope["message"] != "готово" {
		t.Errorf("envelope = %v", envelope)
	}
	data, ok := envelope["data"].(map[string]any)
	if !ok || data["hosts"] != float64(3) {
		t.Errorf("data = %v", envelope["data"])
	}
	if envelope["meta"] == nil {
		t.Error("meta should be present")
	}
}

func TestJSONPresenterPresentError(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantCode    string
		wantMessage string
	}{
		{"not found", fmt.Errorf(`dispatch %w: "x"`, ErrCommandNotFound), "command_not_found", ErrCommandNotFound.Error()},
		{"confirmation", fmt.Errorf(`%w: "wipe"`, ErrConfirmationRequired), "confirmation_required", ErrConfirmationRequired.Error()},
		{"invalid", fmt.Errorf(`register: %w`, ErrInvalidCommand), "invalid_command", ErrInvalidCommand.Error()},
		{"duplicate", fmt.Errorf(`%w: scan`, ErrDuplicateCommand), "duplicate_command", ErrDuplicateCommand.Error()},
		{"apperror code", apperror.New(apperror.CodeNotFound, "хост не найден"), "not_found", "хост не найден"},
		{"apperror wrapped", fmt.Errorf("scan: %w", apperror.New(apperror.CodeTimeout, "дольше 30s")), "timeout", "дольше 30s"},
		{"apperror internal", apperror.New(apperror.CodeInternal, "стек"), "internal", "стек"},
		{"other", errors.New("disk on fire"), "internal_error", "disk on fire"},
		{"nil", nil, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			p := NewJSONPresenter(&out)

			if err := p.PresentError(NewRequest("scan", SourceAPI), tt.err); err != nil {
				t.Fatalf("PresentError: %v", err)
			}

			var envelope map[string]any
			if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
				t.Fatalf("invalid JSON: %v", err)
			}
			if envelope["ok"] != false {
				t.Errorf("ok = %v, want false", envelope["ok"])
			}

			code, _ := envelope["errorCode"].(string)
			if code != tt.wantCode {
				t.Errorf("errorCode = %q, want %q", code, tt.wantCode)
			}

			message, _ := envelope["error"].(string)
			if !strings.Contains(message, tt.wantMessage) {
				t.Errorf("error = %q, want to contain %q", message, tt.wantMessage)
			}
		})
	}
}

func TestJSONPresenterNilWriter(t *testing.T) {
	p := &JSONPresenter{}
	if err := p.Present(NewRequest("x", SourceAPI), NewOK("m", nil)); err != nil {
		t.Errorf("Present with nil writer: %v", err)
	}
	if err := p.PresentError(NewRequest("x", SourceAPI), errors.New("e")); err != nil {
		t.Errorf("PresentError with nil writer: %v", err)
	}
}

func TestSilentPresenter(t *testing.T) {
	var p Presenter = SilentPresenter{}

	if err := p.Present(NewRequest("x", SourceCLI), NewOK("m", "d")); err != nil {
		t.Errorf("Present: %v", err)
	}
	if err := p.PresentError(NewRequest("x", SourceCLI), errors.New("e")); err != nil {
		t.Errorf("PresentError: %v", err)
	}
}

func TestPresentersSatisfyInterface(t *testing.T) {
	presenters := []Presenter{
		NewTextPresenter(io.Discard, io.Discard),
		NewJSONPresenter(io.Discard),
		SilentPresenter{},
	}
	if len(presenters) != 3 {
		t.Fatal("unexpected presenters count")
	}
}
