package snmpcollector

import (
	"testing"

	"network-scanner/internal/scanner"
	"network-scanner/internal/topology"

	"github.com/gosnmp/gosnmp"
)

// ============================================================================
// NewGoSNMPClient — 0% → 100%
// ============================================================================

func TestNewGoSNMPClient_DefaultTimeout(t *testing.T) {
	c := NewGoSNMPClient(0)
	if c == nil {
		t.Fatal("expected non-nil client")
	}
	if c.timeout.Seconds() != 2 {
		t.Fatalf("expected default timeout 2s, got %v", c.timeout)
	}
}

func TestNewGoSNMPClient_CustomTimeout(t *testing.T) {
	c := NewGoSNMPClient(5)
	if c == nil {
		t.Fatal("expected non-nil client")
	}
	if c.timeout.Seconds() != 5 {
		t.Fatalf("expected timeout 5s, got %v", c.timeout)
	}
}

func TestNewGoSNMPClient_NegativeTimeout(t *testing.T) {
	c := NewGoSNMPClient(-1)
	if c.timeout.Seconds() != 2 {
		t.Fatalf("expected default timeout 2s for negative input, got %v", c.timeout)
	}
}

// ============================================================================
// Close — 0% → 100% (nil client branch)
// ============================================================================

func TestGoSNMPClient_Close_NilClient(t *testing.T) {
	c := NewGoSNMPClient(1)
	err := c.Close()
	if err != nil {
		t.Fatalf("expected nil error for nil client close, got %v", err)
	}
}

// ============================================================================
// inferDeviceType — 0% → 100%
// ============================================================================

func TestInferDeviceType_Switch(t *testing.T) {
	if got := inferDeviceType("Cisco Switch", false); got != topology.DeviceTypeSwitch {
		t.Fatalf("expected switch, got %s", got)
	}
}

func TestInferDeviceType_SwitchByDot1d(t *testing.T) {
	if got := inferDeviceType("Some Device", true); got != topology.DeviceTypeSwitch {
		t.Fatalf("expected switch (hasDot1d), got %s", got)
	}
}

func TestInferDeviceType_Router(t *testing.T) {
	if got := inferDeviceType("Linux Router", false); got != topology.DeviceTypeRouter {
		t.Fatalf("expected router, got %s", got)
	}
}

func TestInferDeviceType_Host_Linux(t *testing.T) {
	if got := inferDeviceType("Ubuntu Linux 20.04", false); got != topology.DeviceTypeHost {
		t.Fatalf("expected host, got %s", got)
	}
}

func TestInferDeviceType_Host_Windows(t *testing.T) {
	if got := inferDeviceType("Windows Server 2019", false); got != topology.DeviceTypeHost {
		t.Fatalf("expected host, got %s", got)
	}
}

func TestInferDeviceType_Host_Server(t *testing.T) {
	if got := inferDeviceType("My Server", false); got != topology.DeviceTypeHost {
		t.Fatalf("expected host, got %s", got)
	}
}

func TestInferDeviceType_Host_Host(t *testing.T) {
	if got := inferDeviceType("Host Device", false); got != topology.DeviceTypeHost {
		t.Fatalf("expected host, got %s", got)
	}
}

func TestInferDeviceType_Unknown(t *testing.T) {
	if got := inferDeviceType("Unknown Device", false); got != topology.DeviceTypeUnknown {
		t.Fatalf("expected unknown, got %s", got)
	}
}

func TestInferDeviceType_Empty(t *testing.T) {
	if got := inferDeviceType("", false); got != topology.DeviceTypeUnknown {
		t.Fatalf("expected unknown, got %s", got)
	}
}

func TestInferDeviceType_CaseInsensitive(t *testing.T) {
	if got := inferDeviceType("SWITCH", false); got != topology.DeviceTypeSwitch {
		t.Fatalf("expected switch (case-insensitive), got %s", got)
	}
}

// ============================================================================
// suffixInt — 0% → 100%
// ============================================================================

func TestSuffixInt(t *testing.T) {
	if got := suffixInt(".1.3.6.1.2.1.2.2.1.2.5"); got != 5 {
		t.Fatalf("expected 5, got %d", got)
	}
}

