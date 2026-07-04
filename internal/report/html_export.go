package report

import (
	"bytes"
	"html/template"
	"os"
	"time"

	"network-scanner/internal/contracts"
)

// ScanReportData данные для отчёта о сканировании
type ScanReportData struct {
	GeneratedAt string
	ScanID      string
	Network     string
	HostCount   int
	Results     []ScanResultRow
	Findings    []SecurityFinding
	Topology    *TopologySummary
}

// ScanResultRow строка с результатом сканирования
type ScanResultRow struct {
	IP       string
	Hostname string
	Ports    int
	OS string
	Vendor string
}

// SecurityFinding строка с находкой безопасности
type SecurityFinding struct {
	Severity    string
	Title       string
	Description string
	HostIP      string
	Port        int
}

// TopologySummary сводка по топологии
type TopologySummary struct {
	DeviceCount int
	LinkCount   int
	Devices     []TopologyDevice
}

// TopologyDevice устройство в топологии
type TopologyDevice struct {
	IP       string
	Hostname string
	Vendor string
}

// RenderScanHTML генерирует HTML отчёт о сканировании
func RenderScanHTML(data *ScanReportData) ([]byte, error) {
	tmpl := template.Must(template.New("scan").Parse(scanHTMLTemplate))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// SaveScanHTML сохраняет HTML отчёт о сканировании в файл
func SaveScanHTML(path string, data *ScanReportData) error {
	b, err := RenderScanHTML(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// GenerateScanReportData генерирует данные для отчёта из результатов сканирования
func GenerateScanReportData(scanID, network string, results []contracts.ScanResult, findings []contracts.Finding, topology *contracts.Topology) *ScanReportData {
	reportData := &ScanReportData{
		GeneratedAt: time.Now().Format(time.RFC3339),
		ScanID:      scanID,
		Network:     network,
		HostCount:   len(results),
		Results:     make([]ScanResultRow, 0, len(results)),
		Findings:    make([]SecurityFinding, 0, len(findings)),
	}

	for _, r := range results {
		reportData.Results = append(reportData.Results, ScanResultRow{
			IP:       r.IP,
			Hostname: r.Hostname,
			Ports:    len(r.Ports),
			OS:       r.GuessOS,
			Vendor:   r.DeviceVendor,
		})
	}

	for _, f := range findings {
		reportData.Findings = append(reportData.Findings, SecurityFinding{
			Severity:    string(f.Severity),
			Title:       f.Title,
			Description: f.Recommendation,
			HostIP:      f.Host,
			Port:        0,
		})
	}

	if topology != nil {
		reportData.Topology = &TopologySummary{
			DeviceCount: len(topology.Devices),
			LinkCount:   len(topology.Links),
			Devices:     make([]TopologyDevice, 0, len(topology.Devices)),
		}
		for _, d := range topology.Devices {
			reportData.Topology.Devices = append(reportData.Topology.Devices, TopologyDevice{
				IP:       d.IP,
				Hostname: d.Hostname,
				Vendor:   d.Type,
			})
		}
	}

	return reportData
}

const scanHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Network Scan Report</title>
  <style>
    body { font-family: Arial, sans-serif; margin: 24px; color: #333; }
    h1 { color: #2c3e50; }
    h2 { color: #34495e; border-bottom: 2px solid #3498db; padding-bottom: 8px; }
    .meta { color: #7f8c8d; font-size: 0.9em; margin-bottom: 20px; }
    table { border-collapse: collapse; width: 100%; margin: 15px 0; }
    th, td { border: 1px solid #ddd; padding: 10px; text-align: left; }
    th { background-color: #3498db; color: white; }
    tr:nth-child(even) { background-color: #f2f2f2; }
    .summary { display: flex; gap: 20px; margin: 20px 0; }
    .summary-card { background: #ecf0f1; padding: 15px; border-radius: 5px; flex: 1; text-align: center; }
    .summary-card h3 { margin: 0 0 10px 0; color: #2c3e50; }
    .summary-card .number { font-size: 2em; font-weight: bold; color: #3498db; }
    .finding { background: #fff3cd; border-left: 4px solid #ffc107; padding: 10px; margin: 10px 0; }
    .finding.critical { background: #f8d7da; border-left-color: #dc3545; }
    .finding.high { background: #fff3cd; border-left-color: #fd7e14; }
    .finding.medium { background: #fff3cd; border-left-color: #ffc107; }
    .finding.low { background: #d1ecf1; border-left-color: #17a2b8; }
  </style>
</head>
<body>
  <h1>Network Scan Report</h1>
  <div class="meta">
    <div>Scan ID: {{ .ScanID }}</div>
    <div>Network: {{ .Network }}</div>
    <div>Generated: {{ .GeneratedAt }}</div>
  </div>

  <div class="summary">
    <div class="summary-card">
      <h3>Hosts Scanned</h3>
      <div class="number">{{ .HostCount }}</div>
    </div>
    <div class="summary-card">
      <h3>Security Findings</h3>
      <div class="number">{{ len .Findings }}</div>
    </div>
    {{ if .Topology }}
    <div class="summary-card">
      <h3>Topology Devices</h3>
      <div class="number">{{ .Topology.DeviceCount }}</div>
    </div>
    {{ end }}
  </div>

  <h2>Scan Results</h2>
  <table>
    <thead>
      <tr>
        <th>IP Address</th>
        <th>Hostname</th>
        <th>Open Ports</th>
        <th>OS</th>
        <th>Vendor</th>
      </tr>
    </thead>
    <tbody>
      {{ range .Results }}
      <tr>
        <td>{{ .IP }}</td>
        <td>{{ .Hostname }}</td>
        <td>{{ .Ports }}</td>
        <td>{{ .OS }}</td>
        <td>{{ .Vendor }}</td>
      </tr>
      {{ end }}
    </tbody>
  </table>

  {{ if .Findings }}
  <h2>Security Findings</h2>
  {{ range .Findings }}
  <div class="finding {{ .Severity }}">
    <strong>[{{ .Severity }}]</strong> {{ .Title }}
    <div>Host: {{ .HostIP }}:{{ .Port }}</div>
    <div>{{ .Description }}</div>
  </div>
  {{ end }}
  {{ end }}

  {{ if .Topology }}
  <h2>Network Topology</h2>
  <p>Devices: {{ .Topology.DeviceCount }}, Links: {{ .Topology.LinkCount }}</p>
  <table>
    <thead>
      <tr>
        <th>IP Address</th>
        <th>Hostname</th>
        <th>Vendor</th>
      </tr>
    </thead>
    <tbody>
      {{ range .Topology.Devices }}
      <tr>
        <td>{{ .IP }}</td>
        <td>{{ .Hostname }}</td>
        <td>{{ .Vendor }}</td>
      </tr>
      {{ end }}
    </tbody>
  </table>
  {{ end }}

  <div style="margin-top: 30px; padding: 15px; background: #ecf0f1; border-radius: 5px;">
    <em>This report was automatically generated by network-scanner. Please review all findings with a security specialist.</em>
  </div>
</body>
</html>`


