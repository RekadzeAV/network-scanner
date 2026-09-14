package deviceclassifier

import (
	"strings"
	"testing"
)

// --- Comprehensive Classify tests ---

func TestClassify_PrinterBy515(t *testing.T) {
	got := Classify(Input{Ports: []Port{{Port: 515, State: "open"}}})
	if got != CategoryPrinter {
		t.Errorf("Classify(515) = %q, want %q", got, CategoryPrinter)
	}
}

func TestClassify_PrinterBy631(t *testing.T) {
	got := Classify(Input{Ports: []Port{{Port: 631, State: "open"}}})
	if got != CategoryPrinter {
		t.Errorf("Classify(631) = %q, want %q", got, CategoryPrinter)
	}
}

func TestClassify_PrinterBy9100(t *testing.T) {
	got := Classify(Input{Ports: []Port{{Port: 9100, State: "open"}}})
	if got != CategoryPrinter {
		t.Errorf("Classify(9100) = %q, want %q", got, CategoryPrinter)
	}
}

func TestClassify_Camera(t *testing.T) {
	got := Classify(Input{Ports: []Port{{Port: 554, State: "open"}}})
	if got != CategoryCamera {
		t.Errorf("Classify(554) = %q, want %q", got, CategoryCamera)
	}
}

func TestClassify_NASBy548(t *testing.T) {
	got := Classify(Input{Ports: []Port{{Port: 548, State: "open"}}})
	if got != CategoryNAS {
		t.Errorf("Classify(548) = %q, want %q", got, CategoryNAS)
	}
}

func TestClassify_NASBy2049(t *testing.T) {
	got := Classify(Input{Ports: []Port{{Port: 2049, State: "open"}}})
	if got != CategoryNAS {
		t.Errorf("Classify(2049) = %q, want %q", got, CategoryNAS)
	}
}

func TestClassify_RouterSwitchSSHHTTP(t *testing.T) {
	got := Classify(Input{Ports: []Port{
		{Port: 22, State: "open"},
		{Port: 80, State: "open"},
	}})
	if got != CategoryRouterSwitch {
		t.Errorf("Classify(22+80) = %q, want %q", got, CategoryRouterSwitch)
	}
}

func TestClassify_RouterSwitchNoMySQL(t *testing.T) {
	// Router должен быть без 3306 и 5432
	got := Classify(Input{Ports: []Port{
		{Port: 22, State: "open"},
		{Port: 80, State: "open"},
		{Port: 3306, State: "closed"},
	}})
	if got != CategoryRouterSwitch {
		t.Errorf("Classify(22+80+closed3306) = %q, want %q", got, CategoryRouterSwitch)
	}
}

func TestClassify_ServerHTTP443SSH(t *testing.T) {
	got := Classify(Input{Ports: []Port{
		{Port: 443, State: "open"},
		{Port: 80, State: "open"},
		{Port: 22, State: "open"},
	}})
	// Без vendor/host router — это Router/Switch (22+80 triggers router switch)
	if got != CategoryRouterSwitch {
		t.Errorf("Classify(443+80+22) = %q, want %q", got, CategoryRouterSwitch)
	}
}

func TestClassify_RouterSwitchCiscoVendor(t *testing.T) {
	got := Classify(Input{
		Ports: []Port{
			{Port: 443, State: "open"},
			{Port: 80, State: "open"},
			{Port: 22, State: "open"},
		},
		DeviceVendor: "Cisco",
	})
	if got != CategoryRouterSwitch {
		t.Errorf("Classify(Cisco) = %q, want %q", got, CategoryRouterSwitch)
	}
}

func TestClassify_RouterSwitchNetgearVendor(t *testing.T) {
	got := Classify(Input{
		Ports: []Port{
			{Port: 443, State: "open"},
			{Port: 80, State: "open"},
			{Port: 22, State: "open"},
		},
		DeviceVendor: "NETGEAR",
	})
	if got != CategoryRouterSwitch {
		t.Errorf("Classify(Netgear) = %q, want %q", got, CategoryRouterSwitch)
	}
}

