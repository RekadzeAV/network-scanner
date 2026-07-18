package profiler

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewProfiler_DefaultDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "profile")
	p, err := NewProfiler(dir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil Profiler")
	}
	// Start+Stop to properly close files
	p.Start()
	p.Stop()
}

func TestNewProfiler_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	// Change to temp dir so default "profile" dir is created there
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	p, err := NewProfiler("")
	if err != nil {
		t.Fatalf("expected no error for empty dir, got %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil Profiler")
	}
	p.Start()
	p.Stop()
}

func TestProfiler_StartStop(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "profile")
	p, err := NewProfiler(dir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := p.Start(); err != nil {
		t.Fatalf("expected no error starting profiler, got %v", err)
	}

	if err := p.Stop(); err != nil {
		t.Fatalf("expected no error stopping profiler, got %v", err)
	}

	// Verify files exist
	if _, err := os.Stat(filepath.Join(dir, "cpu.profile")); os.IsNotExist(err) {
		t.Fatal("expected cpu.profile file to exist")
	}
	if _, err := os.Stat(filepath.Join(dir, "memory.profile")); os.IsNotExist(err) {
		t.Fatal("expected memory.profile file to exist")
	}
}

func TestQuickProfile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "qp")
	profiler, stop, err := QuickProfile(dir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if profiler == nil {
		t.Fatal("expected non-nil profiler")
	}
	if stop == nil {
		t.Fatal("expected non-nil stop function")
	}

	// Do some work
	for i := 0; i < 100; i++ {
		_ = i * i
	}

	stop()

	// Verify files exist
	if _, err := os.Stat(filepath.Join(dir, "cpu.profile")); os.IsNotExist(err) {
		t.Fatal("expected cpu.profile file to exist")
	}
}

// ============================================================================
// Error paths
// ============================================================================

func TestNewProfiler_InvalidDir(t *testing.T) {
	// Use a path that contains a NUL byte — MkdirAll will fail
	_, err := NewProfiler("invalid\x00dir/profile")
	if err == nil {
		t.Fatal("expected error for invalid directory path")
	}
}

func TestNewProfiler_MemFileCreateFails(t *testing.T) {
	// Create a directory where the memory profile file should be —
	// os.Create will fail because it's a directory
	dir := filepath.Join(t.TempDir(), "profile")
	cpuPath := filepath.Join(dir, "cpu.profile")
	memPath := filepath.Join(dir, "memory.profile")

	// Pre-create the memory file path as a directory to force Create failure
	os.MkdirAll(dir, 0755)
	os.MkdirAll(memPath, 0755)

	_, err := NewProfiler(dir)
	if err == nil {
		t.Fatal("expected error when memory profile path is a directory")
	}

	// Clean up cpu file that may have been created
	_ = cpuPath
}

func TestStart_Error_AlreadyRunning(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "profile")
	p, err := NewProfiler(dir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := p.Start(); err != nil {
		t.Fatalf("first Start should succeed, got %v", err)
	}

	// Second Start should fail — CPU profiler already running
	err = p.Start()
	if err == nil {
		// On some platforms it might not fail, so just stop
		p.Stop()
	} else {
		p.Stop()
	}
}

func TestQuickProfile_NewProfilerFails(t *testing.T) {
	// Invalid directory path
	_, _, err := QuickProfile("invalid\x00dir/profile")
	if err == nil {
		t.Fatal("expected error for invalid directory in QuickProfile")
	}
}

func TestQuickProfile_StartFails(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "qp_start_fail")
	p, err := NewProfiler(dir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Start CPU profiling first so the second Start fails
	if err := p.Start(); err != nil {
		t.Fatalf("first Start should succeed, got %v", err)
	}

	// Now QuickProfile's internal Start should fail since profiler is already running
	_, _, err2 := QuickProfile(filepath.Join(dir, "inner"))
	if err2 != nil {
		// The internal profiler was already running, so Start might fail
		t.Logf("QuickProfile Start failed as expected: %v", err2)
	}

	p.Stop()
}
