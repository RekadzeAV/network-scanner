package contracts

import (
	"context"
	"time"
)

// ScanConfig конфигурация сканирования
type ScanConfig struct {
	NetworkCIDR string
	PortRange   string
	Timeout     time.Duration
	Threads     int
	ShowClosed  bool
	ScanUDP     bool
	GrabBanners bool
	OSActive    bool
	VerboseLogs bool
}

// ScanResult результат сканирования (упрощённый интерфейс для сервисов)
type ScanResult struct {
	IP           string
	Hostname     string
	MAC          string
	Ports        []PortInfo
	DeviceType   string
	DeviceVendor string
	GuessOS      string
}

// PortInfo информация о порте
type PortInfo struct {
	Port     int
	State    string
	Protocol string
	Service  string
	Banner   string
	Version  string
}

// ScannerService интерфейс для сканирования
type ScannerService interface {
	Scan(ctx context.Context, cfg ScanConfig, onProgress ProgressHandler) ([]ScanResult, error)
	Stop()
}

// ProgressHandler обработчик прогресса сканирования
type ProgressHandler func(stage string, current, total int, message string)

// TopologyOptions опции построения топологии
type TopologyOptions struct {
	SNMPEnabled     bool
	Community       string
	Timeout         time.Duration
	PartialSNMP     map[string]struct{}
}

// TopologyService интерфейс для топологии
type TopologyService interface {
	Build(ctx context.Context, results []ScanResult, opts TopologyOptions) (*Topology, error)
	Export(t *Topology, format string, path string) error
}

// Topology модель топологии
type Topology struct {
	Devices []*Device
	Links   []*Link
}

// Device устройство в топологии
type Device struct {
	IP       string
	Hostname string
	MAC      string
	Type     string
}

// Link связь между устройствами
type Link struct {
	Source     *Device
	SourcePort string
	Target     *Device
	TargetPort string
	SourceType string // lldp|fdb|inferred
	Confidence string // high|medium|low
	Evidence   string
}

// SecurityReport отчёт безопасности
type SecurityReport struct {
	PortAudit   []Finding
	RiskSig     []Finding
	CVEs        []CVE
	Score       int
}

// Finding finding безопасности
type Finding struct {
	Severity       string
	Host           string
	Title          string
	Recommendation string
}

// CVE совпадение с уязвимостью
type CVE struct {
	ID    string
	CVSS  float64
	Title string
}

// SecurityService интерфейс для безопасности
type SecurityService interface {
	AnalyzeRun(ctx context.Context, results []ScanResult) (*SecurityReport, error)
}

// RemoteExecRequest запрос удалённого выполнения
type RemoteExecRequest struct {
	Transport     string
	Target        string
	User          string
	Password      string
	Command       string
	Policy        PolicyConfig
	Consent       string
	DryRun        bool
	Timeout       time.Duration
}

// PolicyConfig конфигурация политики безопасности
type PolicyConfig struct {
	FilePath      string
	Strict        bool
	AllowHosts    []string
	AllowCommands []string
}

// RemoteExecResponse ответ удалённого выполнения
type RemoteExecResponse struct {
	Output  string
	Success bool
}

// RemoteExecService интерфейс для удалённого выполнения
type RemoteExecService interface {
	Execute(ctx context.Context, req RemoteExecRequest) (RemoteExecResponse, error)
	DryRun(ctx context.Context, req RemoteExecRequest) error
}

// InventoryService интерфейс для инвентаризации
type InventoryService interface {
	SaveSnapshot(ctx context.Context, id string, data []ScanResult) error
	ListSnapshots(ctx context.Context, limit int) ([]Snapshot, error)
	Diff(ctx context.Context, idA, idB string) (*Diff, error)
}

// Snapshot снапшот инвентаризации
type Snapshot struct {
	ID        string
	Timestamp time.Time
}

// Diff разница между снапшотами
type Diff struct {
	ScanIDA string
	ScanIDB string
	New     []ScanResult
	Missing []ScanResult
	Changed []Change
}

// Change изменённое поле в снапшоте
type Change struct {
	Key          string
	ChangedField []string
}
