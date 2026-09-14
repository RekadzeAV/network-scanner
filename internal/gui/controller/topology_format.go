package controller

import (
	"fmt"
	"strings"
	"time"

	"network-scanner/internal/snmpcollector"
	"network-scanner/internal/topology"
)

func (c *TopologyController) buildPerformanceReportText() string {
	if c.lastTopo == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("Отчет производительности topology build\n")
	sb.WriteString(fmt.Sprintf("Устройств: %d\n", len(c.lastTopo.Devices)))
	sb.WriteString(fmt.Sprintf("Связей: %d\n", len(c.lastTopo.Links)))
	if c.lastMetrics.snmpDuration > 0 {
		sb.WriteString(fmt.Sprintf("SNMP сбор: %s\n", c.lastMetrics.snmpDuration.Round(time.Millisecond).String()))
	}
	if c.lastMetrics.buildDuration > 0 {
		sb.WriteString(fmt.Sprintf("Построение графа: %s\n", c.lastMetrics.buildDuration.Round(time.Millisecond).String()))
	}
	if c.lastMetrics.totalDuration > 0 {
		sb.WriteString(fmt.Sprintf("Общее время: %s\n", c.lastMetrics.totalDuration.Round(time.Millisecond).String()))
	}
	if c.lastReport != nil {
		sb.WriteString(fmt.Sprintf("SNMP целей: %d\n", c.lastReport.TotalSNMPTargets))
		sb.WriteString(fmt.Sprintf("SNMP ok: %d\n", c.lastReport.Connected))
		sb.WriteString(fmt.Sprintf("SNMP partial: %d\n", c.lastReport.Partial))
		sb.WriteString(fmt.Sprintf("SNMP failed: %d\n", c.lastReport.Failed))
	}
	return sb.String()
}

func topologySuccessStatus(topo *topology.Topology, report *snmpcollector.CollectReport) string {
	if topo == nil {
		return "Топология не построена"
	}
	return fmt.Sprintf("Топология построена: устройств %d, связей %d", len(topo.Devices), len(topo.Links))
}

func formatTopologyPreview(topo *topology.Topology, report *snmpcollector.CollectReport, metrics topologyBuildMetrics) string {
	if topo == nil {
		return "## Топология сети\n\nНет данных для отображения."
	}
	var sb strings.Builder
	sb.WriteString("## Топология сети\n\n")
	if metrics.totalDuration > 0 {
		sb.WriteString("### Время этапов\n\n")
		if metrics.snmpDuration > 0 {
			sb.WriteString(fmt.Sprintf("- SNMP сбор: `%s`\n", metrics.snmpDuration.Round(time.Millisecond).String()))
		}
		if metrics.buildDuration > 0 {
			sb.WriteString(fmt.Sprintf("- Построение графа: `%s`\n", metrics.buildDuration.Round(time.Millisecond).String()))
		}
		sb.WriteString(fmt.Sprintf("- Общее время: `%s`\n\n", metrics.totalDuration.Round(time.Millisecond).String()))
	}
	if report != nil {
		sb.WriteString("### SNMP отчет\n\n")
		sb.WriteString(fmt.Sprintf("- Целей для SNMP: %d\n", report.TotalSNMPTargets))
		sb.WriteString(fmt.Sprintf("- Успешных подключений: %d\n", report.Connected))
		sb.WriteString(fmt.Sprintf("- Частичных опросов: %d\n", report.Partial))
		sb.WriteString(fmt.Sprintf("- Полных отказов: %d\n\n", report.Failed))
	}
	sb.WriteString(fmt.Sprintf("**Устройств:** %d\n\n", len(topo.Devices)))
	sb.WriteString(fmt.Sprintf("**Связей:** %d\n\n", len(topo.Links)))
	sb.WriteString("### Связи\n\n")
	if len(topo.Links) == 0 {
		sb.WriteString("- Связи не найдены.\n")
		return sb.String()
	}
	for _, link := range topo.Links {
		sourceType := strings.TrimSpace(string(link.SourceType))
		confidence := strings.TrimSpace(string(link.Confidence))
		extra := ""
		if sourceType != "" || confidence != "" {
			extra = fmt.Sprintf(" [%s/%s]", sourceType, confidence)
		}
		sb.WriteString(fmt.Sprintf("- `%s (%s)` <-> `%s (%s)`%s\n",
			topoDisplayName(link.Source), topoPortName(link.SourcePort), topoDisplayName(link.Target), topoPortName(link.TargetPort), extra))
	}
	return sb.String()
}

func topoDisplayName(d *topology.Device) string {
	if d == nil {
		return "unknown"
	}
	if d.Hostname != "" {
		return d.Hostname
	}
	if d.IP != "" {
		return d.IP
	}
	if d.MAC != "" {
		return d.MAC
	}
	return "unknown"
}

func topoPortName(p *topology.Port) string {
	if p == nil {
		return "-"
	}
	if p.Name != "" {
		return p.Name
	}
	if p.Index > 0 {
		return fmt.Sprintf("if%d", p.Index)
	}
	return "-"
}

func splitCommaValues(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"public"}
	}
	return out
}

func partialSNMPKeysFromReport(report *snmpcollector.CollectReport) map[string]struct{} {
	if report == nil {
		return nil
	}
	out := make(map[string]struct{})
	for _, f := range report.Failures {
		if f.Kind != snmpcollector.FailureQuery {
			continue
		}
		ip := strings.TrimSpace(strings.ToLower(f.IP))
		if ip != "" {
			out["ip:"+ip] = struct{}{}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
