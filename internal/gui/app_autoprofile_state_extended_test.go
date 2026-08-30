package gui

import (
	"testing"

	"network-scanner/internal/scanner"

	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// --- app.go extended tests: refreshAutoProfileStateLabel, saveResultsViewSettings ---

func TestRefreshAutoProfileStateLabel_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.refreshAutoProfileStateLabel()
}

func TestRefreshAutoProfileStateLabel_NilStateText(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.refreshAutoProfileStateLabel()
}

func TestRefreshAutoProfileStateLabel_WithStateText_NilCheck(t *testing.T) {
	a := &App{}
	a.autoProfileStateText = canvas.NewText("", nil)
	// autoProfileCheck=nil — включено по умолчанию
	a.refreshAutoProfileStateLabel()
	if a.autoProfileStateText.Text != "Автопрофиль: ВКЛ" {
		t.Errorf("expected 'Автопрофиль: ВКЛ', got %q", a.autoProfileStateText.Text)
	}
}

func TestRefreshAutoProfileStateLabel_CheckEnabled(t *testing.T) {
	a := &App{}
	a.autoProfileStateText = canvas.NewText("", nil)
	a.autoProfileCheck = widget.NewCheck("", nil)
	a.autoProfileCheck.SetChecked(true)
	a.refreshAutoProfileStateLabel()
	if a.autoProfileStateText.Text != "Автопрофиль: ВКЛ" {
		t.Errorf("expected 'Автопрофиль: ВКЛ', got %q", a.autoProfileStateText.Text)
	}
}

func TestRefreshAutoProfileStateLabel_CheckDisabled(t *testing.T) {
	a := &App{}
	a.autoProfileStateText = canvas.NewText("", nil)
	a.autoProfileCheck = widget.NewCheck("", nil)
	a.autoProfileCheck.SetChecked(false)
	a.refreshAutoProfileStateLabel()
	if a.autoProfileStateText.Text != "Автопрофиль: ВЫКЛ" {
		t.Errorf("expected 'Автопрофиль: ВЫКЛ', got %q", a.autoProfileStateText.Text)
	}
}

func TestRefreshAutoProfileStateLabel_WithHeaderLabel(t *testing.T) {
	a := &App{}
	a.autoProfileStateText = canvas.NewText("", nil)
	a.autoProfileHeaderLabel = widget.NewLabel("")
	a.autoProfileCheck = widget.NewCheck("", nil)
	a.autoProfileCheck.SetChecked(true)
	a.refreshAutoProfileStateLabel()
	if a.autoProfileHeaderLabel.Text != "Режим сканирования: Автопрофиль ВКЛ" {
		t.Errorf("expected header text, got %q", a.autoProfileHeaderLabel.Text)
	}
}

func TestSaveResultsViewSettings_NilApp(t *testing.T) {
	var a *App
	// Не должен паниковать
	a.saveResultsViewSettings()
}

func TestSaveResultsViewSettings_NilMyApp(t *testing.T) {
	a := &App{}
	// Не должен паниковать
	a.saveResultsViewSettings()
}

func TestSaveResultsViewSettings_WithSettings(t *testing.T) {
	a := &App{}
	a.resultsMode = "Таблица"
	a.resultsSubMode = "Devices"
	a.resultsSort = "IP"
	a.maxPortChips = 24
	a.resultsFilterQuery = "router"
	a.resultsPortStateMode = "all"
	a.quickTypeChecks = map[string]*widget.Check{
		"Router": {Checked: true},
	}
	a.resultsCidrFilterEnt = widget.NewEntry()
	a.resultsCidrFilterEnt.SetText("192.168.1.0/24")
	// Не должен паниковать
	a.saveResultsViewSettings()
}

// --- results_view.go extended tests ---

