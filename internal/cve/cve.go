package cve

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"network-scanner/internal/redact"
	"network-scanner/internal/scanner"
)

// Entry describes a vulnerability record from a CVE source.
type Entry struct {
	ID          string
	Description string
	URL         string
	CVSS        float64
	PublishedAt time.Time
	Service     string
	VersionHint string
}

// Match describes a concrete CVE finding on a host/port.
type Match struct {
	HostIP      string
	HostName    string
	Port        int
	Service     string
	VersionHint string
	Entry       Entry
}

// Options controls filtering of the CVE results.
type Options struct {
	MinCVSS    float64
	MaxAgeDays int
	Now        time.Time
}

// Catalog is a simple in-memory CVE store.
type Catalog struct {
	entries []Entry
}

// NewDefaultCatalog returns a seed dataset for P3 stage-2 bootstrap.
func NewDefaultCatalog() Catalog {
	parse := func(s string) time.Time {
		t, _ := time.Parse("2006-01-02", s)
		return t
	}
	return Catalog{
		entries: []Entry{
			{
				ID:          "CVE-2023-44487",
				Description: "HTTP/2 Rapid Reset DoS vulnerability.",
				URL:         "https://nvd.nist.gov/vuln/detail/CVE-2023-44487",
				CVSS:        7.5,
				PublishedAt: parse("2023-10-10"),
				Service:     "http",
				VersionHint: "nginx/1.25",
			},
			{
				ID:          "CVE-2023-38408",
				Description: "OpenSSH agent forwarding remote code execution issue.",
				URL:         "https://nvd.nist.gov/vuln/detail/CVE-2023-38408",
				CVSS:        9.8,
				PublishedAt: parse("2023-07-19"),
				Service:     "ssh",
				VersionHint: "openssh_9.3",
			},
			{
				ID:          "CVE-2021-44228",
				Description: "Apache Log4j2 JNDI remote code execution (Log4Shell).",
				URL:         "https://nvd.nist.gov/vuln/detail/CVE-2021-44228",
				CVSS:        10.0,
				PublishedAt: parse("2021-12-10"),
				Service:     "http",
				VersionHint: "log4j/2.14",
			},
		},
	}
}

// AnalyzeResults matches scan results with the catalog and applies filters.
func AnalyzeResults(results []scanner.Result, catalog Catalog, opts Options) []Match {
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	out := make([]Match, 0)
	for _, host := range results {
		for _, port := range host.Ports {
			if !strings.EqualFold(port.State, "open") {
				continue
			}
			service := normalizeService(port)
			version := strings.ToLower(strings.TrimSpace(port.Version + " " + port.Banner))
			if service == "" || version == "" {
				continue
			}
			for _, e := range catalog.entries {
				if !strings.EqualFold(service, e.Service) {
					continue
				}
				if !strings.Contains(version, strings.ToLower(e.VersionHint)) {
					continue
				}
				if opts.MinCVSS > 0 && e.CVSS < opts.MinCVSS {
					continue
				}
				if opts.MaxAgeDays > 0 && !e.PublishedAt.IsZero() {
					if now.Sub(e.PublishedAt) > time.Duration(opts.MaxAgeDays)*24*time.Hour {
						continue
					}
				}
				out = append(out, Match{
					HostIP:      host.IP,
					HostName:    host.Hostname,
					Port:        port.Port,
					Service:     service,
					VersionHint: strings.TrimSpace(port.Version),
					Entry:       e,
				})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Entry.CVSS == out[j].Entry.CVSS {
			return out[i].Entry.ID < out[j].Entry.ID
		}
		return out[i].Entry.CVSS > out[j].Entry.CVSS
	})
	return out
}

func normalizeService(p scanner.PortInfo) string {
	name := strings.ToLower(strings.TrimSpace(p.Service))
	switch name {
	case "https":
		return "http"
	case "ssh":
		return "ssh"
	case "http":
		return "http"
	default:
		switch p.Port {
		case 22:
			return "ssh"
		case 80, 443, 8080, 8443:
			return "http"
		}
	}
	return ""
}

// FormatMatches formats findings for CLI output.
func FormatMatches(matches []Match) string {
	if len(matches) == 0 {
		return "CVE анализ: совпадений не найдено."
	}
	lines := []string{fmt.Sprintf("CVE анализ: найдено совпадений: %d", len(matches))}
	for _, m := range matches {
		target := m.HostIP
		if strings.TrimSpace(m.HostName) != "" {
			target = fmt.Sprintf("%s (%s)", strings.TrimSpace(m.HostName), m.HostIP)
		}
		lines = append(lines, fmt.Sprintf("- %s:%d %s -> %s (CVSS %.1f)",
			redact.SanitizeText(target), m.Port, redact.SanitizeText(m.Service), m.Entry.ID, m.Entry.CVSS))
	}
	return strings.Join(lines, "\n")
}
