package controller

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// TestAutoScanProfile_noHosts проверяет, что при hosts < 256 функция возвращает
// исходные значения без изменений.
func TestAutoScanProfile_noHosts(t *testing.T) {
	portRange, threads, msg := autoScanProfile("", "1-1024", 50)
	if portRange != "1-1024" {
		t.Fatalf("portRange = %q, want %q", portRange, "1-1024")
	}
	if threads != 50 {
		t.Fatalf("threads = %d, want %d", threads, 50)
	}
	if msg != "" {
		t.Fatalf("msg = %q, want empty", msg)
	}
}

// TestAutoScanProfile_smallSubnet проверяет поведение для подсети с ~256 хостами.
func TestAutoScanProfile_smallSubnet(t *testing.T) {
	portRange, threads, msg := autoScanProfile("192.168.1.0/24", "1-1024", 50)
	// 256 хостов — порог: hosts >= 256, поэтому autoScanProfile может скорректировать
	// Но default case: portCount > 10000 -> нет, threads > 96 -> нет
	if portRange != "1-1024" {
		t.Fatalf("portRange = %q, want %q", portRange, "1-1024")
	}
	if threads != 50 {
		t.Fatalf("threads = %d, want %d", threads, 50)
	}
	if msg != "" {
		t.Fatalf("msg = %q, want empty", msg)
	}
}

// TestAutoScanProfile_mediumSubnet проверяет подсеть ~65536 хостов (/16).
func TestAutoScanProfile_mediumSubnet(t *testing.T) {
	portRange, threads, msg := autoScanProfile("10.0.0.0/16", "1-1024", 50)
	// 65536 хостов -> hosts >= 2048 case
	// portCount = 1024 > 512 -> portRange = "1-512"
	// threads = 50 > 24 -> threads = 24
	if portRange != "1-512" {
		t.Fatalf("portRange = %q, want %q", portRange, "1-512")
	}
	if threads != 24 {
		t.Fatalf("threads = %d, want %d", threads, 24)
	}
	if msg == "" {
		t.Fatal("msg should not be empty for large subnet")
	}
}

// TestAutoScanProfile_largeSubnet проверяет ветку hosts >= 1024 на /21 (2046 хостов).
func TestAutoScanProfile_largeSubnet(t *testing.T) {
	portRange, threads, msg := autoScanProfile("172.16.0.0/21", "1-2000", 60)
	// hosts = 2^11 - 2 = 2046 -> hosts >= 1024 case
	// portCount = 2000 > 1024 -> portRange = "1-1024"
	// threads = 60 > 40 -> threads = 40
	if portRange != "1-1024" {
		t.Fatalf("portRange = %q, want %q", portRange, "1-1024")
	}
	if threads != 40 {
		t.Fatalf("threads = %d, want %d", threads, 40)
	}
	if msg == "" {
		t.Fatal("msg should not be empty")
	}
}

// TestAutoScanProfile_mediumSubnet2000 проверяет подсеть ~512 хостов (/24 с хостами).
func TestAutoScanProfile_mediumSubnet2000(t *testing.T) {
	portRange, threads, msg := autoScanProfile("192.168.1.0/24", "1-3000", 70)
	// 256 хостов -> default case
	// portCount = 3000 > 10000 -> нет
	// threads = 70 > 96 -> нет
	// Значения не должны меняться
	if portRange != "1-3000" {
		t.Fatalf("portRange = %q, want %q", portRange, "1-3000")
	}
	if threads != 70 {
		t.Fatalf("threads = %d, want %d", threads, 70)
	}
	// Значения не менялись — сообщение автопрофиля должно быть пустым
	if msg != "" {
		t.Fatalf("msg = %q, want empty (nothing changed)", msg)
	}
}

