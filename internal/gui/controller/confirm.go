package controller

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// confirmDangerousAction запрашивает подтверждение необратимого действия
// (перезагрузка устройства и т. п.) до его выполнения.
//
// P0-3: если окно недоступно (headless-режим, тесты) — действие НЕ выполняется.
// Подтверждение получить негде, поэтому безопасный отказ предпочтительнее
// молчаливого «самоподтверждения».
func confirmDangerousAction(win fyne.Window, title, message string, onConfirm func()) {
	if win == nil || onConfirm == nil {
		return
	}
	dialog.ShowConfirm(title, message, func(ok bool) {
		if ok {
			onConfirm()
		}
	}, win)
}

// rebootConfirmMessage — единый текст подтверждения перезагрузки.
func rebootConfirmMessage(target string) string {
	return "Перезагрузить устройство " + target +
		"?\n\nДействие необратимо: сессия устройства и подключённые клиенты будут прерваны."
}
