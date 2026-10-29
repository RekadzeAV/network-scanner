package plugin

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// --- Manifest.Validate ---

func TestManifest_Validate_OK(t *testing.T) {
	m := &Manifest{Name: "Test", Version: "1.0.0", Type: TypeFilter, Entry: "test.so"}
	if err := m.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManifest_Validate_Errors(t *testing.T) {
	cases := []struct {
		name string
		m    Manifest
	}{
		{"empty name", Manifest{Version: "1.0", Type: TypeFilter, Entry: "a.so"}},
		{"blank name", Manifest{Name: "   ", Version: "1.0", Type: TypeFilter, Entry: "a.so"}},
		{"empty version", Manifest{Name: "n", Type: TypeFilter, Entry: "a.so"}},
		{"empty entry", Manifest{Name: "n", Version: "1.0", Type: TypeFilter}},
		{"path traversal", Manifest{Name: "n", Version: "1.0", Type: TypeFilter, Entry: "../evil.so"}},
		{"nested path", Manifest{Name: "n", Version: "1.0", Type: TypeFilter, Entry: "sub/evil.so"}},
		{"bad type", Manifest{Name: "n", Version: "1.0", Type: Type("bogus"), Entry: "a.so"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.m.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

// --- Manifest.IsEnabled ---

func TestManifest_IsEnabled(t *testing.T) {
	yes := true
	no := false
	if !(&Manifest{}).IsEnabled() {
		t.Error("nil enabled must mean enabled")
	}
	if !(&Manifest{Enabled: &yes}).IsEnabled() {
		t.Error("enabled=true must be enabled")
	}
	if (&Manifest{Enabled: &no}).IsEnabled() {
		t.Error("enabled=false must be disabled")
	}
}

// --- LoadManifest / SaveManifest ---

func TestLoadManifest_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	enabled := false
	m := &Manifest{
		Name:        "Example Filter",
		Version:     "1.2.3",
		Description: "тестовый фильтр",
		Author:      "QA",
		Type:        TypeFilter,
		Entry:       "example.so",
		Enabled:     &enabled,
		MinVersion:  "2.4.0",
	}
	if err := SaveManifest(dir, m); err != nil {
		t.Fatalf("SaveManifest: %v", err)
	}
	got, err := LoadManifest(dir)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if got.Name != m.Name || got.Version != m.Version || got.Entry != m.Entry {
		t.Errorf("round trip mismatch: %+v vs %+v", got, m)
	}
	if got.IsEnabled() {
		t.Error("enabled=false must survive round trip")
	}
}

func TestLoadManifest_MissingFile(t *testing.T) {
	dir := t.TempDir()
	if _, err := LoadManifest(dir); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected wrapped fs.ErrNotExist, got %v", err)
	}
}

func TestLoadManifest_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ManifestFile)
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadManifest(dir); err == nil {
		t.Fatal("expected JSON error")
	}
}

func TestLoadManifest_UnknownField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ManifestFile)
	payload := `{"name":"n","version":"1","type":"filter","entry":"a.so","unknownField":true}`
	if err := os.WriteFile(path, []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadManifest(dir); err == nil {
		t.Fatal("unknown fields must be rejected")
	}
}

func TestLoadManifest_InvalidManifest(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ManifestFile)
	payload := `{"name":"","version":"1","type":"filter","entry":"a.so"}`
	if err := os.WriteFile(path, []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadManifest(dir); err == nil {
		t.Fatal("invalid manifest must be rejected")
	}
}

func TestSaveManifest_RejectsInvalid(t *testing.T) {
	dir := t.TempDir()
	if err := SaveManifest(dir, &Manifest{Name: "x"}); err == nil {
		t.Fatal("expected error for invalid manifest")
	}
}

func TestManifest_JSONType(t *testing.T) {
	m := Manifest{Name: "n", Version: "1", Type: TypeExporter, Entry: "e.so"}
	data, err := json.Marshal(&m)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(data) {
		t.Fatal("invalid JSON output")
	}
}