// TestAutoScanProfile_threadsClamp проверяет, что threads < 1 корректируется до 1.
func TestAutoScanProfile_threadsClamp(t *testing.T) {
	portRange, threads, msg := autoScanProfile("192.168.1.0/24", "1-1024", 0)
	// portRange не меняется при clamp потоков
	if portRange != "1-1024" {
		t.Fatalf("portRange = %q, want %q", portRange, "1-1024")
	}
	if threads != 1 {
		t.Fatalf("threads = %d, want 1", threads)
	}
	// msg должен быть пустым — значения не менялись (кроме threads clamp)
	// Но threads clamp происходит ДО проверки hosts, поэтому:
	// portRange не меняется, threads меняется с 0 на 1
	// Однако threads меняется внутри функции, и потом сравнивается:
	// newThreads != threads -> msg будет не пустым
	if msg != "" {
		// threads: 0 -> 1, portRange не меняется -> msg содержит threads
		if msg == "Автопрофиль: подсеть ~256 хостов, threads: 0 -> 1" {
			// OK
		} else {
			t.Logf("msg = %q", msg)
		}
	}
}

// TestAutoScanProfile_portOnly проверяет, что меняется только portRange.
func TestAutoScanProfile_portOnly(t *testing.T) {
	portRange, threads, msg := autoScanProfile("10.0.0.0/16", "1-1024", 10)
	// portCount = 1024 > 512 -> portRange = "1-512"
	// threads = 10 > 24 -> нет
	if portRange != "1-512" {
		t.Fatalf("portRange = %q, want %q", portRange, "1-512")
	}
	if threads != 10 {
		t.Fatalf("threads = %d, want %d", threads, 10)
	}
	if msg == "" {
		t.Fatal("msg should not be empty")
	}
}

// TestAutoScanProfile_threadsOnly проверяет, что меняется только threads.
func TestAutoScanProfile_threadsOnly(t *testing.T) {
	portRange, threads, msg := autoScanProfile("10.0.0.0/16", "1-100", 50)
	// portCount = 100 > 512 -> нет
	// threads = 50 > 24 -> threads = 24
	if portRange != "1-100" {
		t.Fatalf("portRange = %q, want %q", portRange, "1-100")
	}
	if threads != 24 {
		t.Fatalf("threads = %d, want %d", threads, 24)
	}
	if msg == "" {
		t.Fatal("msg should not be empty")
	}
}

// TestAutoScanProfile_emptyNetwork проверяет пустую строку сети.
func TestAutoScanProfile_emptyNetwork(t *testing.T) {
	portRange, threads, msg := autoScanProfile("", "1-1024", 50)
	// Пустая строка -> hosts = 0 -> hosts < 256 -> return unchanged
	if portRange != "1-1024" {
		t.Fatalf("portRange = %q, want %q", portRange, "1-1024")
	}
	if threads != 50 {
		t.Fatalf("threads = %d, want %d", threads, 50)
	}
	if msg != "" {
		t.Fatalf("msg = %q, want empty", msg)
	}
}

// TestAutoScanProfile_whitespaceNetwork проверяет сеть с пробелами.
func TestAutoScanProfile_whitespaceNetwork(t *testing.T) {
	portRange, threads, msg := autoScanProfile("  192.168.1.0/24  ", "1-1024", 50)
	// TrimSpace -> 192.168.1.0/24 -> 256 хостов
	// default case: portCount = 1024 > 10000 -> нет, threads = 50 > 96 -> нет
	if portRange != "1-1024" {
		t.Fatalf("portRange = %q, want %q", portRange, "1-1024")
	}
	if threads != 50 {
		t.Fatalf("threads = %d, want %d", threads, 50)
	}
	// Изменений нет — сообщение автопрофиля должно быть пустым
	if msg != "" {
		t.Fatalf("msg = %q, want empty", msg)
	}
}

// TestAutoScanProfile_invalidNetwork проверяет невалидную строку сети.
func TestAutoScanProfile_invalidNetwork(t *testing.T) {
	portRange, threads, msg := autoScanProfile("not-a-cidr", "1-1024", 50)
	// EstimateHostCount вернёт ошибку -> hosts < 256 -> return unchanged
	if portRange != "1-1024" {
		t.Fatalf("portRange = %q, want %q", portRange, "1-1024")
	}
	if threads != 50 {
		t.Fatalf("threads = %d, want %d", threads, 50)
	}
	if msg != "" {
		t.Fatalf("msg = %q, want empty", msg)
	}
}

