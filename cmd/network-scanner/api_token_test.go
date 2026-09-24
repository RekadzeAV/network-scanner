package main

import (
	"testing"
)

// TestResolveAPIToken_FlagEquals — форма --api-token=<value>.
func TestResolveAPIToken_FlagEquals(t *testing.T) {
	args := []string{"--api", "--api-token=secret-eq"}
	if got := resolveAPIToken(args); got != "secret-eq" {
		t.Fatalf("got %q, want %q", got, "secret-eq")
	}
}

// TestResolveAPIToken_FlagSpace — форма --api-token <value>.
func TestResolveAPIToken_FlagSpace(t *testing.T) {
	args := []string{"--api", "--api-token", "secret-space", "--other"}
	if got := resolveAPIToken(args); got != "secret-space" {
		t.Fatalf("got %q, want %q", got, "secret-space")
	}
}

// TestResolveAPIToken_EnvFallback — переменная окружения как fallback.
func TestResolveAPIToken_EnvFallback(t *testing.T) {
	t.Setenv(envAPIToken, "from-env")
	if got := resolveAPIToken([]string{"--api"}); got != "from-env" {
		t.Fatalf("got %q, want %q", got, "from-env")
	}
}

// TestResolveAPIToken_FlagOverridesEnv — флаг имеет приоритет над env.
func TestResolveAPIToken_FlagOverridesEnv(t *testing.T) {
	t.Setenv(envAPIToken, "from-env")
	args := []string{"--api", "--api-token=from-flag"}
	if got := resolveAPIToken(args); got != "from-flag" {
		t.Fatalf("got %q, want %q", got, "from-flag")
	}
}

// TestResolveAPIToken_Empty — пусто, когда нет ни флага, ни env.
func TestResolveAPIToken_Empty(t *testing.T) {
	t.Setenv(envAPIToken, "")
	if got := resolveAPIToken([]string{"--api"}); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}

// TestResolveAPIToken_DanglingFlag — флаг без значения не заглядывает за границы.
func TestResolveAPIToken_DanglingFlag(t *testing.T) {
	t.Setenv(envAPIToken, "env-fallback")
	if got := resolveAPIToken([]string{"--api", "--api-token"}); got != "env-fallback" {
		t.Fatalf("got %q, want %q", got, "env-fallback")
	}
}
