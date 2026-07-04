package presenter

import (
	"fmt"
	"html/template"
	"os"
	"time"

	"network-scanner/internal/scanner"
)

// HTMLPresenter exports scan results to an HTML report.
type HTMLPresenter struct{}

// htmlReport represents the full report structure for HTML templating.
type htmlReport struct {
	Title       string
	GeneratedAt string
	TotalHosts  int
	OpenPorts   int
	Hosts       []htmlHost
}

// htmlHost represents a single host in the HTML report.
type htmlHost struct {
	IP           string
	MAC          string
	Hostname     string
	DeviceType   string
	DeviceVendor string
	Ports        []htmlPort
	SNMPEnabled  bool
	GuessOS      string
}

// htmlPort represents a port in the HTML report.
type htmlPort struct {
	Port     int
	Protocol string
	State    string
	Service  string
	Version  string
}

// DisplayHeader is a no-op for HTML presenter.
func (p HTMLPresenter) DisplayHeader() {}

// DisplayHost is a no-op for HTML presenter.
func (p HTMLPresenter) DisplayHost(host scanner.HostResult) { _ = host }

// DisplaySummary is a no-op for HTML presenter.
func (p HTMLPresenter) DisplaySummary(totalHosts int, openPortsCount int) {
	_, _ = totalHosts, openPortsCount
}

// Export saves scan results to an HTML file.
func (p HTMLPresenter) Export(results []scanner.HostResult, format string) error {
	if format != "html" {
		return fmt.Errorf("HTMLPresenter supports only html format, got %s", format)
	}

	report := htmlReport{
		Title:       "Network Scanner Report",
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		TotalHosts:  len(results),
		OpenPorts:   countOpenPorts(results),
		Hosts:       make([]htmlHost, 0, len(results)),
	}

	for _, r := range results {
		h := htmlHost{
			IP:           r.IP,
			MAC:          r.MAC,
			Hostname:     r.Hostname,
			DeviceType:   r.DeviceType,
			DeviceVendor: r.DeviceVendor,
			SNMPEnabled:  r.SNMPEnabled,
			GuessOS:      r.GuessOS,
			Ports:        make([]htmlPort, 0, len(r.Ports)),
		}

		for _, port := range r.Ports {
			h.Ports = append(h.Ports, htmlPort{
				Port:     port.Port,
				Protocol: port.Protocol,
				State:    port.State,
				Service:  port.Service,
				Version:  port.Version,
			})
		}

		report.Hosts = append(report.Hosts, h)
	}

	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { color: #333; border-bottom: 2px solid #007bff; padding-bottom: 10px; }
        .summary { background: #e9ecef; padding: 15px; border-radius: 4px; margin-bottom: 20px; }
        table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #dee2e6; }
        th { background: #007bff; color: white; }
        tr:hover { background: #f8f9fa; }
        .open { color: #28a745; font-weight: bold; }
        .closed { color: #dc3545; }
        .filtered { color: #ffc107; }
        .ports { display: flex; flex-wrap: wrap; gap: 4px; }
        .port-chip { background: #007bff; color: white; padding: 2px 8px; border-radius: 12px; font-size: 12px; }
        .port-chip.closed { background: #dc3545; }
        .timestamp { color: #6c757d; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>{{.Title}}</h1>
        <p class="timestamp">Generated at: {{.GeneratedAt}}</p>
        
        <div class="summary">
            <strong>Total hosts:</strong> {{.TotalHosts}} | 
            <strong>Open ports:</strong> {{.OpenPorts}}
        </div>

        <table>
            <thead>
                <tr>
                    <th>IP</th>
                    <th>Hostname</th>
                    <th>MAC</th>
                    <th>Type</th>
                    <th>Vendor</th>
                    <th>OS</th>
                    <th>SNMP</th>
                    <th>Ports</th>
                </tr>
            </thead>
            <tbody>
                {{range .Hosts}}
                <tr>
                    <td>{{.IP}}</td>
                    <td>{{.Hostname}}</td>
                    <td>{{.MAC}}</td>
                    <td>{{.DeviceType}}</td>
                    <td>{{.DeviceVendor}}</td>
                    <td>{{.GuessOS}}</td>
                    <td>{{if .SNMPEnabled}}Yes{{else}}No{{end}}</td>
                    <td>
                        <div class="ports">
                            {{range .Ports}}
                            <span class="port-chip {{.State}}">{{.Port}}/{{.Protocol}}</span>
                            {{end}}
                        </div>
                    </td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>
</body>
</html>`

	t, err := template.New("report").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	file, err := os.Create("scan-report.html")
	if err != nil {
		return fmt.Errorf("failed to create HTML file: %w", err)
	}
	defer file.Close()

	err = t.Execute(file, report)
	if err != nil {
		return fmt.Errorf("failed to render HTML template: %w", err)
	}

	fmt.Println("HTML report saved to: scan-report.html")
	return nil
}

// countOpenPorts counts total open ports across all results.
func countOpenPorts(results []scanner.HostResult) int {
	count := 0
	for _, r := range results {
		for _, port := range r.Ports {
			if port.State == "open" {
				count++
			}
		}
	}
	return count
}
