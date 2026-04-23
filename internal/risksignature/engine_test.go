package risksignature

import (
	"testing"

	"network-scanner/internal/scanner"
)

func TestEvaluate_MatchesByPortAndDeviceType(t *testing.T) {
	db, err := LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault() error = %v", err)
	}
	results := []scanner.Result{
		{
			IP:         "192.168.1.10",
			DeviceType: "Router/Network Device",
			Ports: []scanner.PortInfo{
				{Port: 80, State: "open", Protocol: "tcp", Service: "HTTP"},
				{Port: 23, State: "open", Protocol: "tcp", Service: "Telnet"},
			},
		},
	}
	findings := Evaluate(results, db)
	if len(findings) < 2 {
		t.Fatalf("expected at least 2 findings, got %d", len(findings))
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	_, err := Load([]byte("{bad-json"))
	if err == nil {
		t.Fatal("expected error for invalid json")
	}
}
