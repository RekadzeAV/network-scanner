package snmpcollector

import (
	"context"
	"fmt"
	"runtime"
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
}

type ProgressCallback func(current int, total int, ip string, message string)

type SNMPClient interface {
	Connect(ip, community string) error
	Close() error
	GetSysName() (string, error)
	GetSysDescr() (string, error)
	GetIfTable() (map[int]*IfEntry, error)
	GetMacTable() (map[string]int, error)
	GetLldpNeighbors() (map[int]*LldpNeighbor, error)
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
		Retries:   0,
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
	err := g.walk(oidDot1dTpFdb, func(pdu gosnmp.SnmpPDU) error {
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
	return out, err
}

func (g *GoSNMPClient) GetLldpNeighbors() (map[int]*LldpNeighbor, error) {
	out := make(map[int]*LldpNeighbor)
	if err := g.walk(oidLldpRemSysName, func(pdu gosnmp.SnmpPDU) error {
		localIdx := lldpLocalIfIndexFromOID(pdu.Name)
		if localIdx <= 0 {
			return nil
		}
		n := out[localIdx]
		if n == nil {
			n = &LldpNeighbor{LocalIfIndex: localIdx}
			out[localIdx] = n
		}
		n.RemoteSys = pduValueString(pdu)
		return nil
	}); err != nil {
		return nil, err
	}
	_ = g.walk(oidLldpRemPortID, func(pdu gosnmp.SnmpPDU) error {
		localIdx := lldpLocalIfIndexFromOID(pdu.Name)
		if localIdx <= 0 {
			return nil
		}
		n := out[localIdx]
		if n == nil {
			n = &LldpNeighbor{LocalIfIndex: localIdx}
			out[localIdx] = n
		}
		n.RemotePortID = pduValueString(pdu)
		return nil
	})
	_ = g.walk(oidLldpRemChassis, func(pdu gosnmp.SnmpPDU) error {
		localIdx := lldpLocalIfIndexFromOID(pdu.Name)
		if localIdx <= 0 {
			return nil
		}
		n := out[localIdx]
		if n == nil {
			n = &LldpNeighbor{LocalIfIndex: localIdx}
			out[localIdx] = n
		}
		n.RemoteMac = strings.ToLower(strings.ReplaceAll(pduValueString(pdu), "-", ":"))
		return nil
	})
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
					lldpMap, errLLDP := c.GetLldpNeighbors()
					if errLLDP != nil {
						queryErrs = append(queryErrs, "lldp: "+errLLDP.Error())
						lldpMap = map[int]*LldpNeighbor{}
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
						LldpNeighbors: make(map[int]*topology.LldpNeighbor),
					}
					for idx, ifEntry := range ifTable {
						dev.Ports = append(dev.Ports, topology.Port{
							Index:       idx,
							Name:        ifEntry.Name,
							Description: ifEntry.Description,
						})
					}
					for idx, n := range lldpMap {
						dev.LldpNeighbors[idx] = &topology.LldpNeighbor{
							RemoteChassisID: n.RemoteMac,
							RemotePortID:    n.RemotePortID,
							RemoteSysName:   n.RemoteSys,
						}
					}

					key := dev.MAC
					if key == "" {
						key = dev.IP
					}

					mu.Lock()
					out[key] = dev
					report.Connected++
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

func lldpLocalIfIndexFromOID(oid string) int {
	parts := strings.Split(strings.TrimPrefix(strings.TrimSpace(oid), "."), ".")
	if len(parts) < 3 {
		return -1
	}
	// LLDP OID index ends with timeMark.localPort.remIndex
	n, err := strconv.Atoi(parts[len(parts)-2])
	if err != nil {
		return -1
	}
	return n
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
