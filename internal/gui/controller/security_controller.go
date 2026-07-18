package controller

import (
	"fmt"
	"strings"

	"network-scanner/internal/audit"
	"network-scanner/internal/logger"
	"network-scanner/internal/wol"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// SecurityUI предоставляет доступ к виджетам безопасности и управления устройствами.
type SecurityUI struct {
	AuditMinSeveritySel *widget.Select
	DeviceTargetEntry   *widget.Entry
	DeviceVendorEntry   *widget.Select
	DeviceUserEntry     *widget.Entry
	DevicePassEntry     *widget.Entry
	DeviceStatusBtn     *widget.Button
	DeviceRebootBtn     *widget.Button
	WOLMacEntry         *widget.Entry
	WOLBcastEntry       *widget.Entry
	WOLIfaceEntry       *widget.Entry
	StatusLabel         *widget.Label
	Window              fyne.Window
}

// SecurityController управляет безопасностью и управлением устройствами.
type SecurityController struct {
	app fyne.App
	ui  *SecurityUI
}

// NewSecurityController создает контроллер безопасности.
func NewSecurityController(app fyne.App, ui *SecurityUI) *SecurityController {
	return &SecurityController{
		app: app,
		ui:  ui,
	}
}

// RunAudit запускает аудит открытых портов.
func (c *SecurityController) RunAudit(results []interface{}, minSeverity string) {
	if len(results) == 0 {
		c.setStatus("Нет данных для аудита (выполните сканирование)")
		return
	}
	if minSeverity == "" {
		minSeverity = "low"
	}
	_, ok := audit.NormalizeSeverity(minSeverity)
	if !ok {
		minSeverity = "low"
	}
	c.setStatus("Выполняется аудит открытых портов...")
	logger.Log("Аудит портов: severity=%s, хостов=%d", minSeverity, len(results))

	// Конвертируем interface{} -> scanner.Result для аудита
	// TODO: При полной миграции передать []scanner.Result напрямую
	findings := audit.EvaluateOpenPorts(nil) // TODO: передать реальные данные
	findings = audit.FilterByMinSeverity(findings, minSeverity)
	report := audit.FormatFindings(findings)

	c.setStatus(fmt.Sprintf("Аудит завершен. Найдено проблем: %d", len(findings)))
	logger.Log("Аудит завершен. Проблем: %d", len(findings))
	_ = report // TODO: вывести отчет в UI
}

// CheckDeviceStatus проверяет статус устройства через HTTP API.
func (c *SecurityController) CheckDeviceStatus() {
	target := ""
	if c.ui.DeviceTargetEntry != nil {
		target = strings.TrimSpace(c.ui.DeviceTargetEntry.Text)
	}
	if target == "" {
		c.setStatus("Введите URL устройства")
		return
	}

	vendor := ""
	if c.ui.DeviceVendorEntry != nil {
		vendor = c.ui.DeviceVendorEntry.Selected
	}

	user := ""
	_ = user // TODO: передать в devicecontrol
	pass := ""
	_ = pass // TODO: передать в devicecontrol

	c.setStatus("Проверка статуса устройства...")
	logger.Log("Проверка статуса: target=%s, vendor=%s", target, vendor)

	// TODO: Реализовать вызов devicecontrol.Execute
	// status, err := devicecontrol.GetStatus(target, vendor, user, pass)
	// if err != nil {
	//     c.setStatus(fmt.Sprintf("Ошибка: %v", err))
	//     return
	// }
	// c.setStatus(fmt.Sprintf("Статус: %s", status))
	c.setStatus("Статус устройства: OK (заглушка)")
}

// RebootDevice перезагружает устройство через HTTP API.
func (c *SecurityController) RebootDevice() {
	target := ""
	if c.ui.DeviceTargetEntry != nil {
		target = strings.TrimSpace(c.ui.DeviceTargetEntry.Text)
	}
	if target == "" {
		c.setStatus("Введите URL устройства")
		return
	}

	vendor := ""
	if c.ui.DeviceVendorEntry != nil {
		vendor = c.ui.DeviceVendorEntry.Selected
	}

	user := ""
	_ = user // TODO: передать в devicecontrol
	pass := ""
	_ = pass // TODO: передать в devicecontrol

	c.setStatus("Перезагрузка устройства...")
	logger.Log("Перезагрузка: target=%s, vendor=%s", target, vendor)

	// TODO: Реализовать вызов devicecontrol.Execute
	// err := devicecontrol.Reboot(target, vendor, user, pass)
	// if err != nil {
	//     c.setStatus(fmt.Sprintf("Ошибка: %v", err))
	//     return
	// }
	c.setStatus("Устройство перезагружено (заглушка)")
}

// WakeOnLAN отправляет Wake-on-LAN пакет.
func (c *SecurityController) WakeOnLAN() {
	mac := ""
	if c.ui.WOLMacEntry != nil {
		mac = strings.TrimSpace(c.ui.WOLMacEntry.Text)
	}
	if mac == "" {
		c.setStatus("Введите MAC-адрес для WoL")
		return
	}

	bcast := ""
	if c.ui.WOLBcastEntry != nil {
		bcast = strings.TrimSpace(c.ui.WOLBcastEntry.Text)
	}

	iface := ""
	if c.ui.WOLIfaceEntry != nil {
		iface = strings.TrimSpace(c.ui.WOLIfaceEntry.Text)
	}

	c.setStatus("Отправка WoL-пакета...")
	logger.Log("WoL: MAC=%s, bcast=%s, iface=%s", mac, bcast, iface)

	if err := wol.SendMagicPacket(mac, bcast); err != nil {
		c.setStatus(fmt.Sprintf("Ошибка WoL: %v", err))
		logger.LogError(err, "Wake-on-LAN")
		dialog.ShowError(err, c.ui.Window)
		return
	}

	c.setStatus("WoL-пакет успешно отправлен на MAC: " + mac)
	logger.Log("WoL отправлен на MAC: %s", mac)
}

// RunRiskSignatures запускает проверку сигнатур уязвимостей.
func (c *SecurityController) RunRiskSignatures(results []interface{}) {
	if len(results) == 0 {
		c.setStatus("Нет данных для проверки (выполните сканирование)")
		return
	}

	c.setStatus("Проверка сигнатур уязвимостей...")
	logger.Log("Запуск проверки сигнатур на %d хостах", len(results))

	// TODO: При полной миграции передать []scanner.Result напрямую
	// db, err := risksignature.LoadDefault()
	// if err != nil {
	//     c.setStatus(fmt.Sprintf("Ошибка загрузки базы: %v", err))
	//     return
	// }
	// findings := risksignature.Evaluate(results, db)
	// c.setStatus(fmt.Sprintf("Найдено проблем: %d", len(findings)))

	c.setStatus("Проверка завершена (заглушка)")
	logger.Log("Проверка сигнатур завершена (заглушка)")
}

// setStatus устанавливает статус в UI.
func (c *SecurityController) setStatus(msg string) {
	if c.ui != nil && c.ui.StatusLabel != nil {
		c.ui.StatusLabel.SetText(msg)
		c.ui.StatusLabel.Refresh()
	}
}
