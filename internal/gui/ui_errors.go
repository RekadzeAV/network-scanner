package gui

import (
	"fmt"
	"network-scanner/internal/gui/errors"
	"network-scanner/internal/logger"

	"fyne.io/fyne/v2/dialog"
)

// showError показывает диалог с ошибкой.
func (a *App) showError(title, message string, err error) {
	if a == nil || a.myWindow == nil {
		return
	}
	if err == nil {
		return
	}
	fullMessage := message
	if fullMessage == "" {
		fullMessage = err.Error()
	}
	dialog.ShowError(fmt.Errorf("%s: %v", fullMessage, err), a.myWindow)
}

// showErrorRaw показывает диалог с сырой ошибкой.
func (a *App) showErrorRaw(title string, err error) {
	if a == nil || a.myWindow == nil {
		return
	}
	if err == nil {
		return
	}
	dialog.ShowError(err, a.myWindow)
}

// showInfo показывает информационный диалог.
func (a *App) showInfo(title, message string) {
	if a == nil || a.myWindow == nil {
		return
	}
	dialog.ShowInformation(title, message, a.myWindow)
}

// showConfirm показывает диалог подтверждения.
func (a *App) showConfirm(title, message string, callback func(bool)) {
	if a == nil || a.myWindow == nil {
		return
	}
	dialog.NewConfirm(title, message, callback, a.myWindow).Show()
}

// handleGUIError обрабатывает GUIError через ErrorHandler.
func (a *App) handleGUIError(handler *errors.ErrorHandler, err error, title string) {
	if a == nil {
		return
	}
	if err == nil {
		return
	}
	msg := handler.HandleWithUI(err)
	if msg != "" {
		a.showInfo(title, msg)
	}
}

// safeExec выполняет функцию с recover и логированием ошибок.
func (a *App) safeExec(label string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			logger.LogError(fmt.Errorf("panic: %v", r), label)
		}
	}()
	fn()
}
