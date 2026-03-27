package ports

import "testing"

func TestLookupServiceName(t *testing.T) {
	tests := []struct {
		port int
		want string
	}{
		{22, "SSH"},
		{80, "HTTP"},
		{3389, "RDP"},
		{9999, "Distinct"},
		{65530, "Unknown"},
	}
	for _, tt := range tests {
		if got := LookupServiceName(tt.port); got != tt.want {
			t.Errorf("LookupServiceName(%d) = %q, want %q", tt.port, got, tt.want)
		}
	}
}

func TestDescription(t *testing.T) {
	if d := Description(22); d == "" {
		t.Error("Description(22) empty")
	}
}

func TestProtocolLabel(t *testing.T) {
	if ProtocolLabel(65530) != "" {
		t.Errorf("ProtocolLabel unknown port: want empty")
	}
	if ProtocolLabel(80) != "HTTP" {
		t.Errorf("ProtocolLabel(80) = %q", ProtocolLabel(80))
	}
}
