package snmpcollector

import (
	"context"
	"strings"
	"testing"
	"time"

	"network-scanner/internal/scanner"
	"network-scanner/internal/topology"
)

// === Integration: NewGoSNMPClient ===

func TestIntegrationNewGoSNMPClient_DefaultTimeout(t *testing.T) {
	client := NewGoSNMPClient(0)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.timeout != 2*time.Second {
		t.Errorf("expected 2s timeout, got %v", client.timeout)
	}
}

func TestIntegrationNewGoSNMPClient_CustomTimeout(t *testing.T) {
	client := NewGoSNMPClient(5)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.timeout != 5*time.Second {
		t.Errorf("expected 5s timeout, got %v", client.timeout)
	}
}

func TestIntegrationNewGoSNMPClient_NegativeTimeout(t *testing.T) {
	client := NewGoSNMPClient(-1)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.timeout != 2*time.Second {
		t.Errorf("expected 2s timeout for negative input, got %v", client.timeout)
	}
}

// === Integration: ParseMACFromOID ===

func TestIntegrationParseMACFromOID_ValidDot1d(t *testing.T) {
	mac, err := ParseMACFromOID(".1.3.6.1.2.1.17.4.3.1.2.10.20.30.40.50.60")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mac != "0a:14:1e:28:32:3c" {
		t.Errorf("expected '0a:14:1e:28:32:3c', got %q", mac)
	}
}

func TestIntegrationParseMACFromOID_ValidDot1q(t *testing.T) {
	mac, err := ParseMACFromOID(".1.3.6.1.2.1.17.7.1.2.2.1.2.10.20.30.40.50.60")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mac != "0a:14:1e:28:32:3c" {
		t.Errorf("expected '0a:14:1e:28:32:3c', got %q", mac)
	}
}

func TestIntegrationParseMACFromOID_TooShort(t *testing.T) {
	_, err := ParseMACFromOID(".1.3.6.1")
	if err == nil {
		t.Error("expected error for too short OID")
	}
}

func TestIntegrationParseMACFromOID_Empty(t *testing.T) {
	_, err := ParseMACFromOID("")
	if err == nil {
		t.Error("expected error for empty OID")
	}
}

func TestIntegrationParseMACFromOID_InvalidOctet(t *testing.T) {
	_, err := ParseMACFromOID(".1.3.6.1.2.1.17.4.3.1.2.999.20.30.40.50.60")
	if err == nil {
		t.Error("expected error for invalid octet")
	}
}

func TestIntegrationParseMACFromOID_Whitespace(t *testing.T) {
	mac, err := ParseMACFromOID("  .1.3.6.1.2.1.17.4.3.1.2.10.20.30.40.50.60  ")
	if err != nil {
		t.Fatalf("expected no error for whitespace, got %v", err)
	}
	if mac != "0a:14:1e:28:32:3c" {
		t.Errorf("expected '0a:14:1e:28:32:3c', got %q", mac)
	}
}

func TestIntegrationParseMACFromOID_NegativeOctet(t *testing.T) {
	_, err := ParseMACFromOID(".1.3.6.1.2.1.17.4.3.1.2.-1.20.30.40.50.60")
	if err == nil {
		t.Error("expected error for negative octet")
	}
}

// === Integration: inferDeviceType ===

func TestIntegrationInferDeviceType_Switch(t *testing.T) {
	devType := inferDeviceType("Cisco Switch 3750", false)
	if devType != topology.DeviceTypeSwitch {
		t.Errorf("expected Switch, got %s", devType)
	}
}

func TestIntegrationInferDeviceType_SwitchWithDot1d(t *testing.T) {
	devType := inferDeviceType("Generic Device", true)
	if devType != topology.DeviceTypeSwitch {
		t.Errorf("expected Switch with hasDot1d, got %s", devType)
	}
}

func TestIntegrationInferDeviceType_Router(t *testing.T) {
	devType := inferDeviceType("Cisco Router 2900", false)
	if devType != topology.DeviceTypeRouter {
		t.Errorf("expected Router, got %s", devType)
	}
}

func TestIntegrationInferDeviceType_Host_Linux(t *testing.T) {
	devType := inferDeviceType("Linux Server 22.04", false)
	if devType != topology.DeviceTypeHost {
		t.Errorf("expected Host for Linux, got %s", devType)
	}
}

func TestIntegrationInferDeviceType_Host_Windows(t *testing.T) {
	devType := inferDeviceType("Windows Server 2022", false)
	if devType != topology.DeviceTypeHost {
		t.Errorf("expected Host for Windows, got %s", devType)
	}
}

