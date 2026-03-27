package snmpcollector

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gosnmp/gosnmp"

	"network-scanner/internal/scanner"
	"network-scanner/internal/topology"
)

const (
	oidSysName    = ".1.3.6.1.2.1.1.5.0"
	oidSysDescr   = ".1.3.6.1.2.1.1.1.0"
	oidIfDescr    = ".1.3.6.1.2.1.2.2.1.2"
	oidIfName     = ".1.3.6.1.2.1.31.1.1.1.1"
	oidDot1dTpFdb = ".1.3.6.1.2.1.17.4.3.1.2"
	// Q-BRIDGE-MIB dot1qTpFdbPort — VLAN-aware FDB (часто заполнена, когда dot1d пуста).
	oidDot1qTpFdb = ".1.3.6.1.2.1.17.7.1.2.2.1.2"

	oidLldpRemSysName = ".1.0.8802.1.1.2.1.4.1.1.9"
	oidLldpRemPortID  = ".1.0.8802.1.1.2.1.4.1.1.7"
	oidLldpRemChassis = ".1.0.8802.1.1.2.1.4.1.1.5"
)

type IfEntry struct {
	Index       int
	Name        string
	Description string
}

type LldpNeighbor struct {
	LocalIfIndex int
	RemotePortID string
	RemoteSys    string
	RemoteMac    string
}

type FailureKind string

const (
	FailureConnect FailureKind = "connect_error"
	FailureQuery   FailureKind = "query_error"
)

type DeviceFailure struct {
	IP        string
	Kind      FailureKind
	Message   string
	Community string
}

type CollectReport struct {
	TotalSNMPTargets int
	Connected        int
	Partial          int
	Failed           int
	Failures         []DeviceFailure
	// DeviceSummaries — краткая диагностика по каждому опрошенному IP (порядок по IP).
	DeviceSummaries []DeviceQuerySummary
}

// DeviceQuerySummary помогает понять, почему топология пустая: FDB/LLDP и текст ошибок.
type DeviceQuerySummary struct {
	IP              string
	MACEntries      int
	LLDPNeighbors   int
	QueryErrors     string
}

type ProgressCallback func(current int, total int, ip string, message string)

type SNMPClient interface {
	Connect(ip, community string) error
	Close() error
	GetSysName() (string, error)
	GetSysDescr() (string, error)
	GetIfTable() (map[int]*IfEntry, error)
	GetMacTable() (map[string]int, error)
	GetLldpNeighbors() ([]*LldpNeighbor, error)
}

type GoSNMPClient struct {
	client  *gosnmp.GoSNMP
	timeout time.Duration
}

func NewGoSNMPClient(timeoutSeconds int) *GoSNMPClient {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 2
	}
	return &GoSNMPClient{timeout: time.Duration(timeoutSeconds) * time.Second}
}

func (g *GoSNMPClient) Connect(ip, community string) error {
	c := &gosnmp.GoSNMP{
		Target:    ip,
		Port:      161,
		Community: community,
		Version:   gosnmp.Version2c,
		Timeout:   g.timeout,
		Retries:   2,
	}
	if err := c.Connect(); err != nil {
		return err
	}
	g.client = c
	return nil
}

func (g *GoSNMPClient) Close() error {
	if g.client != nil && g.client.Conn != nil {
		return g.client.Conn.Close()
	}
	return nil
}

func (g *GoSNMPClient) GetSysName() (string, error) {
	val, err := g.getAsString(oidSysName)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(val), nil
}

func (g *GoSNMPClient) GetSysDescr() (string, error) {
	val, err := g.getAsString(oidSysDescr)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(val), nil
}

func (g *GoSNMPClient) GetIfTable() (map[int]*IfEntry, error) {
	out := make(map[int]*IfEntry)
	if err := g.walk(oidIfDescr, func(pdu gosnmp.SnmpPDU) error {
		idx := suffixInt(pdu.Name)
		if idx <= 0 {
			return nil
		}
		entry := out[idx]
		if entry == nil {
			entry = &IfEntry{Index: idx}
			out[idx] = entry
		}
		entry.Description = pduValueString(pdu)
		return nil
	}); err != nil {
		return nil, err
	}
	_ = g.walk(oidIfName, func(pdu gosnmp.SnmpPDU) error {
		idx := suffixInt(pdu.Name)
		if idx <= 0 {
			return nil
		}
		entry := out[idx]
		if entry == nil {
			entry = &IfEntry{Index: idx}
			out[idx] = entry
		}
		entry.Name = pduValueString(pdu)
		return nil
	})
	return out, nil
}

