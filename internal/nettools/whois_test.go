package nettools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBuildRDAPURLDomain(t *testing.T) {
	u := buildRDAPURL("Example.COM")
	if !strings.Contains(u, "/domain/example.com") {
		t.Fatalf("unexpected rdap domain url: %s", u)
	}
}

func TestBuildRDAPURLIP(t *testing.T) {
	u := buildRDAPURL("8.8.8.8")
	if !strings.Contains(u, "/ip/8.8.8.8") {
		t.Fatalf("unexpected rdap ip url: %s", u)
	}
}

func TestBuildRDAPURLUsesCustomBase(t *testing.T) {
	t.Setenv("NETWORK_SCANNER_RDAP_BASE_URL", "http://127.0.0.1:18080/")
	u := buildRDAPURL("example.com")
	if !strings.HasPrefix(u, "http://127.0.0.1:18080/domain/") {
		t.Fatalf("expected custom base URL, got: %s", u)
	}
}

func TestBuildRDAPURLFallsBackOnInvalidCustomBase(t *testing.T) {
	t.Setenv("NETWORK_SCANNER_RDAP_BASE_URL", "not-a-url")
	u := buildRDAPURL("example.com")
	if !strings.HasPrefix(u, "https://rdap.org/domain/") {
		t.Fatalf("expected fallback default base URL, got: %s", u)
	}
}

func TestResolveRDAPBaseURLDefaultWhenUnset(t *testing.T) {
	t.Setenv("NETWORK_SCANNER_RDAP_BASE_URL", "")
	got := resolveRDAPBaseURL()
	if got != "https://rdap.org" {
		t.Fatalf("expected default RDAP base URL, got: %s", got)
	}
}

func TestResolveRDAPBaseURLTrimsTrailingSlash(t *testing.T) {
	t.Setenv("NETWORK_SCANNER_RDAP_BASE_URL", "https://rdap.example.test////")
	got := resolveRDAPBaseURL()
	if got != "https://rdap.example.test" {
		t.Fatalf("expected trailing slash trimmed, got: %s", got)
	}
}

func TestCompactRDAPInvalid(t *testing.T) {
	if _, ok := compactRDAP([]byte("not-json")); ok {
		t.Fatal("expected compactRDAP to fail on invalid json")
	}
}

func TestFormatRDAPSummary(t *testing.T) {
	body := []byte(`{
	  "objectClassName":"domain",
	  "ldhName":"example.com",
	  "status":["active"],
	  "events":[
	    {"eventAction":"registration","eventDate":"1995-08-14T04:00:00Z"},
	    {"eventAction":"last changed","eventDate":"2025-01-01T00:00:00Z"}
	  ],
	  "entities":[
	    {"roles":["registrar"],"vcardArray":["vcard",[["fn",{}, "text","Example Registrar Inc."]]]}
	  ]
	}`)
	s, ok := formatRDAPSummary(body, "example.com")
	if !ok {
		t.Fatal("expected formatted summary")
	}
	if !strings.Contains(s, "example.com") {
		t.Fatalf("expected query/name in summary: %s", s)
	}
	if !strings.Contains(strings.ToLower(s), "registrar") {
		t.Fatalf("expected registrar in summary: %s", s)
	}
}

func TestRunWhoisRDAPReturnsSummary(t *testing.T) {
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
	out, err := runWhoisRDAP(context.Background(), "example.com", 2*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "RDAP summary") {
		t.Fatalf("expected RDAP summary output, got: %s", out)
	}
}

func TestRunWhoisRDAPReturnsRawWhenSummaryUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(`raw-rdap-payload`))
	}))
	defer srv.Close()

	t.Setenv("NETWORK_SCANNER_RDAP_BASE_URL", srv.URL)
	out, err := runWhoisRDAP(context.Background(), "example.com", 2*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "raw-rdap-payload") {
		t.Fatalf("expected raw RDAP output fallback, got: %s", out)
	}
}

func TestRunWhoisRDAPHandlesHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusBadGateway)
	}))
	defer srv.Close()

	t.Setenv("NETWORK_SCANNER_RDAP_BASE_URL", srv.URL)
	_, err := runWhoisRDAP(context.Background(), "example.com", 2*time.Second)
	if err == nil {
		t.Fatal("expected HTTP error from RDAP")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "http 502") {
		t.Fatalf("expected HTTP status in error, got: %v", err)
	}
}

