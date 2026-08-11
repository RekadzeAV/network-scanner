package plugin

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// ============================================================================
// PluginLoader — полное покрытие
// ============================================================================

func TestNewPluginLoader_DefaultDir(t *testing.T) {
	loader := NewPluginLoader("")
	if loader.pluginsDir != "plugins" {
		t.Errorf("expected plugins dir, got %q", loader.pluginsDir)
	}
}

func TestNewPluginLoader_CustomDir(t *testing.T) {
	loader := NewPluginLoader("/custom/path")
	if loader.pluginsDir != "/custom/path" {
		t.Errorf("expected /custom/path, got %q", loader.pluginsDir)
	}
}

func TestPluginLoader_Load_InvalidExtension(t *testing.T) {
	loader := NewPluginLoader("")
	_, err := loader.Load("/path/to/plugin.txt")
	if err == nil {
		t.Fatal("expected error for invalid extension")
	}
}

func TestPluginLoader_Load_FileNotFound(t *testing.T) {
	loader := NewPluginLoader("")
	tmpDir := t.TempDir()
	ext := getPluginExtension()
	path := filepath.Join(tmpDir, "plugin"+ext)
	_, err := loader.Load(path)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestPluginLoader_LoadAll_NonExistentDir(t *testing.T) {
	loader := NewPluginLoader("/nonexistent/path/xyz")
	plugins, err := loader.LoadAll("/nonexistent/path/xyz")
	if err != nil {
		t.Errorf("expected no error for missing dir, got %v", err)
	}
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(plugins))
	}
}

func TestPluginLoader_LoadAll_DirIsFile(t *testing.T) {
	loader := NewPluginLoader("")
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "notadir")
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	_, err = loader.LoadAll(filePath)
	if err == nil {
		t.Fatal("expected error for file path")
	}
}

func TestPluginLoader_Discover_NonExistent(t *testing.T) {
	loader := NewPluginLoader("/nonexistent/discover/path")
	plugins, err := loader.Discover()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(plugins))
	}
}

func TestValidExtensions_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows only")
	}
	extensions := ValidExtensions()
	if len(extensions) != 1 || extensions[0] != ".dll" {
		t.Errorf("expected [.dll], got %v", extensions)
	}
}

func TestValidExtensions_Darwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin only")
	}
	extensions := ValidExtensions()
	if len(extensions) != 1 || extensions[0] != ".dylib" {
		t.Errorf("expected [.dylib], got %v", extensions)
	}
}

func TestValidExtensions_Linux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux only")
	}
	extensions := ValidExtensions()
	if len(extensions) != 1 || extensions[0] != ".so" {
		t.Errorf("expected [.so], got %v", extensions)
	}
}

// ============================================================================
// Helpers
// ============================================================================

func getPluginExtension() string {
	switch runtime.GOOS {
	case "windows":
		return ".dll"
	case "darwin":
		return ".dylib"
	default:
		return ".so"
	}
}

// ============================================================================
// Benchmark
// ============================================================================

func BenchmarkPluginLoader_New(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewPluginLoader("")
	}
}

func BenchmarkValidExtensions(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidExtensions()
	}
}
