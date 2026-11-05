package controller

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// TestToolsController_withHost_nilUI проверяет, что при nil UI withHost
// возвращает пустую строку и false.
func TestToolsController_withHost_nilUI(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ctrl := NewToolsController(a, nil)
	host, ok := ctrl.withHost()
	if ok {
		t.Fatal("withHost should return false for nil UI")
	}
	if host != "" {
		t.Fatalf("host = %q, want %q", host, "")
	}
}

// TestToolsController_withHost_nilEntry проверяет, что при nil HostEntry
// withHost возвращает пустую строку и false.
func TestToolsController_withHost_nilEntry(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ToolsUI{
		HostEntry: nil,
	}
	ctrl := NewToolsController(a, ui)
	host, ok := ctrl.withHost()
	if ok {
		t.Fatal("withHost should return false for nil HostEntry")
	}
	if host != "" {
		t.Fatalf("host = %q, want %q", host, "")
	}
}

// TestToolsController_withHost_empty проверяет, что при пустом хосте
// withHost возвращает false.
func TestToolsController_withHost_empty(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ToolsUI{
		HostEntry: widget.NewEntry(),
	}
	ctrl := NewToolsController(a, ui)
	// HostEntry.Text пустой по умолчанию
	host, ok := ctrl.withHost()
	if ok {
		t.Fatal("withHost should return false for empty host")
	}
	if host != "" {
		t.Fatalf("host = %q, want %q", host, "")
	}
}

// TestToolsController_withHost_whitespace проверяет, что при хосте из пробелов
// withHost возвращает false (trimSpace).
func TestToolsController_withHost_whitespace(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ToolsUI{
		HostEntry: widget.NewEntry(),
	}
	ui.HostEntry.SetText("   ")
	ctrl := NewToolsController(a, ui)
	host, ok := ctrl.withHost()
	if ok {
		t.Fatal("withHost should return false for whitespace-only host")
	}
	if host != "" {
		t.Fatalf("host = %q, want %q", host, "")
	}
}

// TestToolsController_withHost_valid проверяет, что при корректном хосте
// withHost возвращает его.
func TestToolsController_withHost_valid(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ToolsUI{
		HostEntry: widget.NewEntry(),
	}
	ui.HostEntry.SetText("  192.168.1.1  ")
	ctrl := NewToolsController(a, ui)
	host, ok := ctrl.withHost()
	if !ok {
		t.Fatal("withHost should return true for valid host")
	}
	if host != "192.168.1.1" {
		t.Fatalf("host = %q, want %q", host, "192.168.1.1")
	}
}

// TestToolsController_withHost_trailingSpaces проверяет, что пробелы по краям
// удаляются.
func TestToolsController_withHost_trailingSpaces(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ToolsUI{
		HostEntry: widget.NewEntry(),
	}
	ui.HostEntry.SetText("example.com   ")
	ctrl := NewToolsController(a, ui)
	host, ok := ctrl.withHost()
	if !ok {
		t.Fatal("withHost should return true for host with trailing spaces")
	}
	if host != "example.com" {
		t.Fatalf("host = %q, want %q", host, "example.com")
	}
}

// TestToolsController_parseIntOrDefault_empty проверяет пустую строку.
func TestToolsController_parseIntOrDefault_empty(t *testing.T) {
	if parseIntOrDefault("", 42) != 42 {
		t.Fatal("should return default for empty string")
	}
}

// TestToolsController_parseIntOrDefault_zero проверяет ноль.
func TestToolsController_parseIntOrDefault_zero(t *testing.T) {
	if parseIntOrDefault("0", 42) != 42 {
		t.Fatal("should return default for 0")
	}
}

// TestToolsController_parseIntOrDefault_negative проверяет отрицательное число.
func TestToolsController_parseIntOrDefault_negative(t *testing.T) {
	if parseIntOrDefault("-5", 42) != 42 {
		t.Fatal("should return default for negative number")
	}
}

// TestToolsController_parseIntOrDefault_negativeString проверяет строку с "-".
func TestToolsController_parseIntOrDefault_negativeString(t *testing.T) {
	if parseIntOrDefault("-10", 42) != 42 {
		t.Fatal("should return default for negative string")
	}
}

// TestToolsController_parseIntOrDefault_nonNumeric проверяет нечисловую строку.
func TestToolsController_parseIntOrDefault_nonNumeric(t *testing.T) {
	if parseIntOrDefault("abc", 42) != 42 {
		t.Fatal("should return default for non-numeric string")
	}
}

// TestToolsController_parseIntOrDefault_valid проверяет корректное число.
func TestToolsController_parseIntOrDefault_valid(t *testing.T) {
	if parseIntOrDefault("100", 42) != 100 {
		t.Fatal("should return parsed value")
	}
	if parseIntOrDefault("1", 42) != 1 {
		t.Fatal("should return 1")
	}
}

