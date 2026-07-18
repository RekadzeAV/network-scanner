package wol

import (
	"net"
	"testing"
)

// ============================================================================
// parseMAC — edge cases
// ============================================================================

func TestParseMAC_Whitespace(t *testing.T) {
	mac, err := parseMAC("  aa:bb:cc:dd:ee:ff  ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(mac) != 6 {
		t.Fatalf("expected 6 bytes, got %d", len(mac))
	}
}

func TestParseMAC_DashFormat(t *testing.T) {
	mac, err := parseMAC("aa-bb-cc-dd-ee-ff")
	if err != nil {
		t.Fatalf("expected no error for dash format, got %v", err)
	}
	if mac[0] != 0xaa || mac[5] != 0xff {
		t.Fatalf("unexpected MAC bytes: %x", mac)
	}
}

func TestParseMAC_Empty(t *testing.T) {
	_, err := parseMAC("")
	if err == nil {
		t.Fatal("expected error for empty MAC")
	}
}

func TestParseMAC_InvalidFormat(t *testing.T) {
	_, err := parseMAC("zz:zz:zz:zz:zz:zz")
	if err == nil {
		t.Fatal("expected error for invalid MAC")
	}
}

// ============================================================================
// resolveBroadcastAddr — edge cases
// ============================================================================

func TestResolveBroadcastAddr_WhitespaceBcast(t *testing.T) {
	got, err := resolveBroadcastAddr("  ", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != "255.255.255.255:9" {
		t.Fatalf("expected default broadcast, got %q", got)
	}
}

func TestResolveBroadcastAddr_ExplicitWithPort(t *testing.T) {
	got, err := resolveBroadcastAddr("10.0.0.255:7", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != "10.0.0.255:7" {
		t.Fatalf("expected 10.0.0.255:7, got %q", got)
	}
}

func TestResolveBroadcastAddr_EmptyIfaceName(t *testing.T) {
	got, err := resolveBroadcastAddr("", "   ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != "255.255.255.255:9" {
		t.Fatalf("expected default broadcast, got %q", got)
	}
}

// ============================================================================
// broadcastFromInterface — edge cases
// ============================================================================

func TestBroadcastFromInterface_EmptyName(t *testing.T) {
	_, err := broadcastFromInterface("")
	if err == nil {
		t.Fatal("expected error for empty interface name")
	}
}

func TestBroadcastFromInterface_WhitespaceName(t *testing.T) {
	_, err := broadcastFromInterface("   ")
	if err == nil {
		t.Fatal("expected error for whitespace interface name")
	}
}

func TestBroadcastFromInterface_NotFound(t *testing.T) {
	_, err := broadcastFromInterface("__nonexistent_iface__")
	if err == nil {
		t.Fatal("expected error for nonexistent interface")
	}
}

// ============================================================================
// SendMagicPacket — wrapper test
// ============================================================================

func TestSendMagicPacket_InvalidMAC(t *testing.T) {
	err := SendMagicPacket("bad-mac", "255.255.255.255:9")
	if err == nil {
		t.Fatal("expected error for invalid MAC")
	}
}

func TestSendMagicPacketWithInterface_InvalidBcast(t *testing.T) {
	// Valid MAC but invalid broadcast address
	_, err := SendMagicPacketWithInterface("aa:bb:cc:dd:ee:ff", "!!!invalid!!!:99999", "")
	if err == nil {
		t.Fatal("expected error for invalid broadcast address")
	}
}

func TestSendMagicPacketWithInterface_ValidLocal(t *testing.T) {
	// Send to localhost — should succeed (UDP send doesn't require a listener)
	addr, err := SendMagicPacketWithInterface("aa:bb:cc:dd:ee:ff", "127.0.0.1:9", "")
	if err != nil {
		t.Fatalf("expected no error sending to localhost, got %v", err)
	}
	if addr != "127.0.0.1:9" {
		t.Fatalf("expected 127.0.0.1:9, got %q", addr)
	}
}

func TestSendMagicPacket_ValidLocal(t *testing.T) {
	err := SendMagicPacket("aa:bb:cc:dd:ee:ff", "127.0.0.1:9")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// ============================================================================
// broadcastFromInterface — 27.3% → higher
// ============================================================================

func TestBroadcastFromInterface_RealInterface(t *testing.T) {
	ifaces, err := net.Interfaces()
	if err != nil {
		t.Skip("cannot list network interfaces")
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		hasIPv4 := false
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				hasIPv4 = true
				break
			}
		}
		if !hasIPv4 {
			continue
		}
		bcast, err := broadcastFromInterface(iface.Name)
		if err != nil {
			t.Fatalf("expected no error for interface %q, got %v", iface.Name, err)
		}
		if bcast == "" {
			t.Fatalf("expected non-empty broadcast for interface %q", iface.Name)
		}
		return
	}
	t.Skip("no suitable IPv4 interface found")
}

func TestBroadcastFromInterface_NoIPv4(t *testing.T) {
	// Find an interface that only has IPv6 (or no addresses)
	ifaces, err := net.Interfaces()
	if err != nil {
		t.Skip("cannot list network interfaces")
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		hasIPv4 := false
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				hasIPv4 = true
				break
			}
		}
		if !hasIPv4 && len(addrs) > 0 {
			_, err := broadcastFromInterface(iface.Name)
			if err == nil {
				t.Fatalf("expected error for interface without IPv4: %q", iface.Name)
			}
			return
		}
	}
	t.Skip("no IPv6-only interface found")
}
