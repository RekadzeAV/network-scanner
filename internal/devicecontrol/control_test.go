package devicecontrol

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestExecute_StatusOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/status" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	res, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: srv.URL,
		Timeout:   2 * time.Second,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !res.Success {
		t.Fatalf("expected success=true, got false")
	}
}

func TestExecute_RebootNeedsValidURL(t *testing.T) {
	_, err := Execute(context.Background(), Request{
		Action:    ActionReboot,
		TargetURL: "192.168.1.1",
	})
	if err == nil {
		t.Fatal("expected url validation error")
	}
}

func TestExecute_TPLinkAdapterRoutes(t *testing.T) {
	gotPath := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()
	_, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: srv.URL,
		Vendor:    VendorTPLINKHTTP,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if gotPath != "/api/system/status" {
		t.Fatalf("unexpected adapter path: %s", gotPath)
	}
}