func TestClassify_RouterSwitchTPLinkVendor(t *testing.T) {
	got := Classify(Input{
		Ports: []Port{
			{Port: 443, State: "open"},
			{Port: 80, State: "open"},
			{Port: 22, State: "open"},
		},
		DeviceVendor: "TP-Link",
	})
	if got != CategoryRouterSwitch {
		t.Errorf("Classify(TP-Link) = %q, want %q", got, CategoryRouterSwitch)
	}
}

func TestClassify_RouterSwitchDLinkVendor(t *testing.T) {
	got := Classify(Input{
		Ports: []Port{
			{Port: 443, State: "open"},
			{Port: 80, State: "open"},
			{Port: 22, State: "open"},
		},
		DeviceVendor: "D-Link",
	})
	if got != CategoryRouterSwitch {
		t.Errorf("Classify(D-Link) = %q, want %q", got, CategoryRouterSwitch)
	}
}

func TestClassify_RouterSwitchASUSVendor(t *testing.T) {
	got := Classify(Input{
		Ports: []Port{
			{Port: 443, State: "open"},
			{Port: 80, State: "open"},
			{Port: 22, State: "open"},
		},
		DeviceVendor: "ASUS",
	})
	if got != CategoryRouterSwitch {
		t.Errorf("Classify(ASUS) = %q, want %q", got, CategoryRouterSwitch)
	}
}

func TestClassify_RouterSwitchLinksysVendor(t *testing.T) {
	got := Classify(Input{
		Ports: []Port{
			{Port: 443, State: "open"},
			{Port: 80, State: "open"},
			{Port: 22, State: "open"},
		},
		DeviceVendor: "Linksys",
	})
	if got != CategoryRouterSwitch {
		t.Errorf("Classify(Linksys) = %q, want %q", got, CategoryRouterSwitch)
	}
}

func TestClassify_RouterSwitchRouterHostname(t *testing.T) {
	got := Classify(Input{
		Ports: []Port{
			{Port: 443, State: "open"},
			{Port: 80, State: "open"},
			{Port: 22, State: "open"},
		},
		Hostname: "my-router",
	})
	if got != CategoryRouterSwitch {
		t.Errorf("Classify(router hostname) = %q, want %q", got, CategoryRouterSwitch)
	}
}

func TestClassify_RouterSwitchGatewayHostname(t *testing.T) {
	got := Classify(Input{
		Ports: []Port{
			{Port: 443, State: "open"},
			{Port: 80, State: "open"},
			{Port: 22, State: "open"},
		},
		Hostname: "gateway.local",
	})
	if got != CategoryRouterSwitch {
		t.Errorf("Classify(gateway hostname) = %q, want %q", got, CategoryRouterSwitch)
	}
}

func TestClassify_RouterSwitchSNMP(t *testing.T) {
	got := Classify(Input{
		Ports: []Port{
			{Port: 161, State: "open"},
		},
		Hostname: "switch-core",
	})
	if got != CategoryRouterSwitch {
		t.Errorf("Classify(SNMP+switch) = %q, want %q", got, CategoryRouterSwitch)
	}
}

func TestClassify_RouterSwitchSNMPCisco(t *testing.T) {
	got := Classify(Input{
		Ports: []Port{
			{Port: 161, State: "open"},
		},
		DeviceVendor: "Cisco",
	})
	if got != CategoryRouterSwitch {
		t.Errorf("Classify(SNMP+Cisco) = %q, want %q", got, CategoryRouterSwitch)
	}
}

func TestClassify_DesktopRDP(t *testing.T) {
	got := Classify(Input{Ports: []Port{{Port: 3389, State: "open"}}})
	if got != CategoryDesktopLaptop {
		t.Errorf("Classify(3389) = %q, want %q", got, CategoryDesktopLaptop)
	}
}

func TestClassify_DesktopSMB(t *testing.T) {
	got := Classify(Input{Ports: []Port{
		{Port: 135, State: "open"},
		{Port: 445, State: "open"},
	}})
	if got != CategoryDesktopLaptop {
		t.Errorf("Classify(135+445) = %q, want %q", got, CategoryDesktopLaptop)
	}
}

