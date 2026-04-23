package report

import (
	"bytes"
	"html/template"
	"os"
	"strings"
	"time"

	"network-scanner/internal/cve"
	"network-scanner/internal/redact"
	"network-scanner/internal/risksignature"
	"network-scanner/internal/scanner"
)

type securityReportData struct {
	GeneratedAt string
	HostCount   int
	CVECount    int
	RiskCount   int
	Redaction   string
	Unredacted  bool
	Metadata    reportMetadata
	Results     []scanner.Result
	Findings    []cve.Match
	Risks       []risksignature.Finding
}

// Options controls security report rendering behavior.
type Options struct {
	RedactSensitive bool
	PolicyVersion   string
	UnsafeConsent   bool
	GenerationMode  string
	ReportID        string
}

type reportMetadata struct {
	ReportID       string
	GenerationMode string
	PolicyVersion  string
	UnsafeConsent  string
}

const securityTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <title>Network Security Report</title>
  <style>
    body { font-family: Arial, sans-serif; margin: 24px; color: #111; }
    h1,h2 { margin-bottom: 8px; }
    .muted { color: #555; font-size: 0.95rem; }
    table { border-collapse: collapse; width: 100%; margin: 12px 0 20px; }
    th, td { border: 1px solid #ddd; padding: 8px; text-align: left; vertical-align: top; }
    th { background: #f3f3f3; }
    .disclaimer { margin-top: 24px; padding: 12px; border: 1px solid #f0c36d; background: #fff9e6; }
  </style>
</head>
<body>
  <h1>Network Security Report</h1>
  <div class="muted">Generated at: {{ .GeneratedAt }}</div>
  <div class="muted">Hosts scanned: {{ .HostCount }} | CVE matches: {{ .CVECount }} | Risk signatures: {{ .RiskCount }}</div>
  <div class="muted">REDACTION: {{ .Redaction }}</div>
  {{- if .Unredacted }}
  <div class="disclaimer">
    WARNING: This report is generated with redaction disabled and may contain sensitive data.
  </div>
  {{- end }}
  <div class="muted">Metadata: report-id={{ .Metadata.ReportID }} | mode={{ .Metadata.GenerationMode }} | policy={{ .Metadata.PolicyVersion }} | unsafe-consent={{ .Metadata.UnsafeConsent }}</div>

  <h2>CVE Findings</h2>
  <table>
    <thead>
      <tr>
        <th>Host</th><th>Port</th><th>Service</th><th>CVE</th><th>CVSS</th><th>Description</th>
      </tr>
    </thead>
    <tbody>
      {{- if .Findings }}
      {{- range .Findings }}
      <tr>
        <td>{{ if .HostName }}{{ san .HostName }} ({{ san .HostIP }}){{ else }}{{ san .HostIP }}{{ end }}</td>
        <td>{{ .Port }}</td>
        <td>{{ san .Service }}</td>
        <td><a href="{{ .Entry.URL }}">{{ .Entry.ID }}</a></td>
        <td>{{ printf "%.1f" .Entry.CVSS }}</td>
        <td>{{ san .Entry.Description }}</td>
      </tr>
      {{- end }}
      {{- else }}
      <tr><td colspan="6">No CVE matches found for current dataset/filters.</td></tr>
      {{- end }}
    </tbody>
  </table>

  <h2>Risk Signature Findings</h2>
  <table>
    <thead>
      <tr>
        <th>Host</th><th>Severity</th><th>Signature</th><th>Reason</th><th>Recommendation</th>
      </tr>
    </thead>
    <tbody>
      {{- if .Risks }}
      {{- range .Risks }}
      <tr>
        <td>{{ san .HostIP }}</td>
        <td>{{ san .Severity }}</td>
        <td>{{ san .Title }} ({{ san .SignatureID }})</td>
        <td>{{ san .Reason }}</td>
        <td>{{ san .Recommendation }}{{ if .ReferenceURL }} <a href="{{ .ReferenceURL }}">ref</a>{{ end }}</td>
      </tr>
      {{- end }}
      {{- else }}
      <tr><td colspan="5">No risk-signature matches found for current dataset.</td></tr>
      {{- end }}
    </tbody>
  </table>

  <h2>Scanned Hosts</h2>
  <table>
    <thead>
      <tr><th>IP</th><th>Hostname</th><th>Open Ports</th><th>Guessed OS</th></tr>
    </thead>
    <tbody>
      {{- range .Results }}
      <tr>
        <td>{{ san .IP }}</td>
        <td>{{ san .Hostname }}</td>
        <td>{{ openPorts .Ports }}</td>
        <td>{{ san .GuessOS }}</td>
      </tr>
      {{- end }}
    </tbody>
  </table>

  <div class="disclaimer">
    This report is for authorized environments only and should be validated by security specialists before remediation actions.
  </div>
</body>
</html>`

// RenderSecurityHTML returns HTML report bytes.
func RenderSecurityHTML(results []scanner.Result, findings []cve.Match, now time.Time) ([]byte, error) {
	return RenderSecurityHTMLWithRisk(results, findings, nil, now)
}

// RenderSecurityHTMLWithRisk returns HTML report bytes including risk-signature findings.
func RenderSecurityHTMLWithRisk(results []scanner.Result, findings []cve.Match, risks []risksignature.Finding, now time.Time) ([]byte, error) {
	return RenderSecurityHTMLWithRiskOptions(results, findings, risks, now, Options{RedactSensitive: true})
}

// RenderSecurityHTMLWithRiskOptions returns HTML report bytes including risk-signature findings and custom options.
func RenderSecurityHTMLWithRiskOptions(results []scanner.Result, findings []cve.Match, risks []risksignature.Finding, now time.Time, opts Options) ([]byte, error) {
	if now.IsZero() {
		now = time.Now()
	}
	sanitize := func(v string) string {
		if opts.RedactSensitive {
			return redact.SanitizeText(v)
		}
		return v
	}
	tpl, err := template.New("security").Funcs(template.FuncMap{
		"san": func(v string) string {
			return sanitize(v)
		},
		"openPorts": func(ports []scanner.PortInfo) string {
			values := make([]string, 0)
			for _, p := range ports {
				if strings.EqualFold(p.State, "open") {
					values = append(values, sanitize(p.Service))
				}
			}
			if len(values) == 0 {
				return "-"
			}
			return strings.Join(values, ", ")
		},
	}).Parse(securityTemplate)
	if err != nil {
		return nil, err
	}

	data := securityReportData{
		GeneratedAt: now.Format(time.RFC3339),
		HostCount:   len(results),
		CVECount:    len(findings),
		RiskCount:   len(risks),
		Redaction:   "OFF",
		Unredacted:  true,
		Metadata: reportMetadata{
			ReportID:       strings.TrimSpace(opts.ReportID),
			GenerationMode: strings.TrimSpace(opts.GenerationMode),
			PolicyVersion:  strings.TrimSpace(opts.PolicyVersion),
			UnsafeConsent:  "no",
		},
		Results:     results,
		Findings:    findings,
		Risks:       risks,
	}
	if data.Metadata.GenerationMode == "" {
		data.Metadata.GenerationMode = "manual"
	}
	if data.Metadata.ReportID == "" {
		data.Metadata.ReportID = "n/a"
	}
	if data.Metadata.PolicyVersion == "" {
		data.Metadata.PolicyVersion = "v1"
	}
	if opts.RedactSensitive {
		data.Redaction = "ON"
		data.Unredacted = false
	}
	if opts.UnsafeConsent {
		data.Metadata.UnsafeConsent = "yes"
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// SaveSecurityHTML writes HTML report to the target path.
func SaveSecurityHTML(path string, results []scanner.Result, findings []cve.Match, now time.Time) error {
	return SaveSecurityHTMLWithRisk(path, results, findings, nil, now)
}

// SaveSecurityHTMLWithRisk writes HTML report including risk signatures.
func SaveSecurityHTMLWithRisk(path string, results []scanner.Result, findings []cve.Match, risks []risksignature.Finding, now time.Time) error {
	return SaveSecurityHTMLWithRiskOptions(path, results, findings, risks, now, Options{RedactSensitive: true})
}

// SaveSecurityHTMLWithRiskOptions writes HTML report including risk signatures and custom options.
func SaveSecurityHTMLWithRiskOptions(path string, results []scanner.Result, findings []cve.Match, risks []risksignature.Finding, now time.Time, opts Options) error {
	b, err := RenderSecurityHTMLWithRiskOptions(results, findings, risks, now, opts)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
