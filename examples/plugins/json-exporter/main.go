package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"network-scanner/internal/contracts"
	"network-scanner/internal/plugin"
)

// JSONExporter — exporter: экспортирует результаты в JSON Lines.
type JSONExporter struct {
	info plugin.Info
}

// jsonRow — одна строка JSON Lines.
type jsonRow struct {
	IP         string        `json:"ip"`
	Hostname   string        `json:"hostname"`
	DeviceType string        `json:"device_type"`
	GuessOS    string        `json:"guess_os"`
	Ports      []portSummary `json:"ports"`
}

// portSummary — порт в строке экспорта.
type portSummary struct {
	Port     int    `json:"port"`
	State    string `json:"state"`
	Protocol string `json:"protocol"`
	Service  string `json:"service"`
}

func New() (plugin.Plugin, error) {
	return &JSONExporter{
		info: plugin.Info{
			Name:        "JSONExporter",
			Version:     "1.0.0",
			Description: "Экспорт результатов в формат JSON Lines",
			Author:      "Network Scanner",
			Type:        plugin.TypeExporter,
		},
	}, nil
}

func (e *JSONExporter) Info() plugin.Info { return e.info }

func (e *JSONExporter) Init(cfg map[string]interface{}) error { return nil }

func (e *JSONExporter) Run(ctx context.Context, results []contracts.ScanResult) (interface{}, error) {
	rows := make([]jsonRow, len(results))
	for i, r := range results {
		ps := make([]portSummary, len(r.Ports))
		for j, p := range r.Ports {
			ps[j] = portSummary{Port: p.Port, State: p.State, Protocol: p.Protocol, Service: p.Service}
		}
		rows[i] = jsonRow{
			IP: r.IP, Hostname: r.Hostname, DeviceType: r.DeviceType,
			GuessOS: r.GuessOS, Ports: ps,
		}
	}

	var sb strings.Builder
	enc := json.NewEncoder(&sb)
	for _, row := range rows {
		if err := enc.Encode(row); err != nil {
			return nil, fmt.Errorf("json encode: %w", err)
		}
	}
	return sb.String(), nil
}

func (e *JSONExporter) Close() error { return nil }

func (e *JSONExporter) Export(results []contracts.ScanResult, path string) error {
	out, err := e.Run(context.Background(), results)
	if err != nil {
		return err
	}
	data := out.(string)
	return writeToFile(path, data)
}

func (e *JSONExporter) Format() string { return "jsonl" }

func main() {}
