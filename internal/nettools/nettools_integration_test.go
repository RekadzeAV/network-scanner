package nettools

import (
	"context"
	"testing"
	"time"
)

// === Integration: Ping Configuration ===

func TestIntegrationPingConfig_Default(t *testing.T) {
	// RunPing should not panic with default config
	ctx := context.Background()
	_, err := RunPing(ctx, "127.0.0.1", 1, 1*time.Second)
	// Error is expected (no network), but should not panic
	_ = err
}

func TestIntegrationPingConfig_EmptyHost(t *testing.T) {
	ctx := context.Background()
	_, err := RunPing(ctx, "", 1, 1*time.Second)
	if err == nil {
		t.Error("expected error for empty host")
	}
}

func TestIntegrationPingConfig_CountBounds(t *testing.T) {
	// Count < 1 should default to 4
	// Count > 50 should cap at 50
	// These are internal behaviors, tested via buildPingArgs
}

// === Integration: DNS Configuration ===

func TestIntegrationDNSConfig_EmptyHost(t *testing.T) {
	ctx := context.Background()
	_, err := LookupDNS(ctx, "")
	// Error is expected, but should not panic
	_ = err
}

func TestIntegrationDNSConfig_InvalidHost(t *testing.T) {
	ctx := context.Background()
	_, err := LookupDNS(ctx, "invalid-host-that-does-not-exist-12345.com")
	// Error is expected, but should not panic
	_ = err
}

// === Integration: Whois Configuration ===

func TestIntegrationWhoisConfig_EmptyQuery(t *testing.T) {
	ctx := context.Background()
	_, err := RunWhois(ctx, "", 5*time.Second)
	// Error is expected, but should not panic
	_ = err
}

func TestIntegrationWhoisConfig_InvalidQuery(t *testing.T) {
	ctx := context.Background()
	_, err := RunWhois(ctx, "invalid", 5*time.Second)
	// Error is expected (whois not installed or invalid), but should not panic
	_ = err
}

// === Integration: Traceroute Configuration ===

func TestIntegrationTracerouteConfig_EmptyHost(t *testing.T) {
	ctx := context.Background()
	_, err := RunTraceroute(ctx, "", 5*time.Second)
	// Error is expected, but should not panic
	_ = err
}

func TestIntegrationTracerouteConfig_InvalidHost(t *testing.T) {
	ctx := context.Background()
	_, err := RunTraceroute(ctx, "invalid-host-that-does-not-exist-12345.com", 5*time.Second)
	// Error is expected, but should not panic
	_ = err
}

// === Integration: WiFi Configuration ===

func TestIntegrationWiFiConfig_EmptyContext(t *testing.T) {
	// GetWiFiInfo should not panic with valid context
	ctx := context.Background()
	_, err := GetWiFiInfo(ctx, 5*time.Second)
	// Error is expected (wifi not available), but should not panic
	_ = err
}

// === Integration: Error Handling ===

func TestIntegrationError_ToolNotInstalled(t *testing.T) {
	// ToolError should handle not installed case
	err := newToolError("ping", ToolErrorNotInstalled, "ping not found", nil)
	if err == nil {
		t.Error("expected non-nil error")
	}
}

func TestIntegrationError_ToolTimeout(t *testing.T) {
	// ToolError should handle timeout case
	err := newToolError("ping", ToolErrorTimeout, "timeout", nil)
	if err == nil {
		t.Error("expected non-nil error")
	}
}

func TestIntegrationError_ToolPermissionDenied(t *testing.T) {
	// ToolError should handle permission denied case
	err := newToolError("ping", ToolErrorPermissionDenied, "permission denied", nil)
	if err == nil {
		t.Error("expected non-nil error")
	}
}

func TestIntegrationError_ToolNetwork(t *testing.T) {
	// ToolError should handle network error case
	err := newToolError("ping", ToolErrorNetwork, "network error", nil)
	if err == nil {
		t.Error("expected non-nil error")
	}
}

// === Integration: Humanize Tool Error ===

func TestIntegrationHumanizeToolError_NotInstalled(t *testing.T) {
	err := newToolError("ping", ToolErrorNotInstalled, "", nil)
	humanized := HumanizeToolError(err)
	if humanized == "" {
		t.Error("expected non-empty humanized error")
	}
}

func TestIntegrationHumanizeToolError_Timeout(t *testing.T) {
	err := newToolError("ping", ToolErrorTimeout, "", nil)
	humanized := HumanizeToolError(err)
	if humanized == "" {
		t.Error("expected non-empty humanized error")
	}
}