func (g *GoSNMPClient) GetMacTable() (map[string]int, error) {
	out := make(map[string]int)
	errDot1d := g.walk(oidDot1dTpFdb, func(pdu gosnmp.SnmpPDU) error {
		mac, parseErr := ParseMACFromOID(pdu.Name)
		if parseErr != nil {
			return nil
		}
		switch v := pdu.Value.(type) {
		case int:
			out[mac] = v
		case uint:
			out[mac] = int(v)
		case uint32:
			out[mac] = int(v)
		case int64:
			out[mac] = int(v)
		case uint64:
			out[mac] = int(v)
		}
		return nil
	})
	errDot1q := g.walk(oidDot1qTpFdb, func(pdu gosnmp.SnmpPDU) error {
		mac, parseErr := ParseMACFromOID(pdu.Name)
		if parseErr != nil {
			return nil
		}
		switch v := pdu.Value.(type) {
		case int:
			if _, ok := out[mac]; !ok {
				out[mac] = v
			}
		case uint:
			if _, ok := out[mac]; !ok {
				out[mac] = int(v)
			}
		case uint32:
			if _, ok := out[mac]; !ok {
				out[mac] = int(v)
			}
		case int64:
			if _, ok := out[mac]; !ok {
				out[mac] = int(v)
			}
		case uint64:
			if _, ok := out[mac]; !ok {
				out[mac] = int(v)
			}
		}
		return nil
	})
	if len(out) > 0 {
		return out, nil
	}
	if errDot1d != nil {
		return out, errDot1d
	}
	return out, errDot1q
}

func (g *GoSNMPClient) GetLldpNeighbors() ([]*LldpNeighbor, error) {
	rows := make(map[string]*LldpNeighbor)
	if err := g.walk(oidLldpRemSysName, func(pdu gosnmp.SnmpPDU) error {
		key := lldpRowKeyFromOID(pdu.Name)
		if key == "" {
			return nil
		}
		lp := lldpLocalPortFromOID(pdu.Name)
		if lp <= 0 {
			return nil
		}
		n := rows[key]
		if n == nil {
			n = &LldpNeighbor{LocalIfIndex: lp}
			rows[key] = n
		}
		n.RemoteSys = pduValueString(pdu)
		return nil
	}); err != nil {
		return nil, err
	}
	_ = g.walk(oidLldpRemPortID, func(pdu gosnmp.SnmpPDU) error {
		key := lldpRowKeyFromOID(pdu.Name)
		if key == "" {
			return nil
		}
		lp := lldpLocalPortFromOID(pdu.Name)
		n := rows[key]
		if n == nil {
			n = &LldpNeighbor{LocalIfIndex: lp}
			rows[key] = n
		}
		n.RemotePortID = pduValueString(pdu)
		return nil
	})
	_ = g.walk(oidLldpRemChassis, func(pdu gosnmp.SnmpPDU) error {
		key := lldpRowKeyFromOID(pdu.Name)
		if key == "" {
			return nil
		}
		lp := lldpLocalPortFromOID(pdu.Name)
		n := rows[key]
		if n == nil {
			n = &LldpNeighbor{LocalIfIndex: lp}
			rows[key] = n
		}
		n.RemoteMac = lldpChassisToMACString(pdu)
		return nil
	})
	out := make([]*LldpNeighbor, 0, len(rows))
	for _, n := range rows {
		out = append(out, n)
	}
	return out, nil
}

func (g *GoSNMPClient) getAsString(oid string) (string, error) {
	if g.client == nil {
		return "", fmt.Errorf("not connected")
	}
	packet, err := g.client.Get([]string{oid})
	if err != nil {
		return "", err
	}
	if len(packet.Variables) == 0 {
		return "", fmt.Errorf("empty response for %s", oid)
	}
	return pduValueString(packet.Variables[0]), nil
}

func (g *GoSNMPClient) walk(oid string, fn gosnmp.WalkFunc) error {
	if g.client == nil {
		return fmt.Errorf("not connected")
	}
	return g.client.BulkWalk(oid, fn)
}

func Collect(devices []scanner.Result, communities []string, timeout int) (map[string]*topology.Device, error) {
	data, _, err := CollectWithReport(devices, communities, timeout)
	return data, err
}

func CollectWithReport(devices []scanner.Result, communities []string, timeout int) (map[string]*topology.Device, *CollectReport, error) {
	return CollectWithReportProgressContext(context.Background(), devices, communities, timeout, nil)
}

