package risksignature

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"network-scanner/internal/scanner"
)

//go:embed signatures/default-home-risks.v1.json
var defaultSignaturesRaw []byte

// SignatureDB contains a versioned list of risk signatures.
type SignatureDB struct {
	Version    string      `json:"version"`
	Signatures []Signature `json:"signatures"`
}

// Signature describes a single risk rule.
type Signature struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Severity        string   `json:"severity"`
	Recommendation  string   `json:"recommendation"`
	ReferenceURL    string   `json:"reference_url,omitempty"`
	MatchAnyPort    []int    `json:"match_any_port,omitempty"`
	MatchAnyService []string `json:"match_any_service,omitempty"`
	MatchAnyBanner  []string `json:"match_any_banner,omitempty"`
	MatchDeviceType []string `json:"match_device_type,omitempty"`
	MatchVendor     []string `json:"match_vendor,omitempty"`
}

// Finding is a matched signature for a host.
type Finding struct {
	HostIP         string `json:"host_ip"`
	SignatureID    string `json:"signature_id"`
	Title          string `json:"title"`
	Severity       string `json:"severity"`
	Recommendation string `json:"recommendation"`
	ReferenceURL   string `json:"reference_url,omitempty"`
	Reason         string `json:"reason"`
}

// LoadDefault loads bundled signatures.
func LoadDefault() (SignatureDB, error) {
	return Load(defaultSignaturesRaw)
}

// Load parses signature DB from JSON bytes.
func Load(raw []byte) (SignatureDB, error) {
	var db SignatureDB
	if err := json.Unmarshal(raw, &db); err != nil {
		return SignatureDB{}, fmt.Errorf("parse signature db: %w", err)
	}
	if strings.TrimSpace(db.Version) == "" {
		return SignatureDB{}, fmt.Errorf("signature db version is required")
	}
	for i, sig := range db.Signatures {
		if strings.TrimSpace(sig.ID) == "" {
			return SignatureDB{}, fmt.Errorf("signature[%d]: id is required", i)
		}
		if strings.TrimSpace(sig.Title) == "" {
			return SignatureDB{}, fmt.Errorf("signature[%d]: title is required", i)
		}
	}
	return db, nil
}

// Evaluate runs signatures against scan results.
func Evaluate(results []scanner.Result, db SignatureDB) []Finding {
	findings := make([]Finding, 0)
	for _, host := range results {
		for _, sig := range db.Signatures {
			reason, ok := matchSignature(host, sig)
			if !ok {
				continue
			}
			findings = append(findings, Finding{
				HostIP:         strings.TrimSpace(host.IP),
				SignatureID:    sig.ID,
				Title:          sig.Title,
				Severity:       normalizeSeverity(sig.Severity),
				Recommendation: strings.TrimSpace(sig.Recommendation),
				ReferenceURL:   strings.TrimSpace(sig.ReferenceURL),
				Reason:         reason,
			})
		}
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].HostIP == findings[j].HostIP {
			return findings[i].SignatureID < findings[j].SignatureID
		}
		return findings[i].HostIP < findings[j].HostIP
	})
	return findings
}

func matchSignature(host scanner.Result, sig Signature) (string, bool) {
	reasons := make([]string, 0, 3)
	if reason, ok := matchPorts(host, sig.MatchAnyPort); ok {
		reasons = append(reasons, reason)
	}
	if reason, ok := matchServices(host, sig.MatchAnyService); ok {
		reasons = append(reasons, reason)
	}
	if reason, ok := matchBanners(host, sig.MatchAnyBanner); ok {
		reasons = append(reasons, reason)
	}
	if reason, ok := matchStringAny(host.DeviceType, sig.MatchDeviceType, "device_type"); ok {
		reasons = append(reasons, reason)
	}
	if reason, ok := matchStringAny(host.DeviceVendor, sig.MatchVendor, "vendor"); ok {
		reasons = append(reasons, reason)
	}
	if len(reasons) == 0 {
		return "", false
	}
	return strings.Join(reasons, "; "), true
}

func matchPorts(host scanner.Result, ports []int) (string, bool) {
	if len(ports) == 0 {
		return "", false
	}
	for _, p := range host.Ports {
		if p.State != "open" {
			continue
		}
		for _, wanted := range ports {
			if p.Port == wanted {
				return fmt.Sprintf("open_port=%d", p.Port), true
			}
		}
	}
	return "", false
}

func matchServices(host scanner.Result, services []string) (string, bool) {
	if len(services) == 0 {
		return "", false
	}
	for _, p := range host.Ports {
		if p.State != "open" {
			continue
		}
		service := strings.ToLower(strings.TrimSpace(p.Service))
		for _, wanted := range services {
			w := strings.ToLower(strings.TrimSpace(wanted))
			if w != "" && strings.Contains(service, w) {
				return fmt.Sprintf("service=%s", strings.TrimSpace(p.Service)), true
			}
		}
	}
	return "", false
}

func matchBanners(host scanner.Result, patterns []string) (string, bool) {
	if len(patterns) == 0 {
		return "", false
	}
	for _, p := range host.Ports {
		if p.State != "open" {
			continue
		}
		text := strings.ToLower(strings.TrimSpace(p.Banner + " " + p.Version))
		for _, pattern := range patterns {
			w := strings.ToLower(strings.TrimSpace(pattern))
			if w != "" && strings.Contains(text, w) {
				return fmt.Sprintf("banner_pattern=%s", w), true
			}
		}
	}
	return "", false
}

func matchStringAny(value string, wanted []string, key string) (string, bool) {
	if len(wanted) == 0 {
		return "", false
	}
	val := strings.ToLower(strings.TrimSpace(value))
	for _, item := range wanted {
		w := strings.ToLower(strings.TrimSpace(item))
		if w != "" && strings.Contains(val, w) {
			return fmt.Sprintf("%s=%s", key, strings.TrimSpace(value)), true
		}
	}
	return "", false
}

func normalizeSeverity(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "critical":
		return "critical"
	case "high":
		return "high"
	case "medium":
		return "medium"
	default:
		return "low"
	}
}