func TestSuffixInt_Empty(t *testing.T) {
	if got := suffixInt(""); got != -1 {
		t.Fatalf("expected -1, got %d", got)
	}
}

func TestSuffixInt_InvalidNumber(t *testing.T) {
	if got := suffixInt(".1.3.6.abc"); got != -1 {
		t.Fatalf("expected -1, got %d", got)
	}
}

func TestSuffixInt_NoDotPrefix(t *testing.T) {
	if got := suffixInt("1.2.3"); got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}
}

// ============================================================================
// lldpRowKeyFromOID — 0% → 100%
// ============================================================================

func TestLldpRowKeyFromOID(t *testing.T) {
	got := lldpRowKeyFromOID(".1.0.8802.1.1.2.1.4.1.1.9.100.20.1")
	if got != "100.20.1" {
		t.Fatalf("expected '100.20.1', got %q", got)
	}
}

func TestLldpRowKeyFromOID_TooShort(t *testing.T) {
	got := lldpRowKeyFromOID(".1.0")
	if got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestLldpRowKeyFromOID_Empty(t *testing.T) {
	got := lldpRowKeyFromOID("")
	if got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

// ============================================================================
// lldpLocalPortFromOID — 0% → 100%
// ============================================================================

func TestLldpLocalPortFromOID(t *testing.T) {
	got := lldpLocalPortFromOID(".1.0.8802.1.1.2.1.4.1.1.9.100.20.1")
	if got != 20 {
		t.Fatalf("expected 20, got %d", got)
	}
}

func TestLldpLocalPortFromOID_TooShort(t *testing.T) {
	got := lldpLocalPortFromOID(".1.0")
	if got != -1 {
		t.Fatalf("expected -1, got %d", got)
	}
}

func TestLldpLocalPortFromOID_InvalidNumber(t *testing.T) {
	got := lldpLocalPortFromOID(".1.0.8802.abc.1")
	if got != -1 {
		t.Fatalf("expected -1, got %d", got)
	}
}

// ============================================================================
// lldpChassisToMACString — 0% → 100%
// ============================================================================

func TestLldpChassisToMACString_Bytes(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: []byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}}
	got := lldpChassisToMACString(pdu)
	if got != "aa:bb:cc:dd:ee:ff" {
		t.Fatalf("expected aa:bb:cc:dd:ee:ff, got %q", got)
	}
}

func TestLldpChassisToMACString_BytesNotMAC(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: []byte("AA-BB-CC")}
	got := lldpChassisToMACString(pdu)
	if got != "aa:bb:cc" {
		t.Fatalf("expected aa:bb:cc, got %q", got)
	}
}

func TestLldpChassisToMACString_String(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: "AA-BB-CC-DD-EE-FF"}
	got := lldpChassisToMACString(pdu)
	if got != "aa:bb:cc:dd:ee:ff" {
		t.Fatalf("expected aa:bb:cc:dd:ee:ff, got %q", got)
	}
}

func TestLldpChassisToMACString_Integer(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: 42}
	got := lldpChassisToMACString(pdu)
	if got != "42" {
		t.Fatalf("expected 42, got %q", got)
	}
}

// ============================================================================
// bytesToMAC — 0% → 100%
// ============================================================================

func TestBytesToMAC(t *testing.T) {
	got := bytesToMAC([]byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff})
	if got != "aa:bb:cc:dd:ee:ff" {
		t.Fatalf("expected aa:bb:cc:dd:ee:ff, got %q", got)
	}
}

func TestBytesToMAC_Empty(t *testing.T) {
	got := bytesToMAC([]byte{})
	if got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestBytesToMAC_Single(t *testing.T) {
	got := bytesToMAC([]byte{0x00})
	if got != "00" {
		t.Fatalf("expected 00, got %q", got)
	}
}

// ============================================================================
// pduValueString — 0% → 100%
// ============================================================================

func TestPduValueString_Bytes(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: []byte("hello")}
	if got := pduValueString(pdu); got != "hello" {
		t.Fatalf("expected hello, got %q", got)
	}
}

func TestPduValueString_String(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: "world"}
	if got := pduValueString(pdu); got != "world" {
		t.Fatalf("expected world, got %q", got)
	}
}

