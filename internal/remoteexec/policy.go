package remoteexec

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Policy defines allowlist constraints for remote execution.
type Policy struct {
	AllowHosts    []string `json:"allow_hosts"`
	AllowCommands []string `json:"allow_commands"`
}

// LoadPolicy reads policy JSON from file and normalizes entries.
func LoadPolicy(path string) (Policy, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return Policy{}, fmt.Errorf("policy path is required")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Policy{}, fmt.Errorf("read policy: %w", err)
	}
	var p Policy
	if err := json.Unmarshal(data, &p); err != nil {
		return Policy{}, fmt.Errorf("parse policy: %w", err)
	}
	p.AllowHosts = normalizeList(p.AllowHosts)
	p.AllowCommands = normalizeList(p.AllowCommands)
	if hasWildcard(p.AllowHosts) || hasWildcard(p.AllowCommands) {
		return Policy{}, fmt.Errorf("wildcard '*' запрещен в remote-exec policy")
	}
	return p, nil
}

func normalizeList(items []string) []string {
	out := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, v := range items {
		t := strings.TrimSpace(v)
		if t == "" {
			continue
		}
		key := strings.ToLower(t)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, t)
	}
	return out
}

func hasWildcard(items []string) bool {
	for _, v := range items {
		if strings.TrimSpace(v) == "*" {
			return true
		}
	}
	return false
}
