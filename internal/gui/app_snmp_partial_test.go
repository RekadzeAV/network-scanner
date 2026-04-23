package gui

import (
	"testing"

	"network-scanner/internal/snmpcollector"
)

func TestPartialSNMPKeysFromReport_Nil(t *testing.T) {
	keys := partialSNMPKeysFromReport(nil)
	if keys != nil {
		t.Fatalf("expected nil keys for nil report")
	}
}

func TestPartialSNMPKeysFromReport_OnlyQueryFailures(t *testing.T) {
	report := &snmpcollector.CollectReport{
		Failures: []snmpcollector.DeviceFailure{
			{IP: "192.168.1.10", Kind: snmpcollector.FailureConnect, Message: "timeout"},
			{IP: " 192.168.1.11 ", Kind: snmpcollector.FailureQuery, Message: "ifTable failed"},
			{IP: "192.168.1.11", Kind: snmpcollector.FailureQuery, Message: "lldp failed"},
			{IP: "HOST.local", Kind: snmpcollector.FailureQuery, Message: "sysDescr failed"},
		},
	}

	keys := partialSNMPKeysFromReport(report)
	if keys == nil {
		t.Fatalf("expected non-nil keys")
	}
	if _, ok := keys["ip:192.168.1.11"]; !ok {
		t.Fatalf("expected normalized IP key for query error")
	}
	if _, ok := keys["ip:host.local"]; !ok {
		t.Fatalf("expected lowercase host key for query error")
	}
	if _, ok := keys["ip:192.168.1.10"]; ok {
		t.Fatalf("did not expect connect-error key")
	}
	if got := len(keys); got != 2 {
		t.Fatalf("expected 2 unique keys, got %d", got)
	}
}

func TestPartialSNMPKeysFromReport_EmptyResult(t *testing.T) {
	report := &snmpcollector.CollectReport{
		Failures: []snmpcollector.DeviceFailure{
			{IP: "   ", Kind: snmpcollector.FailureQuery, Message: "empty ip"},
		},
	}
	keys := partialSNMPKeysFromReport(report)
	if keys != nil {
		t.Fatalf("expected nil keys when no valid identifiers")
	}
}

