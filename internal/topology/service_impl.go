package topology

import (
	"context"
	"fmt"
	"os"
	"strings"

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
	if t == nil {
		return fmt.Errorf("topology is nil")
	}
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("export path is empty")
	}

	// Конвертируем contracts.Topology во внутренний Topology
	internalTopo := convertFromContractTopology(t)
	if internalTopo == nil {
		return fmt.Errorf("failed to convert topology")
	}

	// Экспортируем в зависимости от формата
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return internalTopo.SaveJSON(path)
	case "graphml", "xml":
		return internalTopo.SaveGraphML(path)
	case "dot":
		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("create file: %w", err)
		}
		defer f.Close()
		return internalTopo.ToDOT(f)
	case "text", "txt":
		return internalTopo.SaveAsText(path)
	default:
		return fmt.Errorf("unsupported export format: %s (supported: json, graphml, dot, text)", format)
	}
}

// convertFromContractTopology конвертирует contracts.Topology во внутренний Topology
func convertFromContractTopology(t *contracts.Topology) *Topology {
	if t == nil {
		return nil
	}

	topo := &Topology{
		Devices: make(map[string]*Device),
		Links:   make([]Link, 0, len(t.Links)),
	}

	// Создаём устройства
	for _, d := range t.Devices {
		if d == nil {
			continue
		}
		key := d.IP
		if key == "" {
			key = d.Hostname
		}
		if key == "" {
			key = d.MAC
		}
		if key == "" {
			key = "unknown"
		}

		dt := DeviceTypeUnknown
		switch strings.ToLower(d.Type) {
		case "router":
			dt = DeviceTypeRouter
		case "switch", "network":
			dt = DeviceTypeSwitch
		case "host", "server", "computer":
			dt = DeviceTypeHost
		}

		topo.Devices[key] = &Device{
			IP:       d.IP,
			MAC:      d.MAC,
			Hostname: d.Hostname,
			Type:     dt,
		}
	}

	// Создаём связи
	for _, l := range t.Links {
		if l == nil || l.Source == nil || l.Target == nil {
			continue
		}

		srcKey := l.Source.IP
		if srcKey == "" {
			srcKey = l.Source.Hostname
		}
		if srcKey == "" {
			srcKey = l.Source.MAC
		}
		if srcKey == "" {
			srcKey = "unknown"
		}

		dstKey := l.Target.IP
		if dstKey == "" {
			dstKey = l.Target.Hostname
		}
		if dstKey == "" {
			dstKey = l.Target.MAC
		}
		if dstKey == "" {
			dstKey = "unknown"
		}

		srcDev := topo.Devices[srcKey]
		dstDev := topo.Devices[dstKey]
		if srcDev == nil {
			srcDev = &Device{IP: l.Source.IP, Hostname: l.Source.Hostname, MAC: l.Source.MAC, Type: DeviceTypeUnknown}
			topo.Devices[srcKey] = srcDev
		}
		if dstDev == nil {
			dstDev = &Device{IP: l.Target.IP, Hostname: l.Target.Hostname, MAC: l.Target.MAC, Type: DeviceTypeUnknown}
			topo.Devices[dstKey] = dstDev
		}

		srcPort := ensurePort(srcDev, 0, l.SourcePort)
		dstPort := ensurePort(dstDev, 0, l.TargetPort)

		sourceType := LinkSourceInferred
		switch strings.ToLower(l.SourceType) {
		case "lldp":
			sourceType = LinkSourceLLDP
		case "fdb":
			sourceType = LinkSourceFDB
		}

		confidence := LinkConfidenceLow
		switch strings.ToLower(l.Confidence) {
		case "high":
			confidence = LinkConfidenceHigh
		case "medium":
			confidence = LinkConfidenceMedium
		}

		topo.Links = append(topo.Links, Link{
			Source:     srcDev,
			SourcePort: srcPort,
			Target:     dstDev,
			TargetPort: dstPort,
			SourceType: sourceType,
			Confidence: confidence,
			Evidence:   l.Evidence,
		})
	}

	return topo
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