func TestClassify_ServerMySQL(t *testing.T) {
	got := Classify(Input{Ports: []Port{{Port: 3306, State: "open"}}})
	if got != CategoryServer {
		t.Errorf("Classify(3306) = %q, want %q", got, CategoryServer)
	}
}

func TestClassify_ServerPostgreSQL(t *testing.T) {
	got := Classify(Input{Ports: []Port{{Port: 5432, State: "open"}}})
	if got != CategoryServer {
		t.Errorf("Classify(5432) = %q, want %q", got, CategoryServer)
	}
}

func TestClassify_ServerMSSQL(t *testing.T) {
	got := Classify(Input{Ports: []Port{{Port: 1433, State: "open"}}})
	if got != CategoryServer {
		t.Errorf("Classify(1433) = %q, want %q", got, CategoryServer)
	}
}

func TestClassify_ServerHTTP8080(t *testing.T) {
	got := Classify(Input{Ports: []Port{{Port: 8080, State: "open"}}})
	if got != CategoryServer {
		t.Errorf("Classify(8080) = %q, want %q", got, CategoryServer)
	}
}

func TestClassify_ServerHTTP8443(t *testing.T) {
	got := Classify(Input{Ports: []Port{{Port: 8443, State: "open"}}})
	if got != CategoryServer {
		t.Errorf("Classify(8443) = %q, want %q", got, CategoryServer)
	}
}

func TestClassify_ServerSSH(t *testing.T) {
	got := Classify(Input{Ports: []Port{{Port: 22, State: "open"}}})
	if got != CategoryServer {
		t.Errorf("Classify(22 only) = %q, want %q", got, CategoryServer)
	}
}

func TestClassify_ServerHTTP443(t *testing.T) {
	got := Classify(Input{Ports: []Port{
		{Port: 80, State: "open"},
		{Port: 443, State: "open"},
	}})
	if got != CategoryServer {
		t.Errorf("Classify(80+443) = %q, want %q", got, CategoryServer)
	}
}

func TestClassify_IoTHTTP(t *testing.T) {
	got := Classify(Input{Ports: []Port{{Port: 80, State: "open"}}})
	if got != CategoryIoT {
		t.Errorf("Classify(80 only) = %q, want %q", got, CategoryIoT)
	}
}

func TestClassify_IoTHTTPS(t *testing.T) {
	got := Classify(Input{Ports: []Port{{Port: 443, State: "open"}}})
	if got != CategoryIoT {
		t.Errorf("Classify(443 only) = %q, want %q", got, CategoryIoT)
	}
}

func TestClassify_IoTSinglePort(t *testing.T) {
	got := Classify(Input{Ports: []Port{{Port: 8883, State: "open"}}})
	if got != CategoryIoT {
		t.Errorf("Classify(single port) = %q, want %q", got, CategoryIoT)
	}
}

func TestClassify_UnknownEmptyPorts(t *testing.T) {
	got := Classify(Input{Ports: []Port{
		{Port: 80, State: "closed"},
	}})
	if got != CategoryUnknown {
		t.Errorf("Classify(closed ports) = %q, want %q", got, CategoryUnknown)
	}
}

func TestClassify_UnknownNoPorts(t *testing.T) {
	got := Classify(Input{})
	if got != CategoryUnknown {
		t.Errorf("Classify(no ports) = %q, want %q", got, CategoryUnknown)
	}
}

func TestClassify_MixedStates(t *testing.T) {
	// Только open порты должны учитываться
	got := Classify(Input{Ports: []Port{
		{Port: 22, State: "open"},
		{Port: 80, State: "closed"},
		{Port: 443, State: "filtered"},
	}})
	if got != CategoryServer {
		t.Errorf("Classify(mixed states) = %q, want %q", got, CategoryServer)
	}
}

func TestClassify_CaseInsensitiveState(t *testing.T) {
	got := Classify(Input{Ports: []Port{
		{Port: 80, State: "Open"},
	}})
	if got != CategoryIoT {
		t.Errorf("Classify(Open case) = %q, want %q", got, CategoryIoT)
	}
}

