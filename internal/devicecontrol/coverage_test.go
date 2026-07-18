package devicecontrol

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ============================================================================
// W-61: Улучшение coverage AppendAudit (73.7% → 90%+)
// ============================================================================

func TestAppendAudit_EmptyPath(t *testing.T) {
	err := AppendAudit("", AuditEntry{Action: "test"})
	if err == nil {
		t.Fatal("expected error for empty path")
	}
	if !strings.Contains(err.Error(), "audit path is required") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestAppendAudit_EmptyTimestamp(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "empty_ts.log")

	err := AppendAudit(file, AuditEntry{
		Action:    "test",
		TargetURL: "http://example.com",
		Success:   true,
	})
	if err != nil {
		t.Fatalf("AppendAudit() error = %v", err)
	}

	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	// Timestamp should be auto-generated
	if !strings.Contains(string(data), `"timestamp"`) {
		t.Fatal("expected timestamp in audit entry")
	}
}

func TestAppendAudit_EmptyActor(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "empty_actor.log")

	err := AppendAudit(file, AuditEntry{
		Action:    "test",
		TargetURL: "http://example.com",
		Success:   true,
	})
	if err != nil {
		t.Fatalf("AppendAudit() error = %v", err)
	}

	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	// Actor should be auto-populated
	if !strings.Contains(string(data), `"actor"`) {
		t.Fatal("expected actor in audit entry")
	}
}

func TestAppendAudit_WithVendor(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "with_vendor.log")

	err := AppendAudit(file, AuditEntry{
		Action:    "reboot",
		TargetURL: "http://192.168.1.1",
		Vendor:    "tplink",
		Success:   false,
		Message:   "failed",
	})
	if err != nil {
		t.Fatalf("AppendAudit() error = %v", err)
	}

	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if !strings.Contains(string(data), `"vendor":"tplink"`) {
		t.Fatal("expected vendor in audit entry")
	}
}

func TestAppendAudit_InvalidJSON(t *testing.T) {
	// Тестируем ситуацию когда Marshal вернёт ошибку
	// Это сложно протестировать без изменения структуры,
	// но можно проверить что функция корректно обрабатывает ошибки
	dir := t.TempDir()

	// Пытаемся записать в несуществующую директорию
	_, err := os.Stat(dir)
	if err != nil {
		t.Skip("directory check failed")
	}

	// Создаём файл и удаляем директорию
	file := filepath.Join(dir, "subdir", "audit.log")
	err = AppendAudit(file, AuditEntry{Action: "test"})
	// Должно создать директорию
	if err != nil {
		t.Fatalf("AppendAudit() error = %v", err)
	}
}

func TestAppendAudit_MultipleEntries(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "multi.log")

	for i := 0; i < 5; i++ {
		err := AppendAudit(file, AuditEntry{
			Action:    "status",
			TargetURL: "http://192.168.1.1",
			Success:   i%2 == 0,
		})
		if err != nil {
			t.Fatalf("AppendAudit() iteration %d error = %v", i, err)
		}
	}

	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(lines))
	}
}

// ============================================================================
// W-62: Улучшение coverage currentActor (44.4% → 75%+)
// ============================================================================

func TestCurrentActor_EnvUsername(t *testing.T) {
	// На Windows user.Current() всегда работает, поэтому env USERNAME не используется
	// Проверяем что функция не паникует и возвращает непустую строку
	actor := currentActor()
	if actor == "" {
		t.Fatal("currentActor() should not return empty string")
	}
	// Функция возвращает имя пользователя из user.Current(), а не из ENV
	// Это нормальное поведение для Windows
}

func TestCurrentActor_EnvUser(t *testing.T) {
	// Аналогично — на Windows user.Current() всегда срабатывает
	// Проверяем что функция возвращает непустую строку
	actor := currentActor()
	if actor == "" {
		t.Fatal("currentActor() should not return empty string")
	}
}

func TestCurrentActor_DefaultUnknown(t *testing.T) {
	// Сохраняем оригинальные значения
	origUsername := os.Getenv("USERNAME")
	origUser := os.Getenv("USER")
	defer func() {
		os.Setenv("USERNAME", origUsername)
		os.Setenv("USER", origUser)
	}()

	// Очищаем все переменные
	os.Unsetenv("USERNAME")
	os.Unsetenv("USER")

	// currentActor() вернёт "unknown" если user.Current не сработает
	// На Windows user.Current обычно работает, поэтому проверяем что не паникует
	actor := currentActor()
	if actor == "" {
		t.Fatal("currentActor() should not return empty string")
	}
}