func CollectWithReportProgress(devices []scanner.Result, communities []string, timeout int, progress ProgressCallback) (map[string]*topology.Device, *CollectReport, error) {
	return CollectWithReportProgressContext(context.Background(), devices, communities, timeout, progress)
}

func CollectWithReportProgressContext(ctx context.Context, devices []scanner.Result, communities []string, timeout int, progress ProgressCallback) (map[string]*topology.Device, *CollectReport, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(communities) == 0 {
		communities = []string{"public"}
	}
	out := make(map[string]*topology.Device)
	report := &CollectReport{
		Failures: make([]DeviceFailure, 0),
	}

	targets := make([]scanner.Result, 0, len(devices))
	for _, d := range devices {
		if d.SNMPEnabled {
			targets = append(targets, d)
		}
	}
	report.TotalSNMPTargets = len(targets)
	if len(targets) == 0 {
		return out, report, nil
	}

	workerCount := runtime.NumCPU()
	if workerCount < 4 {
		workerCount = 4
	}
	if workerCount > 16 {
		workerCount = 16
	}
	if workerCount > len(targets) {
		workerCount = len(targets)
	}

	jobs := make(chan scanner.Result, len(targets))
	var wg sync.WaitGroup
	var mu sync.Mutex
	processed := 0

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for d := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
				}
				var connected bool
				var connectErrs []string
				var queryErrs []string

				for _, community := range communities {
					select {
					case <-ctx.Done():
						return
					default:
					}
					c := NewGoSNMPClient(timeout)
					trimmedCommunity := strings.TrimSpace(community)
					if err := c.Connect(d.IP, trimmedCommunity); err != nil {
						connectErrs = append(connectErrs, fmt.Sprintf("%s: %v", trimmedCommunity, err))
						_ = c.Close()
						continue
					}

					connected = true
					sysName, errSysName := c.GetSysName()
					if errSysName != nil {
						queryErrs = append(queryErrs, "sysName: "+errSysName.Error())
					}
					sysDescr, errSysDescr := c.GetSysDescr()
					if errSysDescr != nil {
						queryErrs = append(queryErrs, "sysDescr: "+errSysDescr.Error())
					}
					ifTable, errIfTable := c.GetIfTable()
					if errIfTable != nil {
						queryErrs = append(queryErrs, "ifTable: "+errIfTable.Error())
						ifTable = map[int]*IfEntry{}
					}
					macTable, errMacTable := c.GetMacTable()
					if errMacTable != nil {
						queryErrs = append(queryErrs, "macTable: "+errMacTable.Error())
						macTable = map[string]int{}
					}
					lldpList, errLLDP := c.GetLldpNeighbors()
					if errLLDP != nil {
						queryErrs = append(queryErrs, "lldp: "+errLLDP.Error())
						lldpList = nil
					}
					_ = c.Close()

					dev := &topology.Device{
						IP:            d.IP,
						MAC:           strings.ToLower(strings.ReplaceAll(d.MAC, "-", ":")),
						Hostname:      sysName,
						Type:          inferDeviceType(sysDescr, len(macTable) > 0),
						SNMPEnabled:   true,
						SNMPCommunity: trimmedCommunity,
						Ports:         make([]topology.Port, 0, len(ifTable)),
						MacTable:      macTable,
						LldpNeighbors: make([]*topology.LldpNeighbor, 0, len(lldpList)),
					}
					for idx, ifEntry := range ifTable {
						dev.Ports = append(dev.Ports, topology.Port{
							Index:       idx,
							Name:        ifEntry.Name,
							Description: ifEntry.Description,
						})
					}
					for _, n := range lldpList {
						if n == nil {
							continue
						}
						dev.LldpNeighbors = append(dev.LldpNeighbors, &topology.LldpNeighbor{
							LocalIfIndex:    n.LocalIfIndex,
							RemoteChassisID: n.RemoteMac,
							RemotePortID:    n.RemotePortID,
							RemoteSysName:   n.RemoteSys,
						})
					}

					key := dev.MAC
					if key == "" {
						key = dev.IP
					}

					summary := DeviceQuerySummary{
						IP:            d.IP,
						MACEntries:    len(macTable),
						LLDPNeighbors: len(dev.LldpNeighbors),
						QueryErrors:   strings.Join(queryErrs, "; "),
					}

					mu.Lock()
					out[key] = dev
					report.Connected++
					report.DeviceSummaries = append(report.DeviceSummaries, summary)
					if len(queryErrs) > 0 {
						report.Partial++
						report.Failures = append(report.Failures, DeviceFailure{
							IP:        d.IP,
							Kind:      FailureQuery,
							Message:   strings.Join(queryErrs, "; "),
							Community: trimmedCommunity,
						})
					}
					mu.Unlock()
					break
				}

				if connected {
					if progress != nil {
						mu.Lock()
						processed++
						current := processed
						total := report.TotalSNMPTargets
						mu.Unlock()
						progress(current, total, d.IP, "SNMP опрос завершен")
					}
					continue
				}

				msg := "unable to connect using provided communities"
				if len(connectErrs) > 0 {
					msg = strings.Join(connectErrs, "; ")
				}
				mu.Lock()
				report.Failed++
				report.Failures = append(report.Failures, DeviceFailure{
					IP:      d.IP,
					Kind:    FailureConnect,
					Message: msg,
				})
				processed++
				current := processed
				total := report.TotalSNMPTargets
				mu.Unlock()
				if progress != nil {
					progress(current, total, d.IP, "SNMP недоступен")
				}
			}
		}()
	}

	for _, d := range targets {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return out, report, ctx.Err()
		default:
		}
		jobs <- d
	}
	close(jobs)
	wg.Wait()
	if ctx.Err() != nil {
		return out, report, ctx.Err()
	}

	sort.Slice(report.DeviceSummaries, func(i, j int) bool {
		return report.DeviceSummaries[i].IP < report.DeviceSummaries[j].IP
	})

	return out, report, nil
}

