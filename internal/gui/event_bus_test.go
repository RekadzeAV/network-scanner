package gui

import (
	"testing"
)

func TestParseIntFromPayload(t *testing.T) {
	tests := []struct {
		name     string
		payload  map[string]interface{}
		key      string
		expected int
	}{
		{"missing key", map[string]interface{}{}, "hostCount", 0},
		{"int value", map[string]interface{}{"hostCount": 42}, "hostCount", 42},
		{"float64 value", map[string]interface{}{"hostCount": float64(7)}, "hostCount", 7},
		{"string value", map[string]interface{}{"hostCount": "100"}, "hostCount", 100},
		{"invalid string", map[string]interface{}{"hostCount": "abc"}, "hostCount", 0},
		{"nil value", map[string]interface{}{"hostCount": nil}, "hostCount", 0},
		{"nil payload", nil, "hostCount", 0},
		{"zero value", map[string]interface{}{"hostCount": 0}, "hostCount", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseIntFromPayload(tt.payload, tt.key)
			if got != tt.expected {
				t.Errorf("parseIntFromPayload(%v, %q) = %d, want %d", tt.payload, tt.key, got, tt.expected)
			}
		})
	}
}

func TestSubscribeToScanEvents_NilBusNoPanic(t *testing.T) {
	// Не должен паниковать при nil шине.
	app := &App{}
	app.SubscribeToScanEvents(nil)
}
