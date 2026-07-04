package diff

import (
	"fmt"
	"sort"

	"network-scanner/internal/scanner"
)

// DiffReport contains the differences between two scan results.
type DiffReport struct {
	NewHosts    []scanner.HostResult // Hosts found in current but not in previous
	GoneHosts   []scanner.HostResult // Hosts found in previous but not in current
	ChangedHosts []ChangedHost       // Hosts present in both but with differences
	TotalNew    int
	TotalGone   int
	TotalChanged int
}

// ChangedHost represents a host that exists in both scans but with differences.
type ChangedHost struct {
	IP        string
	Previous  scanner.HostResult
	Current   scanner.HostResult
	Changes   []Change
}

// Change represents a single difference between previous and current state.
type Change struct {
	Field     string
	Previous  string
	Current   string
}

// CompareScanResults compares two scan results and returns a diff report.
func CompareScanResults(previous, current []scanner.HostResult) *DiffReport {
	report := &DiffReport{}

	// Build lookup maps by IP
	prevMap := make(map[string]scanner.HostResult)
	for _, h := range previous {
		prevMap[h.IP] = h
	}

	currMap := make(map[string]scanner.HostResult)
	for _, h := range current {
		currMap[h.IP] = h
	}

	// Find new and changed hosts
	report.NewHosts = make([]scanner.HostResult, 0)
	report.ChangedHosts = make([]ChangedHost, 0)

	for ip, curr := range currMap {
		prev, exists := prevMap[ip]
		if !exists {
			report.NewHosts = append(report.NewHosts, curr)
		} else {
			changes := detectChanges(prev, curr)
			if len(changes) > 0 {
				report.ChangedHosts = append(report.ChangedHosts, ChangedHost{
					IP:       ip,
					Previous: prev,
					Current:  curr,
					Changes:  changes,
				})
			}
		}
	}

	// Find gone hosts
	report.GoneHosts = make([]scanner.HostResult, 0)
	for ip, prev := range prevMap {
		if _, exists := currMap[ip]; !exists {
			report.GoneHosts = append(report.GoneHosts, prev)
		}
	}

	// Sort for deterministic output
	sortHosts(report.NewHosts)
	sortHosts(report.GoneHosts)
	sortChangedHosts(report.ChangedHosts)

	report.TotalNew = len(report.NewHosts)
	report.TotalGone = len(report.GoneHosts)
	report.TotalChanged = len(report.ChangedHosts)

	return report
}

// detectChanges compares two host results and returns a list of changes.
func detectChanges(prev, curr scanner.HostResult) []Change {
	var changes []Change

	// Check hostname
	if prev.Hostname != curr.Hostname {
		changes = append(changes, Change{
			Field:     "Hostname",
			Previous:  prev.Hostname,
			Current:   curr.Hostname,
		})
	}

	// Check MAC
	if prev.MAC != curr.MAC {
		changes = append(changes, Change{
			Field:     "MAC",
			Previous:  prev.MAC,
			Current:   curr.MAC,
		})
	}

	// Check device type
	if prev.DeviceType != curr.DeviceType {
		changes = append(changes, Change{
			Field:     "DeviceType",
			Previous:  prev.DeviceType,
			Current:   curr.DeviceType,
		})
	}

	// Check device vendor
	if prev.DeviceVendor != curr.DeviceVendor {
		changes = append(changes, Change{
			Field:     "DeviceVendor",
			Previous:  prev.DeviceVendor,
			Current:   curr.DeviceVendor,
		})
	}

	// Check SNMP
	if prev.SNMPEnabled != curr.SNMPEnabled {
		changes = append(changes, Change{
			Field:     "SNMP",
			Previous:  fmt.Sprintf("%v", prev.SNMPEnabled),
			Current:   fmt.Sprintf("%v", curr.SNMPEnabled),
		})
	}

	// Check open ports
	prevPorts := make(map[int]bool)
	for _, p := range prev.Ports {
		if p.State == "open" {
			prevPorts[p.Port] = true
		}
	}

	currPorts := make(map[int]bool)
	for _, p := range curr.Ports {
		if p.State == "open" {
			currPorts[p.Port] = true
		}
	}

	// Find new open ports
	newPorts := make([]int, 0)
	for port := range currPorts {
		if !prevPorts[port] {
			newPorts = append(newPorts, port)
		}
	}
	sort.Ints(newPorts)

	// Find closed ports
	closedPorts := make([]int, 0)
	for port := range prevPorts {
		if !currPorts[port] {
			closedPorts = append(closedPorts, port)
		}
	}
	sort.Ints(closedPorts)

	if len(newPorts) > 0 || len(closedPorts) > 0 {
		prevPortsStr := portsToString(prevPorts)
		currPortsStr := portsToString(currPorts)
		changes = append(changes, Change{
			Field:     "OpenPorts",
			Previous:  prevPortsStr,
			Current:   currPortsStr,
		})
	}

	return changes
}

// portsToString converts a map of open ports to a string.
func portsToString(ports map[int]bool) string {
	if len(ports) == 0 {
		return "none"
	}

	portsList := make([]int, 0, len(ports))
	for port := range ports {
		portsList = append(portsList, port)
	}
	sort.Ints(portsList)

	result := ""
	for i, port := range portsList {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("%d", port)
	}
	return result
}

// sortHosts sorts hosts by IP address.
func sortHosts(hosts []scanner.HostResult) {
	sort.Slice(hosts, func(i, j int) bool {
		return hosts[i].IP < hosts[j].IP
	})
}

// sortChangedHosts sorts changed hosts by IP address.
func sortChangedHosts(hosts []ChangedHost) {
	sort.Slice(hosts, func(i, j int) bool {
		return hosts[i].IP < hosts[j].IP
	})
}

// FormatReport formats the diff report as a human-readable string.
func (r *DiffReport) FormatReport() string {
	result := "=== Network Scan Diff Report ===\n\n"

	result += fmt.Sprintf("Total new hosts: %d\n", r.TotalNew)
	result += fmt.Sprintf("Total gone hosts: %d\n", r.TotalGone)
	result += fmt.Sprintf("Total changed hosts: %d\n\n", r.TotalChanged)

	if len(r.NewHosts) > 0 {
		result += "--- New Hosts ---\n"
		for _, h := range r.NewHosts {
			result += fmt.Sprintf("  + %s (%s)\n", h.IP, h.Hostname)
		}
		result += "\n"
	}

	if len(r.GoneHosts) > 0 {
		result += "--- Gone Hosts ---\n"
		for _, h := range r.GoneHosts {
			result += fmt.Sprintf("  - %s (%s)\n", h.IP, h.Hostname)
		}
		result += "\n"
	}

	if len(r.ChangedHosts) > 0 {
		result += "--- Changed Hosts ---\n"
		for _, ch := range r.ChangedHosts {
			result += fmt.Sprintf("  ~ %s:\n", ch.IP)
			for _, change := range ch.Changes {
				result += fmt.Sprintf("    %s: %s -> %s\n", change.Field, change.Previous, change.Current)
			}
		}
		result += "\n"
	}

	return result
}
