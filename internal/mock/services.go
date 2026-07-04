package mock

import (
	"context"
	"sync"
	"sync/atomic"

	"network-scanner/internal/contracts"
)

// MockScannerService реализация ScannerService для тестов
type MockScannerService struct {
	mu      sync.RWMutex
	Results []contracts.ScanResult
	Error   error
	Called  bool
	scanCnt int64
	stopCnt int64
}

func NewMockScannerService() *MockScannerService {
	return &MockScannerService{}
}

func (m *MockScannerService) Scan(ctx context.Context, cfg contracts.ScanConfig, onProgress contracts.ProgressHandler) ([]contracts.ScanResult, error) {
	atomic.AddInt64(&m.scanCnt, 1)
	m.Called = true
	if onProgress != nil {
		onProgress("test", 0, 1, "mock")
		onProgress("test", 1, 1, "mock done")
	}
	return m.Results, m.Error
}

func (m *MockScannerService) Stop() {
	atomic.AddInt64(&m.stopCnt, 1)
}

func (m *MockScannerService) SetResults(results []contracts.ScanResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Results = results
}

func (m *MockScannerService) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Error = err
}

func (m *MockScannerService) ScanCallCount() int {
	return int(atomic.LoadInt64(&m.scanCnt))
}

func (m *MockScannerService) StopCallCount() int {
	return int(atomic.LoadInt64(&m.stopCnt))
}

func (m *MockScannerService) AssertScanCalled() bool {
	return m.ScanCallCount() > 0
}

// MockTopologyService реализация TopologyService для тестов
type MockTopologyService struct {
	mu       sync.RWMutex
	Topology *contracts.Topology
	Error    error
	Called   bool
	buildCnt int64
}

func NewMockTopologyService() *MockTopologyService {
	return &MockTopologyService{}
}

func (m *MockTopologyService) Build(ctx context.Context, results []contracts.ScanResult, opts contracts.TopologyOptions) (*contracts.Topology, error) {
	atomic.AddInt64(&m.buildCnt, 1)
	m.Called = true
	return m.Topology, m.Error
}

func (m *MockTopologyService) Export(t *contracts.Topology, format string, path string) error {
	return nil
}

func (m *MockTopologyService) SetTopology(topo *contracts.Topology) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Topology = topo
}

func (m *MockTopologyService) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Error = err
}

func (m *MockTopologyService) BuildCallCount() int {
	return int(atomic.LoadInt64(&m.buildCnt))
}

func (m *MockTopologyService) AssertBuildCalled() bool {
	return m.BuildCallCount() > 0
}

// MockSecurityService реализация SecurityService для тестов
type MockSecurityService struct {
	mu         sync.RWMutex
	Report     *contracts.SecurityReport
	Error      error
	Called     bool
	analyzeCnt int64
}

func NewMockSecurityService() *MockSecurityService {
	return &MockSecurityService{}
}

func (m *MockSecurityService) AnalyzeRun(ctx context.Context, results []contracts.ScanResult) (*contracts.SecurityReport, error) {
	atomic.AddInt64(&m.analyzeCnt, 1)
	m.Called = true
	return m.Report, m.Error
}

func (m *MockSecurityService) SetReport(report *contracts.SecurityReport) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Report = report
}

func (m *MockSecurityService) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Error = err
}

func (m *MockSecurityService) AnalyzeCallCount() int {
	return int(atomic.LoadInt64(&m.analyzeCnt))
}

func (m *MockSecurityService) AssertAnalyzeCalled() bool {
	return m.AnalyzeCallCount() > 0
}

// MockRemoteExecService реализация RemoteExecService для тестов
type MockRemoteExecService struct {
	mu         sync.RWMutex
	Response   contracts.RemoteExecResponse
	Error      error
	Called     bool
	executeCnt int64
	dryRunCnt  int64
}

func NewMockRemoteExecService() *MockRemoteExecService {
	return &MockRemoteExecService{}
}

