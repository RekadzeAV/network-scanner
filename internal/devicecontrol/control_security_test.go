package devicecontrol

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestValidateTargetURL_ValidSchemes проверяет, что корректные http/https URL
// с хостом и без user-info проходят валидацию.
func TestValidateTargetURL_ValidSchemes(t *testing.T) {
	valid := []string{
		"http://192.168.1.1",
		"https://192.168.1.1",
		"https://router.example.com:8443",
		"http://10.0.0.1:8080/api",
	}
	for _, raw := range valid {
		if err := validateTargetURL(raw); err != nil {
			t.Errorf("validateTargetURL(%q) unexpected error: %v", raw, err)
		}
	}
}

// TestValidateTargetURL_RejectsUnsafe проверяет отказ для небезопасных целей:
// чужие схемы, отсутствие хоста, наличие user-info, пробелы.
func TestValidateTargetURL_RejectsUnsafe(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"empty", "", "required"},
		{"whitespace-only", "   ", "required"},
		{"file-scheme", "file:///etc/passwd", "http://"},
		{"ftp-scheme", "ftp://example.com", "http://"},
		{"javascript-scheme", "javascript:alert(1)", "http://"},
		{"no-scheme", "192.168.1.1", "http://"},
		{"no-host", "http://", "host"},
		{"user-info", "http://admin:secret@192.168.1.1", "credentials"},
		{"embedded-space", "http://192.168.1.1 /path", "whitespace"},
		{"embedded-newline", "http://192.168.1.1\n", "whitespace"},
		{"embedded-tab", "http://192.168.1.1\t", "whitespace"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTargetURL(tc.raw)
			if err == nil {
				t.Fatalf("validateTargetURL(%q) expected error, got nil", tc.raw)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("validateTargetURL(%q) error = %q, want substring %q", tc.raw, err.Error(), tc.want)
			}
		})
	}
}

// TestExecute_RejectsUnsafeTargetURL проверяет, что Execute не выполняет запрос
// для небезопасных целей и возвращает ошибку до сетевого вызова.
func TestExecute_RejectsUnsafeTargetURL(t *testing.T) {
	unsafe := []string{
		"file:///etc/passwd",
		"ftp://example.com",
		"http://admin:secret@192.168.1.1",
		"http://",
		"not-a-url",
	}
	for _, raw := range unsafe {
		_, err := Execute(context.Background(), Request{
			Action:    ActionStatus,
			TargetURL: raw,
			Vendor:    VendorGenericHTTP,
			Timeout:   time.Second,
		})
		if err == nil {
			t.Errorf("Execute(target=%q) expected error, got nil", raw)
		}
	}
}

// TestExecute_RejectsUserInfoCredentials подтверждает, что учётные данные
// нельзя передать через URL — только через поля Username/Password.
func TestExecute_RejectsUserInfoCredentials(t *testing.T) {
	_, err := Execute(context.Background(), Request{
		Action:    ActionReboot,
		TargetURL: "http://admin:secret@10.0.0.1",
		Vendor:    VendorGenericHTTP,
		Timeout:   time.Second,
	})
	if err == nil {
		t.Fatal("expected error for URL with embedded credentials")
	}
	if !strings.Contains(err.Error(), "credentials") {
		t.Fatalf("error = %q, want mention of credentials", err.Error())
	}
}
