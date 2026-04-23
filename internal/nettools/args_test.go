package nettools

import "testing"

func TestBuildPingArgsByOS(t *testing.T) {
	tests := []struct {
		name string
		goos string
		want []string
	}{
		{name: "windows", goos: "windows", want: []string{"ping", "-n", "4", "example.com"}},
		{name: "linux", goos: "linux", want: []string{"ping", "-c", "4", "example.com"}},
		{name: "darwin", goos: "darwin", want: []string{"ping", "-c", "4", "example.com"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildPingArgs("example.com", 4, tt.goos)
			if len(got) != len(tt.want) {
				t.Fatalf("unexpected args len: got %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("arg[%d] mismatch: got %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestBuildTracerouteArgsByOS(t *testing.T) {
	tests := []struct {
		name string
		goos string
		want []string
	}{
		{name: "windows", goos: "windows", want: []string{"tracert", "-d", "-h", "30", "example.com"}},
		{name: "linux", goos: "linux", want: []string{"traceroute", "-m", "30", "-n", "example.com"}},
		{name: "darwin", goos: "darwin", want: []string{"traceroute", "-m", "30", "-n", "example.com"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTracerouteArgs("example.com", 30, tt.goos)
			if len(got) != len(tt.want) {
				t.Fatalf("unexpected args len: got %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("arg[%d] mismatch: got %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
