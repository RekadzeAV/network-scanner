package topology

import (
	"context"
	"encoding/xml"
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

// Export записывает топологию в файл в запрошенном формате.
//
// Поддерживаемые форматы: json, graphml, dot, text, txt, xml.
// Реализация переиспользует существующие внутренние Save*-методы:
// конвертация contracts -> internal выполняется один здесь раз, а не в
// каждом формате отдельно.
func (s *topologyServiceImpl) Export(t *contracts.Topology, format string, path string) error {
	if t == nil {
		return fmt.Errorf("topology is nil")
	}
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("export path is empty")
	}

	internal := convertFromContractTopology(t)

	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return internal.SaveJSON(path)
	case "graphml":
		return internal.SaveGraphML(path)
	case "dot":
		file, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("create file %s: %w", path, err)
		}
		if err := internal.ToDOT(file); err != nil {
			_ = file.Close()
			return err
		}
		return file.Close()
	case "text", "txt":
		return internal.SaveAsText(path)
	case "xml":
		data, err := marshalTopologyXML(internal)
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			return fmt.Errorf("create file %s: %w", path, err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// marshalTopologyXML сериализует топологию в простой XML-формат:
// topology>devices>device и topology>links>link.
func marshalTopologyXML(t *Topology) ([]byte, error) {
	if err := t.Validate(); err != nil {
		return nil, fmt.Errorf("topology validation failed: %w", err)
	}
	type xmlDevice struct {
		IP       string `xml:"ip,attr"`
		Hostname string `xml:"hostname,attr,omitempty"`
		Type     string `xml:"type,attr,omitempty"`
	}
	type xmlLink struct {
		Source     string `xml:"source,attr"`
		Target     string `xml:"target,attr"`
		SourcePort string `xml:"source_port,attr,omitempty"`
		TargetPort string `xml:"target_port,attr,omitempty"`
		Confidence string `xml:"confidence,attr,omitempty"`
	}
	type xmlTopology struct {
		XMLName xml.Name    `xml:"topology"`
		Devices []xmlDevice `xml:"devices>device"`
		Links   []xmlLink   `xml:"links>link"`
	}

	doc := xmlTopology{}
	for _, key := range sortedDeviceKeys(t.Devices) {
		d := t.Devices[key]
		doc.Devices = append(doc.Devices, xmlDevice{
			IP:       d.IP,
			Hostname: d.Hostname,
			Type:     string(d.Type),
		})
	}
	for _, l := range t.Links {
		doc.Links = append(doc.Links, xmlLink{
			Source:     nodeID(l.Source),
			Target:     nodeID(l.Target),
			SourcePort: portLabel(l.SourcePort),
			TargetPort: portLabel(l.TargetPort),
			Confidence: string(l.Confidence),
		})
	}

	raw, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal topology xml: %w", err)
	}
	return append([]byte(xml.Header), raw...), nil
}

