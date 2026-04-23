package gui

import "testing"

func TestResetUIPanelLayoutNilReceiver(t *testing.T) {
	var a *App
	a.resetUIPanelLayout()
}

func TestResetUIPanelLayoutWithoutApp(t *testing.T) {
	a := &App{}
	a.resetUIPanelLayout()
}