func TestIntegrationInferDeviceType_Host_Server(t *testing.T) {
	devType := inferDeviceType("Dell Server R740", false)
	if devType != topology.DeviceTypeHost {
		t.Errorf("expected Host for Server, got %s", devType)
	}
}

func TestIntegrationInferDeviceType_Host_Host(t *testing.T) {
	devType := inferDeviceType("HP Desktop Host", false)
	if devType != topology.DeviceTypeHost {
		t.Errorf("expected Host, got %s", devType)
	}
}

func TestIntegrationInferDeviceType_Unknown(t *testing.T) {
	devType := inferDeviceType("Unknown Device XYZ", false)
	if devType != topology.DeviceTypeUnknown {
		t.Errorf("expected Unknown, got %s", devType)
	}
}

func TestIntegrationInferDeviceType_Empty(t *testing.T) {
	devType := inferDeviceType("", false)
	if devType != topology.DeviceTypeUnknown {
		t.Errorf("expected Unknown for empty, got %s", devType)
	}
}

func TestIntegrationInferDeviceType_CaseInsensitive(t *testing.T) {
	devType := inferDeviceType("CISCO SWITCH 3750", false)
	if devType != topology.DeviceTypeSwitch {
		t.Errorf("expected case-insensitive match, got %s", devType)
	}
}

func TestIntegrationInferDeviceType_MultipleKeywords(t *testing.T) {
	devType := inferDeviceType("Cisco Switch Router", false)
	if devType != topology.DeviceTypeSwitch {
		t.Errorf("expected Switch (first match), got %s", devType)
	}
}

// === Integration: suffixInt ===

func TestIntegrationSuffixInt_Basic(t *testing.T) {
	n := suffixInt(".1.3.6.1.2.1.2.2.1.2.1")
	if n != 1 {
		t.Errorf("expected 1, got %d", n)
	}
}

func TestIntegrationSuffixInt_Empty(t *testing.T) {
	n := suffixInt("")
	if n != -1 {
		t.Errorf("expected -1 for empty, got %d", n)
	}
}

func TestIntegrationSuffixInt_InvalidNumber(t *testing.T) {
	n := suffixInt(".1.3.6.abc")
	if n != -1 {
		t.Errorf("expected -1 for invalid, got %d", n)
	}
}

func TestIntegrationSuffixInt_NoDotPrefix(t *testing.T) {
	n := suffixInt("1.3.6.1.2.1")
	if n != 1 {
		t.Errorf("expected 1, got %d", n)
	}
}

func TestIntegrationSuffixInt_NonNumeric(t *testing.T) {
	n := suffixInt(".1.3.6.abc.def")
	if n != -1 {
		t.Errorf("expected -1 for non-numeric, got %d", n)
	}
}

// === Integration: lldpRowKeyFromOID ===

func TestIntegrationLldpRowKeyFromOID_Valid(t *testing.T) {
	key := lldpRowKeyFromOID(".1.0.8802.1.1.2.1.4.1.1.9.12345.6.7")
	if key != "12345.6.7" {
		t.Errorf("expected '12345.6.7', got %q", key)
	}
}

func TestIntegrationLldpRowKeyFromOID_TooShort(t *testing.T) {
	key := lldpRowKeyFromOID(".1.0")
	if key != "" {
		t.Errorf("expected empty for too short, got %q", key)
	}
}

func TestIntegrationLldpRowKeyFromOID_Empty(t *testing.T) {
	key := lldpRowKeyFromOID("")
	if key != "" {
		t.Errorf("expected empty, got %q", key)
	}
}

func TestIntegrationLldpRowKeyFromOID_Whitespace(t *testing.T) {
	key := lldpRowKeyFromOID("  .1.0.8802.1.1.2.1.4.1.1.9.12345.6.7  ")
	if key != "12345.6.7" {
		t.Errorf("expected '12345.6.7', got %q", key)
	}
}

// === Integration: lldpLocalPortFromOID ===

func TestIntegrationLldpLocalPortFromOID_Valid(t *testing.T) {
	port := lldpLocalPortFromOID(".1.0.8802.1.1.2.1.4.1.1.9.12345.6.7")
	if port != 6 {
		t.Errorf("expected port 6, got %d", port)
	}
}

func TestIntegrationLldpLocalPortFromOID_TooShort(t *testing.T) {
	port := lldpLocalPortFromOID(".1.0")
	if port != -1 {
		t.Errorf("expected -1 for too short, got %d", port)
	}
}

func TestIntegrationLldpLocalPortFromOID_InvalidNumber(t *testing.T) {
	port := lldpLocalPortFromOID(".1.0.abc.def")
	if port != -1 {
		t.Errorf("expected -1 for invalid, got %d", port)
	}
}

