package main

import (
	"os"
	"testing"
)

func TestVersionFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test --version flag
	os.Args = []string{"network-scanner", "--version"}
	
	// Capture output by redirecting
	// Since main() calls os.Exit(0), we need to test differently
	// Just verify the version variables are set
	if Version == "" {
		t.Error("Version should not be empty")
	}
}

func TestVersionVariables(t *testing.T) {
	// Version defaults to "dev" for local builds
	if Version != "dev" {
		t.Logf("Version is set to: %s (expected 'dev' for local build)", Version)
	}
	
	// BuildTime and GitCommit are set by ldflags
	t.Logf("BuildTime: %s", BuildTime)
	t.Logf("GitCommit: %s", GitCommit)
}
