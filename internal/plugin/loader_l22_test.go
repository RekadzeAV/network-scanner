package plugin

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// writeManifest кладёт валидный enabled-манифест в поддиректорию parent/name без .so файла.
func writeManifest(t *testing.T, parent, name string) {
	t.Helper()
	dir := filepath.Join(parent, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	enabled := true
	m := &Manifest{Name: name, Version: "1.0.0", Type: TypeFilter, Entry: name + ".so", Enabled: &enabled}
	if err := SaveManifest(dir, m); err != nil {
		t.Fatal(err)
	}
}

// setEnvForTest выставляет NETWORK_SCANNER_PLUGINS_DIR и восстанавливает его после теста.
func setEnvForTest(t *testing.T, value string) {
	t.Helper()
	old, had := os.LookupEnv(EnvPluginsDir)
	t.Cleanup(func() {
		if had {
			_ = os.Setenv(EnvPluginsDir, old)
		} else {
			_ = os.Unsetenv(EnvPluginsDir)
		}
	})
	if value == "" {
		_ = os.Unsetenv(EnvPluginsDir)
		return
	}
	_ = os.Setenv(EnvPluginsDir, value)
}

// --- SearchPaths ---

func TestSearchPaths_Defaults(t *testing.T) {
	setEnvForTest(t, "")

	paths := SearchPaths()
	if len(paths) != 2 {
		t.Fatalf("expected 2 default paths, got %d: %v", len(paths), paths)
	}
	want, err := filepath.Abs("plugins")
	if err != nil {
		t.Fatal(err)
	}
	if paths[0] != filepath.Clean(want) {
		t.Errorf("first default path = %q, want %q", paths[0], want)
	}
	cfg, err := os.UserConfigDir()
	if err != nil {
		t.Skip("no user config dir:", err)
	}
	want2, err := filepath.Abs(filepath.Join(cfg, "network-scanner", "plugins"))
	if err != nil {
		t.Fatal(err)
	}
	if paths[1] != filepath.Clean(want2) {
		t.Errorf("second default path = %q, want %q", paths[1], want2)
	}
}

func TestSearchPaths_EnvDirs(t *testing.T) {
	sep := string(filepath.ListSeparator)
	a := t.TempDir()
	b := t.TempDir()
	setEnvForTest(t, sep+a+sep+b+sep)

	paths := SearchPaths()
	if len(paths) < 2 {
		t.Fatalf("expected env dirs first, got %v", paths)
	}
	for i, want := range []string{a, b} {
		got, err := filepath.Abs(want)
		if err != nil {
			t.Fatal(err)
		}
		if paths[i] != filepath.Clean(got) {
			t.Errorf("path[%d] = %q, want %q", i, paths[i], got)
		}
	}
}

func TestSearchPaths_Dedup(t *testing.T) {
	sep := string(filepath.ListSeparator)
	a := t.TempDir()
	setEnvForTest(t, a+sep+a)

	seen := map[string]int{}
	for _, p := range SearchPaths() {
		seen[p]++
	}
	want, err := filepath.Abs(a)
	if err != nil {
		t.Fatal(err)
	}
	if seen[filepath.Clean(want)] != 1 {
		t.Errorf("duplicate dirs must collapse, counts=%v", seen)
	}
}

func TestNewPluginLoaderFromEnv(t *testing.T) {
	a := t.TempDir()
	setEnvForTest(t, a)

	loader := NewPluginLoaderFromEnv()
	dirs := loader.Dirs()
	if len(dirs) == 0 {
		t.Fatal("expected env dir in loader dirs")
	}
	want, err := filepath.Abs(a)
	if err != nil {
		t.Fatal(err)
	}
	if dirs[0] != filepath.Clean(want) {
		t.Errorf("first dir = %q, want %q", dirs[0], want)
	}
}

func TestLoader_DirsReturnsCopy(t *testing.T) {
	sep := string(filepath.ListSeparator)
	setEnvForTest(t, t.TempDir()+sep+t.TempDir())

	loader := NewPluginLoaderFromEnv()
	dirs := loader.Dirs()
	first := dirs[0]
	dirs[0] = "/mutated"
	if loader.Dirs()[0] != first {
		t.Error("Dirs must return a copy")
	}
}

// --- LoadAll ---

func TestLoadAll_SkipsDirWithoutManifest(t *testing.T) {
	root := t.TempDir()
	empty := filepath.Join(root, "empty")
	if err := os.MkdirAll(empty, 0o755); err != nil {
		t.Fatal(err)
	}
	loader := NewPluginLoader(root)
	got, err := loader.LoadAll(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 plugins, got %d", len(got))
	}
	if errs := loader.LoadErrors(); len(errs) != 0 {
		t.Fatalf("expected no load errors, got %v", errs)
	}
}

func TestLoadAll_DisabledSkipsWithoutError(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "off")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	enabled := false
	m := &Manifest{Name: "off", Version: "1.0.0", Type: TypeFilter, Entry: "off.so", Enabled: &enabled}
	if err := SaveManifest(dir, m); err != nil {
		t.Fatal(err)
	}

	loader := NewPluginLoader(root)
	got, err := loader.LoadAll(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("disabled plugin must be skipped, got %d", len(got))
	}
	if errs := loader.LoadErrors(); len(errs) != 0 {
		t.Fatalf("disabled skip must not record error, got %v", errs)
	}
	st := loader.Statuses()
	if len(st) != 1 || st[0].Loaded || st[0].Err != nil {
		t.Errorf("disabled status expected, got %+v", st)
	}
}

