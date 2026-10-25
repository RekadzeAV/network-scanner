package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// AuditEntry описывает одну запись аудит-лога изменяющей операции.
// Формат хранения: JSONL (по одной JSON-строке на запись).
type AuditEntry struct {
	Timestamp string `json:"timestamp"`
	Actor     string `json:"actor,omitempty"`
	Action    string `json:"action"`
	Target    string `json:"target,omitempty"`
	Result    string `json:"result"`
	Detail    string `json:"detail,omitempty"`
}

// now возвращает текущее время в RFC3339 (UTC).
func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// defaultActor возвращает имя текущего пользователя или "unknown".
func defaultActor() string {
	if u, err := os.UserHomeDir(); err == nil && u != "" {
		// Лучшая доступная aproximация'actor без внешних зависимостей.
		if parts := strings.Split(u, string(os.PathSeparator)); len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}
	return "unknown"
}

// AppendAudit добавляет одну запись в JSONL аудит-лог по указанному пути.
// Создаёт родительскую директорию при необходимости. Является потокобезопасным.
var mu sync.Mutex

func AppendAudit(path, actor, action, target, result, detail string) error {
	mu.Lock()
	defer mu.Unlock()

	if path == "" {
		return fmt.Errorf("audit path is required")
	}

	if actor == "" {
		actor = defaultActor()
	}

	entry := AuditEntry{
		Timestamp: now(),
		Actor:     actor,
		Action:    action,
		Target:    target,
		Result:    result,
		Detail:    detail,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal audit entry: %w", err)
	}

	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("create audit dir: %w", err)
		}
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open audit log: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write audit entry: %w", err)
	}
	return nil
}

// AppendAuditWithActor — удобная обёртка, когда acteur известен.
func AppendAuditWithActor(path, actor, action, target, result, detail string) error {
	return AppendAudit(path, actor, action, target, result, detail)
}

// EnsureAuditPath проверяет, что путь к аудит-логу доступен для записи,
// и создаёт файл если он отсутствует.
func EnsureAuditPath(path string) error {
	if path == "" {
		return fmt.Errorf("audit path is required")
	}
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("create audit dir: %w", err)
		}
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open audit log: %w", err)
	}
	f.Close()
	return nil
}
