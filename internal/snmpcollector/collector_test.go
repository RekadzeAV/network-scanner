package snmpcollector

import (
	"testing"

	"network-scanner/internal/scanner"
)

func TestParseMACFromOID(t *testing.T) {
	got, err := ParseMACFromOID(".1.3.6.1.2.1.17.4.3.1.2.170.187.204.221.238.255")
	if err != nil {
		t.Fatalf("ParseMACFromOID error: %v", err)
	}
	want := "aa:bb:cc:dd:ee:ff"
	if got != want {
		t.Fatalf("ParseMACFromOID got %s, want %s", got, want)
	}
}

func TestCollectWithReportSkipsNonSNMPDevices(t *testing.T) {
	devices := []scanner.Result{
		{IP: "192.168.1.10", SNMPEnabled: false},
	}
	data, report, err := CollectWithReport(devices, []string{"public"}, 1)
	if err != nil {
		t.Fatalf("CollectWithReport error: %v", err)
	}
	if len(data) != 0 {
		t.Fatalf("expected no SNMP data, got %d", len(data))
	}
	if report == nil {
		t.Fatalf("expected report, got nil")
	}
	if report.TotalSNMPTargets != 0 || report.Connected != 0 || report.Partial != 0 || report.Failed != 0 {
		t.Fatalf("unexpected report counters: %+v", *report)
	}
}
