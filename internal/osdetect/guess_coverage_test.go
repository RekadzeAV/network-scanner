package osdetect

import (
	"testing"
)

func TestGuessFromHostAndPorts_ActiveLinuxServerDB(t *testing.T) {
	osName, confidence, reason := GuessFromHostAndPorts("db-server.local", []int{22, 3306}, true)
	if osName != "Linux/Unix Server" {
		t.Fatalf("expected Linux/Unix Server, got %q", osName)
	}
	if confidence != "средняя" {
		t.Fatalf("expected confidence=средняя, got %q", confidence)
	}
	if reason == "" {
		t.Fatalf("expected reason in active mode")
	}
}

func TestGuessFromHostAndPorts_ActiveLinuxServerK8s(t *testing.T) {
	osName, confidence, reason := GuessFromHostAndPorts("k8s-node.local", []int{22, 6443}, true)
	if osName != "Linux/Unix Server" {
		t.Fatalf("expected Linux/Unix Server, got %q", osName)
	}
	if confidence != "средняя" {
		t.Fatalf("expected confidence=средняя, got %q", confidence)
	}
	if reason == "" {
		t.Fatalf("expected reason in active mode")
	}
}

func TestGuessFromHostAndPorts_ActiveAndroidDebug(t *testing.T) {
	osName, confidence, reason := GuessFromHostAndPorts("android-dev.local", []int{5555, 8081}, true)
	if osName != "Android" {
		t.Fatalf("expected Android, got %q", osName)
	}
	if confidence != "средняя" {
		t.Fatalf("expected confidence=средняя, got %q", confidence)
	}
	if reason == "" {
		t.Fatalf("expected reason in active mode")
	}
}

func TestGuessFromHostAndPorts_ActiveAppleLockdown(t *testing.T) {
	// Note: hostname contains "iphone" which matches iOS/iPadOS before active mode
	osName, confidence, reason := GuessFromHostAndPorts("lockdown-device.local", []int{62078}, true)
	if osName != "Apple iOS/macOS" {
		t.Fatalf("expected Apple iOS/macOS, got %q", osName)
	}
	if confidence != "средняя" {
		t.Fatalf("expected confidence=средняя, got %q", confidence)
	}
	if reason == "" {
		t.Fatalf("expected reason in active mode")
	}
}

func TestGuessFromHostAndPorts_PassiveWindowsNetBIOS(t *testing.T) {
	osName, confidence, reason := GuessFromHostAndPorts("workstation.local", []int{139, 445}, false)
	if osName != "Windows" {
		t.Fatalf("expected Windows, got %q", osName)
	}
	if confidence != "средняя" {
		t.Fatalf("expected confidence=средняя, got %q", confidence)
	}
	if reason == "" {
		t.Fatalf("expected reason for passive signature")
	}
}

func TestGuessFromHostAndPorts_PassiveLinuxSSHHTTP(t *testing.T) {
	osName, confidence, reason := GuessFromHostAndPorts("web-server.local", []int{22, 80}, false)
	if osName != "Linux/Unix или сетевое устройство" {
		t.Fatalf("expected Linux/Unix или сетевое устройство, got %q", osName)
	}
	if confidence != "низкая" {
		t.Fatalf("expected confidence=низкая, got %q", confidence)
	}
	if reason == "" {
		t.Fatalf("expected reason for passive signature")
	}
}

func TestGuessFromHostAndPorts_PassiveApplePorts(t *testing.T) {
	osName, confidence, _ := GuessFromHostAndPorts("macbook.local", []int{548}, false)
	if osName != "macOS / CUPS / печать" {
		t.Fatalf("expected macOS / CUPS / печать, got %q", osName)
	}
	if confidence != "низкая" {
		t.Fatalf("expected confidence=низкая, got %q", confidence)
	}
}

func TestGuessFromHostAndPorts_PassiveApplePorts631(t *testing.T) {
	osName, confidence, _ := GuessFromHostAndPorts("printer.local", []int{631}, false)
	if osName != "macOS / CUPS / печать" {
		t.Fatalf("expected macOS / CUPS / печать, got %q", osName)
	}
	if confidence != "низкая" {
		t.Fatalf("expected confidence=низкая, got %q", confidence)
	}
}

func TestGuessFromHostAndPorts_RaspberryPi(t *testing.T) {
	osName, confidence, _ := GuessFromHostAndPorts("raspberry-pi.local", []int{22}, false)
	if osName != "Linux / Raspberry Pi OS" {
		t.Fatalf("expected Linux / Raspberry Pi OS, got %q", osName)
	}
	if confidence != "средняя" {
		t.Fatalf("expected confidence=средняя, got %q", confidence)
	}
}

