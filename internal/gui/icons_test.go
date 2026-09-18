package gui

import (
	"testing"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"network-scanner/internal/scanner"
)

// TestThemeIconsNonNil проверяет, что все обёртки icons.go возвращают
// валидный non-nil ресурс (иконки встроены в Fyne, внешних ресурсов нет).
func TestThemeIconsNonNil(t *testing.T) {
	skipHeadless(t)
	_ = test.NewApp()
	cases := map[string]func() fyne.Resource{
		"iconScan":       iconScan,
		"iconStop":       iconStop,
		"iconSave":       iconSave,
		"iconRefresh":    iconRefresh,
		"iconCopy":       iconCopy,
		"iconClear":      iconClear,
		"iconRestore":    iconRestore,
		"iconTopology":   iconTopology,
		"iconSearch":     iconSearch,
		"iconSecurity":   iconSecurity,
		"iconInfo":       iconInfo,
		"iconHelp":       iconHelp,
		"iconAccount":    iconAccount,
		"iconRoute":      iconRoute,
		"iconFast":       iconFast,
		"iconDeep":       iconDeep,
		"iconConfirm":    iconConfirm,
		"iconAdd":        iconAdd,
		"iconInspect":    iconInspect,
		"iconCancel":     iconCancel,
		"iconExport":     iconExport,
		"iconDevice":     iconDevice,
		"iconInventory":  iconInventory,
		"iconSettings":   iconSettings,
		"iconFullScreen": iconFullScreen,
		"iconDocument":   iconDocument,
		"iconMore":       iconMore,
		"iconPalette":    iconPalette,
		"iconAccent":     iconAccent,
	}
	for name, fn := range cases {
		t.Run(name, func(t *testing.T) {
			res := fn()
			if res == nil {
				t.Fatalf("%s() вернул nil", name)
			}
			if res.Name() == "" {
				t.Fatalf("%s() вернул ресурс с пустым именем", name)
			}
			if len(res.Content()) == 0 {
				t.Fatalf("%s() вернул пустой ресурс", name)
			}
		})
	}
}

// TestDeviceRebootButtonDangerImportance — опасное действие визуально выделено.
func TestDeviceRebootButtonDangerImportance(t *testing.T) {
	skipHeadless(t)
	fyneApp := fyneapp.New()
	fyneWindow := fyneApp.NewWindow("Test")

	a := &App{
		myApp:           fyneApp,
		myWindow:        fyneWindow,
		scanResults:     make([]scanner.Result, 0),
		quickTypeChecks: make(map[string]*widget.Check),
	}
	a.initUI()

	if a.toolsDeviceRebootBtn == nil {
		t.Fatal("toolsDeviceRebootBtn не создан")
	}
	if a.toolsDeviceRebootBtn.Importance != widget.DangerImportance {
		t.Fatalf("Device Reboot: Importance = %v, ожидался DangerImportance", a.toolsDeviceRebootBtn.Importance)
	}
	if a.toolsDeviceRebootBtn.Text != "Device Reboot" {
		t.Fatalf("Device Reboot: текст изменился: %q", a.toolsDeviceRebootBtn.Text)
	}
}
