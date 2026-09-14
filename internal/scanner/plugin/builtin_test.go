package plugin

import (
	"context"
	"testing"
)

// ============================================================================
// C1: Тесты для builtin probe плагинов
// ============================================================================

// TestBuiltin_ICMPProbe — ветка: ICMP probe
func TestBuiltin_ICMPProbe(t *testing.T) {
	probe := NewICMPProbe()

	if probe.Name() != "icmp-ping" {
		t.Errorf("expected name icmp-ping, got %s", probe.Name())
	}
	if probe.Phase() != PhaseHostDiscovery {
		t.Error("expected PhaseHostDiscovery")
	}
	if probe.Priority() != 100 {
		t.Error("expected priority 100")
	}
}

// TestBuiltin_PortProbe — ветка: Port probe
func TestBuiltin_PortProbe(t *testing.T) {
	probe := NewDefaultPortProbe()

	if probe.Name() != "port-check" {
		t.Errorf("expected name port-check, got %s", probe.Name())
	}
	if probe.Phase() != PhaseHostDiscovery {
		t.Error("expected PhaseHostDiscovery")
	}
	if probe.Priority() != 200 {
		t.Error("expected priority 200")
	}
	if len(probe.Ports) != 6 {
		t.Errorf("expected 6 ports, got %d", len(probe.Ports))
	}
}

// TestBuiltin_SNMPProbe — ветка: SNMP probe
func TestBuiltin_SNMPProbe(t *testing.T) {
	probe := NewSNMPProbe()

	if probe.Name() != "snmp-check" {
		t.Errorf("expected name snmp-check, got %s", probe.Name())
	}
	if probe.Phase() != PhaseServiceProbe {
		t.Error("expected PhaseServiceProbe")
	}
	if probe.Priority() != 300 {
		t.Error("expected priority 300")
	}
}

// TestBuiltin_SSHProbe — ветка: SSH probe
func TestBuiltin_SSHProbe(t *testing.T) {
	probe := NewSSHProbe()

	if probe.Name() != "ssh-banner" {
		t.Errorf("expected name ssh-banner, got %s", probe.Name())
	}
	if probe.Phase() != PhaseServiceProbe {
		t.Error("expected PhaseServiceProbe")
	}
	if probe.Priority() != 310 {
		t.Error("expected priority 310")
	}
}

// TestBuiltin_HTTPProbe — ветка: HTTP probe
func TestBuiltin_HTTPProbe(t *testing.T) {
	probe := NewHTTPProbe()

	if probe.Name() != "http-fingerprint" {
		t.Errorf("expected name http-fingerprint, got %s", probe.Name())
	}
	if probe.Phase() != PhaseServiceProbe {
		t.Error("expected PhaseServiceProbe")
	}
	if probe.Priority() != 320 {
		t.Error("expected priority 320")
	}
}

// TestBuiltin_RegisterDefaultPlugins — ветка: регистрация всех плагинов
func TestBuiltin_RegisterDefaultPlugins(t *testing.T) {
	registry := NewRegistry()

	RegisterDefaultPlugins(registry)

	if registry.Count() != 5 {
		t.Errorf("expected 5 plugins, got %d", registry.Count())
	}

	// Проверяем что все фазы представлены
	if registry.CountByPhase(PhaseHostDiscovery) != 2 {
		t.Errorf("expected 2 discovery plugins")
	}
	if registry.CountByPhase(PhaseServiceProbe) != 3 {
		t.Errorf("expected 3 service plugins")
	}
}

// TestBuiltin_GetDefaultPlugins — ветка: получение списка плагинов
func TestBuiltin_GetDefaultPlugins(t *testing.T) {
	plugins := GetDefaultPlugins()

	if len(plugins) != 5 {
		t.Errorf("expected 5 plugins, got %d", len(plugins))
	}

	// Проверяем имена
	expectedNames := []string{"icmp-ping", "port-check", "snmp-check", "ssh-banner", "http-fingerprint"}
	for i, name := range expectedNames {
		if plugins[i].Name() != name {
			t.Errorf("expected plugin %d name %s, got %s", i, name, plugins[i].Name())
		}
	}
}

// TestBuiltin_ICMPProbe_Execute — ветка: выполнение ICMP probe
func TestBuiltin_ICMPProbe_Execute(t *testing.T) {
	probe := NewICMPProbe()

	ctx := context.Background()

	// ICMP на localhost должен выполниться (или быть заблокирован firewall)
	result, err := probe.Execute(ctx, "127.0.0.1", 0)
	if err != nil {
		// Ошибка допустима если ICMP заблокирован
		t.Logf("ICMP probe error (expected if firewall blocks): %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// TestBuiltin_PortProbe_Execute — ветка: выполнение Port probe
func TestBuiltin_PortProbe_Execute(t *testing.T) {
	probe := NewDefaultPortProbe()

	ctx := context.Background()

	// Port probe на localhost
	result, err := probe.Execute(ctx, "127.0.0.1", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Если открытые порты есть, они должны быть в Extra
	if extra, ok := result.Extra["open-ports"]; ok {
		t.Logf("Open ports on localhost: %s", extra)
	}
}

// TestBuiltin_SNMPProbe_Execute — ветка: выполнение SNMP probe
func TestBuiltin_SNMPProbe_Execute(t *testing.T) {
	probe := NewSNMPProbe()

	ctx := context.Background()

	result, err := probe.Execute(ctx, "127.0.0.1", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// TestBuiltin_SSHProbe_Execute — ветка: выполнение SSH probe
func TestBuiltin_SSHProbe_Execute(t *testing.T) {
	probe := NewSSHProbe()

	ctx := context.Background()

	result, err := probe.Execute(ctx, "127.0.0.1", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// TestBuiltin_HTTPProbe_Execute — ветка: выполнение HTTP probe
func TestBuiltin_HTTPProbe_Execute(t *testing.T) {
	probe := NewHTTPProbe()

	ctx := context.Background()

	result, err := probe.Execute(ctx, "127.0.0.1", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// TestBuiltin_InvalidPort — ветка: невалидные порты
func TestBuiltin_InvalidPort(t *testing.T) {
	ctx := context.Background()

	// ICMP не должен принимать порт
	probe := Probe(NewICMPProbe())
	_, err := probe.Execute(ctx, "127.0.0.1", 80)
	if err == nil {
		t.Error("expected error for ICMP with non-zero port")
	}

	// SNMP должен использовать порт 161
	probe = Probe(NewSNMPProbe())
	_, err = probe.Execute(ctx, "127.0.0.1", 80)
	if err == nil {
		t.Error("expected error for SNMP with wrong port")
	}

	// SSH должен использовать порт 22
	probe = Probe(NewSSHProbe())
	_, err = probe.Execute(ctx, "127.0.0.1", 80)
	if err == nil {
		t.Error("expected error for SSH with wrong port")
	}
}

// TestBuiltin_ContextCancellation — ветка: отмена контекста
func TestBuiltin_ContextCancellation(t *testing.T) {
	probe := NewICMPProbe()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Немедленная отмена

	_, err := probe.Execute(ctx, "127.0.0.1", 0)
	if err == nil {
		t.Error("expected context cancellation error")
	}
}