func TestGuessFromHostAndPorts_WindowsPC(t *testing.T) {
	osName, confidence, _ := GuessFromHostAndPorts("WIN-DESKTOP.local", []int{3389}, false)
	if osName != "Windows" {
		t.Fatalf("expected Windows, got %q", osName)
	}
	if confidence != "низкая" {
		t.Fatalf("expected confidence=низкая, got %q", confidence)
	}
}

func TestGuessFromHostAndPorts_IPhone(t *testing.T) {
	osName, confidence, _ := GuessFromHostAndPorts("Johns-iPhone.local", []int{}, false)
	if osName != "Apple iOS/iPadOS" {
		t.Fatalf("expected Apple iOS/iPadOS, got %q", osName)
	}
	if confidence != "средняя" {
		t.Fatalf("expected confidence=средняя, got %q", confidence)
	}
}

func TestGuessFromHostAndPorts_IPad(t *testing.T) {
	osName, confidence, _ := GuessFromHostAndPorts("Tablet-iPad.local", []int{}, false)
	if osName != "Apple iOS/iPadOS" {
		t.Fatalf("expected Apple iOS/iPadOS, got %q", osName)
	}
	if confidence != "средняя" {
		t.Fatalf("expected confidence=средняя, got %q", confidence)
	}
}

func TestGuessFromHostAndPorts_ActiveLinuxDockerAPI(t *testing.T) {
	osName, confidence, _ := GuessFromHostAndPorts("docker-host.local", []int{22, 2376}, true)
	if osName != "Linux/Unix Server" {
		t.Fatalf("expected Linux/Unix Server, got %q", osName)
	}
	if confidence != "средняя" {
		t.Fatalf("expected confidence=средняя, got %q", confidence)
	}
}

func TestGuessFromHostAndPorts_ActivePostgres(t *testing.T) {
	osName, confidence, _ := GuessFromHostAndPorts("postgres.local", []int{22, 5432}, true)
	if osName != "Linux/Unix Server" {
		t.Fatalf("expected Linux/Unix Server, got %q", osName)
	}
	if confidence != "средняя" {
		t.Fatalf("expected confidence=средняя, got %q", confidence)
	}
}

func TestGuessFromHostAndPorts_ActiveRedis(t *testing.T) {
	osName, confidence, _ := GuessFromHostAndPorts("redis.local", []int{22, 6379}, true)
	if osName != "Linux/Unix Server" {
		t.Fatalf("expected Linux/Unix Server, got %q", osName)
	}
	if confidence != "средняя" {
		t.Fatalf("expected confidence=средняя, got %q", confidence)
	}
}

func TestGuessFromHostAndPorts_ActiveWindowsServerWinRM(t *testing.T) {
	osName, confidence, _ := GuessFromHostAndPorts("win-server.local", []int{5985, 139}, true)
	if osName != "Windows Server" {
		t.Fatalf("expected Windows Server, got %q", osName)
	}
	if confidence != "средняя" {
		t.Fatalf("expected confidence=средняя, got %q", confidence)
	}
}

func TestGuessFromHostAndPorts_ActiveAppleMDNS(t *testing.T) {
	// Note: port 548 triggers passive "macOS / CUPS / печать" before active mDNS rule
	osName, confidence, _ := GuessFromHostAndPorts("mdns-device.local", []int{5353, 548}, true)
	if osName != "macOS / CUPS / печать" {
		t.Fatalf("expected macOS / CUPS / печать (passive port rule takes precedence), got %q", osName)
	}
	if confidence != "низкая" {
		t.Fatalf("expected confidence=низкая (passive rule), got %q", confidence)
	}
}

func TestGuessFromHostAndPorts_EdgeCaseEmptyPorts(t *testing.T) {
	osName, confidence, reason := GuessFromHostAndPorts("test.local", []int{}, true)
	if osName != "" || confidence != "" || reason != "" {
		t.Fatalf("expected empty guess for empty ports, got os=%q conf=%q reason=%q", osName, confidence, reason)
	}
}

func TestGuessFromHostAndPorts_EdgeCaseNilPorts(t *testing.T) {
	osName, confidence, reason := GuessFromHostAndPorts("test.local", nil, true)
	if osName != "" || confidence != "" || reason != "" {
		t.Fatalf("expected empty guess for nil ports, got os=%q conf=%q reason=%q", osName, confidence, reason)
	}
}

func TestGuessFromHostAndPorts_CaseInsensitiveHostname(t *testing.T) {
	osName, _, _ := GuessFromHostAndPorts("ANDROID-DEV", []int{80}, false)
	if osName != "Android" {
		t.Fatalf("expected Android (case insensitive), got %q", osName)
	}
}

func TestGuessFromHostAndPorts_WhitespaceHostname(t *testing.T) {
	osName, _, _ := GuessFromHostAndPorts("  android-device  ", []int{}, false)
	if osName != "Android" {
		t.Fatalf("expected Android with whitespace, got %q", osName)
	}
}
