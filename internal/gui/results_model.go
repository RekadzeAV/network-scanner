package gui

import (
	"net/netip"
	"sort"
	"strconv"
	"strings"

	"network-scanner/internal/scanner"
)

func sortedResultsForDisplay(results []scanner.Result) []scanner.Result {
	return sortedResultsForDisplayWithMode(results, "IP")
}

func sortedResultsForDisplayWithMode(results []scanner.Result, mode string) []scanner.Result {
	out := append([]scanner.Result(nil), results...)
	sort.SliceStable(out, func(i, j int) bool {
		if strings.EqualFold(strings.TrimSpace(mode), "HostName") {
			hi := strings.ToLower(strings.TrimSpace(out[i].Hostname))
			hj := strings.ToLower(strings.TrimSpace(out[j].Hostname))
			if hi != hj {
				return hi < hj
			}
		}
		ipI := strings.TrimSpace(out[i].IP)
		ipJ := strings.TrimSpace(out[j].IP)
		addrI, errI := netip.ParseAddr(ipI)
		addrJ, errJ := netip.ParseAddr(ipJ)
		if errI == nil && errJ == nil {
			if addrI != addrJ {
				return addrI.Less(addrJ)
			}
		} else if ipI != ipJ {
			return ipI < ipJ
		}
		return strings.TrimSpace(out[i].Hostname) < strings.TrimSpace(out[j].Hostname)
	})
	return out
}

func filterResultsForDisplay(results []scanner.Result, query string) []scanner.Result {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return append([]scanner.Result(nil), results...)
	}
	out := make([]scanner.Result, 0, len(results))
	for _, r := range results {
		fields := []string{
			strings.ToLower(strings.TrimSpace(r.Hostname)),
			strings.ToLower(strings.TrimSpace(r.IP)),
			strings.ToLower(strings.TrimSpace(r.MAC)),
			strings.ToLower(strings.TrimSpace(r.DeviceType)),
		}
		match := false
		for _, f := range fields {
			if strings.Contains(f, q) {
				match = true
				break
			}
		}
		if match {
			out = append(out, r)
		}
	}
	return out
}

func filterResultsForDisplayAdvanced(results []scanner.Result, query string, selectedTypes []string, onlyOpenPorts bool) []scanner.Result {
	base := filterResultsForDisplay(results, query)
	if len(selectedTypes) == 0 && !onlyOpenPorts {
		return base
	}
	normalizedSelection := make(map[string]struct{}, len(selectedTypes))
	for _, t := range selectedTypes {
		key := strings.TrimSpace(t)
		if key == "" {
			continue
		}
		normalizedSelection[key] = struct{}{}
	}

	out := make([]scanner.Result, 0, len(base))
	for _, r := range base {
		if onlyOpenPorts && !hasOpenPorts(r.Ports) {
			continue
		}
		if len(normalizedSelection) > 0 {
			normalizedType := normalizeDeviceTypes(map[string]int{strings.TrimSpace(r.DeviceType): 1})
			matched := false
			for t := range normalizedType {
				if _, ok := normalizedSelection[t]; ok {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		out = append(out, r)
	}
	return out
}

func hasOpenPorts(ports []scanner.PortInfo) bool {
	for _, p := range ports {
		if p.State == "open" {
			return true
		}
	}
	return false
}

func formatDeviceValue(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "-"
	}
	return value
}

func openPortLabels(ports []scanner.PortInfo, maxVisiblePorts int) []string {
	if maxVisiblePorts <= 0 {
		maxVisiblePorts = 24
	}
	labels := make([]string, 0, len(ports))
	openCount := 0
	for _, p := range ports {
		if p.State != "open" {
			continue
		}
		openCount++
		if len(labels) >= maxVisiblePorts {
			continue
		}
		label := strings.TrimSpace(strings.ToUpper(p.Protocol))
		if label == "" {
			label = "TCP"
		}
		text := strings.TrimSpace(strings.Join([]string{formatPortNumber(p.Port) + "/" + label, normalizeServiceName(p.Service)}, " "))
		labels = append(labels, strings.Join(strings.Fields(text), " "))
	}
	if openCount > maxVisiblePorts {
		labels = append(labels, "+"+strconv.Itoa(openCount-maxVisiblePorts))
	}
	return labels
}

func formatPortNumber(port int) string {
	return strconv.Itoa(port)
}

func normalizeServiceName(service string) string {
	s := strings.TrimSpace(service)
	if s == "" || strings.EqualFold(s, "unknown") {
		return ""
	}
	return s
}

func collectAnalytics(results []scanner.Result) (map[string]int, map[string]int) {
	protocols := make(map[string]int)
	deviceTypes := make(map[string]int)
	for _, r := range results {
		for _, protocol := range r.Protocols {
			p := strings.TrimSpace(protocol)
			if p == "" {
				continue
			}
			protocols[p]++
		}
		deviceType := strings.TrimSpace(r.DeviceType)
		if deviceType != "" {
			deviceTypes[deviceType]++
		}
	}
	return protocols, deviceTypes
}

func normalizeDeviceTypes(deviceTypes map[string]int) map[string]int {
	typeMapping := map[string]string{
		"Router/Network Device": "Network Device",
		"Network Device":        "Network Device",
		"Router":                "Network Device",
		"Printer":               "Network Device",
		"IoT Device":            "Network Device",
		"IoT":                   "Network Device",
		"Windows Computer":      "Computer",
		"Computer":              "Computer",
		"Windows":               "Computer",
		"PC":                    "Computer",
		"Desktop":               "Computer",
		"Laptop":                "Computer",
		"Web Server":            "Server",
		"Database Server":       "Server",
		"Linux/Unix Server":     "Server",
		"Server":                "Server",
		"Linux Server":          "Server",
		"Unix Server":           "Server",
		"Linux":                 "Server",
		"Unix":                  "Server",
		"Unknown Device":        "Unknown",
		"Unknown":               "Unknown",
	}
	normalized := make(map[string]int)
	for k, v := range deviceTypes {
		n := typeMapping[k]
		if n == "" {
			n = k
		}
		normalized[n] += v
	}
	return normalized
}
