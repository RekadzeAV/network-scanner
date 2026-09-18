package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// icons.go — единая точка тематических иконок GUI.
//
// Все иконки — встроенные векторные ресурсы Fyne (theme.*Icon()), компилируются
// в бинарь: внешние файлы, шрифты и сетевые загрузки не используются.
// Каждая обёртка возвращает уникальный смысловой алиас, чтобы замена иконки
// не требовала правок по всем местам использования.

// iconScan — запуск сканирования / WOL (пробуждение).
func iconScan() fyne.Resource { return theme.MediaPlayIcon() }

// iconStop — остановка сканирования/топологии.
func iconStop() fyne.Resource { return theme.MediaStopIcon() }

// iconSave — сохранение результатов/отчётов/пресетов.
func iconSave() fyne.Resource { return theme.DocumentSaveIcon() }

// iconRefresh — обновление, повтор (Retry), ping, перезагрузка устройства.
func iconRefresh() fyne.Resource { return theme.ViewRefreshIcon() }

// iconCopy — копирование в буфер обмена.
func iconCopy() fyne.Resource { return theme.ContentCopyIcon() }

// iconClear — очистка полей/сброс фильтров.
func iconClear() fyne.Resource { return theme.ContentClearIcon() }

// iconRestore — сброс расположения UI/карты.
func iconRestore() fyne.Resource { return theme.ViewRestoreIcon() }

// iconTopology — топология (граф узлов).
func iconTopology() fyne.Resource { return theme.GridIcon() }

// iconSearch — поиск хостов/портов, DNS.
func iconSearch() fyne.Resource { return theme.SearchIcon() }

// iconSecurity — риски/безопасность.
func iconSecurity() fyne.Resource { return theme.WarningIcon() }

// iconInfo — информационные действия (Device Status, пояснения автопрофиля).
func iconInfo() fyne.Resource { return theme.InfoIcon() }

// iconHelp — справка/подробности.
func iconHelp() fyne.Resource { return theme.HelpIcon() }

// iconAccount — Whois (регистрант домена).
func iconAccount() fyne.Resource { return theme.AccountIcon() }

// iconRoute — Traceroute (следование по маршруту).
func iconRoute() fyne.Resource { return theme.MailForwardIcon() }

// iconFast — профиль «Быстро».
func iconFast() fyne.Resource { return theme.MediaFastForwardIcon() }

// iconDeep — профиль «Глубоко».
func iconDeep() fyne.Resource { return theme.ZoomInIcon() }

// iconConfirm — подтверждение/применение (рекомендуемые настройки, пресеты).
func iconConfirm() fyne.Resource { return theme.ConfirmIcon() }

// iconAdd — «показать ещё».
func iconAdd() fyne.Resource { return theme.ContentAddIcon() }

// iconInspect — инспекция (аудит портов, открыть детали).
func iconInspect() fyne.Resource { return theme.VisibilityIcon() }

// iconCancel — отмена операции.
func iconCancel() fyne.Resource { return theme.CancelIcon() }

// iconExport — экспорт отчётов.
func iconExport() fyne.Resource { return theme.DownloadIcon() }

// iconDevice — сетевое устройство/адаптер (Wi-Fi).
func iconDevice() fyne.Resource { return theme.ComputerIcon() }

// iconInventory — инвентаризация (снапшоты, БД).
func iconInventory() fyne.Resource { return theme.StorageIcon() }

// iconSettings — инструменты/утилиты.
func iconSettings() fyne.Resource { return theme.SettingsIcon() }

// iconFullScreen — открыть во внешнем окне/полный экран.
func iconFullScreen() fyne.Resource { return theme.ViewFullScreenIcon() }

// iconDocument — документ/каталог (зарегистрированные порты IANA).
func iconDocument() fyne.Resource { return theme.DocumentIcon() }

// iconMore — прочее/частные диапазоны портов.
func iconMore() fyne.Resource { return theme.MoreVerticalIcon() }

// iconPalette — выбор темы.
func iconPalette() fyne.Resource { return theme.ColorPaletteIcon() }

// iconAccent — выбор акцентного цвета.
func iconAccent() fyne.Resource { return theme.ColorChromaticIcon() }
