package scanner

import (
	"testing"
	"time"
)

// ============================================================================
// E7/7.2: настраиваемые UDP-порты.
// ============================================================================

// TestDefaultUDPPorts_ReturnsCopy — вызывающий код не может изменить дефолт.
func TestDefaultUDPPorts_ReturnsCopy(t *testing.T) {
	first := DefaultUDPPorts()
	if len(first) == 0 {
		t.Fatal("DefaultUDPPorts() empty")
	}
	first[0] = 9999

	if second := DefaultUDPPorts(); second[0] == 9999 {
		t.Error("DefaultUDPPorts() must return a copy, not the package slice")
	}
}

// TestNormalizeUDPPorts — очистка, дедупликация и сортировка.
func TestNormalizeUDPPorts(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{"nil-falls-back-to-default", nil, DefaultUDPPorts()},
		{"empty-falls-back-to-default", []int{}, DefaultUDPPorts()},
		{"all-invalid-falls-back-to-default", []int{0, -1, 70000}, DefaultUDPPorts()},
		{"duplicates-removed", []int{161, 53, 161, 53}, []int{53, 161}},
		{"sorted", []int{443, 53, 161}, []int{53, 161, 443}},
		{"boundaries-valid", []int{1, 65535}, []int{1, 65535}},
		{"invalid-mixed-with-valid", []int{0, 53, 99999}, []int{53}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeUDPPorts(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("NormalizeUDPPorts(%v) = %v, want %v", tt.in, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("NormalizeUDPPorts(%v) = %v, want %v", tt.in, got, tt.want)
				}
			}
		})
	}
}

// TestSetUDPPorts_NormalizesAndApplies — сеттер нормализует список.
func TestSetUDPPorts_NormalizesAndApplies(t *testing.T) {
	ns := NewNetworkScanner("192.0.2.0/24", 50*time.Millisecond, "", 10, false)

	// По умолчанию — дефолтный список.
	if got := ns.effectiveUDPPorts(); len(got) != len(DefaultUDPPorts()) {
		t.Fatalf("default UDP ports = %v, want %v", got, DefaultUDPPorts())
	}

	ns.SetUDPPorts([]int{161, 53, 161, 0})
	got := ns.effectiveUDPPorts()
	if len(got) != 2 || got[0] != 53 || got[1] != 161 {
		t.Fatalf("UDP ports = %v, want [53 161]", got)
	}

	// Полностью невалидный список возвращает дефолт (UDP-скан не «пустеет»).
	ns.SetUDPPorts([]int{0, -5})
	if got := ns.effectiveUDPPorts(); len(got) != len(DefaultUDPPorts()) {
		t.Fatalf("invalid list = %v, want defaults", got)
	}
}
