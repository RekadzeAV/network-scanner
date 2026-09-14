// Package commands предоставляет единый слой диспатча команд для всех
// интерфейсов приложения: CLI, GUI и HTTP API.
//
// Вместо дублирования логики в каждом интерфейсе команды регистрируются один
// раз в Registry, а интерфейсы лишь формируют Request и показывают Response.
package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Source — интерфейс-источник вызова команды.
type Source string

const (
	SourceCLI  Source = "cli"
	SourceGUI  Source = "gui"
	SourceAPI  Source = "api"
	SourceTest Source = "test"
)

// Valid — известен ли источник.
func (s Source) Valid() bool {
	switch s {
	case SourceCLI, SourceGUI, SourceAPI, SourceTest:
		return true
	default:
		return false
	}
}

// Risk — уровень опасности команды.
type Risk string

const (
	// RiskRead — только чтение, безопасно из любого источника.
	RiskRead Risk = "read"
	// RiskWrite — изменяет состояние, логируется.
	RiskWrite Risk = "write"
	// RiskDestructive — необратимые действия, требует подтверждения.
	RiskDestructive Risk = "destructive"
)

// NeedsConfirmation — требует ли риск явного подтверждения.
func (r Risk) NeedsConfirmation() bool {
	return r == RiskDestructive
}

// Errors возвращаемые реестром.
var (
	// ErrCommandNotFound — команда не зарегистрирована.
	ErrCommandNotFound = errors.New("command not found")
	// ErrDuplicateCommand — команда с таким именем уже есть.
	ErrDuplicateCommand = errors.New("command already registered")
	// ErrConfirmationRequired — деструктивная команда без подтверждения.
	ErrConfirmationRequired = errors.New("confirmation required for destructive command")
	// ErrInvalidCommand — неверное определение команды.
	ErrInvalidCommand = errors.New("invalid command definition")
)

// Request — унифицированный вызов команды.
type Request struct {
	Command   string
	Args      map[string]string
	Flags     map[string]bool
	Source    Source
	DryRun    bool
	Confirmed bool
}

// NewRequest — запрос с пустыми, но инициализированными картами.
func NewRequest(command string, source Source) Request {
	return Request{
		Command: command,
		Args:    map[string]string{},
		Flags:   map[string]bool{},
		Source:  source,
	}
}

// Arg — значение аргумента или def, если аргумент отсутствует.
func (r Request) Arg(name, def string) string {
	if v, ok := r.Args[name]; ok {
		return v
	}
	return def
}

// RequireArg — значение обязательного аргумента или ошибка.
func (r Request) RequireArg(name string) (string, error) {
	v, ok := r.Args[name]
	if !ok || strings.TrimSpace(v) == "" {
		return "", fmt.Errorf("%w: missing required arg %q", ErrInvalidCommand, name)
	}
	return v, nil
}

// Flag — значение флага (false если отсутствует).
func (r Request) Flag(name string) bool {
	return r.Flags[name]
}

// Clone — глубокая копия запроса (карты не разделяются).
func (r Request) Clone() Request {
	dup := r
	dup.Args = make(map[string]string, len(r.Args))
	for k, v := range r.Args {
		dup.Args[k] = v
	}
	dup.Flags = make(map[string]bool, len(r.Flags))
	for k, v := range r.Flags {
		dup.Flags[k] = v
	}
	return dup
}

// Response — унифицированный результат выполнения команды.
type Response struct {
	OK       bool
	Message  string
	Data     any
	Warnings []string
	Meta     map[string]string
}

// NewOK — успешный результат.
func NewOK(message string, data any) Response {
	return Response{OK: true, Message: message, Data: data}
}

// NewFail — ошибка уровня команды (не error уровня транспорта).
func NewFail(message string) Response {
	return Response{OK: false, Message: message}
}

// WithWarning — добавляет предупреждение.
func (r Response) WithWarning(warning string) Response {
	r.Warnings = append(append([]string{}, r.Warnings...), warning)
	return r
}

// WithMeta — добавляет метаданные.
func (r Response) WithMeta(key, value string) Response {
	meta := make(map[string]string, len(r.Meta)+1)
	for k, v := range r.Meta {
		meta[k] = v
	}
	meta[key] = value
	r.Meta = meta
	return r
}

// Handler — выполнение команды.
//
// error означает сбой транспорта/инфраструктуры; бизнес-результат
// возвращается в Response.OK.
type Handler func(ctx context.Context, req Request) (Response, error)

// Command — описание команды, доступной из любого интерфейса.
type Command struct {
	Name        string
	Aliases     []string
	Description string
	Risk        Risk
	Handler     Handler
}

// Validate — корректно ли определение команды.
func (c Command) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidCommand)
	}
	if c.Handler == nil {
		return fmt.Errorf("%w: handler is required for %q", ErrInvalidCommand, c.Name)
	}
	switch c.Risk {
	case RiskRead, RiskWrite, RiskDestructive:
	default:
		return fmt.Errorf("%w: unknown risk %q for %q", ErrInvalidCommand, c.Risk, c.Name)
	}
	seen := map[string]bool{normalizeName(c.Name): true}
	for _, alias := range c.Aliases {
		key := normalizeName(alias)
		if key == "" {
			return fmt.Errorf("%w: empty alias in %q", ErrInvalidCommand, c.Name)
		}
		if seen[key] {
			return fmt.Errorf("%w: duplicate alias %q in %q", ErrInvalidCommand, alias, c.Name)
		}
		seen[key] = true
	}
	return nil
}

// Matches — относится ли имя к этой команде (имя или алиас).
func (c Command) Matches(name string) bool {
	key := normalizeName(name)
	if key == normalizeName(c.Name) {
		return true
	}
	for _, alias := range c.Aliases {
		if key == normalizeName(alias) {
			return true
		}
	}
	return false
}

func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