// convertToContractTopology конвертирует internal Topology в contracts.Topology
func convertToContractTopology(t *Topology) *contracts.Topology {
	if t == nil {
		return nil
	}

	devices := make([]*contracts.Device, 0, len(t.Devices))
	for _, key := range sortedDeviceKeys(t.Devices) {
		d := t.Devices[key]
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

// convertFromContractTopology конвертирует contracts.Topology во внутренний
// формат Topology. Обратная операция к convertToContractTopology.
//
// Нужна для импорта топологии извне (API/файл): контракты оперируют строками и
// указателями на Device, внутренняя модель — типизированными значениями,
// картой устройств и объектами портов.
func convertFromContractTopology(t *contracts.Topology) *Topology {
	if t == nil {
		return nil
	}

	out := &Topology{
		Devices: make(map[string]*Device, len(t.Devices)),
		Links:   make([]Link, 0, len(t.Links)),
	}
	// Индекс по IP: концы связей в контрактах описаны устройствами, а не ключами.
	byIP := make(map[string]*Device, len(t.Devices))

	for _, cd := range t.Devices {
		if cd == nil {
			continue
		}
		d := &Device{
			IP:       strings.TrimSpace(cd.IP),
			MAC:      strings.TrimSpace(cd.MAC),
			Hostname: strings.TrimSpace(cd.Hostname),
			Type:     deviceTypeFromContract(cd.Type),
		}
		key := contractDeviceKey(cd)
		if _, exists := out.Devices[key]; !exists {
			out.Devices[key] = d
		}
		if d.IP != "" {
			if _, exists := byIP[d.IP]; !exists {
				byIP[d.IP] = d
			}
		}
	}

	for _, cl := range t.Links {
		// Связь без одного из концов бессмысленна — пропускаем, а не роняем импорт.
		if cl == nil || cl.Source == nil || cl.Target == nil {
			continue
		}
		src := resolveContractDevice(out.Devices, byIP, cl.Source)
		dst := resolveContractDevice(out.Devices, byIP, cl.Target)
		if src == nil || dst == nil {
			continue
		}

		link := Link{
			Source:     src,
			Target:     dst,
			SourceType: linkSourceTypeFromContract(cl.SourceType),
			Confidence: linkConfidenceFromContract(cl.Confidence),
			Evidence:   strings.TrimSpace(cl.Evidence),
		}
		if name := strings.TrimSpace(cl.SourcePort); name != "" {
			link.SourcePort = ensurePort(src, -1, name)
		}
		if name := strings.TrimSpace(cl.TargetPort); name != "" {
			link.TargetPort = ensurePort(dst, -1, name)
		}
		out.Links = append(out.Links, link)
	}

	return out
}

// contractDeviceKey — ключ устройства во внутренней карте.
// IP предпочтительнее hostname, hostname — MAC; пустое устройство идёт под "unknown".
func contractDeviceKey(cd *contracts.Device) string {
	if key := strings.TrimSpace(cd.IP); key != "" {
		return key
	}
	if key := strings.TrimSpace(cd.Hostname); key != "" {
		return key
	}
	if key := strings.TrimSpace(cd.MAC); key != "" {
		return key
	}
	return "unknown"
}

// resolveContractDevice находит внутреннее устройство по контрактному, создавая
// его при отсутствии: ссылка связи не должна теряться из-за того, что устройство
// не перечислено в списке Devices.
func resolveContractDevice(devices map[string]*Device, byIP map[string]*Device, cd *contracts.Device) *Device {
	if cd == nil {
		return nil
	}
	key := contractDeviceKey(cd)
	if d, ok := devices[key]; ok {
		return d
	}
	d := &Device{
		IP:       strings.TrimSpace(cd.IP),
		MAC:      strings.TrimSpace(cd.MAC),
		Hostname: strings.TrimSpace(cd.Hostname),
		Type:     deviceTypeFromContract(cd.Type),
	}
	devices[key] = d
	if d.IP != "" {
		byIP[d.IP] = d
	}
	return d
}

// deviceTypeFromContract отображает строковый тип устройства из контракта на
// внутренний тип. Неизвестные значения сводятся к DeviceTypeUnknown.
func deviceTypeFromContract(s string) DeviceType {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "router":
		return DeviceTypeRouter
	case "switch", "network":
		return DeviceTypeSwitch
	case "host", "server", "computer":
		return DeviceTypeHost
	default:
		return DeviceTypeUnknown
	}
}

// linkSourceTypeFromContract отображает источник связи; пустое значение — inferred.
func linkSourceTypeFromContract(s string) LinkSourceType {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "lldp":
		return LinkSourceLLDP
	case "fdb":
		return LinkSourceFDB
	default:
		return LinkSourceInferred
	}
}

// linkConfidenceFromContract отображает достоверность связи; пустое значение — low.
func linkConfidenceFromContract(s string) LinkConfidence {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "high":
		return LinkConfidenceHigh
	case "medium":
		return LinkConfidenceMedium
	default:
		return LinkConfidenceLow
	}
}
