package commands

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// Middleware — обёртка над обработчиком (аудит, recovery, таймауты).
type Middleware func(Handler) Handler

// Registry — реестр команд с единым диспатчем для CLI/GUI/API.
//
// Потокобезопасен: регистрация и вызов возможны из разных горутин.
type Registry struct {
	mu          sync.RWMutex
	commands    map[string]Command
	index       map[string]string // имя/алиас -> каноническое имя
	middlewares []Middleware
}

// NewRegistry — пустой реестр с необязательными middleware по умолчанию.
func NewRegistry(middlewares ...Middleware) *Registry {
	return &Registry{
		commands:    map[string]Command{},
		index:       map[string]string{},
		middlewares: append([]Middleware{}, middlewares...),
	}
}

// Use — добавляет middleware поверх уже настроенных.
func (r *Registry) Use(middlewares ...Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middlewares = append(r.middlewares, middlewares...)
}

// Register — регистрирует команды. Ошибка если команда невалидна или имя
// (включая алиасы) пересекается с уже зарегистрированной командой.
func (r *Registry) Register(cmds ...Command) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем всю пачку до внесения, чтобы не оставлять частичное состояние.
	local := map[string]bool{}
	for _, cmd := range cmds {
		if err := cmd.Validate(); err != nil {
			return err
		}
		for _, key := range commandKeys(cmd) {
			if _, exists := r.index[key]; exists {
				return fmt.Errorf("%w: %q", ErrDuplicateCommand, key)
			}
			if local[key] {
				return fmt.Errorf("%w: %q", ErrDuplicateCommand, key)
			}
			local[key] = true
		}
	}

	for _, cmd := range cmds {
		canonical := normalizeName(cmd.Name)
		r.commands[canonical] = cmd
		for _, key := range commandKeys(cmd) {
			r.index[key] = canonical
		}
	}
	return nil
}

// Get — команда по имени или алиасу.
func (r *Registry) Get(name string) (Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.getLocked(name)
}

func (r *Registry) getLocked(name string) (Command, bool) {
	canonical, ok := r.index[normalizeName(name)]
	if !ok {
		return Command{}, false
	}
	cmd, ok := r.commands[canonical]
	return cmd, ok
}

// Has — зарегистрирована ли команда.
func (r *Registry) Has(name string) bool {
	_, ok := r.Get(name)
	return ok
}

// List — все команды, отсортированные по имени.
func (r *Registry) List() []Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		out = append(out, cmd)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Names — канонические имена команд по возрастанию.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]string, 0, len(r.commands))
	for name := range r.commands {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// Len — число зарегистрированных команд.
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.commands)
}

// Dispatch — выполняет команду по запросу.
//
// Возвращает ErrCommandNotFound для неизвестных команд и
// ErrConfirmationRequired для деструктивных команд без req.Confirmed.
func (r *Registry) Dispatch(ctx context.Context, req Request) (Response, error) {
	r.mu.RLock()
	cmd, ok := r.getLocked(req.Command)
	middlewares := append([]Middleware{}, r.middlewares...)
	r.mu.RUnlock()

	if !ok {
		return Response{}, fmt.Errorf("%w: %q", ErrCommandNotFound, req.Command)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	// Проверка подтверждения выполняется внутри цепочки middleware,
	// чтобы отказ деструктивной команды попадал в журнал аудита.
	guarded := func(ctx context.Context, req Request) (Response, error) {
		if cmd.Risk.NeedsConfirmation() && !req.Confirmed {
			return Response{}, fmt.Errorf("%w: %q", ErrConfirmationRequired, cmd.Name)
		}
		return cmd.Handler(ctx, req)
	}

	handler := Handler(guarded)
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler(ctx, req.Clone())
}

// DispatchNamed — ярлык вызова по имени команды.
func (r *Registry) DispatchNamed(ctx context.Context, name string, source Source) (Response, error) {
	return r.Dispatch(ctx, NewRequest(name, source))
}

func commandKeys(cmd Command) []string {
	keys := make([]string, 0, 1+len(cmd.Aliases))
	keys = append(keys, normalizeName(cmd.Name))
	for _, alias := range cmd.Aliases {
		keys = append(keys, normalizeName(alias))
	}
	return keys
}