func TestLoadAll_EnabledWithoutLibraryRecordsError(t *testing.T) {
	root := t.TempDir()
	writeManifest(t, root, "nolib")

	loader := NewPluginLoader(root)
	got, err := loader.LoadAll(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 plugins, got %d", len(got))
	}
	errs := loader.LoadErrors()
	if len(errs) != 1 {
		t.Fatalf("expected 1 load error, got %v", errs)
	}
	if errs[0].Path == "" {
		t.Error("load error must carry the plugin dir path")
	}
	st := loader.Statuses()
	if len(st) != 1 || st[0].Loaded || st[0].Err == nil {
		t.Errorf("error status expected, got %+v", st)
	}
}

func TestLoadAll_BrokenManifestRecordsError(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "broken")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ManifestFile), []byte("{oops"), 0o600); err != nil {
		t.Fatal(err)
	}

	loader := NewPluginLoader(root)
	if _, err := loader.LoadAll(root); err != nil {
		t.Fatal(err)
	}
	if errs := loader.LoadErrors(); len(errs) != 1 {
		t.Fatalf("expected 1 load error, got %v", errs)
	}
}

func TestDiscoverAndLoad_ReplacesResults(t *testing.T) {
	root := t.TempDir()
	writeManifest(t, root, "a")

	loader := NewPluginLoader(root)
	if _, err := loader.DiscoverAndLoad(); err != nil {
		t.Fatal(err)
	}
	dirs := loader.Dirs()
	if len(dirs) != 1 {
		t.Fatalf("unexpected dirs after discovery: %v", dirs)
	}
	if errs := loader.LoadErrors(); len(errs) != 1 {
		t.Fatalf("expected 1 error (missing library), got %v", errs)
	}

	// второй проход не должен накапливать ошибки
	if _, err := loader.DiscoverAndLoad(); err != nil {
		t.Fatal(err)
	}
	if errs := loader.LoadErrors(); len(errs) != 1 {
		t.Fatalf("errors must be replaced, got %v", errs)
	}
}

func TestReload_UsesRegisteredDirs(t *testing.T) {
	root := t.TempDir()
	writeManifest(t, root, "x")

	loader := NewPluginLoader(root)
	if _, err := loader.Reload(); err != nil {
		t.Fatal(err)
	}
	if errs := loader.LoadErrors(); len(errs) != 1 {
		t.Fatalf("expected 1 load error, got %v", errs)
	}
	if got := loader.LoadedPlugins(); len(got) != 0 {
		t.Fatalf("expected 0 loaded, got %d", len(got))
	}
}

// --- WatchForChanges ---

func TestWatchForChanges_TriggersOnFileChange(t *testing.T) {
	root := t.TempDir()
	writeManifest(t, root, "w")

	loader := NewPluginLoader(root)
	if _, err := loader.LoadAll(root); err != nil {
		t.Fatal(err)
	}

	events := make(chan struct{}, 4)
	stop := loader.WatchForChanges(20*time.Millisecond, func(_ []Plugin, _ error) {
		select {
		case events <- struct{}{}:
		default:
		}
	})
	t.Cleanup(stop)

	// изменение файла внутри плагинной директории меняет mtime директории
	time.Sleep(30 * time.Millisecond)
	target := filepath.Join(root, "w", "extra.txt")
	if err := os.WriteFile(target, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}

	select {
	case <-events:
	case <-time.After(3 * time.Second):
		t.Fatal("watch did not trigger after file change")
	}
	stop() // идемпотентность
}

func TestWatchForChanges_StopIdempotent(t *testing.T) {
	loader := NewPluginLoader(t.TempDir())
	stop := loader.WatchForChanges(10*time.Millisecond, nil)
	stop()
	stop()
}

func TestHasPluginExt(t *testing.T) {
	native := map[string]string{
		"windows": ".dll",
		"darwin":  ".dylib",
	}[runtime.GOOS]
	if native == "" {
		native = ".so"
	}
	other := ".so"
	if native == ".so" {
		other = ".dll"
	}
	if !hasPluginExt("x" + native) {
		t.Errorf("native ext %q must be accepted", native)
	}
	if hasPluginExt("x" + other) {
		t.Errorf("foreign ext %q must be rejected", other)
	}
	if hasPluginExt("notes.txt") {
		t.Error("txt must be rejected")
	}
	if hasPluginExt("main") {
		t.Error("no ext must be rejected")
	}
}

func TestLoadError_Error(t *testing.T) {
	e := LoadError{Path: "/tmp/a.so", Err: os.ErrPermission}
	if e.Error() == "" {
		t.Error("empty error string")
	}
}