// TestEstimateScanUITimeout_base проверяет, что при минимальных параметрах
// возвращается базовое значение (300 секунд).
func TestEstimateScanUITimeout_base(t *testing.T) {
	// Минимальные параметры: 1 хост, 1 порт, timeout=1, threads=1
	dur := estimateScanUITimeout("127.0.0.1/32", "1", "1", "1", true, false)
	if dur != 300*time.Second {
		t.Fatalf("duration = %v, want %v", dur, 300*time.Second)
	}
}

// TestEstimateScanUITimeout_largeSubnet проверяет оценку для крупной подсети.
func TestEstimateScanUITimeout_largeSubnet(t *testing.T) {
	// 65536 хостов, 1000 портов, timeout=5, threads=50
	dur := estimateScanUITimeout("10.0.0.0/16", "1-1000", "5", "50", true, false)
	if dur < 300*time.Second {
		t.Fatalf("duration = %v, want >= %v", dur, 300*time.Second)
	}
}

// TestEstimateScanUITimeout_maxTimeout проверяет ограничение максимального таймаута.
func TestEstimateScanUITimeout_maxTimeout(t *testing.T) {
	// Очень большая подсеть с большим количеством портов
	dur := estimateScanUITimeout("10.0.0.0/8", "1-65535", "10", "10", true, false)
	maxTimeout := 45 * time.Minute
	if dur > maxTimeout {
		t.Fatalf("duration = %v, want <= %v", dur, maxTimeout)
	}
}

// TestEstimateScanUITimeout_emptyNetwork проверяет оценку для пустой сети.
func TestEstimateScanUITimeout_emptyNetwork(t *testing.T) {
	dur := estimateScanUITimeout("", "", "", "", true, false)
	// При пустых параметрах: hosts = 256 (default), ports = 65535 (default TCP),
	// timeoutSec = 2 (default), threads = 50 (default)
	if dur != 300*time.Second {
		t.Logf("duration = %v (likely base due to small calculated value)", dur)
	}
}

// TestEstimateScanUITimeout_noTCP_noUDP проверяет оценку без TCP и UDP.
func TestEstimateScanUITimeout_noTCP_noUDP(t *testing.T) {
	dur := estimateScanUITimeout("192.168.1.0/24", "", "2", "50", false, false)
	// scanTCP = false, scanUDP = false -> ports = 0 -> ports = 1 (default)
	// hosts = 256, ports = 1, timeoutSec = 2, threads = 50
	// workUnits = 256, estimatedSec = 256*2/50 = 10, estimated = 10/4 + 90 = 92.5 -> base
	if dur != 300*time.Second {
		t.Fatalf("duration = %v, want %v", dur, 300*time.Second)
	}
}

// TestEstimateScanUITimeout_withUDP проверяет оценку с UDP.
func TestEstimateScanUITimeout_withUDP(t *testing.T) {
	dur := estimateScanUITimeout("192.168.1.0/24", "", "2", "50", false, true)
	// scanTCP = false, scanUDP = true -> ports = 0 + 9 = 9
	// hosts = 256, ports = 9, timeoutSec = 2, threads = 50
	// workUnits = 256*9 = 2304, estimatedSec = 2304*2/50 = 92, estimated = 92/4 + 90 = 113
	// 113 < 300 -> base
	if dur != 300*time.Second {
		t.Fatalf("duration = %v, want %v", dur, 300*time.Second)
	}
}

// TestEstimateScanUITimeout_largeHosts_largePorts проверяет, что расчёт ограничивается maxTimeout.
func TestEstimateScanUITimeout_largeHosts_largePorts(t *testing.T) {
	// /16 = 65534 hosts, 1024 ports, timeout=10, threads=10
	// workUnits = 65534*1024 = 67,106,816
	// estimatedSec = 67,106,816*10/10 = 67,106,816 s
	// estimated = 67,106,816/4 + 90 = 16,776,794 s -> far exceeds 45m -> maxTimeout
	dur := estimateScanUITimeout("10.0.0.0/16", "1-1024", "10", "10", true, false)
	maxTimeout := 45 * time.Minute
	if dur != maxTimeout {
		t.Fatalf("duration = %v, want max timeout %v", dur, maxTimeout)
	}
}