func TestOsGuessLine_WithConfidence(t *testing.T) {
	r := scanner.Result{
		GuessOS:           "Linux",
		GuessOSConfidence: "high",
	}
	result := osGuessLine(r)
	if result == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestOsGuessLine_WithReason(t *testing.T) {
	r := scanner.Result{
		GuessOS:           "Linux",
		GuessOSConfidence: "high",
		GuessOSReason:     "TTL=64",
	}
	result := osGuessLine(r)
	if result == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestOsGuessLine_WithOSOnly(t *testing.T) {
	r := scanner.Result{
		GuessOS: "Windows",
	}
	result := osGuessLine(r)
	if result != "Windows" {
		t.Errorf("expected 'Windows', got %q", result)
	}
}

func TestDeviceTypeWithBadge_Router(t *testing.T) {
	result := deviceTypeWithBadge("Router")
	if result != "[NET] Router" {
		t.Errorf("expected '[NET] Router', got %q", result)
	}
}

func TestDeviceTypeWithBadge_Switch(t *testing.T) {
	result := deviceTypeWithBadge("Switch")
	if result != "[NET] Switch" {
		t.Errorf("expected '[NET] Switch', got %q", result)
	}
}

func TestDeviceTypeWithBadge_AccessPoint(t *testing.T) {
	result := deviceTypeWithBadge("Access Point")
	if result != "[AP] Access Point" {
		t.Errorf("expected '[AP] Access Point', got %q", result)
	}
}

func TestDeviceTypeWithBadge_Printer(t *testing.T) {
	result := deviceTypeWithBadge("Printer")
	if result != "[PRN] Printer" {
		t.Errorf("expected '[PRN] Printer', got %q", result)
	}
}

func TestDeviceTypeWithBadge_Camera(t *testing.T) {
	result := deviceTypeWithBadge("Camera")
	if result != "[CAM] Camera" {
		t.Errorf("expected '[CAM] Camera', got %q", result)
	}
}

func TestDeviceTypeWithBadge_NAS(t *testing.T) {
	result := deviceTypeWithBadge("NAS")
	if result != "[NAS] NAS" {
		t.Errorf("expected '[NAS] NAS', got %q", result)
	}
}

func TestDeviceTypeWithBadge_IoT(t *testing.T) {
	result := deviceTypeWithBadge("IoT Device")
	if result != "[IOT] IoT Device" {
		t.Errorf("expected '[IOT] IoT Device', got %q", result)
	}
}

func TestDeviceTypeWithBadge_Desktop(t *testing.T) {
	result := deviceTypeWithBadge("Desktop")
	if result != "[PC] Desktop" {
		t.Errorf("expected '[PC] Desktop', got %q", result)
	}
}

func TestDeviceTypeWithBadge_Laptop(t *testing.T) {
	result := deviceTypeWithBadge("Laptop")
	if result != "[PC] Laptop" {
		t.Errorf("expected '[PC] Laptop', got %q", result)
	}
}

func TestDeviceTypeWithBadge_Server(t *testing.T) {
	result := deviceTypeWithBadge("Server")
	if result != "[SRV] Server" {
		t.Errorf("expected '[SRV] Server', got %q", result)
	}
}

func TestDeviceTypeWithBadge_Phone(t *testing.T) {
	result := deviceTypeWithBadge("Phone")
	if result != "[MOB] Phone" {
		t.Errorf("expected '[MOB] Phone', got %q", result)
	}
}

func TestDeviceTypeWithBadge_Tablet(t *testing.T) {
	result := deviceTypeWithBadge("Tablet")
	if result != "[MOB] Tablet" {
		t.Errorf("expected '[MOB] Tablet', got %q", result)
	}
}

func TestDeviceTypeWithBadge_UnknownExtended(t *testing.T) {
	result := deviceTypeWithBadge("Unknown Device")
	if result != "[UNK] Unknown Device" {
		t.Errorf("expected '[UNK] Unknown Device', got %q", result)
	}
}

func TestDeviceTypeWithBadge_EmptyExtended(t *testing.T) {
	result := deviceTypeWithBadge("")
	if result != "-" {
		t.Errorf("expected '-', got %q", result)
	}
}

func TestDeviceTypeWithBadge_MixedCase(t *testing.T) {
	result := deviceTypeWithBadge("ROUTER")
	if result != "[NET] ROUTER" {
		t.Errorf("expected '[NET] ROUTER', got %q", result)
	}
}

func TestBuildResultsAnalyticsView_EmptyApp(t *testing.T) {
	a := &App{}
	result := a.buildResultsAnalyticsView(nil)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestBuildResultsAnalyticsView_WithData(t *testing.T) {
	a := &App{}
	data := []scanner.Result{
		{
			IP:         "192.168.1.1",
			DeviceType: "Router",
			Protocols:  []string{"TCP", "UDP"},
		},
		{
			IP:         "192.168.1.2",
			DeviceType: "Server",
			Protocols:  []string{"TCP"},
		},
	}
	result := a.buildResultsAnalyticsView(data)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestBuildResultsAnalyticsView_CardsMode(t *testing.T) {
	a := &App{}
	a.resultsMode = "Карточки"
	data := []scanner.Result{
		{
			IP:         "192.168.1.1",
			DeviceType: "Router",
			Protocols:  []string{"TCP"},
		},
	}
	result := a.buildResultsAnalyticsView(data)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}
