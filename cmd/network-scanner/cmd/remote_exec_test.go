package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"network-scanner/internal/auditpath"
	"network-scanner/internal/contracts"
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

// ============================================================================
// E7/7.9: аудит-лог изменяющих операций.
//
// Проверяем, что журнал ведётся для всех операций (dry-run, успех, ошибка),
// команда санитизируется, а путь не попадает в CWD.
// ============================================================================

// TestWriteRemoteExecAudit_WritesJSONL — запись появляется в указанном файле.
func TestWriteRemoteExecAudit_WritesJSONL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "remote-exec.log")

	opts := remoteExecOptions{transport: "ssh", target: "10.0.0.1", command: "uptime"}
	gotPath, err := writeRemoteExecAudit(path, opts, true, contracts.RemoteExecResponse{Success: true}, nil)
	if err != nil {
		t.Fatalf("writeRemoteExecAudit() error = %v", err)
	}
	if gotPath != path {
		t.Fatalf("path = %q, want %q", gotPath, path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(data)
	for _, want := range []string{`"transport":"ssh"`, `"target":"10.0.0.1"`, `"dry_run":true`, `"success":true`} {
		if !strings.Contains(text, want) {
			t.Errorf("audit log missing %s, got: %s", want, text)
		}
	}
}

// TestWriteRemoteExecAudit_SanitizesSecrets — секреты в команде не попадают в журнал.
func TestWriteRemoteExecAudit_SanitizesSecrets(t *testing.T) {
	path := filepath.Join(t.TempDir(), "remote-exec.log")
	opts := remoteExecOptions{
		transport: "ssh",
		target:    "10.0.0.1",
		command:   "mysql --password SuperSecret123 -e 'select 1'",
	}
	if _, err := writeRemoteExecAudit(path, opts, false, contracts.RemoteExecResponse{Success: true}, nil); err != nil {
		t.Fatalf("writeRemoteExecAudit() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.Contains(string(data), "SuperSecret123") {
		t.Fatalf("audit log leaked a secret: %s", string(data))
	}
	if !strings.Contains(string(data), "***") {
		t.Fatalf("audit log should contain redacted marker, got: %s", string(data))
	}
}

// TestWriteRemoteExecAudit_ErrorMessage — ошибка выполнения фиксируется в журнале.
func TestWriteRemoteExecAudit_ErrorMessage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "remote-exec.log")
	opts := remoteExecOptions{transport: "ssh", target: "10.0.0.1", command: "uptime"}

	if _, err := writeRemoteExecAudit(path, opts, false, contracts.RemoteExecResponse{}, errors.New("ssh: connect failed")); err != nil {
		t.Fatalf("writeRemoteExecAudit() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(data)
	if !strings.Contains(text, `"success":false`) {
		t.Errorf("expected success=false, got: %s", text)
	}
	if !strings.Contains(text, "connect failed") {
		t.Errorf("expected error message in audit, got: %s", text)
	}
}

// TestWriteRemoteExecAudit_DefaultPathOutsideCWD — путь по умолчанию не в CWD.
func TestWriteRemoteExecAudit_DefaultPathOutsideCWD(t *testing.T) {
	opts := remoteExecOptions{transport: "ssh", target: "10.0.0.1", command: "uptime"}
	gotPath, err := writeRemoteExecAudit("", opts, true, contracts.RemoteExecResponse{Success: true}, nil)
	if err != nil {
		t.Fatalf("writeRemoteExecAudit() error = %v", err)
	}
	defer func() { _ = os.Remove(gotPath) }()

	if gotPath != auditpath.RemoteExecPath() {
		t.Fatalf("path = %q, want default %q", gotPath, auditpath.RemoteExecPath())
	}
	cwd, err := os.Getwd()
	if err == nil && cwd != "" {
		if rel, relErr := filepath.Rel(cwd, gotPath); relErr == nil && !strings.HasPrefix(rel, "..") {
			t.Fatalf("default audit path %q must not live inside CWD %q", gotPath, cwd)
		}
	}
}

// TestRunRemoteExecCLI_DryRunWritesAudit — сквозной путь: dry-run пишет журнал.
func TestRunRemoteExecCLI_DryRunWritesAudit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "remote-exec.log")
	args := append(baseRemoteExecArgs(),
		"--allow-hosts", "10.0.0.1",
		"--allow-commands", "uptime",
		"--audit-log", path,
	)
	if err := RunRemoteExecCLI(testCfg(t), args...); err != nil {
		t.Fatalf("RunRemoteExecCLI() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("audit log not written: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, `"dry_run":true`) {
		t.Errorf("dry-run audit expected, got: %s", text)
	}
	if !strings.Contains(text, `"actor":`) {
		t.Errorf("audit entry must contain actor, got: %s", text)
	}
}

// TestRunRemoteExecCLI_PolicyFailureStillAudited — отказ политики тоже в журнале.
func TestRunRemoteExecCLI_PolicyFailureStillAudited(t *testing.T) {
	path := filepath.Join(t.TempDir(), "remote-exec.log")
	// 10.0.0.99 не в allowlist → dry-run вернёт ошибку политики.
	args := []string{
		"--transport", "ssh",
		"--target", "10.0.0.99",
		"--command", "uptime",
		"--allow-hosts", "10.0.0.1",
		"--allow-commands", "uptime",
		"--audit-log", path,
	}
	if err := RunRemoteExecCLI(testCfg(t), args...); err == nil {
		t.Fatal("ожидалась ошибка политики")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("audit log not written for failed policy check: %v", err)
	}
	if !strings.Contains(string(data), `"success":false`) {
		t.Errorf("expected success=false in audit, got: %s", string(data))
	}
}

// ============================================================================
// E7/7.7: импорт списков хостов (hostlist importer → CLI).
// ============================================================================

func TestLoadTargetsFromFile_CSV(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hosts.csv")
	content := "192.168.1.10,router\n192.168.1.20,switch,Основной\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write csv: %v", err)
	}

	ips, format, err := loadTargetsFromFile(path, hostsFromFileOptions{})
	if err != nil {
		t.Fatalf("loadTargetsFromFile() error = %v", err)
	}
	if format != "csv" {
		t.Errorf("format = %q, want csv (auto по расширению)", format)
	}
	if len(ips) != 2 || ips[0] != "192.168.1.10" || ips[1] != "192.168.1.20" {
		t.Fatalf("ips = %v, want [192.168.1.10 192.168.1.20]", ips)
	}
}

