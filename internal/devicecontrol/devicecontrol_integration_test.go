package devicecontrol

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// === Integration: Execute — Request Validation ===

func TestIntegrationExecute_EmptyAction(t *testing.T) {
	_, err := Execute(context.Background(), Request{
		Action:    "",
		TargetURL: "http://127.0.0.1:9999",
	})
	if err == nil {
		t.Error("expected error for empty action")
	}
}

func TestIntegrationExecute_EmptyTargetURL(t *testing.T) {
	_, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: "",
	})
	if err == nil {
		t.Error("expected error for empty target URL")
	}
}

func TestIntegrationExecute_InvalidURLScheme(t *testing.T) {
	_, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: "ftp://192.168.1.1",
	})
	if err == nil {
		t.Error("expected error for non-http URL")
	}
}

func TestIntegrationExecute_InvalidURLScheme2(t *testing.T) {
	_, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: "192.168.1.1",
	})
	if err == nil {
		t.Error("expected error for URL without scheme")
	}
}

func TestIntegrationExecute_UnknownAction(t *testing.T) {
	_, err := Execute(context.Background(), Request{
		Action:    "shutdown",
		TargetURL: "http://127.0.0.1:9999",
	})
	if err == nil {
		t.Error("expected error for unknown action")
	}
}

func TestIntegrationExecute_UnknownVendor(t *testing.T) {
	_, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: "http://127.0.0.1:9999",
		Vendor:    "unknown-vendor",
	})
	if err == nil {
		t.Error("expected error for unknown vendor")
	}
}

// === Integration: Execute — HTTP Tests ===

func TestIntegrationExecute_GenericHTTP_Status(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/status" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("status ok"))
	}))
	defer srv.Close()

	res, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: srv.URL,
		Timeout:   2 * time.Second,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !res.Success {
		t.Error("expected success=true")
	}
	if res.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", res.StatusCode)
	}
	if res.Message != "status ok" {
		t.Errorf("expected message 'status ok', got %q", res.Message)
	}
	if res.Action != ActionStatus {
		t.Errorf("expected action %q, got %q", ActionStatus, res.Action)
	}
}

func TestIntegrationExecute_GenericHTTP_Reboot(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/reboot" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	res, err := Execute(context.Background(), Request{
		Action:    ActionReboot,
		TargetURL: srv.URL,
		Timeout:   2 * time.Second,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !res.Success {
		t.Error("expected success=true")
	}
}

func TestIntegrationExecute_TPLinkHTTP_Status(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/system/status" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	res, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: srv.URL,
		Vendor:    VendorTPLINKHTTP,
		Timeout:   2 * time.Second,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !res.Success {
		t.Error("expected success=true")
	}
}

func TestIntegrationExecute_TPLinkHTTP_Reboot(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/system/reboot" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	res, err := Execute(context.Background(), Request{
		Action:    ActionReboot,
		TargetURL: srv.URL,
		Vendor:    VendorTPLINKHTTP,
		Timeout:   2 * time.Second,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !res.Success {
		t.Error("expected success=true")
	}
}

func TestIntegrationExecute_TPLinkHTTP_UnknownAction(t *testing.T) {
	_, err := Execute(context.Background(), Request{
		Action:    "shutdown",
		TargetURL: "http://127.0.0.1:9999",
		Vendor:    VendorTPLINKHTTP,
	})
	if err == nil {
		t.Error("expected error for unsupported action with tplink vendor")
	}
}

func TestIntegrationExecute_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server error"))
	}))
	defer srv.Close()

	res, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: srv.URL,
		Timeout:   2 * time.Second,
	})
	if err == nil {
		t.Error("expected error for 500 status")
	}
	if res.Success {
		t.Error("expected success=false for 500 status")
	}
	if res.StatusCode != 500 {
		t.Errorf("expected status 500, got %d", res.StatusCode)
	}
}

func TestIntegrationExecute_404Status(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	res, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: srv.URL,
		Timeout:   2 * time.Second,
	})
	if err == nil {
		t.Error("expected error for 404 status")
	}
	if res.Success {
		t.Error("expected success=false for 404")
	}
}

