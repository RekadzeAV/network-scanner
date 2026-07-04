package topology

import (
	"context"
	"fmt"

	"network-scanner/internal/contracts"
	"network-scanner/internal/scanner"
)

// topologyServiceImpl реализация TopologyService
type topologyServiceImpl struct{}

// NewService создаёт TopologyService
func NewService() contracts.TopologyService {
	return &topologyServiceImpl{}
}

func (s *topologyServiceImpl) Build(ctx context.Context, results []contracts.ScanResult, opts contracts.TopologyOptions) (*contracts.Topology, error) {
	// Преобразуем результаты во внутренний формат
	internalResults := make([]scanner.Result, 0, len(results))
	for _, r := range results {
		ports := make([]scanner.PortInfo, 0, len(r.Ports))
		for _, p := range r.Ports {
			ports = append(ports, scanner.PortInfo{
				Port:     p.Port,
				State:    p.State,
				Protocol: p.Protocol,
				Service:  p.Service,
				Banner:   p.Banner,
				Version:  p.Version,
			})
		}

		internalResults = append(internalResults, scanner.Result{
			IP:           r.IP,
			Hostname:     r.Hostname,
			MAC:          r.MAC,
			Ports:        ports,
			DeviceType:   r.DeviceType,
			DeviceVendor: r.DeviceVendor,
			GuessOS:      r.GuessOS,
		})
	}

	// Построение топологии без SNMP (упрощённый режим)
	topo, err := BuildTopology(internalResults, nil)
	if err != nil {
		return nil, fmt.Errorf("построение топологии: %w", err)
	}

	// Конвертируем в contracts.Topology
	return convertToContractTopology(topo), nil
}

func (s *topologyServiceImpl) Export(t *contracts.Topology, format string, path string) error {
	// TODO: реализация экспорта
	return nil
}

// convertToContractTopology конвертирует internal Topology в contracts.Topology
func convertToContractTopology(t *Topology) *contracts.Topology {
	if t == nil {
		return nil
	}

	devices := make([]*contracts.Device, 0, len(t.Devices))
	for _, d := range t.Devices {
		devices = append(devices, &contracts.Device{
			IP:       d.IP,
			Hostname: d.Hostname,
			Type:     string(d.Type),
		})
	}

	links := make([]*contracts.Link, 0, len(t.Links))
	for _, l := range t.Links {
		src := convertToDevice(l.Source)
		dst := convertToDevice(l.Target)
		links = append(links, &contracts.Link{
			Source:     src,
			SourcePort: portLabel(l.SourcePort),
			Target:     dst,
			TargetPort: portLabel(l.TargetPort),
			Confidence: string(l.Confidence),
		})
	}

	return &contracts.Topology{
		Devices: devices,
		Links:   links,
	}
}

func convertToDevice(d *Device) *contracts.Device {
	if d == nil {
		return nil
	}
	return &contracts.Device{
		IP:       d.IP,
		Hostname: d.Hostname,
		Type:     string(d.Type),
	}
}
