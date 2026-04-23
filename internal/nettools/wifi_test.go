package nettools

import (
	"strings"
	"testing"
)

func TestParseWindowsNetsh(t *testing.T) {
	raw := `
Name                   : Wi-Fi
State                  : connected
SSID                   : HomeWiFi
BSSID                  : 01:23:45:67:89:ab
Radio type             : 802.11ax
Authentication         : WPA2-Personal
Channel                : 6
Receive rate (Mbps)    : 120
Transmit rate (Mbps)   : 60
Signal                 : 87%
`
	s := parseWindowsNetsh(raw)
	if s["ssid"] != "HomeWiFi" {
		t.Fatalf("unexpected ssid: %q", s["ssid"])
	}
	if s["state"] != "connected" {
		t.Fatalf("unexpected state: %q", s["state"])
	}
}

func TestParseWindowsNetshRussianLocale(t *testing.T) {
	raw := `
Имя                    : Wi-Fi
Состояние              : подключено
SSID                   : ДомашняяСеть
BSSID                  : 01:23:45:67:89:ab
Тип радиомодуля        : 802.11ax
Проверка подлинности   : WPA2-Personal
Канал                  : 6
Скорость приема (Мбит/с) : 120
Скорость передачи (Мбит/с): 60
Сигнал                 : 87%
`
	s := parseWindowsNetsh(raw)
	if s["interface"] != "Wi-Fi" {
		t.Fatalf("unexpected interface: %q", s["interface"])
	}
	if s["ssid"] != "ДомашняяСеть" {
		t.Fatalf("unexpected ssid: %q", s["ssid"])
	}
	if s["state"] != "connected" {
		t.Fatalf("unexpected state: %q", s["state"])
	}
	if s["security"] != "WPA2-Personal" {
		t.Fatalf("unexpected security: %q", s["security"])
	}
	if s["mode"] != "802.11ax" {
		t.Fatalf("unexpected mode: %q", s["mode"])
	}
	if s["rate"] != "120 Mbps / 60 Mbps" {
		t.Fatalf("unexpected rate: %q", s["rate"])
	}
}

func TestParseWindowsNetshDisconnectedState(t *testing.T) {
	raw := `
Name                   : Wi-Fi
State                  : disconnected
`
	s := parseWindowsNetsh(raw)
	if s["interface"] != "Wi-Fi" {
		t.Fatalf("unexpected interface: %q", s["interface"])
	}
	if s["state"] != "disconnected" {
		t.Fatalf("unexpected normalized state: %q", s["state"])
	}
}

func TestParseWindowsNetshUnknownStateFallback(t *testing.T) {
	raw := `
Name                   : Wi-Fi
`
	s := parseWindowsNetsh(raw)
	if s["state"] != "unknown" {
		t.Fatalf("unexpected fallback state: %q", s["state"])
	}
}

func TestParseLinuxNmcli(t *testing.T) {
	raw := `
*:HomeWiFi:Infra:11:130 Mbit/s:78:WPA2
 :OtherWiFi:Infra:1:65 Mbit/s:40:WPA1
`
	s := parseLinuxNmcli(raw)
	if s["ssid"] != "HomeWiFi" {
		t.Fatalf("unexpected ssid: %q", s["ssid"])
	}
	if s["state"] != "connected" {
		t.Fatalf("unexpected state: %q", s["state"])
	}
}

func TestParseLinuxNmcliEscapedColonInSSID(t *testing.T) {
	raw := `
*:Office\:Guest:Infra:36:540 Mbit/s:91:WPA2
`
	s := parseLinuxNmcli(raw)
	if s["ssid"] != "Office:Guest" {
		t.Fatalf("unexpected escaped ssid parse: %q", s["ssid"])
	}
	if s["state"] != "connected" {
		t.Fatalf("unexpected state: %q", s["state"])
	}
}

func TestSplitNmcliFieldsEscapedSeparators(t *testing.T) {
	line := `*:A\:B:Infra:1:65 Mbit/s:40:WPA2`
	parts := splitNmcliFields(line)
	if len(parts) != 7 {
		t.Fatalf("unexpected fields count: %d", len(parts))
	}
	if parts[1] != "A:B" {
		t.Fatalf("unexpected ssid field: %q", parts[1])
	}
}

func TestParseDarwinAirport(t *testing.T) {
	raw := `
SSID: HomeWiFi
BSSID: 01:23:45:67:89:ab
agrCtlRSSI: -51
lastTxRate: 351
link auth: wpa2-psk
channel: 44
`
	s := parseDarwinAirport(raw)
	if s["ssid"] != "HomeWiFi" {
		t.Fatalf("unexpected ssid: %q", s["ssid"])
	}
	if !strings.Contains(s["signal"], "dBm") {
		t.Fatalf("unexpected signal: %q", s["signal"])
	}
	if s["state"] != "connected" {
		t.Fatalf("unexpected state: %q", s["state"])
	}
}

func TestParseDarwinAirportDisconnectedState(t *testing.T) {
	raw := `
state: init
SSID:
channel: --
`
	s := parseDarwinAirport(raw)
	if s["state"] != "disconnected" {
		t.Fatalf("unexpected state: %q", s["state"])
	}
	if _, ok := s["ssid"]; ok {
		t.Fatalf("expected empty ssid to be ignored, got: %q", s["ssid"])
	}
}

func TestParseDarwinAirportUnknownWhenPartialWithoutSSID(t *testing.T) {
	raw := `
bssid: 01:23:45:67:89:ab
lastTxRate: 351
`
	s := parseDarwinAirport(raw)
	if s["state"] != "unknown" {
		t.Fatalf("unexpected state: %q", s["state"])
	}
}

func TestFormatWiFiSummary(t *testing.T) {
	raw := "*:Office:Infra:6:130 Mbit/s:70:WPA2"
	out := formatWiFiSummary("linux", raw)
	if !strings.Contains(out, "Wi-Fi summary:") {
		t.Fatalf("missing summary header: %s", out)
	}
	if !strings.Contains(out, "Office") {
		t.Fatalf("missing ssid: %s", out)
	}
}

func TestFormatWiFiSummaryWindowsDisconnected(t *testing.T) {
	raw := `
Name                   : Wi-Fi
State                  : disconnected
`
	out := formatWiFiSummary("windows", raw)
	if !strings.Contains(out, "Wi-Fi summary:") {
		t.Fatalf("missing summary header: %s", out)
	}
	if !strings.Contains(out, "- State: disconnected") {
		t.Fatalf("expected normalized disconnected state, got: %s", out)
	}
	if !strings.Contains(out, "- Interface: Wi-Fi") {
		t.Fatalf("expected interface in summary, got: %s", out)
	}
	if !strings.Contains(out, "Raw output:") {
		t.Fatalf("expected raw output section, got: %s", out)
	}
}