func TestIntegrationExecute_ContextCancellation(t *testing.T) {
	// Use a slow server that takes longer than the context timeout
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := Execute(ctx, Request{
		Action:    ActionStatus,
		TargetURL: srv.URL,
		Timeout:   2 * time.Second,
	})
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestIntegrationExecute_BasicAuth(t *testing.T) {
	var receivedAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			receivedAuth = username + ":" + password
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	_, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: srv.URL,
		Username:  "admin",
		Password:  "secret",
		Timeout:   2 * time.Second,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if receivedAuth != "admin:secret" {
		t.Errorf("expected auth 'admin:secret', got %q", receivedAuth)
	}
}

func TestIntegrationExecute_NoAuth(t *testing.T) {
	var receivedAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, _, ok := r.BasicAuth()
		if ok {
			receivedAuth = username
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	_, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: srv.URL,
		Timeout:   2 * time.Second,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if receivedAuth != "" {
		t.Error("expected no auth when username is empty")
	}
}

// === Integration: Response Structure ===

func TestIntegrationResponse_Structure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test message"))
	}))
	defer srv.Close()

	res, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: srv.URL,
		Timeout:   2 * time.Second,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res.Action == "" {
		t.Error("expected non-empty action")
	}
	if res.TargetURL == "" {
		t.Error("expected non-empty target URL")
	}
	if !res.Success {
		t.Error("expected success=true for 200 status")
	}
	if res.StatusCode == 0 {
		t.Error("expected non-zero status code")
	}
}

// === Integration: Constants ===

func TestIntegrationConstants(t *testing.T) {
	if ActionStatus != "status" {
		t.Errorf("expected ActionStatus='status', got %q", ActionStatus)
	}
	if ActionReboot != "reboot" {
		t.Errorf("expected ActionReboot='reboot', got %q", ActionReboot)
	}
	if VendorGenericHTTP != "generic-http" {
		t.Errorf("expected VendorGenericHTTP='generic-http', got %q", VendorGenericHTTP)
	}
	if VendorTPLINKHTTP != "tp-link-http" {
		t.Errorf("expected VendorTPLINKHTTP='tp-link-http', got %q", VendorTPLINKHTTP)
	}
}

// === Integration: AuditEntry ===

func TestIntegrationAuditEntry_Structure(t *testing.T) {
	entry := AuditEntry{
		Timestamp: "2026-01-01T00:00:00Z",
		Actor:     "admin",
		Action:    "reboot",
		TargetURL: "http://192.168.1.1",
		Vendor:    "tplink-http",
		Success:   true,
		Message:   "rebooted",
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty JSON")
	}

	var parsed AuditEntry
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if parsed.Actor != "admin" {
		t.Errorf("expected actor 'admin', got %q", parsed.Actor)
	}
	if parsed.Success != true {
		t.Errorf("expected success true, got %v", parsed.Success)
	}
}

// === Integration: AppendAudit ===

func TestIntegrationAppendAudit_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	auditPath := filepath.Join(tmpDir, "audit.jsonl")

	entry := AuditEntry{
		Timestamp: "2026-01-01T00:00:00Z",
		Actor:     "test-user",
		Action:    "reboot",
		TargetURL: "http://192.168.1.1",
		Success:   true,
		Message:   "ok",
	}

	err := AppendAudit(auditPath, entry)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	content, err := os.ReadFile(auditPath)
	if err != nil {
		t.Fatalf("failed to read audit file: %v", err)
	}
	if len(content) == 0 {
		t.Error("expected non-empty audit file")
	}

	var parsed AuditEntry
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("failed to parse audit entry: %v", err)
	}
	if parsed.Action != "reboot" {
		t.Errorf("expected action 'reboot', got %q", parsed.Action)
	}
}

func TestIntegrationAppendAudit_DefaultTimestamp(t *testing.T) {
	tmpDir := t.TempDir()
	auditPath := filepath.Join(tmpDir, "audit2.jsonl")

	entry := AuditEntry{
		Actor:   "test-user",
		Action:  "status",
		Message: "ok",
	}

	err := AppendAudit(auditPath, entry)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	content, err := os.ReadFile(auditPath)
	if err != nil {
		t.Fatalf("failed to read audit file: %v", err)
	}

	var parsed AuditEntry
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("failed to parse audit entry: %v", err)
	}
	if parsed.Timestamp == "" {
		t.Error("expected non-empty timestamp (should be auto-generated)")
	}
}

