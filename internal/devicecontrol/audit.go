package devicecontrol

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

// AuditEntry stores action trail for sensitive operations.
type AuditEntry struct {
	Timestamp string `json:"timestamp"`
	Actor     string `json:"actor"`
	Action    string `json:"action"`
	TargetURL string `json:"target_url"`
	Vendor    string `json:"vendor,omitempty"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
}

// AppendAudit writes a single JSONL record.
func AppendAudit(path string, entry AuditEntry) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("audit path is required")
	}
	if strings.TrimSpace(entry.Timestamp) == "" {
		entry.Timestamp = time.Now().Format(time.RFC3339)
	}
	if strings.TrimSpace(entry.Actor) == "" {
		entry.Actor = currentActor()
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal audit entry: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create audit dir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open audit file: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write audit entry: %w", err)
	}
	return nil
}

func currentActor() string {
	u, err := user.Current()
	if err == nil {
		if strings.TrimSpace(u.Username) != "" {
			return u.Username
		}
	}
	if v := strings.TrimSpace(os.Getenv("USERNAME")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("USER")); v != "" {
		return v
	}
	return "unknown"
}
