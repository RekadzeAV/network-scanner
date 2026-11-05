//go:build !gui_only

package main

import (
	"os"
	"testing"
)

// TestResolveAPIFlags_TokenOnly — только --api-token.
func TestResolveAPIFlags_TokenOnly(t *testing.T) {
	f := resolveAPIFlags([]string{"--api", "--api-token", "mysecret"})
	if f.token != "mysecret" {
		t.Errorf("token = %q, want %q", f.token, "mysecret")
	}
	if f.tlsCertFile != "" || f.tlsKeyFile != "" {
		t.Errorf("TLS fields should be empty, got cert=%q key=%q", f.tlsCertFile, f.tlsKeyFile)
	}
}

// TestResolveAPIFlags_TokenEquals — форма --api-token=value.
func TestResolveAPIFlags_TokenEquals(t *testing.T) {
	f := resolveAPIFlags([]string{"--api", "--api-token=eqval"})
	if f.token != "eqval" {
		t.Errorf("token = %q, want %q", f.token, "eqval")
	}
}

// TestResolveAPIFlags_TLSFiles — оба флага TLS.
func TestResolveAPIFlags_TLSFiles(t *testing.T) {
	f := resolveAPIFlags([]string{
		"--api",
		"--api-tls-cert", "/etc/ssl/cert.pem",
		"--api-tls-key", "/etc/ssl/key.pem",
	})
	if f.tlsCertFile != "/etc/ssl/cert.pem" {
		t.Errorf("tlsCertFile = %q, want %q", f.tlsCertFile, "/etc/ssl/cert.pem")
	}
	if f.tlsKeyFile != "/etc/ssl/key.pem" {
		t.Errorf("tlsKeyFile = %q, want %q", f.tlsKeyFile, "/etc/ssl/key.pem")
	}
}

// TestResolveAPIFlags_TLSFilesEquals — форма --flag=value.
func TestResolveAPIFlags_TLSFilesEquals(t *testing.T) {
	f := resolveAPIFlags([]string{
		"--api",
		"--api-tls-cert=/cert/path",
		"--api-tls-key=/key/path",
	})
	if f.tlsCertFile != "/cert/path" {
		t.Errorf("tlsCertFile = %q, want %q", f.tlsCertFile, "/cert/path")
	}
	if f.tlsKeyFile != "/key/path" {
		t.Errorf("tlsKeyFile = %q, want %q", f.tlsKeyFile, "/key/path")
	}
}

// TestResolveAPIFlags_MissingKey — resolveAPIFlags не валидирует полноту
// пары TLS; он просто разбирает аргументы. Проверка полноты (оба cert+key)
// делается в main.go перед запуском ListenAndServeTLS.
func TestResolveAPIFlags_MissingKey(t *testing.T) {
	f := resolveAPIFlags([]string{
		"--api",
		"--api-tls-cert", "/cert.pem",
	})
	// resolveAPIFlags не фильтрует поля — caller (main.go) проверяет полноту.
	if f.tlsCertFile != "/cert.pem" {
		t.Errorf("tlsCertFile = %q, want %q", f.tlsCertFile, "/cert.pem")
	}
	if f.tlsKeyFile != "" {
		t.Errorf("tlsKeyFile should be empty, got %q", f.tlsKeyFile)
	}
}

// TestResolveAPIFlags_EnvFallback — env переменная для токена.
func TestResolveAPIFlags_EnvFallback(t *testing.T) {
	os.Setenv("NETWORK_SCANNER_API_TOKEN", "envsecret")
	defer os.Unsetenv("NETWORK_SCANNER_API_TOKEN")

	f := resolveAPIFlags([]string{"--api"})
	if f.token != "envsecret" {
		t.Errorf("token = %q, want %q", f.token, "envsecret")
	}
}

// TestResolveAPIFlags_FlagOverridesEnv — флаг имеет приоритет.
func TestResolveAPIFlags_FlagOverridesEnv(t *testing.T) {
	os.Setenv("NETWORK_SCANNER_API_TOKEN", "envsecret")
	defer os.Unsetenv("NETWORK_SCANNER_API_TOKEN")

	f := resolveAPIFlags([]string{"--api", "--api-token", "flagsecret"})
	if f.token != "flagsecret" {
		t.Errorf("token = %q, want %q", f.token, "flagsecret")
	}
}