func (m *MockRemoteExecService) Execute(ctx context.Context, req contracts.RemoteExecRequest) (contracts.RemoteExecResponse, error) {
	atomic.AddInt64(&m.executeCnt, 1)
	m.Called = true
	return m.Response, m.Error
}

func (m *MockRemoteExecService) DryRun(ctx context.Context, req contracts.RemoteExecRequest) error {
	atomic.AddInt64(&m.dryRunCnt, 1)
	return m.Error
}

func (m *MockRemoteExecService) SetResponse(resp contracts.RemoteExecResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Response = resp
}

func (m *MockRemoteExecService) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Error = err
}

func (m *MockRemoteExecService) ExecuteCallCount() int {
	return int(atomic.LoadInt64(&m.executeCnt))
}

func (m *MockRemoteExecService) DryRunCallCount() int {
	return int(atomic.LoadInt64(&m.dryRunCnt))
}

// MockInventoryService реализация InventoryService для тестов
type MockInventoryService struct {
	mu        sync.RWMutex
	Snapshots []contracts.Snapshot
	DiffData  *contracts.Diff
	Error     error
	Called    bool
	saveCnt   int64
	listCnt   int64
	diffCnt   int64
}

func NewMockInventoryService() *MockInventoryService {
	return &MockInventoryService{}
}

func (m *MockInventoryService) SaveSnapshot(ctx context.Context, id string, data []contracts.ScanResult) error {
	atomic.AddInt64(&m.saveCnt, 1)
	m.Called = true
	return m.Error
}

func (m *MockInventoryService) ListSnapshots(ctx context.Context, limit int) ([]contracts.Snapshot, error) {
	atomic.AddInt64(&m.listCnt, 1)
	m.Called = true
	return m.Snapshots, m.Error
}

func (m *MockInventoryService) Diff(ctx context.Context, idA, idB string) (*contracts.Diff, error) {
	atomic.AddInt64(&m.diffCnt, 1)
	m.Called = true
	return m.DiffData, m.Error
}

func (m *MockInventoryService) SetSnapshots(snaps []contracts.Snapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Snapshots = snaps
}

func (m *MockInventoryService) SetDiff(diff *contracts.Diff) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DiffData = diff
}

func (m *MockInventoryService) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Error = err
}

func (m *MockInventoryService) SaveCallCount() int {
	return int(atomic.LoadInt64(&m.saveCnt))
}

func (m *MockInventoryService) ListCallCount() int {
	return int(atomic.LoadInt64(&m.listCnt))
}

func (m *MockInventoryService) DiffCallCount() int {
	return int(atomic.LoadInt64(&m.diffCnt))
}

func (m *MockInventoryService) AssertSaveCalled() bool {
	return m.SaveCallCount() > 0
}

// Helper для создания тестовых данных
func NewMockScanResult(ip, hostname string) contracts.ScanResult {
	return contracts.ScanResult{
		IP:       ip,
		Hostname: hostname,
		Ports: []contracts.PortInfo{
			{Port: 80, State: "open", Protocol: "tcp", Service: "http"},
			{Port: 443, State: "open", Protocol: "tcp", Service: "https"},
		},
	}
}

func NewMockSecurityReport() *contracts.SecurityReport {
	return &contracts.SecurityReport{
		PortAudit: []contracts.Finding{
			{Severity: "high", Host: "192.168.1.1", Title: "Open SSH port"},
		},
		Score: 85,
	}
}

// TestHelper - утилита для создания mock-контейнера
type TestContainer struct {
	Scanner    *MockScannerService
	Topology   *MockTopologyService
	Security   *MockSecurityService
	RemoteExec *MockRemoteExecService
	Inventory  *MockInventoryService
}

func NewTestContainer() *TestContainer {
	return &TestContainer{
		Scanner:    NewMockScannerService(),
		Topology:   NewMockTopologyService(),
		Security:   NewMockSecurityService(),
		RemoteExec: NewMockRemoteExecService(),
		Inventory:  NewMockInventoryService(),
	}
}
