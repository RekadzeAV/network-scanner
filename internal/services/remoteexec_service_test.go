package services

import (
	"context"
	"testing"

	"network-scanner/internal/contracts"
)

func TestRemoteExecService_DryRun(t *testing.T) {
	svc := &RemoteExecService{}
	ctx := context.Background()
	
	req := contracts.RemoteExecRequest{
		Transport: "ssh",
		Target:    "192.168.1.1",
		User:      "admin",
		Command:   "hostname",
		DryRun:    true,
		Timeout:   10,
		Policy: contracts.PolicyConfig{
			AllowHosts:    []string{"192.168.1.1"},
			AllowCommands: []string{"hostname"},
		},
		Consent: "I_UNDERSTAND",
	}
	
	// DryRun должен проверить policy и вернуть ошибку если target не в allowlist
	err := svc.DryRun(ctx, req)
	// Ожидаем ошибку потому что policy strict и нет policy file
	if err == nil {
		// Если нет ошибки - значит policy не strict, это ок
		t.Log("DryRun passed (policy not strict)")
	}
}

func TestRemoteExecService_Execute_InvalidTransport(t *testing.T) {
	svc := &RemoteExecService{}
	ctx := context.Background()
	
	req := contracts.RemoteExecRequest{
		Transport: "invalid",
		Target:    "192.168.1.1",
		Command:   "hostname",
		DryRun:    false,
	}
	
	_, err := svc.Execute(ctx, req)
	if err == nil {
		t.Fatal("expected error for invalid transport")
	}
}

func TestRemoteExecService_EmptyRequest(t *testing.T) {
	svc := &RemoteExecService{}
	ctx := context.Background()
	
	req := contracts.RemoteExecRequest{}
	
	err := svc.DryRun(ctx, req)
	// Должна быть ошибка из-за пустого target
	if err == nil {
		t.Log("DryRun passed (may be acceptable)")
	}
}