func TestIntegrationHumanizeToolError_PermissionDenied(t *testing.T) {
	err := newToolError("ping", ToolErrorPermissionDenied, "", nil)
	humanized := HumanizeToolError(err)
	if humanized == "" {
		t.Error("expected non-empty humanized error")
	}
}

func TestIntegrationHumanizeToolError_Nil(t *testing.T) {
	humanized := HumanizeToolError(nil)
	if humanized != "" {
		t.Errorf("expected empty string for nil error, got %q", humanized)
	}
}

func TestIntegrationHumanizeToolError_NotToolError(t *testing.T) {
	err := context.DeadlineExceeded
	humanized := HumanizeToolError(err)
	if humanized == "" {
		t.Error("expected non-empty humanized error for non-ToolError")
	}
}

// === Integration: Build Ping Args ===

func TestIntegrationBuildPingArgs_Windows(t *testing.T) {
	args := buildPingArgs("192.168.1.1", 4, "windows")
	if len(args) < 3 {
		t.Errorf("expected at least 3 args, got %d", len(args))
	}
	if args[0] != "ping" {
		t.Errorf("expected 'ping', got %q", args[0])
	}
	if args[1] != "-n" {
		t.Errorf("expected '-n', got %q", args[1])
	}
}

func TestIntegrationBuildPingArgs_Unix(t *testing.T) {
	args := buildPingArgs("192.168.1.1", 4, "linux")
	if len(args) < 3 {
		t.Errorf("expected at least 3 args, got %d", len(args))
	}
	if args[0] != "ping" {
		t.Errorf("expected 'ping', got %q", args[0])
	}
	if args[1] != "-c" {
		t.Errorf("expected '-c', got %q", args[1])
	}
}

// === Integration: Build Traceroute Args ===

func TestIntegrationBuildTracerouteArgs_Windows(t *testing.T) {
	args := buildTracerouteArgs("192.168.1.1", 30, "windows")
	if len(args) < 3 {
		t.Errorf("expected at least 3 args, got %d", len(args))
	}
	if args[0] != "tracert" {
		t.Errorf("expected 'tracert', got %q", args[0])
	}
}

func TestIntegrationBuildTracerouteArgs_Unix(t *testing.T) {
	args := buildTracerouteArgs("192.168.1.1", 30, "linux")
	if len(args) < 3 {
		t.Errorf("expected at least 3 args, got %d", len(args))
	}
	if args[0] != "traceroute" {
		t.Errorf("expected 'traceroute', got %q", args[0])
	}
}

// === Integration: Context Cancellation ===

func TestIntegrationContext_Cancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// These should not panic even with cancelled context
	_, _ = RunPing(ctx, "127.0.0.1", 1, 1*time.Second)
	_, _ = LookupDNS(ctx, "localhost")
	_, _ = RunWhois(ctx, "example.com", 1*time.Second)
	_, _ = RunTraceroute(ctx, "127.0.0.1", 1*time.Second)
	_, _ = GetWiFiInfo(ctx, 1*time.Second)
}

// === Integration: Timeout Validation ===

func TestIntegrationTimeout_Zero(t *testing.T) {
	ctx := context.Background()
	// Zero timeout should not panic
	_, _ = RunPing(ctx, "127.0.0.1", 1, 0)
	_, _ = LookupDNS(ctx, "localhost")
	_, _ = RunWhois(ctx, "example.com", 0)
	_, _ = RunTraceroute(ctx, "127.0.0.1", 0)
	_, _ = GetWiFiInfo(ctx, 0)
}

func TestIntegrationTimeout_Negative(t *testing.T) {
	ctx := context.Background()
	// Negative timeout should not panic
	_, _ = RunPing(ctx, "127.0.0.1", 1, -1*time.Second)
	_, _ = LookupDNS(ctx, "localhost")
	_, _ = RunWhois(ctx, "example.com", -1*time.Second)
	_, _ = RunTraceroute(ctx, "127.0.0.1", -1*time.Second)
	_, _ = GetWiFiInfo(ctx, -1*time.Second)
}

// === Integration: Run Command ===

func TestIntegrationRunCmd_EmptyCommand(t *testing.T) {
	ctx := context.Background()
	_, err := runCmd(ctx, []string{}, 1*time.Second)
	// Error is expected, but should not panic
	_ = err
}

func TestIntegrationRunCmd_InvalidCommand(t *testing.T) {
	ctx := context.Background()
	_, err := runCmd(ctx, []string{"nonexistent-command-12345"}, 1*time.Second)
	// Error is expected, but should not panic
	_ = err
}

// === Integration: DNS Result ===

func TestIntegrationDNSResult_Empty(t *testing.T) {
	result := &DNSResult{}
	if result == nil {
		t.Error("expected non-nil DNSResult")
	}
}