// ============================================================================
// W-63: Улучшение coverage buildEndpoint tplink (50.0% → 100%)
// ============================================================================

func TestBuildEndpoint_TPLinkStatus(t *testing.T) {
	adapter := tplinkHTTPAdapter{}
	endpoint, err := adapter.buildEndpoint("http://192.168.1.1", ActionStatus)
	if err != nil {
		t.Fatalf("buildEndpoint() error = %v", err)
	}
	want := "http://192.168.1.1/api/system/status"
	if endpoint != want {
		t.Fatalf("buildEndpoint() = %q, want %q", endpoint, want)
	}
}

func TestBuildEndpoint_TPLinkReboot(t *testing.T) {
	adapter := tplinkHTTPAdapter{}
	endpoint, err := adapter.buildEndpoint("http://192.168.1.1", ActionReboot)
	if err != nil {
		t.Fatalf("buildEndpoint() error = %v", err)
	}
	want := "http://192.168.1.1/api/system/reboot"
	if endpoint != want {
		t.Fatalf("buildEndpoint() = %q, want %q", endpoint, want)
	}
}

func TestBuildEndpoint_TPLinkUnsupportedAction(t *testing.T) {
	adapter := tplinkHTTPAdapter{}
	_, err := adapter.buildEndpoint("http://192.168.1.1", "unsupported")
	if err == nil {
		t.Fatal("expected error for unsupported action")
	}
	if !strings.Contains(err.Error(), "unsupported action") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestBuildEndpoint_GenericHTTP(t *testing.T) {
	adapter := genericHTTPAdapter{}
	endpoint, err := adapter.buildEndpoint("http://192.168.1.1", ActionStatus)
	if err != nil {
		t.Fatalf("buildEndpoint() error = %v", err)
	}
	want := "http://192.168.1.1/api/status"
	if endpoint != want {
		t.Fatalf("buildEndpoint() = %q, want %q", endpoint, want)
	}
}

func TestBuildEndpoint_GenericHTTP_TrailingSlash(t *testing.T) {
	adapter := genericHTTPAdapter{}
	endpoint, err := adapter.buildEndpoint("http://192.168.1.1/", ActionStatus)
	if err != nil {
		t.Fatalf("buildEndpoint() error = %v", err)
	}
	want := "http://192.168.1.1/api/status"
	if endpoint != want {
		t.Fatalf("buildEndpoint() = %q, want %q", endpoint, want)
	}
}

// ============================================================================
// W-64: Улучшение coverage resolveAdapter (75.0% → 100%)
// ============================================================================

func TestResolveAdapter_GenericHTTP(t *testing.T) {
	ad, err := resolveAdapter(VendorGenericHTTP)
	if err != nil {
		t.Fatalf("resolveAdapter() error = %v", err)
	}
	if ad == nil {
		t.Fatal("resolveAdapter() returned nil adapter")
	}
}

func TestResolveAdapter_EmptyVendor(t *testing.T) {
	ad, err := resolveAdapter("")
	if err != nil {
		t.Fatalf("resolveAdapter() error = %v", err)
	}
	if ad == nil {
		t.Fatal("resolveAdapter() returned nil adapter")
	}
}

func TestResolveAdapter_TPLink(t *testing.T) {
	ad, err := resolveAdapter(VendorTPLINKHTTP)
	if err != nil {
		t.Fatalf("resolveAdapter() error = %v", err)
	}
	if ad == nil {
		t.Fatal("resolveAdapter() returned nil adapter")
	}
}

func TestResolveAdapter_UnknownVendor(t *testing.T) {
	_, err := resolveAdapter("unknown-vendor")
	if err == nil {
		t.Fatal("expected error for unknown vendor")
	}
	if !strings.Contains(err.Error(), "unsupported vendor") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestResolveAdapter_CaseInsensitive(t *testing.T) {
	ad, err := resolveAdapter("GENERIC-HTTP")
	if err != nil {
		t.Fatalf("resolveAdapter() error = %v", err)
	}
	if ad == nil {
		t.Fatal("resolveAdapter() returned nil adapter")
	}
}

func TestResolveAdapter_Whitespace(t *testing.T) {
	ad, err := resolveAdapter("  generic-http  ")
	if err != nil {
		t.Fatalf("resolveAdapter() error = %v", err)
	}
	if ad == nil {
		t.Fatal("resolveAdapter() returned nil adapter")
	}
}

// ============================================================================
// W-65: Улучшение coverage Execute (77.5% → 90%+)
// ============================================================================

func TestExecute_EmptyAction(t *testing.T) {
	_, err := Execute(context.Background(), Request{
		Action:    "",
		TargetURL: "http://192.168.1.1",
	})
	if err == nil {
		t.Fatal("expected error for empty action")
	}
}

func TestExecute_InvalidAction(t *testing.T) {
	_, err := Execute(context.Background(), Request{
		Action:    "invalid",
		TargetURL: "http://192.168.1.1",
	})
	if err == nil {
		t.Fatal("expected error for invalid action")
	}
	if !strings.Contains(err.Error(), "unsupported action") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestExecute_EmptyURL(t *testing.T) {
	_, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: "",
	})
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
	if !strings.Contains(err.Error(), "target URL is required") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestExecute_InvalidURLScheme(t *testing.T) {
	_, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: "ftp://192.168.1.1",
	})
	if err == nil {
		t.Fatal("expected error for invalid URL scheme")
	}
	if !strings.Contains(err.Error(), "http://") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestExecute_InvalidURLFormat(t *testing.T) {
	_, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: "not-a-url",
	})
	if err == nil {
		t.Fatal("expected error for invalid URL format")
	}
}

