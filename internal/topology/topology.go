package topology

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"network-scanner/internal/scanner"
)

type DeviceType string

const (
	DeviceTypeSwitch  DeviceType = "switch"
	DeviceTypeRouter  DeviceType = "router"
	DeviceTypeHost    DeviceType = "host"
	DeviceTypeUnknown DeviceType = "unknown"
)

type Port struct {
	Index            int
	Name             string
	Description      string
	Neighbor         *Device
	NeighborPort     string
	ConnectedDevices []*Device
}

type Device struct {
	IP            string
	MAC           string
	Hostname      string
	Type          DeviceType
	SNMPEnabled   bool
	SNMPCommunity string
	Ports         []Port
	MacTable      map[string]int
	LldpNeighbors []*LldpNeighbor
}

type Link struct {
	Source     *Device
	SourcePort *Port
	Target     *Device
	TargetPort *Port
	SourceType LinkSourceType
	Confidence LinkConfidence
	Evidence   string
}

type Topology struct {
	Devices map[string]*Device
	Links   []Link
}

type LinkSourceType string

const (
	LinkSourceLLDP     LinkSourceType = "lldp"
	LinkSourceFDB      LinkSourceType = "fdb"
	LinkSourceInferred LinkSourceType = "inferred"
)

type LinkConfidence string

const (
	LinkConfidenceHigh   LinkConfidence = "high"
	LinkConfidenceMedium LinkConfidence = "medium"
	LinkConfidenceLow    LinkConfidence = "low"
)

type LldpNeighbor struct {
	LocalIfIndex    int
	RemoteChassisID string
	RemotePortID    string
	RemotePortDescr string
	RemoteSysName   string
}

func BuildTopology(results []scanner.Result, snmpData map[string]*Device) (*Topology, error) {
	t := &Topology{
		Devices: make(map[string]*Device),
		Links:   make([]Link, 0),
	}

	byIP := make(map[string]*Device)
	byMAC := make(map[string]*Device)
	byHostname := make(map[string]*Device)

	for _, r := range results {
		key := normalizedKey(r.MAC, r.IP)
		dev := &Device{
			IP:          r.IP,
			MAC:         normalizeMAC(r.MAC),
			Hostname:    strings.TrimSpace(r.Hostname),
			Type:        classifyFromScannerResult(r.DeviceType),
			SNMPEnabled: r.SNMPEnabled,
		}
		t.Devices[key] = dev
		if dev.IP != "" {
			byIP[dev.IP] = dev
		}
		if dev.MAC != "" {
			byMAC[dev.MAC] = dev
		}
		if dev.Hostname != "" {
			byHostname[strings.ToLower(dev.Hostname)] = dev
		}
	}

	for _, d := range snmpData {
		if d == nil {
			continue
		}
		mac := normalizeMAC(d.MAC)
		target := (*Device)(nil)
		if mac != "" {
			target = byMAC[mac]
		}
		if target == nil && d.IP != "" {
			target = byIP[d.IP]
		}
		if target == nil && d.Hostname != "" {
			target = byHostname[strings.ToLower(d.Hostname)]
		}
		if target == nil {
			key := normalizedKey(mac, d.IP)
			copyDev := *d
			copyDev.MAC = mac
			t.Devices[key] = &copyDev
			target = &copyDev
			if copyDev.IP != "" {
				byIP[copyDev.IP] = target
			}
			if copyDev.MAC != "" {
				byMAC[copyDev.MAC] = target
			}
			if copyDev.Hostname != "" {
				byHostname[strings.ToLower(copyDev.Hostname)] = target
			}
		} else {
			target.SNMPEnabled = true
			if target.Hostname == "" {
				target.Hostname = d.Hostname
			}
			if target.Type == DeviceTypeUnknown && d.Type != DeviceTypeUnknown {
				target.Type = d.Type
			}
			target.Ports = d.Ports
			target.MacTable = d.MacTable
			target.LldpNeighbors = d.LldpNeighbors
		}
	}

	linkDedup := make(map[string]int)
	linkByEndpoint := make(map[string]int)
	for _, dev := range t.Devices {
		if !dev.SNMPEnabled {
			continue
		}
		// LLDP links
		for _, n := range dev.LldpNeighbors {
			if n == nil {
				continue
			}
			localIf := n.LocalIfIndex
			remote := findNeighbor(byMAC, byHostname, n)
			if remote == nil || remote == dev {
				continue
			}
			addLink(linkDedup, linkByEndpoint, t, dev, localIf, "", remote, -1, n.RemotePortID, LinkSourceLLDP, LinkConfidenceHigh, "lldp_neighbor_match")
		}
		// FDB/MAC links
		for mac, ifIndex := range dev.MacTable {
			normalized := normalizeMAC(mac)
			if normalized == "" || isBroadcastOrMulticast(normalized) || isZeroMAC(normalized) || normalized == dev.MAC {
				continue
			}
			remote := byMAC[normalized]
			if remote == nil {
				remote = &Device{MAC: normalized, Type: DeviceTypeHost}
				t.Devices[normalized] = remote
				byMAC[normalized] = remote
			}
			if remote == dev {
				continue
			}
			addLink(linkDedup, linkByEndpoint, t, dev, ifIndex, "", remote, -1, "", LinkSourceFDB, LinkConfidenceMedium, "fdb_mac_match")
		}
	}

	sort.Slice(t.Links, func(i, j int) bool {
		a := t.Links[i]
		b := t.Links[j]
		aKey := linkKey(
			deviceDisplayName(a.Source),
			portLabel(a.SourcePort),
			deviceDisplayName(a.Target),
			portLabel(a.TargetPort),
		)
		bKey := linkKey(
			deviceDisplayName(b.Source),
			portLabel(b.SourcePort),
			deviceDisplayName(b.Target),
			portLabel(b.TargetPort),
		)
		return aKey < bKey
	})

	return t, nil
}

