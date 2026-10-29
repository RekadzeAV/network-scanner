package main

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"

	"network-scanner/internal/contracts"
	"network-scanner/internal/plugin"
)

// CriticalPorts для генерации отчётов.
var CriticalPorts = []int{21, 22, 23, 80, 443, 445, 3389}

// alertTmpl — шаблон текстового отчёта.
var alertTmpl = template.Must(template.New("alert").Parse(`Alert Report — {{ .Date }}
Total hosts: {{ .Total }}
Hosts with critical ports: {{ .Alerts }}

Critical port alerts:
{{ range .Hosts }}  {{ .IP }} ({{ .Hostname }}):
{{ range .Ports }}    Port {{ .Port }}/{{ .Proto }} {{ .State }} ({{ .Service }})
{{ end }}
{{ end }}`))

// AlertReport — reporter: генерирует текстовый отчёт о критических портах.
type AlertReport struct {
	info plugin.Info
}

func New() (plugin.Plugin, error) {
	return &AlertReport{
		info: plugin.Info{
			Name:        "AlertReport",
			Version:     "1.0.0",
			Description: "Генерация отчёта о критических портах (21/22/23/80/443/445/3389)",
			Author:      "Network Scanner",
			Type:        plugin.TypeReporter,
		},
	}, nil
}

func (r *AlertReport) Info() plugin.Info                     { return r.info }
func (r *AlertReport) Init(cfg map[string]interface{}) error { return nil }
func (r *AlertReport) Close() error                          { return nil }

func (r *AlertReport) Run(ctx context.Context, results []contracts.ScanResult) (interface{}, error) {
	data, err := r.Generate(results, "text")
	return string(data), err
}

type alertHost struct {
	IP       string
	Hostname string
	Ports    []alertPort
}

type alertPort struct {
	Port    int
	Proto   string
	State   string
	Service string
}

func (r *AlertReport) Generate(results []contracts.ScanResult, format string) ([]byte, error) {
	switch format {
	case "text":
		return r.generateText(results)
	case "json":
		return nil, fmt.Errorf("json format not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported format: %q", format)
	}
}

func (r *AlertReport) generateText(results []contracts.ScanResult) ([]byte, error) {
	var hosts []alertHost
	for _, res := range results {
		var ports []alertPort
		for _, p := range res.Ports {
			for _, cp := range CriticalPorts {
				if p.Port == cp {
					ports = append(ports, alertPort{Port: p.Port, Proto: p.Protocol, State: p.State, Service: p.Service})
					break
				}
			}
		}
		if len(ports) > 0 {
			hosts = append(hosts, alertHost{IP: res.IP, Hostname: res.Hostname, Ports: ports})
		}
	}

	var buf bytes.Buffer
	if err := alertTmpl.Execute(&buf, map[string]interface{}{
		"Date":   time.Now().Format("2006-01-02 15:04:05"),
		"Total":  len(results),
		"Alerts": len(hosts),
		"Hosts":  hosts,
	}); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}
	return buf.Bytes(), nil
}

func main() {}
