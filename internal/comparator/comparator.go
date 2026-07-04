package comparator

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"network-scanner/internal/scanner"
)

// ScanHistoryEntry запись в истории сканирований
type ScanHistoryEntry struct {
	ID        string            `json:"id"`
	Network   string            `json:"network"`
	HostCount int               `json:"host_count"`
	StartedAt time.Time         `json:"started_at"`
	Completed time.Time         `json:"completed"`
	Ports     map[string]int    `json:"ports"`
	OSMap     map[string]int    `json:"os_map"`
	VendorMap map[string]int    `json:"vendor_map"`
}

// ComparisonResult результат сравнения двух сканирований
type ComparisonResult struct {
	ScanIDA      string             `json:"scan_id_a"`
	ScanIDB      string             `json:"scan_id_b"`
	NewHosts     []scanner.Result   `json:"new_hosts"`
	RemovedHosts []scanner.Result   `json:"removed_hosts"`
	ChangedHosts []ChangedHost      `json:"changed_hosts"`
	PortChanges  []PortChange       `json:"port_changes"`
	TotalDiff    int                `json:"total_diff"`
}

// ChangedHost изменённый хост
type ChangedHost struct {
	IP        string           `json:"ip"`
	Hostname  string           `json:"hostname"`
	Before    scanner.Result   `json:"before"`
	After     scanner.Result   `json:"after"`
	ChangedIn []string         `json:"changed_in"`
}

// PortChange изменение портов
type PortChange struct {
	HostIP      string `json:"host_ip"`
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	ChangedFrom string `json:"changed_from"`
	ChangedTo   string `json:"changed_to"`
}

// BuildHistoryEntry создаёт запись истории из результатов сканирования
func BuildHistoryEntry(scanID, network string, hosts []scanner.Result, startedAt, completed time.Time) ScanHistoryEntry {
	entry := ScanHistoryEntry{
		ID:        scanID,
		Network:   network,
		HostCount: len(hosts),
		StartedAt: startedAt,
		Completed: completed,
		Ports:     make(map[string]int),
		OSMap:     make(map[string]int),
		VendorMap: make(map[string]int),
	}

	for _, h := range hosts {
		if h.GuessOS != "" {
			entry.OSMap[h.GuessOS]++
		}
		if h.DeviceVendor != "" {
			entry.VendorMap[h.DeviceVendor]++
		}
		for _, p := range h.Ports {
			if strings.EqualFold(p.State, "open") {
				key := fmt.Sprintf("%d/%s", p.Port, p.Protocol)
				entry.Ports[key]++
			}
		}
	}

	return entry
}

// CompareSnapshots сравнивает два набора результатов
func CompareSnapshots(scanIDA, scanIDB string, a, b []scanner.Result) *ComparisonResult {
	aMap := hostsByID(a)
	bMap := hostsByID(b)

	res := &ComparisonResult{
		ScanIDA:      scanIDA,
		ScanIDB:      scanIDB,
		NewHosts:     make([]scanner.Result, 0),
		RemovedHosts: make([]scanner.Result, 0),
		ChangedHosts: make([]ChangedHost, 0),
		PortChanges:  make([]PortChange, 0),
	}

	// Новые хосты (есть в B, нет в A)
	for id, hostB := range bMap {
		if _, exists := aMap[id]; !exists {
			res.NewHosts = append(res.NewHosts, hostB)
		}
	}

	// Удалённые хосты (есть в A, нет в B)
	for id, hostA := range aMap {
		if _, exists := bMap[id]; !exists {
			res.RemovedHosts = append(res.RemovedHosts, hostA)
		}
	}

	// Изменённые хосты
	for id, hostA := range aMap {
		hostB, exists := bMap[id]
		if !exists {
			continue
		}
		changes := detectChanges(hostA, hostB, &res.PortChanges)
		if len(changes) > 0 {
			res.ChangedHosts = append(res.ChangedHosts, ChangedHost{
				IP:        hostA.IP,
				Hostname:  hostA.Hostname,
				Before:    hostA,
				After:     hostB,
				ChangedIn: changes,
			})
		}
	}

	// Сортировка для детерминированности
	sortResults(res)

	// Подсчёт общего дифа
	res.TotalDiff = len(res.NewHosts) + len(res.RemovedHosts) + len(res.ChangedHosts) + len(res.PortChanges)

	return res
}

// detectChanges находит различия между двумя результатами
func detectChanges(a, b scanner.Result, portChanges *[]PortChange) []string {
	changes := make([]string, 0)

	if a.Hostname != b.Hostname {
		changes = append(changes, "hostname")
	}
	if a.DeviceType != b.DeviceType {
		changes = append(changes, "device_type")
	}
	if a.DeviceVendor != b.DeviceVendor {
		changes = append(changes, "device_vendor")
	}
	if a.GuessOS != b.GuessOS {
		changes = append(changes, "os")
	}

	// Проверка портов
	aPorts := portsByNumber(a.Ports)
	bPorts := portsByNumber(b.Ports)

	for port, portB := range bPorts {
		portA, exists := aPorts[port]
		if !exists {
			*portChanges = append(*portChanges, PortChange{
				HostIP:      b.IP,
				Port:        port,
				Protocol:    portB.Protocol,
				ChangedFrom: "closed",
				ChangedTo:   portB.State,
			})
			changes = append(changes, "ports")
			continue
		}
		if portA.State != portB.State {
			*portChanges = append(*portChanges, PortChange{
				HostIP:      a.IP,
				Port:        port,
				Protocol:    portA.Protocol,
				ChangedFrom: portA.State,
				ChangedTo:   portB.State,
			})
			changes = append(changes, "ports")
		}
	}

	for port, portA := range aPorts {
		if _, exists := bPorts[port]; !exists {
			*portChanges = append(*portChanges, PortChange{
				HostIP:      a.IP,
				Port:        port,
				Protocol:    portA.Protocol,
				ChangedFrom: portA.State,
				ChangedTo:   "closed",
			})
		}
	}

	return changes
}

// portsByNumber создаёт мапу port -> PortInfo
func portsByNumber(ports []scanner.PortInfo) map[int]scanner.PortInfo {
	out := make(map[int]scanner.PortInfo)
	for _, p := range ports {
		out[p.Port] = p
	}
	return out
}

// hostsByID создаёт мапу IP -> Result
func hostsByID(hosts []scanner.Result) map[string]scanner.Result {
	out := make(map[string]scanner.Result)
	for _, h := range hosts {
		out[h.IP] = h
	}
	return out
}

// sortResults сортирует результаты для детерминированности
func sortResults(res *ComparisonResult) {
	sort.Slice(res.NewHosts, func(i, j int) bool {
		return res.NewHosts[i].IP < res.NewHosts[j].IP
	})
	sort.Slice(res.RemovedHosts, func(i, j int) bool {
		return res.RemovedHosts[i].IP < res.RemovedHosts[j].IP
	})
	sort.Slice(res.ChangedHosts, func(i, j int) bool {
		return res.ChangedHosts[i].IP < res.ChangedHosts[j].IP
	})
}
