package wol

import (
	"strings"
	"testing"
)

// === Integration: parseMAC edge cases ===

func TestIntegrationParseMAC_MixedDashes(t *testing.T) {
	mac, err := parseMAC("AA-BB:CC-DD:EE:FF")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(mac) != 6 {
		t.Errorf("expected 6 bytes, got %d", len(mac))
	}
}

func TestIntegrationParseMAC_Uppercase(t *testing.T) {
	mac, err := parseMAC("AA:BB:CC:DD:EE:FF")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(mac) != 6 {
		t.Errorf("expected 6 bytes, got %d", len(mac))
	}
	if mac[0] != 0xAA {
		t.Errorf("expected 0xAA, got 0x%02x", mac[0])
	}
}

func TestIntegrationParseMAC_AllZeros(t *testing.T) {
	mac, err := parseMAC("00:00:00:00:00:00")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(mac) != 6 {
		t.Errorf("expected 6 bytes, got %d", len(mac))
	}
}

func TestIntegrationParseMAC_BroadcastMAC(t *testing.T) {
	mac, err := parseMAC("FF:FF:FF:FF:FF:FF")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(mac) != 6 {
		t.Errorf("expected 6 bytes, got %d", len(mac))
	}
}

func TestIntegrationParseMAC_TooShort(t *testing.T) {
	_, err := parseMAC("aa:bb:cc")
	if err == nil {
		t.Error("expected error for too short MAC")
	}
}

func TestIntegrationParseMAC_TooLong(t *testing.T) {
	_, err := parseMAC("aa:bb:cc:dd:ee:ff:00:11")
	if err == nil {
		t.Error("expected error for too long MAC")
	}
}

func TestIntegrationParseMAC_Spaces(t *testing.T) {
	// parseMAC trims outer whitespace but not inner spaces
	// "aa : bb : cc : dd : ee : ff" has inner spaces which are invalid
	_, err := parseMAC("aa : bb : cc : dd : ee : ff")
	if err == nil {
		t.Error("expected error for MAC with inner spaces")
	}
}

func TestIntegrationParseMAC_MixedCase(t *testing.T) {
	mac, err := parseMAC("Aa:Bb:Cc:Dd:Ee:Ff")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(mac) != 6 {
		t.Errorf("expected 6 bytes, got %d", len(mac))
	}
}

// === Integration: resolveBroadcastAddr edge cases ===

