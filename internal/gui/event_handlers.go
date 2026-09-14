package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2/dialog"

	"network-scanner/internal/devicecontrol"
)

// setupEventHandlers настраивает обработчики событий
func (a *App) setupEventHandlers() {
	a.scanButton.OnTapped = func() {
		a.scanCtrl.StartScan(a.scanResults)
	}
	a.stopButton.OnTapped = func() {
		a.scanCtrl.StopScan()
	}
	a.presetQuickBtn.OnTapped = func() {
		a.scanCtrl.ApplyPreset("quick")
	}
	a.presetBalBtn.OnTapped = func() {
		a.scanCtrl.ApplyPreset("balanced")
	}
	a.presetDeepBtn.OnTapped = func() {
		a.scanCtrl.ApplyPreset("deep")
	}
	if a.recommendedProfileBtn != nil {
		a.recommendedProfileBtn.OnTapped = func() {
			a.scanCtrl.ApplyRecommendedProfile(a.networkEntry.Text)
		}
	}
	if a.recommendedProfileInfoBtn != nil {
		a.recommendedProfileInfoBtn.OnTapped = func() {
			dialog.ShowInformation(
				"Логика рекомендованного профиля",
				"Кнопка подбирает безопасные параметры под размер подсети:\n\n"+
					"- небольшие сети: чуть глубже диапазон портов;\n"+
					"- средние/крупные: диапазон и параллелизм умеренные;\n"+
					"- очень крупные: минимально нагружающий профиль.\n\n"+
					"Во всех случаях для стабильности отключаются: UDP, баннеры,\n"+
					"active-эвристики ОС и детальные портовые логи.",
				a.myWindow,
			)
		}
	}
	a.networkEntry.OnChanged = func(_ string) {
		a.saveScanSettings()
	}
	a.portRangeEntry.OnChanged = func(_ string) {
		a.saveScanSettings()
	}
	a.timeoutEntry.OnChanged = func(_ string) {
		a.saveScanSettings()
	}
	a.threadsEntry.OnChanged = func(_ string) {
		a.saveScanSettings()
	}
	if a.inventoryDBEntry != nil {
		a.inventoryDBEntry.OnChanged = func(_ string) {
			a.saveScanSettings()
		}
	}
	if a.inventoryAutoSaveCheck != nil {
		a.inventoryAutoSaveCheck.OnChanged = func(_ bool) {
			a.saveScanSettings()
		}
	}
	a.scanUDPCheck.OnChanged = func(_ bool) {
		a.saveScanSettings()
	}
	a.scanBannersCheck.OnChanged = func(_ bool) {
		a.saveScanSettings()
	}
	a.scanOSActiveCheck.OnChanged = func(_ bool) {
		a.saveScanSettings()
	}
	if a.scanVerboseLogsCheck != nil {
		a.scanVerboseLogsCheck.OnChanged = func(_ bool) {
			a.saveScanSettings()
		}
	}
	if a.scanVerboseInfoBtn != nil {
		a.scanVerboseInfoBtn.OnTapped = func() {
			dialog.ShowInformation(
				"Детальные логи по портам",
				"Режим включает debug-логи по отдельным TCP/UDP probe.\n\n"+
					"Используйте только для диагностики:\n"+
					"- заметно увеличивает размер лога;\n"+
					"- может замедлить сканирование на больших диапазонах.\n\n"+
					"Для обычной работы оставляйте опцию выключенной.",
				a.myWindow,
			)
		}
	}
	if a.autoProfileCheck != nil {
		a.autoProfileCheck.OnChanged = func(_ bool) {
			a.refreshAutoProfileStateLabel()
			a.saveScanSettings()
		}
	}
	if a.autoProfileInfoBtn != nil {
		a.autoProfileInfoBtn.OnTapped = func() {
			dialog.ShowInformation(
				"Автопрофиль сканирования",
				fmt.Sprintf(
					"Автопрофиль снижает риск перегрузки сети и UI на больших диапазонах.\n\n"+
						"Логика:\n"+
						"- от ~%d хостов: мягкое ограничение при слишком тяжелых настройках\n"+
						"- от ~%d хостов: до ports=1-2000 и threads<=64\n"+
						"- от ~%d хостов: до ports=1-1024 и threads<=40\n"+
						"- от ~%d хостов: до ports=1-512 и threads<=24\n\n"+
						"Опцию можно отключить, если нужен полный ручной контроль.",
					autoProfileHostWarn,
					autoProfileHostLarge,
					autoProfileHostXLarge,
					autoProfileHostXXLarge,
				),
				a.myWindow,
			)
		}
	}
	a.portWellKnownBtn.OnTapped = func() {
		a.portRangeEntry.SetText("0-1023")
		a.saveScanSettings()
	}
	a.portRegisteredBtn.OnTapped = func() {
		a.portRangeEntry.SetText("1024-49151")
		a.saveScanSettings()
	}
	a.portDynamicBtn.OnTapped = func() {
		a.portRangeEntry.SetText("49152-65535")
		a.saveScanSettings()
	}

	a.saveButton.OnTapped = func() {
		a.saveResults()
	}
	a.buildTopoBtn.OnTapped = func() {
		a.topoCtrl.BuildTopology(a.scanResults, a.myWindow)
	}
	a.stopTopoBtn.OnTapped = func() {
		a.topoCtrl.StopTopologyBuild()
	}
	a.saveTopoBtn.OnTapped = func() {
		a.topoCtrl.SaveTopology(a.lastTopology, a.myWindow)
	}
	a.copyPerfBtn.OnTapped = func() {
		a.topoCtrl.CopyPerformanceReport(a.myWindow)
	}
	if a.copyDiagnosticsBtn != nil {
		a.copyDiagnosticsBtn.OnTapped = func() {
			if a.diagnosticsLabel != nil {
				a.scanCtrl.CopyScanDiagnostics(a.diagnosticsLabel.Text)
			}
		}
	}
	if a.saveDiagnosticsBtn != nil {
		a.saveDiagnosticsBtn.OnTapped = func() {
			if a.diagnosticsLabel != nil {
				a.scanCtrl.SaveScanDiagnostics(a.diagnosticsLabel.Text)
			}
		}
	}
	a.savePerfBtn.OnTapped = func() {
		a.topoCtrl.SavePerformanceReport(a.myWindow)
	}
	a.refreshPreviewBtn.OnTapped = func() {
		a.topoCtrl.RefreshTopologyPreview(a.lastTopology, a.myWindow)
	}
	a.openPreviewBtn.OnTapped = func() {
		a.topoCtrl.OpenPreviewExternal(a.previewPath, a.myWindow)
	}
	a.zoomSelect.OnChanged = func(value string) {
		a.topoCtrl.ApplyTopologyZoom(value, a.myWindow)
	}
	if a.topologySearchEntry != nil {
		a.topologySearchEntry.OnChanged = func(v string) {
			a.topologyViewState.query = strings.TrimSpace(v)
			a.renderTopologyInteractiveMap(a.lastTopology)
		}
	}
	if a.topologyTypeFilterSel != nil {
		a.topologyTypeFilterSel.OnChanged = func(v string) {
			a.topologyViewState.typeFilter = strings.TrimSpace(strings.ToLower(v))
			a.renderTopologyInteractiveMap(a.lastTopology)
		}
	}
	if a.topologyConfidenceFilterSel != nil {
		a.topologyConfidenceFilterSel.OnChanged = func(v string) {
			a.topologyViewState.confFilter = strings.TrimSpace(strings.ToLower(v))
			a.renderTopologyInteractiveMap(a.lastTopology)
		}
	}
	if a.topologyResetMapBtn != nil {
		a.topologyResetMapBtn.OnTapped = func() {
			a.topologyViewState = topologyMapState{}
			if a.topologySearchEntry != nil {
				a.topologySearchEntry.SetText("")
			}
			if a.topologyTypeFilterSel != nil {
				a.topologyTypeFilterSel.SetSelected("all")
			}
			if a.topologyConfidenceFilterSel != nil {
				a.topologyConfidenceFilterSel.SetSelected("all")
			}
			a.renderTopologyInteractiveMap(a.lastTopology)
		}
	}
	if a.toolsHostEntry != nil {
		a.toolsHostEntry.OnChanged = func(_ string) {
			a.saveScanSettings()
		}
	}
	if a.toolsPingCountEnt != nil {
		a.toolsPingCountEnt.OnChanged = func(_ string) {
			a.saveScanSettings()
		}
	}
	if a.toolsTimeoutEnt != nil {
		a.toolsTimeoutEnt.OnChanged = func(_ string) {
			a.saveScanSettings()
		}
	}
	if a.toolsTraceHopsEnt != nil {
		a.toolsTraceHopsEnt.OnChanged = func(_ string) {
			a.saveScanSettings()
		}
	}
	if a.toolsDNSResolverEnt != nil {
		a.toolsDNSResolverEnt.OnChanged = func(_ string) {
			a.saveScanSettings()
		}
	}
	if a.toolsDeviceTargetEntry != nil {
		a.toolsDeviceTargetEntry.OnChanged = func(_ string) {
			a.saveScanSettings()
		}
	}
	if a.toolsDeviceUserEntry != nil {
		a.toolsDeviceUserEntry.OnChanged = func(_ string) {
			a.saveScanSettings()
		}
	}
	if a.toolsDevicePassEntry != nil {
		a.toolsDevicePassEntry.OnChanged = func(_ string) {
			a.saveScanSettings()
		}
	}
	a.toolsPingBtn.OnTapped = func() {
		a.toolsCtrl.RunPingTool()
	}
	a.toolsTraceBtn.OnTapped = func() {
		a.toolsCtrl.RunTracerouteTool()
	}
	a.toolsDNSBtn.OnTapped = func() {
		a.toolsCtrl.RunDNSTool()
	}
	a.toolsWhoisBtn.OnTapped = func() {
		a.toolsCtrl.RunWhoisTool()
	}
	a.toolsWiFiBtn.OnTapped = func() {
		a.toolsCtrl.RunWiFiTool()
	}
	a.toolsWOLBtn.OnTapped = func() {
		a.toolsCtrl.RunWOLTool()
	}
	a.toolsAuditBtn.OnTapped = func() {
		a.toolsCtrl.RunPortAuditTool(a.scanResults)
	}
	a.toolsRiskBtn.OnTapped = func() {
		a.toolsCtrl.RunRiskSignaturesTool(a.scanResults)
	}
	a.toolsDeviceStatusBtn.OnTapped = func() {
		a.toolsCtrl.RunDeviceControlTool(devicecontrol.ActionStatus)
	}
	a.toolsDeviceRebootBtn.OnTapped = func() {
		dialog.NewConfirm(
			"Подтверждение опасного действия",
			"Подтвердите reboot устройства. Действие может привести к кратковременной недоступности сети.",
			func(ok bool) {
				if !ok {
					return
				}
				a.runDeviceControlTool(devicecontrol.ActionReboot)
			},
			a.myWindow,
		).Show()
	}
}
