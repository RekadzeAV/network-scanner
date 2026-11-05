package controller

import (
	"testing"
)

// --- ScanController: setResultsState ---

func TestScanController_setResultsState_WithActions(t *testing.T) {
	mock := &mockScanActions{}
	c := &ScanController{actions: mock}

	states := []string{"scanning", "stopped", "timeout", "done", "idle"}
	for _, s := range states {
		c.setResultsState(s)
		if mock.lastState != s {
			t.Errorf("SetResultsState called with %q, want %q", mock.lastState, s)
		}
	}
}

func TestScanController_setResultsState_NilUI(t *testing.T) {
	c := &ScanController{ui: nil, actions: nil}
	c.setResultsState("scanning") // no panic
}

func TestScanController_setResultsState_NilActionsNilUI(t *testing.T) {
	c := &ScanController{ui: nil, actions: nil}
	c.setResultsState("stopped") // no panic
}

// --- ScanController: CopyScanDiagnostics nil checks ---

func TestScanController_CopyDiagnostics_NoWindow(t *testing.T) {
	c := &ScanController{ui: &ScanUI{Window: nil}, actions: nil}
	c.CopyScanDiagnostics("some diagnostics text") // should return early
}

func TestScanController_CopyDiagnostics_EmptyText(t *testing.T) {
	c := &ScanController{ui: &ScanUI{Window: nil}, actions: nil}
	c.CopyScanDiagnostics("") // should return early
}

// --- ScanController: SaveScanDiagnostics nil checks ---

func TestScanController_SaveDiagnostics_NoWindow(t *testing.T) {
	c := &ScanController{ui: &ScanUI{Window: nil}, actions: nil}
	c.SaveScanDiagnostics("some diagnostics text") // should return early
}

func TestScanController_SaveDiagnostics_EmptyText(t *testing.T) {
	c := &ScanController{ui: &ScanUI{Window: nil}, actions: nil}
	c.SaveScanDiagnostics("") // should return early
}