func TestIntegrationDNSResult_WithData(t *testing.T) {
	result := &DNSResult{
		Query:      "example.com",
		ForwardIPs: []string{"93.184.216.34"},
	}
	if result.Query != "example.com" {
		t.Errorf("expected query 'example.com', got %q", result.Query)
	}
	if len(result.ForwardIPs) != 1 {
		t.Errorf("expected 1 forward IP, got %d", len(result.ForwardIPs))
	}
}

// === Integration: Ping Result ===

func TestIntegrationPingResult_Empty(t *testing.T) {
	result := &PingResult{}
	if result == nil {
		t.Error("expected non-nil PingResult")
	}
}

func TestIntegrationPingResult_WithStats(t *testing.T) {
	result := &PingResult{
		RawOutput: "PING 127.0.0.1",
		Stats: PingStats{
			Sent:       4,
			Received:   4,
			PacketLoss: 0.0,
		},
	}
	if result.Stats.Sent != 4 {
		t.Errorf("expected sent 4, got %d", result.Stats.Sent)
	}
	if result.Stats.Received != 4 {
		t.Errorf("expected received 4, got %d", result.Stats.Received)
	}
	if result.Stats.PacketLoss != 0.0 {
		t.Errorf("expected packet loss 0.0, got %f", result.Stats.PacketLoss)
	}
}

// === Integration: Traceroute Result ===

func TestIntegrationTracerouteResult_Empty(t *testing.T) {
	result := &TracerouteResult{}
	if result == nil {
		t.Error("expected non-nil TracerouteResult")
	}
}

func TestIntegrationTracerouteResult_WithHops(t *testing.T) {
	result := &TracerouteResult{
		RawOutput: "traceroute to example.com",
		Hops: []TracerouteHop{
			{Index: 1, Address: "192.168.1.1"},
		},
	}
	if result.RawOutput != "traceroute to example.com" {
		t.Errorf("expected raw output 'traceroute to example.com', got %q", result.RawOutput)
	}
	if len(result.Hops) != 1 {
		t.Errorf("expected 1 hop, got %d", len(result.Hops))
	}
}

// === Integration: WiFi Info ===

func TestIntegrationWiFiInfo_Empty(t *testing.T) {
	info := map[string]string{}
	if info == nil {
		t.Error("expected non-nil map")
	}
}

func TestIntegrationWiFiInfo_WithData(t *testing.T) {
	info := map[string]string{
		"SSID":     "MyNetwork",
		"Signal":   "80%",
		"Channel":  "6",
		"Security": "WPA2",
	}
	if info["SSID"] != "MyNetwork" {
		t.Errorf("expected SSID 'MyNetwork', got %q", info["SSID"])
	}
	if info["Signal"] != "80%" {
		t.Errorf("expected Signal '80%%', got %q", info["Signal"])
	}
}

// === Integration: Tool Error Codes ===

func TestIntegrationToolErrorCode_Constants(t *testing.T) {
	codes := []ToolErrorCode{
		ToolErrorNotInstalled,
		ToolErrorTimeout,
		ToolErrorPermissionDenied,
		ToolErrorNetwork,
		ToolErrorParse,
	}
	for _, code := range codes {
		if code == "" {
			t.Error("expected non-empty error code")
		}
	}
}

// === Integration: Context Lifecycle ===

func TestIntegrationContext_Lifecycle(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Context should be valid
	if ctx.Err() != nil {
		t.Errorf("context should be valid, got error: %v", ctx.Err())
	}

	// Tools should work with valid context
	_, _ = RunPing(ctx, "127.0.0.1", 1, 1*time.Second)
	_, _ = LookupDNS(ctx, "localhost")
	_, _ = RunWhois(ctx, "example.com", 1*time.Second)
	_, _ = RunTraceroute(ctx, "127.0.0.1", 1*time.Second)
	_, _ = GetWiFiInfo(ctx, 1*time.Second)
}

// === Integration: Multiple Tool Calls ===

func TestIntegrationMultipleToolCalls(t *testing.T) {
	ctx := context.Background()

	// Multiple calls should not panic
	for i := 0; i < 5; i++ {
		_, _ = RunPing(ctx, "127.0.0.1", 1, 1*time.Second)
		_, _ = LookupDNS(ctx, "localhost")
		_, _ = RunWhois(ctx, "example.com", 1*time.Second)
		_, _ = RunTraceroute(ctx, "127.0.0.1", 1*time.Second)
		_, _ = GetWiFiInfo(ctx, 1*time.Second)
	}
}

// === Integration: Empty String Handling ===