func (t *Topology) ToDOT(w io.Writer) error {
	if t == nil {
		return fmt.Errorf("topology is nil")
	}
	_, _ = fmt.Fprintln(w, "graph network {")
	_, _ = fmt.Fprintln(w, `  rankdir="LR";`)
	_, _ = fmt.Fprintln(w, `  node [shape=box, style="rounded,filled", fillcolor="#eef4ff"];`)
	for _, d := range t.Devices {
		label := deviceDisplayName(d)
		_, _ = fmt.Fprintf(w, "  %q [label=%q];\n", nodeID(d), label)
	}
	for _, l := range t.Links {
		src := nodeID(l.Source)
		dst := nodeID(l.Target)
		edgeLabel := strings.TrimSpace(strings.Join([]string{
			portLabel(l.SourcePort),
			portLabel(l.TargetPort),
		}, " <> "))
		_, _ = fmt.Fprintf(w, "  %q -- %q [label=%q];\n", src, dst, edgeLabel)
	}
	_, _ = fmt.Fprintln(w, "}")
	return nil
}

func (t *Topology) SaveJSON(filename string) error {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal topology json: %w", err)
	}
	return os.WriteFile(filename, data, 0644)
}

func (t *Topology) SaveGraphML(filename string) error {
	type Data struct {
		Key   string `xml:"key,attr"`
		Value string `xml:",chardata"`
	}
	type Node struct {
		ID   string `xml:"id,attr"`
		Data []Data `xml:"data"`
	}
	type Edge struct {
		ID     string `xml:"id,attr"`
		Source string `xml:"source,attr"`
		Target string `xml:"target,attr"`
		Data   []Data `xml:"data"`
	}
	type Graph struct {
		XMLName xml.Name `xml:"graph"`
		ID      string   `xml:"id,attr"`
		EdgeDef string   `xml:"edgedefault,attr"`
		Nodes   []Node   `xml:"node"`
		Edges   []Edge   `xml:"edge"`
	}
	type GraphML struct {
		XMLName xml.Name `xml:"graphml"`
		Xmlns   string   `xml:"xmlns,attr"`
		Graph   Graph    `xml:"graph"`
	}

	g := Graph{ID: "network", EdgeDef: "undirected"}
	for _, d := range t.Devices {
		g.Nodes = append(g.Nodes, Node{
			ID: nodeID(d),
			Data: []Data{
				{Key: "label", Value: deviceDisplayName(d)},
				{Key: "type", Value: string(d.Type)},
			},
		})
	}
	for i, l := range t.Links {
		g.Edges = append(g.Edges, Edge{
			ID:     fmt.Sprintf("e%d", i+1),
			Source: nodeID(l.Source),
			Target: nodeID(l.Target),
			Data: []Data{
				{Key: "src_port", Value: portLabel(l.SourcePort)},
				{Key: "dst_port", Value: portLabel(l.TargetPort)},
				{Key: "source_type", Value: string(l.SourceType)},
				{Key: "confidence", Value: string(l.Confidence)},
			},
		})
	}
	raw, err := xml.MarshalIndent(GraphML{
		Xmlns: "http://graphml.graphdrawing.org/xmlns",
		Graph: g,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal graphml: %w", err)
	}
	return os.WriteFile(filename, append([]byte(xml.Header), raw...), 0644)
}

func (t *Topology) RenderWithGraphviz(outputFormat, outputFile string) error {
	dotPath, err := exec.LookPath("dot")
	if err != nil {
		return fmt.Errorf("graphviz dot не найден в PATH, установите Graphviz: %w", err)
	}
	tmp, err := os.CreateTemp("", "network-topology-*.dot")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if err = t.ToDOT(tmp); err != nil {
		return err
	}

	cmd := exec.Command(dotPath, "-T"+outputFormat, tmp.Name(), "-o", outputFile)
	if out, runErr := cmd.CombinedOutput(); runErr != nil {
		return fmt.Errorf("ошибка запуска dot: %w (%s)", runErr, string(out))
	}
	return nil
}

func addLink(
	dedup map[string]int,
	byEndpoint map[string]int,
	t *Topology,
	src *Device,
	srcIf int,
	srcPortName string,
	dst *Device,
	dstIf int,
	dstPortName string,
	sourceType LinkSourceType,
	confidence LinkConfidence,
	evidence string,
) {
	if src == nil || dst == nil {
		return
	}
	srcPort := ensurePort(src, srcIf, srcPortName)
	dstPort := ensurePort(dst, dstIf, dstPortName)

	key := linkKey(
		nodeID(src), portLabel(srcPort),
		nodeID(dst), portLabel(dstPort),
	)
	endpointKey := linkKey(nodeID(src), "", nodeID(dst), "")
	if existingIndex, ok := dedup[key]; ok {
		existing := t.Links[existingIndex]
		if confidenceRank(confidence) <= confidenceRank(existing.Confidence) {
			return
		}
		t.Links[existingIndex] = Link{
			Source:     src,
			SourcePort: srcPort,
			Target:     dst,
			TargetPort: dstPort,
			SourceType: sourceType,
			Confidence: confidence,
			Evidence:   evidence,
		}
		return
	}
	if existingIndex, ok := byEndpoint[endpointKey]; ok {
		existing := t.Links[existingIndex]
		existingSrcPort := strings.TrimSpace(portLabel(existing.SourcePort))
		existingDstPort := strings.TrimSpace(portLabel(existing.TargetPort))
		newSrcPort := strings.TrimSpace(portLabel(srcPort))
		newDstPort := strings.TrimSpace(portLabel(dstPort))
		existingHasFullPortPair := existingSrcPort != "" && existingDstPort != ""
		newHasFullPortPair := newSrcPort != "" && newDstPort != ""
		// Keep multiple links for the same devices only when both links have
		// explicit port info and the port pairs differ.
		if !(existingHasFullPortPair && newHasFullPortPair &&
			(existingSrcPort != newSrcPort || existingDstPort != newDstPort)) {
			if confidenceRank(confidence) <= confidenceRank(existing.Confidence) {
				return
			}
			t.Links[existingIndex] = Link{
				Source:     src,
				SourcePort: srcPort,
				Target:     dst,
				TargetPort: dstPort,
				SourceType: sourceType,
				Confidence: confidence,
				Evidence:   evidence,
			}
			dedup[key] = existingIndex
			return
		}
	}
	t.Links = append(t.Links, Link{
		Source:     src,
		SourcePort: srcPort,
		Target:     dst,
		TargetPort: dstPort,
		SourceType: sourceType,
		Confidence: confidence,
		Evidence:   evidence,
	})
	newIndex := len(t.Links) - 1
	dedup[key] = newIndex
	if _, exists := byEndpoint[endpointKey]; !exists {
		byEndpoint[endpointKey] = newIndex
	}
}

func ensurePort(d *Device, index int, name string) *Port {
	for i := range d.Ports {
		if (index > 0 && d.Ports[i].Index == index) || (name != "" && d.Ports[i].Name == name) {
			return &d.Ports[i]
		}
	}
	p := Port{Index: index, Name: name}
	if p.Name == "" && p.Index > 0 {
		p.Name = fmt.Sprintf("if%d", p.Index)
	}
	d.Ports = append(d.Ports, p)
	return &d.Ports[len(d.Ports)-1]
}

func findNeighbor(byMAC, byHostname map[string]*Device, n *LldpNeighbor) *Device {
	if n == nil {
		return nil
	}
	if m := normalizeMAC(n.RemoteChassisID); m != "" {
		if d := byMAC[m]; d != nil {
			return d
		}
	}
	if hn := strings.ToLower(strings.TrimSpace(n.RemoteSysName)); hn != "" {
		if d := byHostname[hn]; d != nil {
			return d
		}
	}
	return nil
}

func classifyFromScannerResult(s string) DeviceType {
	v := strings.ToLower(s)
	switch {
	case strings.Contains(v, "router"):
		return DeviceTypeRouter
	case strings.Contains(v, "switch"), strings.Contains(v, "network"):
		return DeviceTypeSwitch
	case strings.Contains(v, "server"), strings.Contains(v, "computer"), strings.Contains(v, "host"):
		return DeviceTypeHost
	default:
		return DeviceTypeUnknown
	}
}

func normalizedKey(mac, ip string) string {
	nm := normalizeMAC(mac)
	if nm != "" {
		return nm
	}
	return strings.TrimSpace(ip)
}

func normalizeMAC(mac string) string {
	m := strings.ToLower(strings.TrimSpace(mac))
	m = strings.ReplaceAll(m, "-", ":")
	parts := strings.Split(m, ":")
	if len(parts) == 6 {
		for i := range parts {
			if len(parts[i]) == 1 {
				parts[i] = "0" + parts[i]
			}
		}
		return strings.Join(parts, ":")
	}
	return ""
}

func isBroadcastOrMulticast(mac string) bool {
	b, ok := parseMACBytes(mac)
	if !ok {
		return false
	}
	allFF := true
	for _, v := range b {
		if v != 0xff {
			allFF = false
			break
		}
	}
	if allFF {
		return true
	}
	// I/G bit set means multicast.
	return (b[0] & 0x01) == 0x01
}

func isZeroMAC(mac string) bool {
	b, ok := parseMACBytes(mac)
	if !ok {
		return false
	}
	for _, v := range b {
		if v != 0x00 {
			return false
		}
	}
	return true
}

func parseMACBytes(mac string) ([6]byte, bool) {
	var out [6]byte
	parts := strings.Split(normalizeMAC(mac), ":")
	if len(parts) != 6 {
		return out, false
	}
	for i := 0; i < 6; i++ {
		v, err := strconv.ParseUint(parts[i], 16, 8)
		if err != nil {
			return out, false
		}
		out[i] = byte(v)
	}
	return out, true
}

func confidenceRank(c LinkConfidence) int {
	switch c {
	case LinkConfidenceHigh:
		return 3
	case LinkConfidenceMedium:
		return 2
	case LinkConfidenceLow:
		return 1
	default:
		return 0
	}
}

func nodeID(d *Device) string {
	if d == nil {
		return "unknown"
	}
	if d.MAC != "" {
		return "mac_" + strings.ReplaceAll(d.MAC, ":", "_")
	}
	if d.IP != "" {
		return "ip_" + strings.ReplaceAll(d.IP, ".", "_")
	}
	if d.Hostname != "" {
		return "hn_" + strings.ReplaceAll(strings.ToLower(d.Hostname), " ", "_")
	}
	return "unknown"
}

func deviceDisplayName(d *Device) string {
	if d == nil {
		return "unknown"
	}
	if d.Hostname != "" {
		return d.Hostname
	}
	if d.IP != "" {
		return d.IP
	}
	if d.MAC != "" {
		return d.MAC
	}
	return "unknown"
}

func portLabel(p *Port) string {
	if p == nil {
		return ""
	}
	if p.Name != "" {
		return p.Name
	}
	if p.Index > 0 {
		return fmt.Sprintf("if%d", p.Index)
	}
	return ""
}

func linkKey(aNode, aPort, bNode, bPort string) string {
	left := aNode + "|" + aPort
	right := bNode + "|" + bPort
	if left > right {
		left, right = right, left
	}
	return left + "<->" + right
}