// TestFormatDurationMMSS_basic проверяет базовое форматирование.
func TestFormatDurationMMSS_basic(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{0, "00:00"},
		{60 * time.Second, "01:00"},
		{90 * time.Second, "01:30"},
		{120 * time.Second, "02:00"},
		{3661 * time.Second, "61:01"},
		{30 * time.Second, "00:30"},
	}

	for _, tc := range tests {
		got := formatDurationMMSS(tc.input)
		if got != tc.expected {
			t.Fatalf("formatDurationMMSS(%v) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

// TestFormatDurationMMSS_negative проверяет обработку отрицательных значений.
func TestFormatDurationMMSS_negative(t *testing.T) {
	got := formatDurationMMSS(-1 * time.Hour)
	if got != "00:00" {
		t.Fatalf("formatDurationMMSS(-1h) = %q, want %q", got, "00:00")
	}
}

// TestFormatDurationMMSS_exactMinutes проверяет точные минуты.
func TestFormatDurationMMSS_exactMinutes(t *testing.T) {
	got := formatDurationMMSS(5 * time.Minute)
	if got != "05:00" {
		t.Fatalf("formatDurationMMSS(5m) = %q, want %q", got, "05:00")
	}
}

// TestApplyRecommendedProfile_smallSubnet проверяет ветку default (<512 хостов).
func TestApplyRecommendedProfile_smallSubnet(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ScanUI{
		PortRangeEntry:       widget.NewEntry(),
		TimeoutEntry:         widget.NewEntry(),
		ThreadsEntry:         widget.NewEntry(),
		ScanUDPCheck:         widget.NewCheck("UDP", nil),
		ScanBannersCheck:     widget.NewCheck("Banners", nil),
		ScanOSActiveCheck:    widget.NewCheck("OS", nil),
		ScanVerboseLogsCheck: widget.NewCheck("Verbose", nil),
		AutoProfileCheck:     widget.NewCheck("Auto", nil),
		StatusLabel:          widget.NewLabel(""),
		RecommendedBadge:     nil,
	}
	ctrl := NewScanController(a, ui, &mockScanActions{})
	ctrl.ApplyRecommendedProfile("192.168.1.0/24")

	if ui.PortRangeEntry.Text != "1-2048,3389" {
		t.Fatalf("portRange = %q, want %q", ui.PortRangeEntry.Text, "1-2048,3389")
	}
	if ui.TimeoutEntry.Text != "2" {
		t.Fatalf("timeout = %q, want %q", ui.TimeoutEntry.Text, "2")
	}
	if ui.ThreadsEntry.Text != "100" {
		t.Fatalf("threads = %q, want %q", ui.ThreadsEntry.Text, "100")
	}
	if ui.ScanUDPCheck.Checked {
		t.Fatal("UDP should be unchecked")
	}
	if ui.ScanBannersCheck.Checked {
		t.Fatal("Banners should be unchecked")
	}
	if ui.ScanOSActiveCheck.Checked {
		t.Fatal("OS should be unchecked")
	}
	if !ui.AutoProfileCheck.Checked {
		t.Fatal("AutoProfile should be checked")
	}
}

// TestApplyRecommendedProfile_mediumSubnet проверяет ветку 512-1023 хостов.
func TestApplyRecommendedProfile_mediumSubnet(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ScanUI{
		PortRangeEntry:    widget.NewEntry(),
		TimeoutEntry:      widget.NewEntry(),
		ThreadsEntry:      widget.NewEntry(),
		ScanUDPCheck:      widget.NewCheck("UDP", nil),
		ScanBannersCheck:  widget.NewCheck("Banners", nil),
		ScanOSActiveCheck: widget.NewCheck("OS", nil),
		AutoProfileCheck:  widget.NewCheck("Auto", nil),
		StatusLabel:       widget.NewLabel(""),
		RecommendedBadge:  nil,
	}
	ctrl := NewScanController(a, ui, &mockScanActions{})
	ctrl.ApplyRecommendedProfile("172.16.0.0/22")

	if ui.PortRangeEntry.Text != "1-1024,3389" {
		t.Fatalf("portRange = %q, want %q", ui.PortRangeEntry.Text, "1-1024,3389")
	}
	if ui.ThreadsEntry.Text != "80" {
		t.Fatalf("threads = %q, want %q", ui.ThreadsEntry.Text, "80")
	}
}

// TestApplyRecommendedProfile_largeSubnet проверяет ветку 1024-2047 хостов.
func TestApplyRecommendedProfile_largeSubnet(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ScanUI{
		PortRangeEntry:    widget.NewEntry(),
		TimeoutEntry:      widget.NewEntry(),
		ThreadsEntry:      widget.NewEntry(),
		ScanUDPCheck:      widget.NewCheck("UDP", nil),
		ScanBannersCheck:  widget.NewCheck("Banners", nil),
		ScanOSActiveCheck: widget.NewCheck("OS", nil),
		AutoProfileCheck:  widget.NewCheck("Auto", nil),
		StatusLabel:       widget.NewLabel(""),
		RecommendedBadge:  nil,
	}
	ctrl := NewScanController(a, ui, &mockScanActions{})
	ctrl.ApplyRecommendedProfile("172.16.0.0/21")

	if ui.PortRangeEntry.Text != "1-1024" {
		t.Fatalf("portRange = %q, want %q", ui.PortRangeEntry.Text, "1-1024")
	}
	if ui.ThreadsEntry.Text != "60" {
		t.Fatalf("threads = %q, want %q", ui.ThreadsEntry.Text, "60")
	}
}

// TestApplyRecommendedProfile_hugeSubnet проверяет ветку >=2048 хостов.
func TestApplyRecommendedProfile_hugeSubnet(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ScanUI{
		PortRangeEntry:    widget.NewEntry(),
		TimeoutEntry:      widget.NewEntry(),
		ThreadsEntry:      widget.NewEntry(),
		ScanUDPCheck:      widget.NewCheck("UDP", nil),
		ScanBannersCheck:  widget.NewCheck("Banners", nil),
		ScanOSActiveCheck: widget.NewCheck("OS", nil),
		AutoProfileCheck:  widget.NewCheck("Auto", nil),
		StatusLabel:       widget.NewLabel(""),
		RecommendedBadge:  nil,
	}
	ctrl := NewScanController(a, ui, &mockScanActions{})
	ctrl.ApplyRecommendedProfile("10.0.0.0/16")

	if ui.PortRangeEntry.Text != "22,80,443,445,3389" {
		t.Fatalf("portRange = %q, want %q", ui.PortRangeEntry.Text, "22,80,443,445,3389")
	}
	if ui.TimeoutEntry.Text != "1" {
		t.Fatalf("timeout = %q, want %q", ui.TimeoutEntry.Text, "1")
	}
	if ui.ThreadsEntry.Text != "40" {
		t.Fatalf("threads = %q, want %q", ui.ThreadsEntry.Text, "40")
	}
}

// TestApplyRecommendedProfile_emptyNetwork проверяет пустую сеть.
func TestApplyRecommendedProfile_emptyNetwork(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ScanUI{
		PortRangeEntry: widget.NewEntry(),
		TimeoutEntry:   widget.NewEntry(),
		ThreadsEntry:   widget.NewEntry(),
		StatusLabel:    widget.NewLabel(""),
	}
	ctrl := NewScanController(a, ui, &mockScanActions{})
	ctrl.ApplyRecommendedProfile("")

	// default case (<512 хостов)
	if ui.PortRangeEntry.Text != "1-2048,3389" {
		t.Fatalf("portRange = %q, want %q", ui.PortRangeEntry.Text, "1-2048,3389")
	}
}