func inferDeviceType(sysDescr string, hasDot1d bool) topology.DeviceType {
	d := strings.ToLower(sysDescr)
	switch {
	case strings.Contains(d, "switch") || hasDot1d:
		return topology.DeviceTypeSwitch
	case strings.Contains(d, "router"):
		return topology.DeviceTypeRouter
	case strings.Contains(d, "host"), strings.Contains(d, "server"), strings.Contains(d, "linux"), strings.Contains(d, "windows"):
		return topology.DeviceTypeHost
	default:
		return topology.DeviceTypeUnknown
	}
}

func ParseMACFromOID(oid string) (string, error) {
	parts := strings.Split(strings.TrimPrefix(strings.TrimSpace(oid), "."), ".")
	if len(parts) < 6 {
		return "", fmt.Errorf("oid is too short for mac suffix")
	}
	suffix := parts[len(parts)-6:]
	out := make([]string, 0, 6)
	for _, s := range suffix {
		n, err := strconv.Atoi(s)
		if err != nil || n < 0 || n > 255 {
			return "", fmt.Errorf("invalid mac suffix")
		}
		out = append(out, fmt.Sprintf("%02x", n))
	}
	return strings.Join(out, ":"), nil
}

func suffixInt(oid string) int {
	parts := strings.Split(strings.TrimPrefix(strings.TrimSpace(oid), "."), ".")
	if len(parts) == 0 {
		return -1
	}
	n, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return -1
	}
	return n
}

// lldpRowKeyFromOID — уникальный ключ строки lldpRemTable (timeMark, localPortNum, remIndex).
func lldpRowKeyFromOID(oid string) string {
	parts := strings.Split(strings.TrimPrefix(strings.TrimSpace(oid), "."), ".")
	if len(parts) < 3 {
		return ""
	}
	return strings.Join(parts[len(parts)-3:], ".")
}

// lldpLocalPortFromOID — lldpRemLocalPortNum (не обязательно совпадает с ifIndex).
func lldpLocalPortFromOID(oid string) int {
	parts := strings.Split(strings.TrimPrefix(strings.TrimSpace(oid), "."), ".")
	if len(parts) < 3 {
		return -1
	}
	n, err := strconv.Atoi(parts[len(parts)-2])
	if err != nil {
		return -1
	}
	return n
}

func lldpChassisToMACString(pdu gosnmp.SnmpPDU) string {
	switch v := pdu.Value.(type) {
	case []byte:
		if len(v) == 6 {
			return bytesToMAC(v)
		}
		s := strings.TrimSpace(string(v))
		return strings.ToLower(strings.ReplaceAll(s, "-", ":"))
	default:
		s := pduValueString(pdu)
		return strings.ToLower(strings.ReplaceAll(s, "-", ":"))
	}
}

func bytesToMAC(b []byte) string {
	parts := make([]string, len(b))
	for i := range b {
		parts[i] = fmt.Sprintf("%02x", b[i])
	}
	return strings.Join(parts, ":")
}

func pduValueString(pdu gosnmp.SnmpPDU) string {
	switch v := pdu.Value.(type) {
	case []byte:
		return strings.TrimSpace(string(v))
	case string:
		return strings.TrimSpace(v)
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}
