package snmpcollector

import (
	"testing"

	"github.com/gosnmp/gosnmp"
)

// ============================================================================
// bytesToMAC — полное покрытие
// ============================================================================

func TestBytesToMAC_Basic(t *testing.T) {
	b := []byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	got := bytesToMAC(b)
	if got != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("bytesToMAC = %q, want aa:bb:cc:dd:ee:ff", got)
	}
}

func TestBytesToMAC_SingleByte(t *testing.T) {
	got := bytesToMAC([]byte{0xab})
	if got != "ab" {
		t.Errorf("bytesToMAC([0xab]) = %q, want ab", got)
	}
}

func TestBytesToMAC_ZeroBytes(t *testing.T) {
	b := []byte{0, 0, 0, 0, 0, 0}
	got := bytesToMAC(b)
	if got != "00:00:00:00:00:00" {
		t.Errorf("bytesToMAC(zeros) = %q, want 00:00:00:00:00:00", got)
	}
}

// ============================================================================
// pduValueString — полное покрытие
// ============================================================================

func TestPDUValueString_Bytes(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: []byte("test bytes")}
	got := pduValueString(pdu)
	if got != "test bytes" {
		t.Errorf("pduValueString(bytes) = %q, want test bytes", got)
	}
}

func TestPDUValueString_String(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: "test string"}
	got := pduValueString(pdu)
	if got != "test string" {
		t.Errorf("pduValueString(string) = %q, want test string", got)
	}
}

func TestPDUValueString_Int(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: 42}
	got := pduValueString(pdu)
	if got != "42" {
		t.Errorf("pduValueString(int) = %q, want 42", got)
	}
}

func TestPDUValueString_Float(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: 3.14}
	got := pduValueString(pdu)
	if got != "3.14" {
		t.Errorf("pduValueString(float) = %q, want 3.14", got)
	}
}

func TestPDUValueString_Whitespace(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: "  test  "}
	got := pduValueString(pdu)
	if got != "test" {
		t.Errorf("pduValueString(whitespace) = %q, want test", got)
	}
}

func TestPDUValueString_Empty(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: ""}
	got := pduValueString(pdu)
	if got != "" {
		t.Errorf("pduValueString(empty) = %q, want empty", got)
	}
}

func TestPDUValueString_NilBytes(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: []byte(nil)}
	got := pduValueString(pdu)
	// Должно работать без паники
	_ = got
}

// ============================================================================
// lldpChassisToMACString — полное покрытие
// ============================================================================

func TestLLDPChassisToMACString_Bytes6(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: []byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}}
	got := lldpChassisToMACString(pdu)
	if got != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("lldpChassisToMACString(6 bytes) = %q, want aa:bb:cc:dd:ee:ff", got)
	}
}

func TestLLDPChassisToMACString_BytesNot6(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: []byte("test")}
	got := lldpChassisToMACString(pdu)
	if got != "test" {
		t.Errorf("lldpChassisToMACString(4 bytes) = %q, want test", got)
	}
}

func TestLLDPChassisToMACString_String(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: "test-string"}
	got := lldpChassisToMACString(pdu)
	if got != "test:string" {
		t.Errorf("lldpChassisToMACString(string) = %q, want test:string", got)
	}
}

func TestLLDPChassisToMACString_Integer(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: 123}
	got := lldpChassisToMACString(pdu)
	if got != "123" {
		t.Errorf("lldpChassisToMACString(int) = %q, want 123", got)
	}
}

func TestLLDPChassisToMACString_EmptyString(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: ""}
	got := lldpChassisToMACString(pdu)
	if got != "" {
		t.Errorf("lldpChassisToMACString(empty) = %q, want empty", got)
	}
}

func TestLLDPChassisToMACString_WhitespaceString(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: "  test  "}
	got := lldpChassisToMACString(pdu)
	if got != "test" {
		t.Errorf("lldpChassisToMACString(whitespace) = %q, want test", got)
	}
}

func TestLLDPChassisToMACString_DashInString(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: "a-b-c-d-e-f"}
	got := lldpChassisToMACString(pdu)
	if got != "a:b:c:d:e:f" {
		t.Errorf("lldpChassisToMACString(dashes) = %q, want a:b:c:d:e:f", got)
	}
}

func TestLLDPChassisToMACString_UppercaseString(t *testing.T) {
	pdu := gosnmp.SnmpPDU{Value: "AA:BB:CC:DD:EE:FF"}
	got := lldpChassisToMACString(pdu)
	if got != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("lldpChassisToMACString(upper) = %q, want aa:bb:cc:dd:ee:ff", got)
	}
}

// ============================================================================
// Benchmark
// ============================================================================

func BenchmarkBytesToMAC(b *testing.B) {
	data := []byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bytesToMAC(data)
	}
}

func BenchmarkPDUValueString(b *testing.B) {
	pdu := gosnmp.SnmpPDU{Value: "test string"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pduValueString(pdu)
	}
}