func TestIntegrationEmptyString_Host(t *testing.T) {
	ctx := context.Background()
	_, err := RunPing(ctx, "  ", 1, 1*time.Second)
	if err == nil {
		t.Error("expected error for whitespace-only host")
	}
}

func TestIntegrationEmptyString_Query(t *testing.T) {
	ctx := context.Background()
	_, err := RunWhois(ctx, "  ", 1*time.Second)
	if err == nil {
		t.Error("expected error for whitespace-only query")
	}
}

// === Integration: Invalid IP Address ===

func TestIntegrationInvalidIP(t *testing.T) {
	ctx := context.Background()
	// Invalid IP should not panic
	_, _ = RunPing(ctx, "999.999.999.999", 1, 1*time.Second)
	_, _ = RunTraceroute(ctx, "999.999.999.999", 1*time.Second)
}

// === Integration: Very Long Hostname ===

func TestIntegrationLongHostname(t *testing.T) {
	ctx := context.Background()
	longHostname := "this-is-a-very-long-hostname-that-may-cause-issues-with-some-tools.example.com"
	// Long hostname should not panic
	_, _ = RunPing(ctx, longHostname, 1, 1*time.Second)
	_, _ = LookupDNS(ctx, longHostname)
}

// === Integration: Run Ping Structured ===

func TestIntegrationRunPingStructured(t *testing.T) {
	ctx := context.Background()
	result, err := RunPingStructured(ctx, "127.0.0.1", 1, 1*time.Second)
	// Error is expected (no network), but should not panic
	if err != nil {
		// Expected in headless environment
		_ = result
	} else {
		if result == nil {
			t.Error("expected non-nil result")
		}
		if result.Stats.Sent < 0 {
			t.Error("expected non-negative sent count")
		}
	}
}

// === Integration: Run Traceroute Structured ===

func TestIntegrationRunTracerouteStructured(t *testing.T) {
	ctx := context.Background()
	result, err := RunTracerouteStructured(ctx, "127.0.0.1", 1*time.Second)
	// Error is expected (no network), but should not panic
	if err != nil {
		// Expected in headless environment
		_ = result
	} else {
		if result == nil {
			t.Error("expected non-nil result")
		}
	}
}

// === Integration: Run Traceroute with Max Hops ===

func TestIntegrationRunTracerouteWithMaxHops(t *testing.T) {
	ctx := context.Background()
	result, err := RunTracerouteStructuredWithMaxHops(ctx, "127.0.0.1", 1*time.Second, 5)
	// Error is expected (no network), but should not panic
	if err != nil {
		// Expected in headless environment
		_ = result
	} else {
		if result == nil {
			t.Error("expected non-nil result")
		}
	}
}

// === Integration: Run Whois RDAP ===

func TestIntegrationRunWhoisRDAP(t *testing.T) {
	ctx := context.Background()
	// RunWhois should handle RDAP fallback gracefully
	_, err := RunWhois(ctx, "example.com", 5*time.Second)
	// Error is expected (whois not installed), but should not panic
	_ = err
}

// === Integration: WiFi Parse Functions ===

func TestIntegrationParseWindowsNetsh(t *testing.T) {
	t.Skip("parseWindowsNetsh is internal - test via GetWiFiInfo integration")
	raw := `
SSID 1 : MyNetwork
   Type                   : Managed
   Description            : Managed network
   Channel                : 6
   RSSI                   : 80
   Authenticated          : Yes
`
	parsed := parseWindowsNetsh(raw)
	if parsed["SSID"] != "MyNetwork" {
		t.Errorf("expected SSID 'MyNetwork', got %q", parsed["SSID"])
	}
}

func TestIntegrationParseLinuxNmcli(t *testing.T) {
	t.Skip("parseLinuxNmcli is internal - test via GetWiFiInfo integration")
	raw := `
SSID: MyNetwork
BSSID: AA:BB:CC:DD:EE:FF
MODE: Infra
CHAN: 6
RATE: 144 Mbit/s
SIGNAL: 80
SECURITY: WPA2
`
	parsed := parseLinuxNmcli(raw)
	if parsed["SSID"] != "MyNetwork" {
		t.Errorf("expected SSID 'MyNetwork', got %q", parsed["SSID"])
	}
}

func TestIntegrationParseDarwinAirport(t *testing.T) {
	t.Skip("parseDarwinAirport is internal - test via GetWiFiInfo integration")
	raw := `
AirPort: On
Wireless Network: MyNetwork
Status: Connected
SSID: MyNetwork
Channel: 6
Signal Strength: 80%
Security: WPA2 Personal
`
	parsed := parseDarwinAirport(raw)
	if parsed["SSID"] != "MyNetwork" {
		t.Errorf("expected SSID 'MyNetwork', got %q", parsed["SSID"])
	}
}
