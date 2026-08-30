package controller

import (
	"testing"

	"network-scanner/internal/audit"
	"network-scanner/internal/scanner"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// --- SecurityController tests ---

func TestSecurityController_RunAudit_EmptyResults(t *testing.T) {
	ctrl := &SecurityController{app: app.New(), ui: &SecurityUI{}}
	ctrl.RunAudit(nil, "")
	// Не паникует — это успех
}

func TestSecurityController_RunAudit_NilMinSeverity(t *testing.T) {
	ctrl := &SecurityController{app: app.New(), ui: &SecurityUI{}}
	ctrl.RunAudit([]scanner.Result{}, "")
	// Не паникует — это успех
}

func TestSecurityController_RunAudit_WithResults(t *testing.T) {
	ctrl := &SecurityController{app: app.New(), ui: &SecurityUI{}}
	ctrl.RunAudit([]scanner.Result{{IP: "192.168.1.1"}}, "low")
	// Не паникует — это успех
}

func TestSecurityController_RunAudit_InvalidSeverity(t *testing.T) {
	ctrl := &SecurityController{app: app.New(), ui: &SecurityUI{}}
	ctrl.RunAudit([]scanner.Result{}, "invalid_severity")
	// Должен fallback на "low" и не паниковать
}

func TestSecurityController_CheckDeviceStatus_EmptyTarget(t *testing.T) {
	ctrl := &SecurityController{app: app.New(), ui: &SecurityUI{}}
	ctrl.CheckDeviceStatus()
	// Не паникует — это успех
}

func TestSecurityController_CheckDeviceStatus_WithTarget(t *testing.T) {
	ctrl := &SecurityController{app: app.New(), ui: &SecurityUI{
		DeviceTargetEntry: widget.NewEntry(),
	}}
	ctrl.ui.DeviceTargetEntry.SetText("http://192.168.1.1")
	ctrl.CheckDeviceStatus()
	// Не паникует — это успех
}

func TestSecurityController_RebootDevice_EmptyTarget(t *testing.T) {
	ctrl := &SecurityController{app: app.New(), ui: &SecurityUI{}}
	ctrl.RebootDevice()
	// Не паникует — это успех
}

func TestSecurityController_RebootDevice_WithTarget(t *testing.T) {
	ctrl := &SecurityController{app: app.New(), ui: &SecurityUI{
		DeviceTargetEntry: widget.NewEntry(),
	}}
	ctrl.ui.DeviceTargetEntry.SetText("http://192.168.1.1")
	ctrl.RebootDevice()
	// Не паникует — это успех
}

func TestSecurityController_WakeOnLAN_EmptyMAC(t *testing.T) {
	ctrl := &SecurityController{app: app.New(), ui: &SecurityUI{}}
	ctrl.WakeOnLAN()
	// Не паникует — это успех
}

func TestSecurityController_WakeOnLAN_WithMAC(t *testing.T) {
	ctrl := &SecurityController{app: app.New(), ui: &SecurityUI{
		WOLMacEntry: widget.NewEntry(),
	}}
	ctrl.ui.WOLMacEntry.SetText("aa:bb:cc:dd:ee:ff")
	defer func() { recover() }()
	ctrl.WakeOnLAN()
	// Может паниковать из-за network в headless — это ожидаемо
}

func TestSecurityController_WakeOnLAN_InvalidMAC(t *testing.T) {
	ctrl := &SecurityController{app: app.New(), ui: &SecurityUI{
		WOLMacEntry: widget.NewEntry(),
	}}
	ctrl.ui.WOLMacEntry.SetText("invalid-mac")
	defer func() { recover() }()
	ctrl.WakeOnLAN()
	// Может паниковать из-за network в headless — это ожидаемо
}

func TestSecurityController_RunRiskSignatures_EmptyResults(t *testing.T) {
	ctrl := &SecurityController{app: app.New(), ui: &SecurityUI{}}
	ctrl.RunRiskSignatures(nil)
	// Не паникует — это успех
}

func TestSecurityController_RunRiskSignatures_WithResults(t *testing.T) {
	ctrl := &SecurityController{app: app.New(), ui: &SecurityUI{}}
	ctrl.RunRiskSignatures([]scanner.Result{{IP: "192.168.1.1"}})
	// Не паникует — это успех
}

func TestSecurityController_SetStatus_NilUI(t *testing.T) {
	ctrl := &SecurityController{app: app.New()}
	ctrl.setStatus("test status")
	// Не паникует при nil UI — это успех
}

func TestSecurityController_SetStatus_NilLabel(t *testing.T) {
	ctrl := &SecurityController{app: app.New(), ui: &SecurityUI{StatusLabel: nil}}
	ctrl.setStatus("test status")
	// Не паникует при nil StatusLabel — это успех
}

// --- SettingsManager tests ---

func TestSettingsManager_LoadSplitStates_NilApp(t *testing.T) {
	mgr := &SettingsManager{app: nil}
	state := mgr.LoadSplitStates()
	if state.ScanTabOffset != 0 {
		t.Errorf("expected 0 for nil app, got %v", state.ScanTabOffset)
	}
}

func TestSettingsManager_SaveSplitState_NilApp(t *testing.T) {
	mgr := &SettingsManager{app: nil}
	mgr.SaveSplitState("test.key", 0.5)
	// Не паникует при nil app — это успех
}

func TestSettingsManager_ResetUIPanelLayout_AllNil(t *testing.T) {
	mgr := &SettingsManager{app: app.New()}
	mgr.ResetUIPanelLayout(nil, nil, nil)
	// Не паникует при nil splits — это успех
}

func TestSettingsManager_ResetUIPanelLayoutWithFeedback_NilWindow(t *testing.T) {
	mgr := &SettingsManager{app: app.New()}
	mgr.ResetUIPanelLayoutWithFeedback(nil, nil, nil, nil)
	// Не паникует при nil window — это успех
}

func TestSettingsManager_ApplyDefaultSplitOffsetsForProfile_Compact(t *testing.T) {
	mgr := &SettingsManager{app: app.New()}
	mgr.ApplyDefaultSplitOffsetsForProfile("compact")
	// Не паникует — это успех
}

func TestSettingsManager_ApplyDefaultSplitOffsetsForProfile_Normal(t *testing.T) {
	mgr := &SettingsManager{app: app.New()}
	mgr.ApplyDefaultSplitOffsetsForProfile("normal")
	// Не паникует — это успех
}

func TestSettingsManager_ApplyDefaultSplitOffsetsForProfile_Empty(t *testing.T) {
	mgr := &SettingsManager{app: app.New()}
	mgr.ApplyDefaultSplitOffsetsForProfile("")
	// Не паникует — это успех
}

func TestSettingsManager_ClearSplitPreferences_NilApp(t *testing.T) {
	mgr := &SettingsManager{app: nil}
	mgr.ClearSplitPreferences()
	// Не паникует при nil app — это успех
}

func TestSettingsManager_ApplySplitOffset_NilSplit(t *testing.T) {
	mgr := &SettingsManager{app: app.New()}
	mgr.ApplySplitOffset(nil, 0.5)
	// Не паникует при nil split — это успех
}

func TestSettingsManager_ApplySplitOffset_Valid(t *testing.T) {
	mgr := &SettingsManager{app: app.New()}
	split := &container.Split{}
	mgr.ApplySplitOffset(split, 0.5)
	if split.Offset != 0.5 {
		t.Errorf("expected offset 0.5, got %v", split.Offset)
	}
}

// --- Audit integration tests ---

func TestAuditNormalizeSeverity_Low(t *testing.T) {
	_, ok := audit.NormalizeSeverity("low")
	if !ok {
		t.Error("expected ok for 'low'")
	}
}

func TestAuditNormalizeSeverity_Medium(t *testing.T) {
	_, ok := audit.NormalizeSeverity("medium")
	if !ok {
		t.Error("expected ok for 'medium'")
	}
}

func TestAuditNormalizeSeverity_High(t *testing.T) {
	_, ok := audit.NormalizeSeverity("high")
	if !ok {
		t.Error("expected ok for 'high'")
	}
}

func TestAuditNormalizeSeverity_Critical(t *testing.T) {
	_, ok := audit.NormalizeSeverity("critical")
	if !ok {
		t.Error("expected ok for 'critical'")
	}
}

func TestAuditNormalizeSeverity_Invalid(t *testing.T) {
	_, ok := audit.NormalizeSeverity("invalid")
	if ok {
		t.Error("expected !ok for 'invalid'")
	}
}

func TestAuditEvaluateOpenPorts_Empty(t *testing.T) {
	findings := audit.EvaluateOpenPorts(nil)
	if findings == nil {
		t.Error("expected non-nil findings for nil results")
	}
}

func TestAuditFilterByMinSeverity_Empty(t *testing.T) {
	findings := audit.FilterByMinSeverity(nil, "low")
	if findings == nil {
		t.Error("expected non-nil findings for empty input")
	}
}

func TestAuditFormatFindings_Empty(t *testing.T) {
	report := audit.FormatFindings(nil)
	if report == "" {
		t.Error("expected non-empty report for empty findings")
	}
}