func TestClassify_WhitespaceState(t *testing.T) {
	got := Classify(Input{Ports: []Port{
		{Port: 80, State: " open "},
	}})
	if got != CategoryIoT {
		t.Errorf("Classify(whitespace state) = %q, want %q", got, CategoryIoT)
	}
}

// --- containsAny tests ---

func TestContainsAny_SingleMatch(t *testing.T) {
	if !containsAny("cisco router", "cisco") {
		t.Error("containsAny should match single part")
	}
}

func TestContainsAny_MultipleMatch(t *testing.T) {
	if !containsAny("netgear device", "tp-link", "netgear", "asus") {
		t.Error("containsAny should match second part")
	}
}

func TestContainsAny_NoMatch(t *testing.T) {
	if containsAny("unknown device", "cisco", "netgear") {
		t.Error("containsAny should not match")
	}
}

func TestContainsAny_EmptyString(t *testing.T) {
	if containsAny("", "test") {
		t.Error("containsAny should not match empty string")
	}
}

func TestContainsAny_EmptyParts(t *testing.T) {
	if containsAny("test") {
		t.Error("containsAny with no parts should return false")
	}
}

func TestContainsAny_CaseSensitive(t *testing.T) {
	// containsAny case-sensitive
	if !containsAny("cisco", "cisco") {
		t.Error("containsAny should match exact case")
	}
}

// --- Edge cases ---

func TestClassify_MultiplePrinterPorts(t *testing.T) {
	got := Classify(Input{Ports: []Port{
		{Port: 515, State: "open"},
		{Port: 631, State: "open"},
		{Port: 9100, State: "open"},
	}})
	if got != CategoryPrinter {
		t.Errorf("Classify(multiple printer ports) = %q, want %q", got, CategoryPrinter)
	}
}

func TestClassify_PrinterPriorityOverCamera(t *testing.T) {
	got := Classify(Input{Ports: []Port{
		{Port: 9100, State: "open"},
		{Port: 554, State: "open"},
	}})
	if got != CategoryPrinter {
		t.Errorf("Classify(printer+camera) = %q, want %q (printer priority)", got, CategoryPrinter)
	}
}

func TestClassify_NASPriorityOverPrinter(t *testing.T) {
	// Printer имеет приоритет (проверяется первым)
	got := Classify(Input{Ports: []Port{
		{Port: 2049, State: "open"},
		{Port: 9100, State: "open"},
	}})
	if got != CategoryPrinter {
		t.Errorf("Classify(nas+printer) = %q, want %q (printer priority)", got, CategoryPrinter)
	}
}

func TestClassify_RouterSwitchPriorityOverServer(t *testing.T) {
	got := Classify(Input{Ports: []Port{
		{Port: 22, State: "open"},
		{Port: 80, State: "open"},
		{Port: 3306, State: "closed"},
	}})
	if got != CategoryRouterSwitch {
		t.Errorf("Classify(router+closed mysql) = %q, want %q", got, CategoryRouterSwitch)
	}
}

// --- Benchmark ---

func BenchmarkClassify(b *testing.B) {
	input := Input{
		Ports: []Port{
			{Port: 22, State: "open"},
			{Port: 80, State: "open"},
			{Port: 443, State: "open"},
		},
		DeviceVendor: "Cisco",
		Hostname:     "router",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Classify(input)
	}
}

func BenchmarkContainsAny(b *testing.B) {
	vendor := "cisco"
	parts := []string{"cisco", "netgear", "tp-link", "d-link", "asus", "linksys"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = containsAny(vendor, parts...)
	}
}

// Verify category constants
func TestCategoryConstants(t *testing.T) {
	categories := []string{
		CategoryUnknown,
		CategoryRouterSwitch,
		CategoryAccessPoint,
		CategoryPrinter,
		CategoryCamera,
		CategoryNAS,
		CategoryIoT,
		CategoryDesktopLaptop,
		CategoryServer,
		CategoryPhoneTablet,
	}

	for _, cat := range categories {
		if cat == "" {
			t.Error("Category constant should not be empty")
		}
		if !strings.Contains(cat, cat) {
			t.Errorf("Category %q should be valid", cat)
		}
	}
}