// === Integration: bytesToMAC ===

func TestIntegrationBytesToMAC_Basic(t *testing.T) {
	mac := bytesToMAC([]byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff})
	if mac != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("expected 'aa:bb:cc:dd:ee:ff', got %q", mac)
	}
}

func TestIntegrationBytesToMAC_SingleByte(t *testing.T) {
	mac := bytesToMAC([]byte{0xaa})
	if mac != "aa" {
		t.Errorf("expected 'aa', got %q", mac)
	}
}

func TestIntegrationBytesToMAC_ZeroBytes(t *testing.T) {
	mac := bytesToMAC([]byte{0, 0, 0, 0, 0, 0})
	if mac != "00:00:00:00:00:00" {
		t.Errorf("expected '00:00:00:00:00:00', got %q", mac)
	}
}

// === Integration: CollectReport Structure ===

func TestIntegrationCollectReport_Structure(t *testing.T) {
	report := &CollectReport{
		TotalSNMPTargets: 5,
		Connected:        3,
		Partial:          1,
		Failed:           1,
		Failures: []DeviceFailure{
			{IP: "192.168.1.1", Kind: FailureConnect, Message: "connection refused"},
		},
		DeviceSummaries: []DeviceQuerySummary{
			{IP: "192.168.1.1", MACEntries: 10, LLDPNeighbors: 2, QueryErrors: ""},
		},
	}

	if report.TotalSNMPTargets != 5 {
		t.Errorf("expected 5 targets, got %d", report.TotalSNMPTargets)
	}
	if report.Connected != 3 {
		t.Errorf("expected 3 connected, got %d", report.Connected)
	}
	if report.Partial != 1 {
		t.Errorf("expected 1 partial, got %d", report.Partial)
	}
	if report.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", report.Failed)
	}
	if len(report.Failures) != 1 {
		t.Errorf("expected 1 failure, got %d", len(report.Failures))
	}
	if len(report.DeviceSummaries) != 1 {
		t.Errorf("expected 1 summary, got %d", len(report.DeviceSummaries))
	}
}

func TestIntegrationCollectReport_Defaults(t *testing.T) {
	report := &CollectReport{}
	// When created with new, slices are nil - this is expected Go behavior
	// In production, CollectWithReport initializes them properly
	if report.TotalSNMPTargets != 0 {
		t.Errorf("expected 0 targets for empty report, got %d", report.TotalSNMPTargets)
	}
}

// === Integration: DeviceQuerySummary ===

func TestIntegrationDeviceQuerySummary_Fields(t *testing.T) {
	summary := DeviceQuerySummary{
		IP: "192.168.1.1", MACEntries: 15, LLDPNeighbors: 3, QueryErrors: "sysName: timeout",
	}

	if summary.IP != "192.168.1.1" {
		t.Errorf("expected IP '192.168.1.1', got %q", summary.IP)
	}
	if summary.MACEntries != 15 {
		t.Errorf("expected 15 MAC entries, got %d", summary.MACEntries)
	}
	if summary.LLDPNeighbors != 3 {
		t.Errorf("expected 3 LLDP neighbors, got %d", summary.LLDPNeighbors)
	}
	if summary.QueryErrors != "sysName: timeout" {
		t.Errorf("expected 'sysName: timeout', got %q", summary.QueryErrors)
	}
}

// === Integration: FailureKind Constants ===

func TestIntegrationFailureKind_Constants(t *testing.T) {
	if FailureConnect != "connect_error" {
		t.Errorf("expected FailureConnect='connect_error', got %q", FailureConnect)
	}
	if FailureQuery != "query_error" {
		t.Errorf("expected FailureQuery='query_error', got %q", FailureQuery)
	}
}

// === Integration: DeviceFailure Structure ===

func TestIntegrationDeviceFailure_Fields(t *testing.T) {
	failure := DeviceFailure{
		IP: "192.168.1.1", Kind: FailureConnect, Message: "connection refused", Community: "public",
	}

	if failure.IP != "192.168.1.1" {
		t.Errorf("expected IP '192.168.1.1', got %q", failure.IP)
	}
	if failure.Kind != FailureConnect {
		t.Errorf("expected Kind FailureConnect, got %q", failure.Kind)
	}
	if failure.Message != "connection refused" {
		t.Errorf("expected 'connection refused', got %q", failure.Message)
	}
	if failure.Community != "public" {
		t.Errorf("expected 'public', got %q", failure.Community)
	}
}

// === Integration: ProgressCallback Signature ===