func TestPduValueString_Integer(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: 12345}
	if got := pduValueString(pdu); got != "12345" {
		t.Fatalf("expected 12345, got %q", got)
	}
}

func TestPduValueString_BytesWithWhitespace(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: []byte("  hello  ")}
	if got := pduValueString(pdu); got != "hello" {
		t.Fatalf("expected trimmed hello, got %q", got)
	}
}

// ============================================================================
// ParseMACFromOID — edge cases (81.8% → 100%)
// ============================================================================

func TestParseMACFromOID_TooShort(t *testing.T) {
	_, err := ParseMACFromOID(".1.2.3")
	if err == nil {
		t.Fatal("expected error for short OID")
	}
}

func TestParseMACFromOID_InvalidOctet(t *testing.T) {
	_, err := ParseMACFromOID(".1.2.3.4.5.6.256")
	if err == nil {
		t.Fatal("expected error for invalid octet > 255")
	}
}

func TestParseMACFromOID_NegativeOctet(t *testing.T) {
	_, err := ParseMACFromOID(".1.2.3.4.5.6.-1")
	if err == nil {
		t.Fatal("expected error for negative octet")
	}
}

func TestParseMACFromOID_NoDotPrefix(t *testing.T) {
	got, err := ParseMACFromOID("1.2.3.4.5.6.170.187.204.221.238.255")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != "aa:bb:cc:dd:ee:ff" {
		t.Fatalf("expected aa:bb:cc:dd:ee:ff, got %q", got)
	}
}

func TestParseMACFromOID_Empty(t *testing.T) {
	_, err := ParseMACFromOID("")
	if err == nil {
		t.Fatal("expected error for empty OID")
	}
}

func TestParseMACFromOID_Whitespace(t *testing.T) {
	got, err := ParseMACFromOID("  .1.2.3.4.5.6.170.187.204.221.238.255  ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != "aa:bb:cc:dd:ee:ff" {
		t.Fatalf("expected aa:bb:cc:dd:ee:ff, got %q", got)
	}
}

// ============================================================================
// Collect / CollectWithReportProgress — edge cases
// ============================================================================

func TestCollect_NoSNMPDevices(t *testing.T) {
	devices := []scanner.Result{
		{IP: "10.0.0.1", SNMPEnabled: false},
	}
	data, err := Collect(devices, []string{"public"}, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(data) != 0 {
		t.Fatalf("expected 0 devices, got %d", len(data))
	}
}

func TestCollectWithReport_EmptyCommunities(t *testing.T) {
	devices := []scanner.Result{
		{IP: "10.0.0.1", SNMPEnabled: false},
	}
	data, report, err := CollectWithReport(devices, nil, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(data) != 0 {
		t.Fatalf("expected 0 devices, got %d", len(data))
	}
	if report.TotalSNMPTargets != 0 {
		t.Fatalf("expected 0 targets, got %d", report.TotalSNMPTargets)
	}
}

func TestCollectWithReportProgress_NoSNMPDevices(t *testing.T) {
	devices := []scanner.Result{
		{IP: "10.0.0.1", SNMPEnabled: false},
	}
	called := false
	progress := func(current, total int, ip, msg string) {
		called = true
	}
	data, report, err := CollectWithReportProgress(devices, []string{"public"}, 1, progress)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(data) != 0 {
		t.Fatalf("expected 0 devices, got %d", len(data))
	}
	if report.TotalSNMPTargets != 0 {
		t.Fatalf("expected 0 targets, got %d", report.TotalSNMPTargets)
	}
	if called {
		t.Fatal("expected progress callback to not be called")
	}
}

func TestCollectWithReportProgressContext_NilContext(t *testing.T) {
	devices := []scanner.Result{
		{IP: "10.0.0.1", SNMPEnabled: false},
	}
	called := false
	progress := func(current, total int, ip, msg string) {
		called = true
	}
	data, report, err := CollectWithReportProgress(devices, []string{"public"}, 1, progress)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(data) != 0 {
		t.Fatalf("expected 0 devices, got %d", len(data))
	}
	if report.TotalSNMPTargets != 0 {
		t.Fatalf("expected 0 targets, got %d", report.TotalSNMPTargets)
	}
	if called {
		t.Fatal("expected progress callback to not be called")
	}
}
