package osdetect

import "testing"

func TestGuessFromHostAndPorts_PassiveByHostname(t *testing.T) {
	osName, confidence, reason := GuessFromHostAndPorts("android-tv.local", []int{80}, false)
	if osName != "Android" {
		t.Fatalf("expected Android, got %q", osName)
	}
	if confidence == "" {
		t.Fatalf("expected non-empty confidence")
	}
	if reason == "" {
		t.Fatalf("expected non-empty reason")
	}
}

func TestGuessFromHostAndPorts_PassiveByPorts(t *testing.T) {
	osName, confidence, reason := GuessFromHostAndPorts("host.local", []int{135, 445}, false)
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

func TestGuessFromHostAndPorts_ActiveSignature(t *testing.T) {
	osName, confidence, reason := GuessFromHostAndPorts("unknown.local", []int{3389, 445}, true)
	if osName != "Windows" {
		t.Fatalf("expected Windows in active mode, got %q", osName)
	}
	if confidence != "высокая" {
		t.Fatalf("expected confidence=высокая, got %q", confidence)
	}
	if reason == "" {
		t.Fatalf("expected reason in active mode")
	}
}

func TestGuessFromHostAndPorts_ActiveWindowsWinRM(t *testing.T) {
	osName, confidence, reason := GuessFromHostAndPorts("srv.local", []int{5985, 445}, true)
	if osName != "Windows Server" {
		t.Fatalf("expected Windows Server, got %q", osName)
	}
	if confidence != "средняя" {
		t.Fatalf("expected confidence=средняя, got %q", confidence)
	}
	if reason == "" {
		t.Fatalf("expected reason in active mode")
	}
}

func TestGuessFromHostAndPorts_ActiveLinuxDocker(t *testing.T) {
	osName, confidence, reason := GuessFromHostAndPorts("node.local", []int{22, 2375}, true)
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

func TestGuessFromHostAndPorts_ActiveAppleHighConfidence(t *testing.T) {
	osName, confidence, reason := GuessFromHostAndPorts("apple.local", []int{5353, 62078}, true)
	if osName != "Apple iOS/macOS" {
		t.Fatalf("expected Apple iOS/macOS, got %q", osName)
	}
	if confidence != "высокая" {
		t.Fatalf("expected confidence=высокая, got %q", confidence)
	}
	if reason == "" {
		t.Fatalf("expected reason in active mode")
	}
}

func TestGuessFromHostAndPorts_ReasonAlwaysEmptyWhenNoGuess(t *testing.T) {
	osName, confidence, reason := GuessFromHostAndPorts("unknown.local", []int{65000}, true)
	if osName != "" || confidence != "" || reason != "" {
		t.Fatalf("expected empty tuple for unknown signature, got os=%q conf=%q reason=%q", osName, confidence, reason)
	}
}

func TestGuessFromHostAndPorts_ActiveDisabled(t *testing.T) {
	osName, confidence, reason := GuessFromHostAndPorts("unknown.local", []int{3389, 445}, false)
	if osName != "" || confidence != "" || reason != "" {
		t.Fatalf("expected empty guess when active disabled, got os=%q conf=%q reason=%q", osName, confidence, reason)
	}
}

func TestGuessFromHostAndPorts_EmptyInput(t *testing.T) {
	osName, confidence, reason := GuessFromHostAndPorts("", nil, true)
	if osName != "" || confidence != "" || reason != "" {
		t.Fatalf("expected empty guess for empty input, got os=%q conf=%q reason=%q", osName, confidence, reason)
	}
}

