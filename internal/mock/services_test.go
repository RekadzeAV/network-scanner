package mock

import (
	"context"
	"errors"
	"testing"

	"network-scanner/internal/contracts"
)

func TestMockScannerService(t *testing.T) {
	mock := NewMockScannerService()

	// Test default state
	if mock.ScanCallCount() != 0 {
		t.Error("expected scan count 0")
	}

	// Test Scan
	mock.SetResults([]contracts.ScanResult{{IP: "1.2.3.4"}})
	mock.SetError(nil)

	results, err := mock.Scan(context.Background(), contracts.ScanConfig{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if mock.ScanCallCount() != 1 {
		t.Errorf("expected scan count 1, got %d", mock.ScanCallCount())
	}
	if !mock.AssertScanCalled() {
		t.Error("expected Scan to be called")
	}

	// Test Stop
	mock.Stop()
	if mock.StopCallCount() != 1 {
		t.Errorf("expected stop count 1, got %d", mock.StopCallCount())
	}
}

func TestMockScannerService_Error(t *testing.T) {
	mock := NewMockScannerService()
	mock.SetError(errors.New("test error"))

	_, err := mock.Scan(context.Background(), contracts.ScanConfig{}, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockTopologyService(t *testing.T) {
	mock := NewMockTopologyService()

	// Test default state
	if mock.BuildCallCount() != 0 {
		t.Error("expected build count 0")
	}

	// Test Build
	mock.SetTopology(&contracts.Topology{Devices: []*contracts.Device{{IP: "1.2.3.4"}}})

	topo, err := mock.Build(context.Background(), nil, contracts.TopologyOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if topo == nil || len(topo.Devices) != 1 {
		t.Fatal("expected topology with 1 device")
	}

	if mock.BuildCallCount() != 1 {
		t.Errorf("expected build count 1, got %d", mock.BuildCallCount())
	}
	if !mock.AssertBuildCalled() {
		t.Error("expected Build to be called")
	}
}

func TestMockSecurityService(t *testing.T) {
	mock := NewMockSecurityService()

	// Test default state
	if mock.AnalyzeCallCount() != 0 {
		t.Error("expected analyze count 0")
	}

	// Test AnalyzeRun
	mock.SetReport(&contracts.SecurityReport{Score: 90})

	report, err := mock.AnalyzeRun(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Score != 90 {
		t.Errorf("expected score 90, got %d", report.Score)
	}

	if mock.AnalyzeCallCount() != 1 {
		t.Errorf("expected analyze count 1, got %d", mock.AnalyzeCallCount())
	}
	if !mock.AssertAnalyzeCalled() {
		t.Error("expected Analyze to be called")
	}
}

func TestMockInventoryService(t *testing.T) {
	mock := NewMockInventoryService()

	// Test default state
	if mock.SaveCallCount() != 0 || mock.ListCallCount() != 0 || mock.DiffCallCount() != 0 {
		t.Error("expected all counts 0")
	}

	// Test SaveSnapshot
	mock.SetError(nil)
	err := mock.SaveSnapshot(context.Background(), "test-id", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.SaveCallCount() != 1 {
		t.Errorf("expected save count 1, got %d", mock.SaveCallCount())
	}
	if !mock.AssertSaveCalled() {
		t.Error("expected SaveSnapshot to be called")
	}

	// Test ListSnapshots
	mock.SetSnapshots([]contracts.Snapshot{{ID: "snap1"}})
	snaps, err := mock.ListSnapshots(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snaps) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snaps))
	}
	if mock.ListCallCount() != 1 {
		t.Errorf("expected list count 1, got %d", mock.ListCallCount())
	}

	// Test Diff
	mock.SetDiff(&contracts.Diff{ScanIDA: "snap1", ScanIDB: "snap2"})
	diff, err := mock.Diff(context.Background(), "snap1", "snap2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diff.ScanIDA != "snap1" || diff.ScanIDB != "snap2" {
		t.Error("expected diff data")
	}
	if mock.DiffCallCount() != 1 {
		t.Errorf("expected diff count 1, got %d", mock.DiffCallCount())
	}
}

func TestMockRemoteExecService(t *testing.T) {
	mock := NewMockRemoteExecService()

	// Test default state
	if mock.ExecuteCallCount() != 0 || mock.DryRunCallCount() != 0 {
		t.Error("expected all counts 0")
	}

	// Test Execute
	mock.SetResponse(contracts.RemoteExecResponse{Success: true, Output: "hello"})
	resp, err := mock.Execute(context.Background(), contracts.RemoteExecRequest{Command: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success || resp.Output != "hello" {
		t.Error("expected successful response")
	}
	if mock.ExecuteCallCount() != 1 {
		t.Errorf("expected execute count 1, got %d", mock.ExecuteCallCount())
	}

	// Test DryRun
	mock.SetError(nil)
	err = mock.DryRun(context.Background(), contracts.RemoteExecRequest{Command: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.DryRunCallCount() != 1 {
		t.Errorf("expected dry run count 1, got %d", mock.DryRunCallCount())
	}
}

func TestTestContainer(t *testing.T) {
	container := NewTestContainer()

	if container.Scanner == nil || container.Topology == nil || container.Security == nil {
		t.Fatal("expected all services in container")
	}

	// Verify all mocks are independent
	container.Scanner.SetResults([]contracts.ScanResult{{IP: "1.1.1.1"}})
	container.Scanner.Scan(context.Background(), contracts.ScanConfig{}, nil)

	if container.Scanner.ScanCallCount() != 1 {
		t.Error("expected scanner call count 1")
	}

	if container.Topology.BuildCallCount() != 0 {
		t.Error("expected topology call count 0")
	}
}

func TestNewMockScanResult(t *testing.T) {
	result := NewMockScanResult("192.168.1.1", "test-host")

	if result.IP != "192.168.1.1" {
		t.Errorf("expected IP 192.168.1.1, got %s", result.IP)
	}
	if result.Hostname != "test-host" {
		t.Errorf("expected hostname test-host, got %s", result.Hostname)
	}
	if len(result.Ports) != 2 {
		t.Errorf("expected 2 ports, got %d", len(result.Ports))
	}
}

func TestNewMockSecurityReport(t *testing.T) {
	report := NewMockSecurityReport()

	if report.Score != 85 {
		t.Errorf("expected score 85, got %d", report.Score)
	}
	if len(report.PortAudit) != 1 {
		t.Errorf("expected 1 finding, got %d", len(report.PortAudit))
	}
}