func TestLoadTargetsFromFile_JSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hosts.json")
	content := `[{"ip":"10.0.0.5","hostname":"srv"},{"ip":"10.0.0.6"}]`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write json: %v", err)
	}

	ips, format, err := loadTargetsFromFile(path, hostsFromFileOptions{})
	if err != nil {
		t.Fatalf("loadTargetsFromFile() error = %v", err)
	}
	if format != "json" {
		t.Errorf("format = %q, want json", format)
	}
	if len(ips) != 2 || ips[1] != "10.0.0.6" {
		t.Fatalf("ips = %v, want [10.0.0.5 10.0.0.6]", ips)
	}
}

func TestLoadTargetsFromFile_CSV_CIDRExpanded(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cidr.csv")
	// CIDR-запись в CSV раскрывается импортером в отдельные IP.
	if err := os.WriteFile(path, []byte("192.168.5.0/30\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	ips, _, err := loadTargetsFromFile(path, hostsFromFileOptions{Format: "csv"})
	if err != nil {
		t.Fatalf("loadTargetsFromFile() error = %v", err)
	}
	if len(ips) != 4 {
		t.Fatalf("ips = %v, want 4 адреса из /30", ips)
	}
}

func TestLoadTargetsFromFile_TargetsFormatKeepsRanges(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "targets.txt")
	content := "# comment\n192.168.1.1-3\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	ips, format, err := loadTargetsFromFile(path, hostsFromFileOptions{})
	if err != nil {
		t.Fatalf("loadTargetsFromFile() error = %v", err)
	}
	if format != "targets" {
		t.Errorf("format = %q, want targets (обратная совместимость .txt)", format)
	}
	if len(ips) != 3 {
		t.Fatalf("ips = %v, want 3 адреса из диапазона", ips)
	}
}

func TestLoadTargetsFromFile_ExplicitTargetsFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hosts.csv")
	// Даже при .csv явный --hosts-format targets использует прежний парсер.
	if err := os.WriteFile(path, []byte("192.0.2.1\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	ips, format, err := loadTargetsFromFile(path, hostsFromFileOptions{Format: "targets"})
	if err != nil {
		t.Fatalf("loadTargetsFromFile() error = %v", err)
	}
	if format != "targets" || len(ips) != 1 || ips[0] != "192.0.2.1" {
		t.Fatalf("ips = %v, format = %q, want [192.0.2.1]/targets", ips, format)
	}
}

func TestLoadTargetsFromFile_UnsupportedFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hosts.xml")
	if err := os.WriteFile(path, []byte("<hosts/>"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, _, err := loadTargetsFromFile(path, hostsFromFileOptions{Format: "xml"})
	if err == nil || !strings.Contains(err.Error(), "неподдерживаемый формат") {
		t.Fatalf("expected unsupported format error, got: %v", err)
	}
}

func TestScanCmd_HasHostsFormatFlag(t *testing.T) {
	for _, name := range []string{"hosts-file", "hosts-format"} {
		if scanCmd.Flags().Lookup(name) == nil {
			t.Errorf("scan не имеет флага --%s", name)
		}
	}
}

// TestScanTargetsPriority — цели из файла приоритетнее автоопределения сети,
// иначе адреса из файла игнорировались бы (E7/7.7).
func TestScanTargetsPriority(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "hosts.csv")
	if err := os.WriteFile(csvPath, []byte("192.168.99.10,router\n192.168.99.0/30\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	ips, format, err := loadTargetsFromFile(csvPath, hostsFromFileOptions{})
	if err != nil {
		t.Fatalf("loadTargetsFromFile() error = %v", err)
	}
	if format != "csv" {
		t.Errorf("format = %q, want csv", format)
	}
	// 1 одиночный + 4 из /30 = 5 целей
	if len(ips) != 5 {
		t.Fatalf("ips = %v, want 5 целей", ips)
	}
	if ips[0] != "192.168.99.10" {
		t.Errorf("ips[0] = %q, want 192.168.99.10 (первая цель — источник CIDR)", ips[0])
	}
}

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
