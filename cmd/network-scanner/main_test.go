package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveRemoteAllowlists_StrictRequiresPolicyFile(t *testing.T) {
	_, _, err := resolveRemoteAllowlists("", "", "", true)
	if err == nil {
		t.Fatalf("expected strict policy error")
	}
}

func TestResolveRemoteAllowlists_StrictRejectsInline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.json")
	if err := os.WriteFile(path, []byte(`{"allow_hosts":["10.0.0.10"],"allow_commands":["hostname"]}`), 0o644); err != nil {
		t.Fatalf("write policy: %v", err)
	}
	_, _, err := resolveRemoteAllowlists("10.0.0.10", "", path, true)
	if err == nil {
		t.Fatalf("expected strict inline rejection")
	}
}

func TestResolveRemoteAllowlists_MergePolicyAndInline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.json")
	if err := os.WriteFile(path, []byte(`{"allow_hosts":["10.0.0.20"],"allow_commands":["uname -a"]}`), 0o644); err != nil {
		t.Fatalf("write policy: %v", err)
	}
	hosts, cmds, err := resolveRemoteAllowlists("10.0.0.10", "hostname", path, false)
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}
	if len(hosts) != 2 {
		t.Fatalf("unexpected hosts len: %d", len(hosts))
	}
	if len(cmds) != 2 {
		t.Fatalf("unexpected commands len: %d", len(cmds))
	}
}

func TestValidateSecurityReportRedaction_DefaultSafe(t *testing.T) {
	if err := validateSecurityReportRedaction(true, ""); err != nil {
		t.Fatalf("expected no error for default safe mode, got: %v", err)
	}
}

func TestValidateSecurityReportRedaction_RequiresConsentWhenDisabled(t *testing.T) {
	err := validateSecurityReportRedaction(false, "")
	if err == nil {
		t.Fatalf("expected consent error")
	}
}

func TestValidateSecurityReportRedaction_AllowsExplicitUnsafeConsent(t *testing.T) {
	if err := validateSecurityReportRedaction(false, securityReportUnsafeConsentToken); err != nil {
		t.Fatalf("expected consent acceptance, got: %v", err)
	}
}

func TestResolveSecurityReportPath_RespectsExplicitPath(t *testing.T) {
	got := resolveSecurityReportPath("report.html", true, "r1")
	if got != "report.html" {
		t.Fatalf("unexpected path: %s", got)
	}
}

func TestResolveSecurityReportPath_AutoRedacted(t *testing.T) {
	got := resolveSecurityReportPath("auto", true, "r1")
	if got != "security-report-redacted-r1.html" {
		t.Fatalf("unexpected auto redacted path: %s", got)
	}
}

func TestResolveSecurityReportPath_AutoUnredacted(t *testing.T) {
	got := resolveSecurityReportPath("auto", false, "r1")
	if got != "security-report-unredacted-r1.html" {
		t.Fatalf("unexpected auto unredacted path: %s", got)
	}
}

func TestBuildSecurityReportID_Format(t *testing.T) {
	id := buildSecurityReportID()
	if len(id) != len("20060102T150405Z") {
		t.Fatalf("unexpected id length: %s", id)
	}
	if !strings.Contains(id, "T") || !strings.HasSuffix(id, "Z") {
		t.Fatalf("unexpected id format: %s", id)
	}
}

func TestRunToolsMode_WhoisUsesRDAPFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/domain/example.com" {
			http.Error(w, "unexpected path", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/rdap+json")
		_, _ = w.Write([]byte(`{
			"objectClassName":"domain",
			"ldhName":"example.com",
			"status":["active"]
		}`))
	}))
	defer srv.Close()

	t.Setenv("NETWORK_SCANNER_RDAP_BASE_URL", srv.URL)
	// Force fallback to RDAP in tests regardless of local whois availability.
	t.Setenv("PATH", "")

	stdout := captureStdout(t, func() {
		ok := runToolsMode(
			"", "", "", "",
			4, 10, 30, false,
			"example.com", false,
			"", "", "",
			"", "", "", "", "", "",
			10,
			"",
			"", "", "", "", "", "", "", "", false, "", true, 10, "",
		)
		if !ok {
			t.Fatalf("expected tools mode to run")
		}
	})

	if !strings.Contains(stdout, "Whois: example.com") {
		t.Fatalf("expected whois header in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "RDAP summary") {
		t.Fatalf("expected RDAP summary in output, got: %s", stdout)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = old
	}()

	fn()
	_ = w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	return string(out)
}
