package scanner

import "net"

// HostResult is a backward-compatible alias for scan results.
type HostResult = Result

// NetworkProber defines host liveness and MAC resolution operations.
type NetworkProber interface {
	Ping(ip string) (bool, error)
	ResolveMAC(ip string) (net.HardwareAddr, error)
}

// ContextNetworkProber extends NetworkProber with context-aware liveness probe.
// Implementers should honor cancellation to stop long-running checks quickly.
type ContextNetworkProber interface {
	NetworkProber
	PingContext(ip string, done <-chan struct{}) (bool, error)
}

// PortScanner scans ports for a given host and protocol.
type PortScanner interface {
	ScanPort(ip string, port int, proto string) (bool, error)
	ScanPorts(ip string, ports []int, proto string) ([]int, error)
}

// ResultPresenter displays and exports scan results.
type ResultPresenter interface {
	DisplayHeader()
	DisplayHost(host HostResult)
	DisplaySummary(totalHosts int, openPortsCount int)
	Export(results []HostResult, format string) error
}

// SNMPCollector abstracts SNMP collection.
type SNMPCollector interface {
	Collect(hosts []HostResult) error
}

// TopologyBuilder abstracts topology generation.
type TopologyBuilder interface {
	Build(hosts []HostResult) error
}
