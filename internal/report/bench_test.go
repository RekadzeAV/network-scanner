package report

import (
	"fmt"
	"testing"
	"time"

	"network-scanner/internal/contracts"
	"network-scanner/internal/cve"
	"network-scanner/internal/risksignature"
	"network-scanner/internal/scanner"
)

// benchFixedTime — фиксированное время для воспроизводимых замеров.
var benchFixedTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// newScanData создаёт ScanReportData с заданным числом хостов и находок.
// Данные генерируются детерминированно через арифметику по индексу, без rand.
func newScanData(hosts, findings int) *ScanReportData {
	rows := make([]ScanResultRow, hosts)
	for i := range hosts {
		rows[i] = ScanResultRow{
			IP:       fmt.Sprintf("192.168.1.%d", (i%254)+1),
			Hostname: fmt.Sprintf("host-%03d", i),
			Ports:    i % 10,
			OS:       []string{"linux", "windows", "ios", ""}[i%4],
			Vendor:   []string{"cisco", "juniper", "huawei", ""}[i%4],
		}
	}
	sevs := []string{"critical", "high", "medium", "low"}
	ttls := []string{"Open Telnet", "Open FTP", "SMB exposed", "RDP exposed"}
	fnd := make([]SecurityFinding, findings)
	for i := range findings {
		fnd[i] = SecurityFinding{
			Severity:    sevs[i%len(sevs)],
			Title:       ttls[i%len(ttls)],
			Description: fmt.Sprintf("Recommendation for finding %d: restrict access", i),
			HostIP:      fmt.Sprintf("192.168.1.%d", (i%254)+1),
			Port:        23 + i%100,
		}
	}
	return &ScanReportData{
		GeneratedAt: benchFixedTime.Format(time.RFC3339),
		ScanID:      "bench-scan",
		Network:     "192.168.1.0/24",
		HostCount:   hosts,
		Results:     rows,
		Findings:    fnd,
		Topology: &TopologySummary{
			DeviceCount: hosts,
			LinkCount:   hosts / 2,
		},
	}
}

// securityResults создаёт срез scanner.Result для бенчмарков security-отчёта.
func securityResults(n int) []scanner.Result {
	results := make([]scanner.Result, n)
	for i := range n {
		results[i] = scanner.Result{
			IP:       fmt.Sprintf("10.0.0.%d", (i%254)+1),
			Hostname: fmt.Sprintf("srv-%03d", i),
			GuessOS:  []string{"linux", "windows", "ios", ""}[i%4],
			Ports: []scanner.PortInfo{
				{Port: 22, State: "open", Protocol: "tcp", Service: "ssh"},
				{Port: 80, State: "open", Protocol: "tcp", Service: "http"},
			},
		}
	}
	return results
}

// securityMatches создаёт срез cve.Match для бенчмарков security-отчёта.
func securityMatches(n int) []cve.Match {
	cveIDs := []string{"CVE-2023-44487", "CVE-2023-38408", "CVE-2021-44228"}
	cveDescs := []string{"HTTP/2 Rapid Reset DoS", "OpenSSH agent forwarding RCE", "Log4Shell RCE"}
	cveCVSS := []float64{7.5, 9.8, 10.0}
	matches := make([]cve.Match, n)
	for i := range n {
		j := i % len(cveIDs)
		matches[i] = cve.Match{
			HostIP:   fmt.Sprintf("10.0.0.%d", (i%254)+1),
			HostName: fmt.Sprintf("srv-%03d", i),
			Port:     22 + i%80,
			Service:  []string{"ssh", "http"}[i%2],
			Entry: cve.Entry{
				ID:          cveIDs[j],
				Description: cveDescs[j],
				URL:         "https://nvd.nist.gov/vuln/detail/CVE-2023-00000",
				CVSS:        cveCVSS[j],
			},
		}
	}
	return matches
}