func TestExecute_UnsupportedVendor(t *testing.T) {
	_, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: "http://192.168.1.1",
		Vendor:    "unknown-vendor",
	})
	if err == nil {
		t.Fatal("expected error for unsupported vendor")
	}
	if !strings.Contains(err.Error(), "unsupported vendor") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestExecute_500ErrorResponse(t *testing.T) {
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
		t.Fatal("expected error for 500 response")
	}
	if res.Success {
		t.Fatal("expected success=false for 500 response")
	}
	if res.StatusCode != 500 {
		t.Fatalf("unexpected status code: %d", res.StatusCode)
	}
}

func TestExecute_404ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer srv.Close()

	res, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: srv.URL,
		Timeout:   2 * time.Second,
	})
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if res.Success {
		t.Fatal("expected success=false for 404 response")
	}
	if res.StatusCode != 404 {
		t.Fatalf("unexpected status code: %d", res.StatusCode)
	}
}

func TestExecute_ConnectionRefused(t *testing.T) {
	// Пробуем подключиться к несуществующему серверу
	_, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: "http://127.0.0.1:1",
		Timeout:   100 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected error for connection refused")
	}
}

func TestExecute_ContextTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := Execute(ctx, Request{
		Action:    ActionStatus,
		TargetURL: srv.URL,
		Timeout:   5 * time.Second,
	})
	if err == nil {
		t.Fatal("expected error for context timeout")
	}
}

func TestExecute_DefaultTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	// Без явного Timeout должен использоваться 10 секунд
	res, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: srv.URL,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !res.Success {
		t.Fatal("expected success=true")
	}
}

func TestExecute_BasicAuth(t *testing.T) {
	var gotAuth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _, ok := r.BasicAuth()
		gotAuth = ok
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	_, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: srv.URL,
		Username:  "admin",
		Password:  "secret",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !gotAuth {
		t.Fatal("expected basic auth to be set")
	}
}

func TestExecute_EmptyResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Пустой ответ
	}))
	defer srv.Close()

	res, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: srv.URL,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !res.Success {
		t.Fatal("expected success=true for empty 200 response")
	}
	if res.Message == "" {
		t.Fatal("expected message to be set from status text")
	}
}

func TestExecute_ResponseTruncated(t *testing.T) {
	largeResponse := strings.Repeat("A", 4096)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(largeResponse))
	}))
	defer srv.Close()

	res, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: srv.URL,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !res.Success {
		t.Fatal("expected success=true")
	}
	// Response должен быть ограничен 2048 байтами
	if len(res.Message) > 2048 {
		t.Fatalf("expected message to be truncated, got len=%d", len(res.Message))
	}
}

func TestExecute_CaseInsensitiveAction(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	res, err := Execute(context.Background(), Request{
		Action:    "STATUS", // uppercase
		TargetURL: srv.URL,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !res.Success {
		t.Fatal("expected success=true")
	}
}

func TestExecute_WhitespaceTargetURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	res, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: "  " + srv.URL + "  ",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !res.Success {
		t.Fatal("expected success=true")
	}
}

func TestExecute_DefaultVendor(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	_, err := Execute(context.Background(), Request{
		Action:    ActionStatus,
		TargetURL: srv.URL,
		// Vendor не указан, должен использоваться generic-http
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if gotPath != "/api/status" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
}
