package report

import (
	"strings"
	"testing"
	"time"

	"network-scanner/internal/cve"
	"network-scanner/internal/risksignature"
	"network-scanner/internal/scanner"
)

func TestRenderSecurityHTML_RendersSummaryAndFindings(t *testing.T) {
	results := []scanner.Result{
		{
			IP:       "10.0.0.5",
			Hostname: "srv-01",
			Ports: []scanner.PortInfo{
				{Port: 22, State: "open", Service: "ssh"},
			},
		},
	}
	findings := []cve.Match{
		{
			HostIP:   "10.0.0.5",
			HostName: "srv-01",
			Port:     22,
			Service:  "ssh",
			Entry: cve.Entry{
				ID:          "CVE-2023-38408",
				Description: "OpenSSH issue",
				URL:         "https://nvd.nist.gov/vuln/detail/CVE-2023-38408",
				CVSS:        9.8,
			},
		},
	}
	got, err := RenderSecurityHTML(results, findings, time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	html := string(got)
	if !strings.Contains(html, "Network Security Report") {
		t.Fatalf("expected title in html")
	}
	if !strings.Contains(html, "CVE-2023-38408") {
		t.Fatalf("expected cve row in html")
	}
	if !strings.Contains(html, "REDACTION: ON") {
		t.Fatalf("expected redaction status ON in html")
	}
	if !strings.Contains(html, "Metadata: report-id=n/a | mode=manual | policy=v1 | unsafe-consent=no") {
		t.Fatalf("expected default metadata in html")
	}
	if strings.Contains(html, "redaction disabled") {
		t.Fatalf("did not expect unredacted warning in safe mode")
	}
}

func TestRenderSecurityHTMLWithRisk_RendersRiskSection(t *testing.T) {
	results := []scanner.Result{{IP: "10.0.0.8"}}
	findings := []cve.Match{}
	risks := []risksignature.Finding{
		{
			HostIP:         "10.0.0.8",
			SignatureID:    "home.telnet.open",
			Title:          "Telnet открыт",
			Severity:       "high",
			Reason:         "open_port=23",
			Recommendation: "Disable telnet",
		},
	}
	got, err := RenderSecurityHTMLWithRisk(results, findings, risks, time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	html := string(got)
	if !strings.Contains(html, "Risk Signature Findings") {
		t.Fatalf("expected risk section in html")
	}
	if !strings.Contains(html, "home.telnet.open") {
		t.Fatalf("expected risk row in html")
	}
}

func TestRenderSecurityHTML_SanitizesSensitiveText(t *testing.T) {
	results := []scanner.Result{
		{
			IP:       "10.0.0.5",
			Hostname: "srv token=abc123",
			Ports: []scanner.PortInfo{
				{Port: 22, State: "open", Service: "ssh password=secret"},
			},
		},
	}
	findings := []cve.Match{
		{
			HostIP:   "10.0.0.5",
			HostName: "srv token=abc123",
			Port:     22,
			Service:  "ssh password=secret",
			Entry: cve.Entry{
				ID:          "CVE-TEST",
				Description: "token:abcd",
				URL:         "https://example.test/cve",
				CVSS:        5.0,
			},
		},
	}
	got, err := RenderSecurityHTML(results, findings, time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	html := string(got)
	for _, leaked := range []string{"abc123", "secret", "abcd"} {
		if strings.Contains(html, leaked) {
			t.Fatalf("found leaked secret %q in html", leaked)
		}
	}
	if !strings.Contains(html, "***") {
		t.Fatalf("expected masked token marker in html")
	}
}

func TestRenderSecurityHTMLWithRiskOptions_AllowsDisableRedaction(t *testing.T) {
	results := []scanner.Result{
		{
			IP:       "10.0.0.5",
			Hostname: "srv token=abc123",
			Ports: []scanner.PortInfo{
				{Port: 22, State: "open", Service: "ssh password=secret"},
			},
		},
	}
	findings := []cve.Match{
		{
			HostIP:   "10.0.0.5",
			HostName: "srv token=abc123",
			Port:     22,
			Service:  "ssh password=secret",
			Entry: cve.Entry{
				ID:          "CVE-TEST",
				Description: "token:abcd",
				URL:         "https://example.test/cve",
				CVSS:        5.0,
			},
		},
	}
	got, err := RenderSecurityHTMLWithRiskOptions(results, findings, nil, time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC), Options{
		RedactSensitive: false,
		UnsafeConsent:   true,
		GenerationMode:  "auto",
		PolicyVersion:   "v1",
		ReportID:        "r1",
	})
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	html := string(got)
	if !strings.Contains(html, "abc123") || !strings.Contains(html, "secret") || !strings.Contains(html, "abcd") {
		t.Fatalf("expected raw values with redaction disabled, got: %s", html)
	}
	if !strings.Contains(html, "REDACTION: OFF") {
		t.Fatalf("expected redaction status OFF in html")
	}
	if !strings.Contains(html, "Metadata: report-id=r1 | mode=auto | policy=v1 | unsafe-consent=yes") {
		t.Fatalf("expected metadata for unredacted mode")
	}
	if !strings.Contains(html, "redaction disabled") {
		t.Fatalf("expected unredacted warning in html")
	}
}