func TestIntegrationAppendAudit_DefaultActor(t *testing.T) {
	tmpDir := t.TempDir()
	auditPath := filepath.Join(tmpDir, "audit3.jsonl")

	entry := AuditEntry{
		Timestamp: "2026-01-01T00:00:00Z",
		Action:    "status",
		Message:   "ok",
	}

	err := AppendAudit(auditPath, entry)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	content, err := os.ReadFile(auditPath)
	if err != nil {
		t.Fatalf("failed to read audit file: %v", err)
	}

	var parsed AuditEntry
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("failed to parse audit entry: %v", err)
	}
	if parsed.Actor == "" {
		t.Error("expected non-empty actor (should be auto-detected)")
	}
}

func TestIntegrationAppendAudit_EmptyPath(t *testing.T) {
	entry := AuditEntry{
		Action: "status",
	}
	err := AppendAudit("", entry)
	if err == nil {
		t.Error("expected error for empty path")
	}
}

func TestIntegrationAppendAudit_WhitespacePath(t *testing.T) {
	entry := AuditEntry{
		Action: "status",
	}
	err := AppendAudit("   ", entry)
	if err == nil {
		t.Error("expected error for whitespace path")
	}
}

func TestIntegrationAppendAudit_SubdirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()
	auditPath := filepath.Join(tmpDir, "subdir", "nested", "audit.jsonl")

	entry := AuditEntry{
		Action: "reboot",
		Actor:  "test",
	}

	err := AppendAudit(auditPath, entry)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, err := os.Stat(auditPath); os.IsNotExist(err) {
		t.Error("expected audit file to exist")
	}
}

// === Integration: Full Control Pipeline ===

func TestIntegrationFullControlPipeline_Valid(t *testing.T) {
	var receivedBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 256)
		n, _ := r.Body.Read(buf)
		receivedBody = string(buf[:n])
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))
	defer srv.Close()

	// Step 1: Execute control action
	res, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: srv.URL,
		Vendor:    VendorGenericHTTP,
		Username:  "admin",
		Password:  "pass",
		Timeout:   2 * time.Second,
	})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if !res.Success {
		t.Error("expected success")
	}

	// Step 2: Verify request body
	if receivedBody == "" {
		t.Error("expected non-empty request body")
	}

	// Step 3: Write audit entry
	tmpDir := t.TempDir()
	auditPath := filepath.Join(tmpDir, "audit.jsonl")
	auditEntry := AuditEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Actor:     "admin",
		Action:    res.Action,
		TargetURL: res.TargetURL,
		Success:   res.Success,
		Message:   res.Message,
	}
	err = AppendAudit(auditPath, auditEntry)
	if err != nil {
		t.Fatalf("AppendAudit error: %v", err)
	}

	// Step 4: Verify audit file
	content, err := os.ReadFile(auditPath)
	if err != nil {
		t.Fatalf("failed to read audit file: %v", err)
	}
	if len(content) == 0 {
		t.Error("expected non-empty audit file")
	}
}

// === Integration: Request Field Normalization ===

func TestIntegrationRequest_NormalizeAction(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Uppercase action should work
	res, err := Execute(context.Background(), Request{
		Action:    "STATUS",
		TargetURL: srv.URL,
		Timeout:   2 * time.Second,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !res.Success {
		t.Error("expected success")
	}
}

func TestIntegrationRequest_NormalizeVendor(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Mixed case vendor should work
	_, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: srv.URL,
		Vendor:    "TP-LINK-HTTP",
		Timeout:   2 * time.Second,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestIntegrationRequest_DefaultVendor(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/status" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	_, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: srv.URL,
		Timeout:   2 * time.Second,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// === Integration: Response Fields ===

func TestIntegrationResponse_AllFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("custom message"))
	}))
	defer srv.Close()

	res, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: srv.URL,
		Timeout:   2 * time.Second,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res.Action == "" {
		t.Error("expected non-empty action")
	}
	if res.TargetURL == "" {
		t.Error("expected non-empty target URL")
	}
	if !res.Success {
		t.Error("expected success")
	}
	if res.StatusCode == 0 {
		t.Error("expected non-zero status code")
	}
	if res.Message == "" {
		t.Error("expected non-empty message")
	}
}