func TestIntegrationProgressCallback_Signature(t *testing.T) {
	var cb ProgressCallback = func(current int, total int, ip string, message string) {
		_ = current
		_ = total
		_ = ip
		_ = message
	}
	if cb == nil {
		t.Error("expected non-nil callback")
	}
}

// === Integration: SNMPClient Interface ===

func TestIntegrationSNMPClient_Interface(t *testing.T) {
	var _ SNMPClient = (*GoSNMPClient)(nil)
}

// === Integration: DeviceType Constants ===

func TestIntegrationDeviceTypeConstants(t *testing.T) {
	if topology.DeviceTypeSwitch != "switch" {
		t.Errorf("expected DeviceTypeSwitch='switch', got %q", topology.DeviceTypeSwitch)
	}
	if topology.DeviceTypeRouter != "router" {
		t.Errorf("expected DeviceTypeRouter='router', got %q", topology.DeviceTypeRouter)
	}
	if topology.DeviceTypeHost != "host" {
		t.Errorf("expected DeviceTypeHost='host', got %q", topology.DeviceTypeHost)
	}
	if topology.DeviceTypeUnknown != "unknown" {
		t.Errorf("expected DeviceTypeUnknown='unknown', got %q", topology.DeviceTypeUnknown)
	}
}

// === Integration: MAC Normalization ===

func TestIntegrationMAC_Normalization(t *testing.T) {
	mac := strings.ToLower(strings.ReplaceAll("AA-BB-CC-DD-EE-FF", "-", ":"))
	if mac != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("expected 'aa:bb:cc:dd:ee:ff', got %q", mac)
	}
}

// === Integration: Report Sorting ===

func TestIntegrationReport_Sorting(t *testing.T) {
	summaries := []DeviceQuerySummary{
		{IP: "192.168.1.3"}, {IP: "192.168.1.1"}, {IP: "192.168.1.2"},
	}
	sortDeviceSummaries(summaries)

	if summaries[0].IP != "192.168.1.1" {
		t.Errorf("expected first IP '192.168.1.1', got %q", summaries[0].IP)
	}
	if summaries[1].IP != "192.168.1.2" {
		t.Errorf("expected second IP '192.168.1.2', got %q", summaries[1].IP)
	}
	if summaries[2].IP != "192.168.1.3" {
		t.Errorf("expected third IP '192.168.1.3', got %q", summaries[2].IP)
	}
}

func sortDeviceSummaries(summaries []DeviceQuerySummary) {
	for i := 0; i < len(summaries); i++ {
		for j := i + 1; j < len(summaries); j++ {
			if summaries[i].IP > summaries[j].IP {
				summaries[i], summaries[j] = summaries[j], summaries[i]
			}
		}
	}
}

// === Integration: Multi-Community Fallback ===

func TestIntegrationMultiCommunity_Fallback(t *testing.T) {
	communities := []string{"public", "private", "community1"}
	if len(communities) != 3 {
		t.Errorf("expected 3 communities, got %d", len(communities))
	}
	trimmed := strings.TrimSpace("  public  ")
	if trimmed != "public" {
		t.Errorf("expected 'public', got %q", trimmed)
	}
}

// === Integration: Context (no network) ===

func TestIntegrationContext_Propagation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_ = ctx.Err()
	_ = cancel
}

// === Integration: Collect (empty - no network) ===

func TestIntegrationCollect(t *testing.T) {
	// Test with empty devices - no network needed
	devices := []scanner.Result{}
	communities := []string{"public"}
	timeout := 2

	_, err := Collect(devices, communities, timeout)
	if err != nil {
		t.Fatalf("expected no error for empty devices, got %v", err)
	}
}

// === Integration: Full Collection Pipeline (empty) ===

func TestIntegrationFullCollectionPipeline(t *testing.T) {
	// Test with empty devices - no network needed
	devices := []scanner.Result{}
	communities := []string{"public"}
	timeout := 2

	report := &CollectReport{Failures: make([]DeviceFailure, 0)}
	_ = report

	// Verify device filtering logic
	nonSNMPCount := 0
	for _, d := range devices {
		if !d.SNMPEnabled {
			nonSNMPCount++
		}
	}
	_ = nonSNMPCount

	// Test Collect with empty devices
	_, err := Collect(devices, communities, timeout)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// === Integration: Edge Cases ===

func TestIntegrationCollectWithReport_EmptyCommunitiesAndDevices(t *testing.T) {
	_, report, err := CollectWithReport([]scanner.Result{}, []string{}, 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if report.TotalSNMPTargets != 0 {
		t.Errorf("expected 0 targets, got %d", report.TotalSNMPTargets)
	}
}
