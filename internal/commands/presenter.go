package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"network-scanner/internal/apperror"
	"sort"
)

// Presenter — способ показа результата команды конкретным интерфейсом.
//
// CLI печатает текст, GUI отдаёт данные в виджет, API сериализует JSON.
type Presenter interface {
	Present(req Request, resp Response) error
	PresentError(req Request, err error) error
}

// TextPresenter — человекочитаемый вывод для CLI.
type TextPresenter struct {
	Out io.Writer
	Err io.Writer
}

// NewTextPresenter — текстовый презентер в заданные потоки.
func NewTextPresenter(out, errOut io.Writer) *TextPresenter {
	return &TextPresenter{Out: out, Err: errOut}
}

// Present — реализует Presenter.
func (p *TextPresenter) Present(req Request, resp Response) error {
	w := p.Out
	if w == nil {
		w = io.Discard
	}

	if resp.Message != "" {
		if _, err := fmt.Fprintf(w, "%s\n", resp.Message); err != nil {
			return err
		}
	}
	if resp.Data != nil {
		if _, err := fmt.Fprintf(w, "%v\n", resp.Data); err != nil {
			return err
		}
	}
	for _, warning := range resp.Warnings {
		if _, err := fmt.Fprintf(w, "warning: %s\n", warning); err != nil {
			return err
		}
	}
	for _, key := range sortedMetaKeys(resp.Meta) {
		if _, err := fmt.Fprintf(w, "%s: %s\n", key, resp.Meta[key]); err != nil {
			return err
		}
	}
	return nil
}

// PresentError — реализует Presenter.
func (p *TextPresenter) PresentError(req Request, err error) error {
	w := p.Err
	if w == nil {
		w = io.Discard
	}
	if err == nil {
		return nil
	}
	_, writeErr := fmt.Fprintf(w, "error: %s\n", err)
	return writeErr
}

// JSONPresenter — машинный вывод для HTTP API.
type JSONPresenter struct {
	Out io.Writer
}

// NewJSONPresenter — JSON-президентер в заданный поток.
func NewJSONPresenter(out io.Writer) *JSONPresenter {
	return &JSONPresenter{Out: out}
}

// jsonEnvelope — единая обёртка ответа API.
type jsonEnvelope struct {
	Command   string            `json:"command"`
	OK        bool              `json:"ok"`
	Message   string            `json:"message,omitempty"`
	Data      any               `json:"data,omitempty"`
	Warnings  []string          `json:"warnings,omitempty"`
	Meta      map[string]string `json:"meta,omitempty"`
	Error     string            `json:"error,omitempty"`
	ErrorCode string            `json:"errorCode,omitempty"`
}

// Present — реализует Presenter.
func (p *JSONPresenter) Present(req Request, resp Response) error {
	return p.write(jsonEnvelope{
		Command:  req.Command,
		OK:       resp.OK,
		Message:  resp.Message,
		Data:     resp.Data,
		Warnings: resp.Warnings,
		Meta:     resp.Meta,
	})
}

// PresentError — реализует Presenter.
func (p *JSONPresenter) PresentError(req Request, err error) error {
	envelope := jsonEnvelope{Command: req.Command, OK: false}
	if err != nil {
		envelope.Error = err.Error()
		envelope.ErrorCode = classifyError(err)
	}
	return p.write(envelope)
}

func (p *JSONPresenter) write(envelope jsonEnvelope) error {
	out := p.Out
	if out == nil {
		out = io.Discard
	}
	encoder := json.NewEncoder(out)
	return encoder.Encode(envelope)
}

// SilentPresenter — подавляет вывод (фоновые задачи, тесты).
type SilentPresenter struct{}

// Present — реализует Presenter.
func (SilentPresenter) Present(Request, Response) error { return nil }

// PresentError — реализует Presenter.
func (SilentPresenter) PresentError(Request, error) error { return nil }

// classifyError — стабильный код ошибки для клиента API.
//
// Сначала проверяются собственные маркеры реестра, затем коды apperror
// из цепочки обёрток; всё остальное считается внутренней ошибкой.
func classifyError(err error) string {
	if err == nil {
		return ""
	}

	switch {
	case errors.Is(err, ErrCommandNotFound):
		return "command_not_found"
	case errors.Is(err, ErrConfirmationRequired):
		return "confirmation_required"
	case errors.Is(err, ErrInvalidCommand):
		return "invalid_command"
	case errors.Is(err, ErrDuplicateCommand):
		return "duplicate_command"
	}

	if code := apperror.Of(err); code != "" && code != apperror.CodeUnknown {
		return string(code)
	}

	return "internal_error"
}

func sortedMetaKeys(meta map[string]string) []string {
	keys := make([]string, 0, len(meta))
	for key := range meta {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
