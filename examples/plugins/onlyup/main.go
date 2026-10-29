package main

import (
	"context"

	"network-scanner/internal/contracts"
	"network-scanner/internal/plugin"
)

// OnlyUpFilter — filter: пропускает только хосты с хотя бы одним открытым портом.
type OnlyUpFilter struct {
	info plugin.Info
}

// New конструктор для динамического загрузчика.
func New() (plugin.Plugin, error) {
	return &OnlyUpFilter{
		info: plugin.Info{
			Name:        "OnlyUp",
			Version:     "1.0.0",
			Description: "Фильтр: пропускает только хосты с открытыми портами (state == open)",
			Author:      "Network Scanner",
			Type:        plugin.TypeFilter,
		},
	}, nil
}

func (f *OnlyUpFilter) Info() plugin.Info { return f.info }

func (f *OnlyUpFilter) Init(cfg map[string]interface{}) error { return nil }

func (f *OnlyUpFilter) Run(ctx context.Context, results []contracts.ScanResult) (interface{}, error) {
	return f.Filter(results)
}

func (f *OnlyUpFilter) Close() error { return nil }

func (f *OnlyUpFilter) Filter(results []contracts.ScanResult) ([]contracts.ScanResult, error) {
	out := make([]contracts.ScanResult, 0, len(results))
	for _, r := range results {
		for _, p := range r.Ports {
			if p.State == "open" {
				out = append(out, r)
				break
			}
		}
	}
	return out, nil
}

func main() {}
