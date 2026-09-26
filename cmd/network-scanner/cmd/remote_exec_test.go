package cmd

import (
	"strings"
	"testing"
)

// ============================================================================
// P0-2: тесты защиты remote-exec (dry-run по умолчанию, consent).
//
// Проверяют, что реальное удалённое выполнение невозможно «случайно»:
// без --execute команда остаётся в dry-run, а реальное выполнение требует
// явного --consent I_UNDERSTAND.
// ============================================================================

// baseRemoteExecArgs — минимальный валидный набор аргументов.
func baseRemoteExecArgs() []string {
	return []string{"--transport", "ssh", "--target", "10.0.0.1", "--command", "uptime"}
}

func TestParseRemoteExecArgs_DryRunByDefault(t *testing.T) {
	opts, err := parseRemoteExecArgs(baseRemoteExecArgs())
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if !opts.dryRun {
		t.Error("dry-run должен быть включён по умолчанию (P0-2)")
	}
}

// TestParseRemoteExecArgs_RequireTLS — E7/7.10: флаг строгого TLS-канала.
func TestParseRemoteExecArgs_RequireTLS(t *testing.T) {
	// По умолчанию выключен.
	opts, err := parseRemoteExecArgs(baseRemoteExecArgs())
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if opts.requireTLS {
		t.Error("require-tls должен быть выключен по умолчанию")
	}

	// --require-tls включает строгий режим.
	opts, err = parseRemoteExecArgs(append(baseRemoteExecArgs(), "--require-tls"))
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if !opts.requireTLS {
		t.Error("--require-tls должен включать строгий TLS")
	}

	// Синоним --strict-tls.
	opts, err = parseRemoteExecArgs(append(baseRemoteExecArgs(), "--strict-tls"))
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if !opts.requireTLS {
		t.Error("--strict-tls должен включать строгий TLS")
	}

	// Явное отключение.
	opts, err = parseRemoteExecArgs(append(baseRemoteExecArgs(), "--require-tls=false"))
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if opts.requireTLS {
		t.Error("--require-tls=false должен выключать строгий TLS")
	}
}

func TestParseRemoteExecArgs_ExecuteWithoutConsentRejected(t *testing.T) {
	args := append(baseRemoteExecArgs(), "--execute")
	_, err := parseRemoteExecArgs(args)
	if err == nil {
		t.Fatal("ожидалась ошибка: --execute без --consent")
	}
	if !strings.Contains(err.Error(), "I_UNDERSTAND") {
		t.Errorf("неожиданная ошибка: %v", err)
	}
}

func TestParseRemoteExecArgs_ExecuteWithConsentAllowed(t *testing.T) {
	args := append(baseRemoteExecArgs(), "--execute", "--consent", remoteExecConsentToken)
	opts, err := parseRemoteExecArgs(args)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if opts.dryRun {
		t.Error("--execute должен выключать dry-run")
	}
	if opts.consent != remoteExecConsentToken {
		t.Errorf("consent = %q, ожидался %q", opts.consent, remoteExecConsentToken)
	}
}

func TestParseRemoteExecArgs_LegacyDryRunFalseEnablesExecute(t *testing.T) {
	args := append(baseRemoteExecArgs(), "--dry-run=false", "--consent", remoteExecConsentToken)
	opts, err := parseRemoteExecArgs(args)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if opts.dryRun {
		t.Error("--dry-run=false должен включать реальное выполнение")
	}
}

func TestParseRemoteExecArgs_RequiresCoreFlags(t *testing.T) {
	cases := map[string][]string{
		"без transport": {"--target", "10.0.0.1", "--command", "uptime"},
		"без target":    {"--transport", "ssh", "--command", "uptime"},
		"без command":   {"--transport", "ssh", "--target", "10.0.0.1"},
	}
	for name, args := range cases {
		if _, err := parseRemoteExecArgs(args); err == nil {
			t.Errorf("%s: ожидалась ошибка валидации", name)
		}
	}
}

func TestParseRemoteExecArgs_ConsentAutoFilledInDryRun(t *testing.T) {
	opts, err := parseRemoteExecArgs(baseRemoteExecArgs())
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if opts.consent != remoteExecConsentToken {
		t.Errorf("в dry-run consent должен автозаполняться: got %q", opts.consent)
	}
}
