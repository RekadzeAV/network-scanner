package gui

import (
	"fmt"
	"strings"

	"network-scanner/internal/snmpcollector"
	"network-scanner/internal/topology"
)

func (a *App) applyTopologyRunStart() {
	if a == nil {
		return
	}
	a.buildTopoBtn.Disable()
	a.stopTopoBtn.Enable()
	a.saveTopoBtn.Disable()
	a.copyPerfBtn.Disable()
	a.savePerfBtn.Disable()
	a.refreshPreviewBtn.Disable()
	a.openPreviewBtn.Disable()
	a.statusLabel.SetText("Сбор SNMP данных и построение топологии...")
	a.topologyStatus.SetText("Сбор SNMP данных и построение топологии...")
	a.snmpStageLabel.SetText("SNMP: подготовка...")
	a.snmpStageLabel.Show()
	a.snmpProgress.SetValue(0)
	a.snmpProgress.Show()
	a.statusLabel.Refresh()
	a.topologyStatus.Refresh()
	a.snmpStageLabel.Refresh()
	a.snmpProgress.Refresh()
}

func (a *App) applyTopologyProgress(status string, progressValue float64) {
	if a == nil {
		return
	}
	a.statusLabel.SetText(status)
	a.topologyStatus.SetText(status)
	a.snmpStageLabel.SetText(status)
	a.snmpProgress.SetValue(progressValue)
	a.statusLabel.Refresh()
	a.topologyStatus.Refresh()
	a.snmpStageLabel.Refresh()
	a.snmpProgress.Refresh()
}

func (a *App) applyTopologyCanceled() {
	if a == nil {
		return
	}
	a.statusLabel.SetText("Построение топологии остановлено пользователем")
	a.topologyStatus.SetText("Построение топологии остановлено пользователем")
	a.snmpStageLabel.SetText("SNMP: остановлено")
	a.snmpProgress.Hide()
	a.buildTopoBtn.Enable()
	a.stopTopoBtn.Disable()
	a.topologyCancel = nil
	a.statusLabel.Refresh()
	a.topologyStatus.Refresh()
	a.snmpStageLabel.Refresh()
	a.snmpProgress.Refresh()
}

func (a *App) applyTopologyFailure(stage string) {
	if a == nil {
		return
	}
	a.buildTopoBtn.Enable()
	a.stopTopoBtn.Disable()
	a.topologyCancel = nil
	switch strings.TrimSpace(stage) {
	case "snmp":
		a.topologyStatus.SetText("Ошибка SNMP опроса")
		a.snmpStageLabel.SetText("SNMP: ошибка")
	default:
		a.topologyStatus.SetText("Ошибка построения топологии")
		a.snmpStageLabel.SetText("Построение топологии: ошибка")
	}
	a.snmpProgress.Hide()
	a.topologyStatus.Refresh()
	a.snmpStageLabel.Refresh()
	a.snmpProgress.Refresh()
}

func (a *App) applyTopologySuccess(topoStatus string, topoPreview string, topo *topology.Topology, report *snmpcollector.CollectReport, metrics topologyBuildMetrics) {
	if a == nil {
		return
	}
	a.lastTopology = topo
	a.lastSNMPReport = report
	a.lastTopoMetric = metrics
	a.saveTopoBtn.Enable()
	a.copyPerfBtn.Enable()
	a.savePerfBtn.Enable()
	a.refreshPreviewBtn.Enable()
	a.openPreviewBtn.Enable()
	a.statusLabel.SetText(topoStatus)
	a.topologyStatus.SetText(topoStatus)
	a.snmpStageLabel.SetText("SNMP: завершено")
	a.snmpProgress.SetValue(1)
	a.snmpProgress.Hide()
	a.topologyText.ParseMarkdown(topoPreview)
	a.topologyScroll.ScrollToTop()
	a.statusLabel.Refresh()
	a.topologyStatus.Refresh()
	a.snmpStageLabel.Refresh()
	a.snmpProgress.Refresh()
	a.topologyScroll.Refresh()
	a.topologyText.Refresh()
	a.mainTabs.SelectTabIndex(1)
	a.buildTopoBtn.Enable()
	a.stopTopoBtn.Disable()
	a.topologyCancel = nil
}

func topologySuccessStatus(topo *topology.Topology, report *snmpcollector.CollectReport) string {
	status := fmt.Sprintf("Топология построена: устройств %d, связей %d", len(topo.Devices), len(topo.Links))
	if report != nil {
		status = fmt.Sprintf("%s | SNMP: целей %d, ok %d, partial %d, failed %d",
			status, report.TotalSNMPTargets, report.Connected, report.Partial, report.Failed)
	}
	return status
}