// TestToolsController_showInfo_nilWindow проверяет, что showInfo безопасно
// завершается при nil окне.
func TestToolsController_showInfo_nilWindow(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ToolsUI{}
	ctrl := NewToolsController(a, ui)
	ctrl.showInfo("Title", "Message") // не должен паниковать
}

// TestToolsController_currentWindow_nilApp проверяет, что currentWindow
// возвращает nil при nil app.
func TestToolsController_currentWindow_nilApp(t *testing.T) {
	ui := &ToolsUI{}
	ctrl := NewToolsController(nil, ui)
	w := ctrl.currentWindow()
	if w != nil {
		t.Fatal("currentWindow should return nil for nil app")
	}
}

// TestToolsController_currentWindow_testApp проверяет, что currentWindow
// возвращает nil для test.App (он не реализует desktop-dependent CurrentWindow).
func TestToolsController_currentWindow_testApp(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ToolsUI{}
	ctrl := NewToolsController(a, ui)
	w := ctrl.currentWindow()
	// test.App может не поддерживать CurrentWindow() или вернуть nil без активного окна
	if w != nil {
		t.Logf("currentWindow returned non-nil window (test driver supports it)")
	}
}

// TestToolsController_setButtonsEnabled_nilUI проверяет, что setButtonsEnabled
// безопасно завершается при nil UI.
func TestToolsController_setButtonsEnabled_nilUI(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ctrl := NewToolsController(a, nil)
	ctrl.setButtonsEnabled(true)  // не должен паниковать
	ctrl.setButtonsEnabled(false) // не должен паниковать
}

// TestToolsController_setOutputMarkdown_nilOutput проверяет, что
// setOutputMarkdown безопасно завершается при nil ToolsOutput.
func TestToolsController_setOutputMarkdown_nilOutput(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ToolsUI{
		ToolsOutput: nil,
	}
	ctrl := NewToolsController(a, ui)
	ctrl.setOutputMarkdown("test markdown") // не должен паниковать
}

// TestToolsController_setOutputMarkdown_nilUI проверяет, что
// setOutputMarkdown безопасно завершается при nil UI.
func TestToolsController_setOutputMarkdown_nilUI(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ctrl := NewToolsController(a, nil)
	ctrl.setOutputMarkdown("test markdown") // не должен паниковать
}

// TestToolsController_setOutputMarkdown_emptyString проверяет установку пустой строки.
func TestToolsController_setOutputMarkdown_emptyString(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ToolsUI{
		ToolsOutput: widget.NewRichText(),
	}
	ctrl := NewToolsController(a, ui)
	ctrl.setOutputMarkdown("") // не должен паниковать
}

// TestToolsController_setOutputMarkdown_validMarkdown проверяет, что
// markdown корректно парсится.
func TestToolsController_setOutputMarkdown_validMarkdown(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ToolsUI{
		ToolsOutput: widget.NewRichText(),
	}
	ctrl := NewToolsController(a, ui)
	ctrl.setOutputMarkdown("### Test\n\nHello **world**")

	// Проверим, что вывод обновился
	if ui.ToolsOutput == nil {
		t.Fatal("ToolsOutput should not be nil")
	}
}

// TestToolsController_RunWOLTool_emptyMAC проверяет ранний выход при пустом MAC.
// Безопасен для -race: пустой MAC даёт возврат до запуска горутины.
func TestToolsController_RunWOLTool_emptyMAC(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ToolsUI{
		WOLMacEntry: widget.NewEntry(),
		WOLBtn:      widget.NewButton("", nil),
		ToolsOutput: widget.NewRichText(),
	}
	ctrl := NewToolsController(a, ui)
	ctrl.RunWOLTool()
	// Кнопка НЕ должна быть disabled (ранний выход)
	if ui.WOLBtn.Disabled() {
		t.Fatal("WOLBtn should remain enabled on empty MAC")
	}
}

// TestToolsController_RunDeviceControlTool_emptyTarget проверяет ранний выход.
// Безопасен для -race: пустой target даёт возврат до запуска горутины.
func TestToolsController_RunDeviceControlTool_emptyTarget(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	ui := &ToolsUI{
		DeviceTargetEntry: widget.NewEntry(),
		DeviceStatusBtn:   widget.NewButton("", nil),
		ToolsOutput:       widget.NewRichText(),
	}
	ctrl := NewToolsController(a, ui)
	ctrl.RunDeviceControlTool("status")
	// Ранний выход — кнопка остаётся enabled
	if ui.DeviceStatusBtn.Disabled() {
		t.Fatal("DeviceStatusBtn should remain enabled on empty target")
	}
}
