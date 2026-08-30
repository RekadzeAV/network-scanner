package gui

import (
	"testing"

	"network-scanner/internal/scanner"
)

// --- results_view.go tests ---

func TestCountOpenPorts_Empty(t *testing.T) {
	n := countOpenPorts(nil)
	if n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}

func TestCountOpenPorts_EmptySlice(t *testing.T) {
	n := countOpenPorts([]scanner.PortInfo{})
	if n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}

func TestCountOpenPorts_NoOpen(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 22, State: "closed"},
		{Port: 80, State: "filtered"},
	}
	n := countOpenPorts(ports)
	if n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}

func TestCountOpenPorts_SingleOpen(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 80, State: "open"},
	}
	n := countOpenPorts(ports)
	if n != 1 {
		t.Errorf("expected 1, got %d", n)
	}
}

func TestCountOpenPorts_MultipleOpen(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 80, State: "open"},
		{Port: 443, State: "open"},
		{Port: 22, State: "closed"},
		{Port: 8080, State: "open"},
	}
	n := countOpenPorts(ports)
	if n != 3 {
		t.Errorf("expected 3, got %d", n)
	}
}

func TestCountOpenPorts_CaseInsensitive(t *testing.T) {
	ports := []scanner.PortInfo{
		{Port: 80, State: "Open"},
		{Port: 443, State: "OPEN"},
	}
	n := countOpenPorts(ports)
	if n != 2 {
		t.Errorf("expected 2 (case-insensitive), got %d", n)
	}
}