// securityRisks создаёт срез risksignature.Finding для бенчмарков.
func securityRisks(n int) []risksignature.Finding {
	risks := make([]risksignature.Finding, n)
	for i := range n {
		risks[i] = risksignature.Finding{
			HostIP:         fmt.Sprintf("10.0.0.%d", (i%254)+1),
			SignatureID:    fmt.Sprintf("sig.%03d", i),
			Title:          fmt.Sprintf("Risk %d", i),
			Severity:       []string{"critical", "high", "medium", "low"}[i%4],
			Reason:         fmt.Sprintf("open_port=%d", 23+i%80),
			Recommendation: "Disable service or restrict access via firewall",
		}
	}
	return risks
}

// contractsResults создаёт срез contracts.ScanResult для PDF-бенчмарков.
func contractsResults(n int) []contracts.ScanResult {
	results := make([]contracts.ScanResult, n)
	for i := range n {
		results[i] = contracts.ScanResult{
			IP:           fmt.Sprintf("192.168.1.%d", (i%254)+1),
			Hostname:     fmt.Sprintf("host-%03d", i),
			MAC:          fmt.Sprintf("aa:bb:cc:dd:%02x:%02x", (i/256)%256, i%256),
			DeviceVendor: []string{"cisco", "juniper", "huawei", ""}[i%4],
			GuessOS:      []string{"linux", "windows", "ios", ""}[i%4],
			Ports: []contracts.PortInfo{
				{Port: 22, State: "open", Protocol: "tcp", Service: "ssh"},
				{Port: 80, State: "open", Protocol: "tcp", Service: "http"},
			},
		}
	}
	return results
}

// contractsFindings создаёт срез contracts.Finding для PDF-бенчмарков.
func contractsFindings(n int) []contracts.Finding {
	sevs := []string{"critical", "high", "medium", "low"}
	ttls := []string{"Open Telnet", "Open FTP", "SMB exposed", "RDP exposed"}
	findings := make([]contracts.Finding, n)
	for i := range n {
		findings[i] = contracts.Finding{
			Severity:       sevs[i%len(sevs)],
			Host:           fmt.Sprintf("192.168.1.%d", (i%254)+1),
			Title:          ttls[i%len(ttls)],
			Recommendation: fmt.Sprintf("Recommendation %d: restrict access to the port", i),
		}
	}
	return findings
}

// BenchmarkHTMLExport замеряет рендеринг HTML-отчёта о сканировании
// через RenderScanHTML (возвращает []byte, файлы не создаются).
func BenchmarkHTMLExport(b *testing.B) {
	sizes := []struct {
		name     string
		hosts    int
		findings int
	}{
		{"Small", 5, 3},
		{"Medium", 50, 20},
		{"Large", 300, 100},
	}
	for _, tc := range sizes {
		b.Run(tc.name, func(b *testing.B) {
			data := newScanData(tc.hosts, tc.findings)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = RenderScanHTML(data)
			}
		})
	}
}

// BenchmarkSecurityHTML замеряет рендеринг security HTML-отчёта
// через RenderSecurityHTMLWithRisk (включая CVE-находки и risk-сигнатуры).
func BenchmarkSecurityHTML(b *testing.B) {
	sizes := []struct {
		name     string
		hosts    int
		findings int
		risks    int
	}{
		{"Small", 5, 3, 2},
		{"Medium", 50, 20, 10},
		{"Large", 300, 100, 50},
	}
	for _, tc := range sizes {
		b.Run(tc.name, func(b *testing.B) {
			results := securityResults(tc.hosts)
			findings := securityMatches(tc.findings)
			risks := securityRisks(tc.risks)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = RenderSecurityHTMLWithRisk(results, findings, risks, benchFixedTime)
			}
		})
	}
}

// BenchmarkPDFExport замеряет генерацию PDF-отчёта через PDFReport.Bytes().
// Файлы не записываются — используется Bytes() вместо Save().
func BenchmarkPDFExport(b *testing.B) {
	sizes := []struct {
		name     string
		hosts    int
		findings int
	}{
		{"Small", 5, 3},
		{"Medium", 50, 20},
		{"Large", 300, 100},
	}
	for _, tc := range sizes {
		b.Run(tc.name, func(b *testing.B) {
			results := contractsResults(tc.hosts)
			findings := contractsFindings(tc.findings)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r := NewPDFReport("Benchmark Report")
				r.AddScanResults(results)
				r.AddSecurityFindings(findings)
				_, _ = r.Bytes()
			}
		})
	}
}
