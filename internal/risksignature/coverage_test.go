package risksignature

import (
	"testing"

	"network-scanner/internal/scanner"
)

// ============================================================================
// Load — edge cases (72.7% → 100%)
// ============================================================================

func TestLoad_MissingVersion(t *testing.T) {
	raw := []byte(`{"signatures":[{"id":"s1","title":"Test"}]}`)
	_, err := Load(raw)
	if err == nil {
		t.Fatal("expected error for missing version")
	}
}

func TestLoad_MissingID(t *testing.T) {
	raw := []byte(`{"version":"1.0","signatures":[{"title":"Test"}]}`)
	_, err := Load(raw)
	if err == nil {
		t.Fatal("expected error for missing signature ID")
	}
}

func TestLoad_MissingTitle(t *testing.T) {
	raw := []byte(`{"version":"1.0","signatures":[{"id":"s1"}]}`)
	_, err := Load(raw)
	if err == nil {
		t.Fatal("expected error for missing title")
	}
}

func TestLoad_EmptySignatures(t *testing.T) {
	raw := []byte(`{"version":"1.0","signatures":[]}`)
	db, err := Load(raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(db.Signatures) != 0 {
		t.Fatalf("expected 0 signatures, got %d", len(db.Signatures))
	}
}

func TestLoad_Valid(t *testing.T) {
	raw := []byte(`{"version":"1.0","signatures":[{"id":"s1","title":"Test","severity":"high"}]}`)
	db, err := Load(raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if db.Version != "1.0" {
		t.Fatalf("expected version 1.0, got %q", db.Version)
	}
	if len(db.Signatures) != 1 {
		t.Fatalf("expected 1 signature, got %d", len(db.Signatures))
	}
}

// ============================================================================
// matchBanners — 18.2% → 100%
// ============================================================================

func TestMatchBanners_MatchBanner(t *testing.T) {
	host := scanner.Result{
		Ports: []scanner.PortInfo{
			{State: "open", Banner: "Apache/2.4.41", Version: ""},
		},
	}
	reason, ok := matchBanners(host, []string{"apache"})
	if !ok {
		t.Fatal("expected match for apache banner")
	}
	if reason == "" {
		t.Fatal("expected non-empty reason")
	}
}

func TestMatchBanners_MatchVersion(t *testing.T) {
	host := scanner.Result{
		Ports: []scanner.PortInfo{
			{State: "open", Banner: "", Version: "OpenSSH_7.4"},
		},
	}
	_, ok := matchBanners(host, []string{"openssh"})
	if !ok {
		t.Fatal("expected match for openssh version")
	}
}

func TestMatchBanners_MatchCombined(t *testing.T) {
	host := scanner.Result{
		Ports: []scanner.PortInfo{
			{State: "open", Banner: "nginx", Version: "1.18.0"},
		},
	}
	_, ok := matchBanners(host, []string{"nginx 1.18"})
	if !ok {
		t.Fatal("expected match for combined banner+version")
	}
}

func TestMatchBanners_NoMatch(t *testing.T) {
	host := scanner.Result{
		Ports: []scanner.PortInfo{
			{State: "open", Banner: "Apache", Version: ""},
		},
	}
	_, ok := matchBanners(host, []string{"nginx"})
	if ok {
		t.Fatal("expected no match for nginx pattern on Apache banner")
	}
}

func TestMatchBanners_ClosedPort(t *testing.T) {
	host := scanner.Result{
		Ports: []scanner.PortInfo{
			{State: "closed", Banner: "Apache"},
		},
	}
	_, ok := matchBanners(host, []string{"apache"})
	if ok {
		t.Fatal("expected no match on closed port")
	}
}

func TestMatchBanners_EmptyPatterns(t *testing.T) {
	host := scanner.Result{
		Ports: []scanner.PortInfo{
			{State: "open", Banner: "Apache"},
		},
	}
	_, ok := matchBanners(host, []string{})
	if ok {
		t.Fatal("expected no match for empty patterns")
	}
}

func TestMatchBanners_CaseInsensitive(t *testing.T) {
	host := scanner.Result{
		Ports: []scanner.PortInfo{
			{State: "open", Banner: "APACHE/2.4"},
		},
	}
	_, ok := matchBanners(host, []string{"apache"})
	if !ok {
		t.Fatal("expected case-insensitive match")
	}
}

func TestMatchBanners_EmptyPatternSkipped(t *testing.T) {
	host := scanner.Result{
		Ports: []scanner.PortInfo{
			{State: "open", Banner: "Apache"},
		},
	}
	_, ok := matchBanners(host, []string{"", "nginx"})
	if ok {
		t.Fatal("expected no match when only empty and non-matching patterns")
	}
}

// ============================================================================
// matchPorts — 77.8% → 100%
// ============================================================================

func TestMatchPorts_ClosedPort(t *testing.T) {
	host := scanner.Result{
		Ports: []scanner.PortInfo{
			{Port: 80, State: "closed"},
		},
	}
	_, ok := matchPorts(host, []int{80})
	if ok {
		t.Fatal("expected no match on closed port")
	}
}

func TestMatchPorts_NoPorts(t *testing.T) {
	host := scanner.Result{}
	_, ok := matchPorts(host, []int{80})
	if ok {
		t.Fatal("expected no match when host has no ports")
	}
}

// ============================================================================
// matchStringAny — 87.5% → 100%
// ============================================================================

func TestMatchStringAny_EmptyValue(t *testing.T) {
	_, ok := matchStringAny("", []string{"router"}, "device_type")
	if ok {
		t.Fatal("expected no match for empty value")
	}
}

func TestMatchStringAny_EmptyWantedItem(t *testing.T) {
	_, ok := matchStringAny("router", []string{""}, "device_type")
	if ok {
		t.Fatal("expected no match for empty wanted item")
	}
}

// ============================================================================
// normalizeSeverity — 60.0% → 100%
// ============================================================================

func TestNormalizeSeverity_Critical(t *testing.T) {
	if got := normalizeSeverity("critical"); got != "critical" {
		t.Fatalf("expected critical, got %q", got)
	}
}

func TestNormalizeSeverity_High(t *testing.T) {
	if got := normalizeSeverity("high"); got != "high" {
		t.Fatalf("expected high, got %q", got)
	}
}

func TestNormalizeSeverity_Medium(t *testing.T) {
	if got := normalizeSeverity("medium"); got != "medium" {
		t.Fatalf("expected medium, got %q", got)
	}
}

func TestNormalizeSeverity_Low(t *testing.T) {
	if got := normalizeSeverity("low"); got != "low" {
		t.Fatalf("expected low, got %q", got)
	}
}

func TestNormalizeSeverity_Unknown(t *testing.T) {
	if got := normalizeSeverity("unknown"); got != "low" {
		t.Fatalf("expected low for unknown, got %q", got)
	}
}

func TestNormalizeSeverity_Empty(t *testing.T) {
	if got := normalizeSeverity(""); got != "low" {
		t.Fatalf("expected low for empty, got %q", got)
	}
}

func TestNormalizeSeverity_Uppercase(t *testing.T) {
	if got := normalizeSeverity("CRITICAL"); got != "critical" {
		t.Fatalf("expected critical for uppercase, got %q", got)
	}
}

func TestNormalizeSeverity_Whitespace(t *testing.T) {
	if got := normalizeSeverity("  HIGH  "); got != "high" {
		t.Fatalf("expected high for whitespace, got %q", got)
	}
}

// ============================================================================
// Evaluate — edge cases
// ============================================================================

func TestEvaluate_EmptyResults(t *testing.T) {
	db, _ := LoadDefault()
	findings := Evaluate([]scanner.Result{}, db)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestEvaluate_EmptySignatures(t *testing.T) {
	db := SignatureDB{Version: "1.0", Signatures: []Signature{}}
	results := []scanner.Result{{IP: "10.0.0.1"}}
	findings := Evaluate(results, db)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestEvaluate_BannerMatch(t *testing.T) {
	db := SignatureDB{
		Version: "1.0",
		Signatures: []Signature{
			{
				ID:             "sig-banner",
				Title:          "Insecure Banner",
				Severity:       "high",
				MatchAnyBanner: []string{"apache"},
			},
		},
	}
	results := []scanner.Result{
		{
			IP: "10.0.0.1",
			Ports: []scanner.PortInfo{
				{Port: 80, State: "open", Banner: "Apache/2.4.41"},
			},
		},
	}
	findings := Evaluate(results, db)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].SignatureID != "sig-banner" {
		t.Fatalf("expected sig-banner, got %q", findings[0].SignatureID)
	}
}

func TestEvaluate_MultipleHostsSorted(t *testing.T) {
	db := SignatureDB{
		Version: "1.0",
		Signatures: []Signature{
			{ID: "sig-1", Title: "Test", Severity: "high", MatchAnyPort: []int{23}},
		},
	}
	results := []scanner.Result{
		{IP: "10.0.0.2", Ports: []scanner.PortInfo{{Port: 23, State: "open"}}},
		{IP: "10.0.0.1", Ports: []scanner.PortInfo{{Port: 23, State: "open"}}},
	}
	findings := Evaluate(results, db)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
	if findings[0].HostIP != "10.0.0.1" {
		t.Fatalf("expected first host 10.0.0.1, got %q", findings[0].HostIP)
	}
}

func TestEvaluate_VendorMatch(t *testing.T) {
	db := SignatureDB{
		Version: "1.0",
		Signatures: []Signature{
			{ID: "sig-vendor", Title: "Vendor Risk", Severity: "medium", MatchVendor: []string{"cisco"}},
		},
	}
	results := []scanner.Result{
		{IP: "10.0.0.1", DeviceVendor: "Cisco Systems"},
	}
	findings := Evaluate(results, db)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}