// === Integration: JSON Content-Type ===

func TestIntegrationRequest_ContentType(t *testing.T) {
	var receivedContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	_, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: srv.URL,
		Timeout:   2 * time.Second,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if receivedContentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", receivedContentType)
	}
}

// === Integration: JSON Payload Structure ===

func TestIntegrationRequest_JSONPayload(t *testing.T) {
	var receivedBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 256)
		n, _ := r.Body.Read(buf)
		receivedBody = string(buf[:n])
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	_, err := Execute(context.Background(), Request{
		Action:    ActionReboot,
		TargetURL: srv.URL,
		Vendor:    VendorTPLINKHTTP,
		Timeout:   2 * time.Second,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var payload map[string]string
	if err := json.Unmarshal([]byte(receivedBody), &payload); err != nil {
		t.Fatalf("failed to parse JSON payload: %v", err)
	}
	if payload["action"] != "reboot" {
		t.Errorf("expected action 'reboot', got %q", payload["action"])
	}
}

// === Integration: URL Trailing Slash ===

func TestIntegrationExecute_TrailingSlash(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// URL with trailing slash should work
	_, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: srv.URL + "/",
		Timeout:   2 * time.Second,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// === Integration: Default Timeout ===

func TestIntegrationExecute_DefaultTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Zero timeout should default to 10 seconds
	res, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: srv.URL,
		Timeout:   0,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !res.Success {
		t.Error("expected success")
	}
}

// === Integration: HTTPS Support ===

func TestIntegrationExecute_HTTPS(t *testing.T) {
	// httptest.NewTLSServer creates an HTTPS server with self-signed cert
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// After TASK-028.3, InsecureTLS is implemented in Execute
	// The request should succeed because InsecureTLS skips cert verification
	resp, err := Execute(context.Background(), Request{
		Action:      ActionStatus,
		TargetURL:   srv.URL,
		Timeout:     2 * time.Second,
		InsecureTLS: true,
	})
	if err != nil {
		t.Fatalf("Execute should succeed with InsecureTLS=true, got error: %v", err)
	}
	if !resp.Success {
		t.Error("expected Success=true")
	}
}

// === Integration: Response Message from Status Text ===

func TestIntegrationResponse_MessageFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// No body - should fallback to http.StatusText
	}))
	defer srv.Close()

	res, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: srv.URL,
		Timeout:   2 * time.Second,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Message == "" {
		t.Error("expected non-empty message (should fallback to StatusText)")
	}
}

// === Integration: Bulk Operations ===

func TestIntegrationBulkOperations_Multiple(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Execute multiple operations
	for i := 0; i < 3; i++ {
		_, err := Execute(context.Background(), Request{
			Action:    ActionStatus,
			TargetURL: srv.URL,
			Timeout:   2 * time.Second,
		})
		if err != nil {
			t.Fatalf("iteration %d: expected no error, got %v", i, err)
		}
	}

	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

// === Integration: Audit Entry JSONL Format ===

func TestIntegrationAudit_JSONLFormat(t *testing.T) {
	tmpDir := t.TempDir()
	auditPath := filepath.Join(tmpDir, "audit.jsonl")

	// Write multiple entries
	for i := 0; i < 3; i++ {
		entry := AuditEntry{
			Timestamp: "2026-01-01T00:00:00Z",
			Actor:     "test",
			Action:    "status",
			TargetURL: "http://192.168.1.1",
			Success:   true,
			Message:   "ok",
		}
		err := AppendAudit(auditPath, entry)
		if err != nil {
			t.Fatalf("entry %d: expected no error, got %v", i, err)
		}
	}

	content, err := os.ReadFile(auditPath)
	if err != nil {
		t.Fatalf("failed to read audit file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 JSONL lines, got %d", len(lines))
	}

	// Each line should be valid JSON
	for i, line := range lines {
		var entry AuditEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Errorf("line %d: failed to parse JSON: %v", i, err)
		}
	}
}