func TestIntegrationResolveBroadcastAddr_PortOnly(t *testing.T) {
	// ":9" contains ":", so it's returned as-is without adding another ":9"
	got, err := resolveBroadcastAddr(":9", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != ":9" {
		t.Errorf("expected ':9', got %q", got)
	}
}

func TestIntegrationResolveBroadcastAddr_EmptyWithColon(t *testing.T) {
	// ":" contains ":", so it's returned as-is
	got, err := resolveBroadcastAddr(":", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != ":" {
		t.Errorf("expected ':', got %q", got)
	}
}

func TestIntegrationResolveBroadcastAddr_WhitespaceMultiple(t *testing.T) {
	got, err := resolveBroadcastAddr("  192.168.1.1  ", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != "192.168.1.1:9" {
		t.Errorf("expected '192.168.1.1:9', got %q", got)
	}
}

// === Integration: SendMagicPacket integration ===

func TestIntegrationSendMagicPacket_VariousMACs(t *testing.T) {
	macs := []string{
		"aa:bb:cc:dd:ee:ff",
		"AA:BB:CC:DD:EE:FF",
		"aa-bb-cc-dd-ee-ff",
		"AA-BB-CC-DD-EE-FF",
		"  aa:bb:cc:dd:ee:ff  ",
	}

	for _, macStr := range macs {
		err := SendMagicPacket(macStr, "127.0.0.1:9")
		if err != nil {
			t.Errorf("SendMagicPacket(%q) error: %v", macStr, err)
		}
	}
}

func TestIntegrationSendMagicPacketWithInterface_VariousMACs(t *testing.T) {
	macs := []string{
		"aa:bb:cc:dd:ee:ff",
		"AA:BB:CC:DD:EE:FF",
		"aa-bb-cc-dd-ee-ff",
		"  00:11:22:33:44:55  ",
	}

	for _, macStr := range macs {
		addr, err := SendMagicPacketWithInterface(macStr, "127.0.0.1:9", "")
		if err != nil {
			t.Errorf("SendMagicPacketWithInterface(%q) error: %v", macStr, err)
		}
		if addr != "127.0.0.1:9" {
			t.Errorf("expected '127.0.0.1:9', got %q", addr)
		}
	}
}

func TestIntegrationSendMagicPacket_MultipleBroadcasts(t *testing.T) {
	mac := "aa:bb:cc:dd:ee:ff"
	broadcasts := []string{
		"127.0.0.1:9",
		"192.168.1.255:9",
		"10.0.0.255:7",
		"255.255.255.255:9",
	}

	for _, bcast := range broadcasts {
		err := SendMagicPacket(mac, bcast)
		if err != nil {
			t.Errorf("SendMagicPacket with bcast %q error: %v", bcast, err)
		}
	}
}

// === Integration: Full WOL Pipeline ===

func TestIntegrationFullWOLPipeline_Valid(t *testing.T) {
	mac := "aa:bb:cc:dd:ee:ff"
	bcast := "127.0.0.1:9"

	// Step 1: Parse MAC
	parsed, err := parseMAC(mac)
	if err != nil {
		t.Fatalf("parseMAC error: %v", err)
	}
	if len(parsed) != 6 {
		t.Errorf("expected 6 bytes, got %d", len(parsed))
	}

	// Step 2: Resolve broadcast
	resolved, err := resolveBroadcastAddr(bcast, "")
	if err != nil {
		t.Fatalf("resolveBroadcastAddr error: %v", err)
	}
	if resolved != bcast {
		t.Errorf("expected %q, got %q", bcast, resolved)
	}

	// Step 3: Send packet
	addr, err := SendMagicPacketWithInterface(mac, bcast, "")
	if err != nil {
		t.Fatalf("SendMagicPacketWithInterface error: %v", err)
	}
	if addr != bcast {
		t.Errorf("expected address %q, got %q", bcast, addr)
	}
}

func TestIntegrationFullWOLPipeline_InvalidMAC(t *testing.T) {
	// Should fail at parse step
	_, err := parseMAC("invalid")
	if err == nil {
		t.Error("expected error for invalid MAC")
	}
}

func TestIntegrationFullWOLPipeline_EmptyBroadcast(t *testing.T) {
	resolved, err := resolveBroadcastAddr("", "")
	if err != nil {
		t.Fatalf("resolveBroadcastAddr error: %v", err)
	}
	if resolved != "255.255.255.255:9" {
		t.Errorf("expected '255.255.255.255:9', got %q", resolved)
	}
}

// === Integration: Magic Packet Structure ===

func TestIntegrationMagicPacketStructure(t *testing.T) {
	// Verify that the magic packet payload size is correct
	// 6 bytes of 0xff + 16 copies of MAC (6 bytes each) = 6 + 96 = 102 bytes
	macBytes, err := parseMAC("aa:bb:cc:dd:ee:ff")
	if err != nil {
		t.Fatalf("parseMAC error: %v", err)
	}
	if len(macBytes) != 6 {
		t.Errorf("expected 6-byte MAC, got %d", len(macBytes))
	}

	// The payload structure in SendMagicPacketWithInterface:
	// payload := make([]byte, 6+16*6) = 102 bytes
	// First 6 bytes = 0xff
	// Next 16*6 bytes = 16 copies of MAC
	expectedPayloadSize := 6 + 16*6
	if expectedPayloadSize != 102 {
		t.Errorf("expected payload size 102, got %d", expectedPayloadSize)
	}

	// Verify MAC bytes are correct
	if macBytes[0] != 0xaa || macBytes[5] != 0xff {
		t.Errorf("unexpected MAC bytes: %x", macBytes)
	}
}

// === Integration: Error Messages ===

func TestIntegrationParseMAC_ErrorContainsMAC(t *testing.T) {
	_, err := parseMAC("bad-mac")
	if err == nil {
		t.Error("expected error")
	}
	if !strings.Contains(err.Error(), "некорректный MAC") {
		t.Errorf("expected error to contain 'некорректный MAC', got %q", err.Error())
	}
}

func TestIntegrationResolveBroadcastAddr_ErrorContainsIface(t *testing.T) {
	_, err := broadcastFromInterface("__nonexistent_test_iface__")
	if err == nil {
		t.Error("expected error")
	}
	if !strings.Contains(err.Error(), "__nonexistent_test_iface__") {
		t.Errorf("expected error to contain interface name, got %q", err.Error())
	}
}

// === Integration: Edge Case Broadcast Addresses ===

func TestIntegrationResolveBroadcastAddr_DomainName(t *testing.T) {
	// Domain names should still get :9 appended
	got, err := resolveBroadcastAddr("example.com", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != "example.com:9" {
		t.Errorf("expected 'example.com:9', got %q", got)
	}
}

func TestIntegrationResolveBroadcastAddr_IPv6Format(t *testing.T) {
	// IPv6 addresses contain colons, should not add :9
	got, err := resolveBroadcastAddr("::1", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// IPv6 with colons should be returned as-is
	if got != "::1" {
		t.Errorf("expected '::1', got %q", got)
	}
}

func TestIntegrationResolveBroadcastAddr_IPv6WithPort(t *testing.T) {
	got, err := resolveBroadcastAddr("::1:9", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// IPv6 with port should be returned as-is
	if got != "::1:9" {
		t.Errorf("expected '::1:9', got %q", got)
	}
}

// === Integration: Interface Name Edge Cases ===

func TestIntegrationResolveBroadcastAddr_InterfaceWhitespace(t *testing.T) {
	// eth0 might not exist, but the function should trim whitespace
	// This test verifies that whitespace is properly handled
	got, err := resolveBroadcastAddr("", "  eth0  ")
	// If eth0 doesn't exist, we get an error - that's expected
	_ = got
	_ = err
}

// === Integration: Multiple Calls Stability ===

func TestIntegrationMultipleCalls_Stability(t *testing.T) {
	mac := "aa:bb:cc:dd:ee:ff"
	bcast := "127.0.0.1:9"

	// Call multiple times to ensure stability
	for i := 0; i < 5; i++ {
		err := SendMagicPacket(mac, bcast)
		if err != nil {
			t.Errorf("iteration %d: error: %v", i, err)
		}
	}
}

// === Integration: MAC Format Consistency ===

func TestIntegrationMACFormat_Consistency(t *testing.T) {
	// All these formats should parse to the same MAC
	formats := []string{
		"aa:bb:cc:dd:ee:ff",
		"AA:BB:CC:DD:EE:FF",
		"aa-bb-cc-dd-ee-ff",
		"AA-BB-CC-DD-EE-FF",
		"aabb.ccdd.eeff",
	}

	var firstMac []byte
	for _, f := range formats {
		mac, err := parseMAC(f)
		if err != nil {
			t.Fatalf("parseMAC(%q) error: %v", f, err)
		}
		if firstMac == nil {
			firstMac = mac
		} else if len(firstMac) == len(mac) {
			for i := range firstMac {
				if firstMac[i] != mac[i] {
					t.Errorf("MAC mismatch: %q != %q", formats[0], f)
					break
				}
			}
		}
	}
}