func TestTruncateStr_Empty(t *testing.T) {
	result := truncateStr("", 10)
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestTruncateStr_Shorter(t *testing.T) {
	result := truncateStr("hello", 10)
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestTruncateStr_ExactLength(t *testing.T) {
	result := truncateStr("hello", 5)
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestTruncateStr_Longer(t *testing.T) {
	result := truncateStr("hello world", 8)
	if result != "hello..." {
		t.Errorf("expected 'hello...', got %q", result)
	}
}

func TestTruncateStr_MaxLen3(t *testing.T) {
	result := truncateStr("hello", 3)
	if result != "hel" {
		t.Errorf("expected 'hel', got %q", result)
	}
}

func TestTruncateStr_MaxLen2(t *testing.T) {
	result := truncateStr("hello", 2)
	if result != "he" {
		t.Errorf("expected 'he', got %q", result)
	}
}

func TestTruncateStr_MaxLen0(t *testing.T) {
	result := truncateStr("hello", 0)
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestTruncateStr_WhitespaceTrimmed(t *testing.T) {
	result := truncateStr("  hello  ", 8)
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestClampFloat32_BelowMinimum(t *testing.T) {
	result := clampFloat32(0.1, 0.5, 1.0)
	if result != 0.5 {
		t.Errorf("expected 0.5, got %v", result)
	}
}

func TestClampFloat32_AboveMaximum(t *testing.T) {
	result := clampFloat32(1.5, 0.5, 1.0)
	if result != 1.0 {
		t.Errorf("expected 1.0, got %v", result)
	}
}

func TestClampFloat32_InRange(t *testing.T) {
	result := clampFloat32(0.75, 0.5, 1.0)
	if result != 0.75 {
		t.Errorf("expected 0.75, got %v", result)
	}
}

func TestClampFloat32_EdgeMinimum(t *testing.T) {
	result := clampFloat32(0.5, 0.5, 1.0)
	if result != 0.5 {
		t.Errorf("expected 0.5, got %v", result)
	}
}

func TestClampFloat32_EdgeMaximum(t *testing.T) {
	result := clampFloat32(1.0, 0.5, 1.0)
	if result != 1.0 {
		t.Errorf("expected 1.0, got %v", result)
	}
}

func TestClampFloat32_Negative(t *testing.T) {
	result := clampFloat32(-1.0, -0.5, 0.5)
	if result != -0.5 {
		t.Errorf("expected -0.5, got %v", result)
	}
}

func TestClampFloat64_BelowMinimum(t *testing.T) {
	result := clampFloat64(0.1, 0.5, 1.0)
	if result != 0.5 {
		t.Errorf("expected 0.5, got %v", result)
	}
}

func TestClampFloat64_AboveMaximum(t *testing.T) {
	result := clampFloat64(1.5, 0.5, 1.0)
	if result != 1.0 {
		t.Errorf("expected 1.0, got %v", result)
	}
}

func TestClampFloat64_InRange(t *testing.T) {
	result := clampFloat64(0.75, 0.5, 1.0)
	if result != 0.75 {
		t.Errorf("expected 0.75, got %v", result)
	}
}

func TestAbsFloat32_Positive(t *testing.T) {
	result := absFloat32(5.0)
	if result != 5.0 {
		t.Errorf("expected 5.0, got %v", result)
	}
}

func TestAbsFloat32_Negative(t *testing.T) {
	result := absFloat32(-5.0)
	if result != 5.0 {
		t.Errorf("expected 5.0, got %v", result)
	}
}

func TestAbsFloat32_Zero(t *testing.T) {
	result := absFloat32(0.0)
	if result != 0.0 {
		t.Errorf("expected 0.0, got %v", result)
	}
}

func TestAbsFloat32_SmallNegative(t *testing.T) {
	result := absFloat32(-0.001)
	if result != 0.001 {
		t.Errorf("expected 0.001, got %v", result)
	}
}

func TestLayoutAdaptiveMultiplier_ZeroSize(t *testing.T) {
	result := layoutAdaptiveMultiplier(0, 0, 1.0)
	if result != 1 {
		t.Errorf("expected 1 for zero size, got %v", result)
	}
}

func TestLayoutAdaptiveMultiplier_ZeroScale(t *testing.T) {
	result := layoutAdaptiveMultiplier(800, 600, 0)
	if result <= 0 {
		t.Errorf("expected positive multiplier, got %v", result)
	}
}

func TestLayoutAdaptiveMultiplier_NormalSize(t *testing.T) {
	result := layoutAdaptiveMultiplier(1920, 1080, 1.0)
	if result <= 0 {
		t.Errorf("expected positive multiplier, got %v", result)
	}
	if result > 2 {
		t.Errorf("expected reasonable multiplier, got %v", result)
	}
}

func TestSuggestedScanTabOffset_ZeroHeight(t *testing.T) {
	result := suggestedScanTabOffset("normal", 800, 0, 1.0, 1.0)
	if result != 0.38 {
		t.Errorf("expected 0.38 for zero height, got %v", result)
	}
}

func TestSuggestedScanTabOffset_CompactProfile(t *testing.T) {
	result := suggestedScanTabOffset("compact", 800, 800, 1.0, 1.0)
	if result < 0.26 || result > 0.54 {
		t.Errorf("expected value in [0.26, 0.54], got %v", result)
	}
}

func TestSuggestedScanTabOffset_WideProfile(t *testing.T) {
	result := suggestedScanTabOffset("wide", 800, 800, 1.0, 1.0)
	if result < 0.26 || result > 0.54 {
		t.Errorf("expected value in [0.26, 0.54], got %v", result)
	}
}

func TestSuggestedScanTabOffset_TallScreen(t *testing.T) {
	result := suggestedScanTabOffset("normal", 800, 1080, 1.0, 1.0)
	if result < 0.26 || result > 0.54 {
		t.Errorf("expected value in [0.26, 0.54], got %v", result)
	}
}

func TestSuggestedScanTabOffset_ShortScreen(t *testing.T) {
	result := suggestedScanTabOffset("normal", 800, 600, 1.0, 1.0)
	if result < 0.26 || result > 0.54 {
		t.Errorf("expected value in [0.26, 0.54], got %v", result)
	}
}

func TestDefaultTopologySplitOffset_Normal(t *testing.T) {
	result := defaultTopologySplitOffset("normal")
	if result != 0.62 {
		t.Errorf("expected 0.62, got %v", result)
	}
}

func TestDefaultTopologySplitOffset_Compact(t *testing.T) {
	result := defaultTopologySplitOffset("compact")
	if result != 0.72 {
		t.Errorf("expected 0.72, got %v", result)
	}
}

func TestDefaultTopologySplitOffset_Wide(t *testing.T) {
	result := defaultTopologySplitOffset("wide")
	if result != 0.6 {
		t.Errorf("expected 0.6, got %v", result)
	}
}

func TestDefaultTopologySplitOffset_Unknown(t *testing.T) {
	result := defaultTopologySplitOffset("unknown")
	if result != 0.62 {
		t.Errorf("expected 0.62 (default), got %v", result)
	}
}

func TestDefaultToolsSplitOffset_Normal(t *testing.T) {
	result := defaultToolsSplitOffset("normal")
	if result != 0.44 {
		t.Errorf("expected 0.44, got %v", result)
	}
}

func TestDefaultToolsSplitOffset_Compact(t *testing.T) {
	result := defaultToolsSplitOffset("compact")
	if result != 0.48 {
		t.Errorf("expected 0.48, got %v", result)
	}
}

func TestDefaultToolsSplitOffset_Wide(t *testing.T) {
	result := defaultToolsSplitOffset("wide")
	if result != 0.40 {
		t.Errorf("expected 0.40, got %v", result)
	}
}

func TestDefaultToolsSplitOffset_Unknown(t *testing.T) {
	result := defaultToolsSplitOffset("unknown")
	if result != 0.44 {
		t.Errorf("expected 0.44 (default), got %v", result)
	}
}

func TestNullDash_Empty(t *testing.T) {
	result := nullDash("")
	if result != "-" {
		t.Errorf("expected '-', got %q", result)
	}
}

func TestNullDash_NonEmpty(t *testing.T) {
	result := nullDash("router")
	if result != "router" {
		t.Errorf("expected 'router', got %q", result)
	}
}

func TestNullDash_WhitespaceOnly(t *testing.T) {
	result := nullDash("   ")
	if result != "-" {
		t.Errorf("expected '-', got %q", result)
	}
}

func TestDeviceTypeWithBadge_Empty(t *testing.T) {
	result := deviceTypeWithBadge("")
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestDeviceTypeWithBadge_NonEmpty(t *testing.T) {
	result := deviceTypeWithBadge("Router")
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestDeviceTypeWithBadge_Unknown(t *testing.T) {
	result := deviceTypeWithBadge("Unknown")
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestOsGuessLine_Empty(t *testing.T) {
	result := osGuessLine(scanner.Result{})
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestOsGuessLine_WithOS(t *testing.T) {
	result := osGuessLine(scanner.Result{
		GuessOS: "Linux",
	})
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestOsGuessLine_EmptyOS(t *testing.T) {
	result := osGuessLine(scanner.Result{
		GuessOS: "",
	})
	if result == "" {
		t.Error("expected non-empty result")
	}
}
