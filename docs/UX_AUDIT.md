# Статус внедрения UX-аудита GUI

Документ — **реестр статуса** правок по мотивам [`GUI_DESIGN_RECOMMENDATIONS.md`](GUI_DESIGN_RECOMMENDATIONS.md)
(аналитический аудит, ID дефектов вида `C-01`, `R-02`, `S-02`). Здесь фиксируется только то,
что уже проверено в коде и тестах; план следующих шагов не дублирует рекомендации.

| Параметр | Значение |
|---|---|
| Ветвь | `feature/v2.3.0-plan-continuation` |
| Стек | Go 1.25, `fyne.io/fyne/v2 v2.7.1` |
| Охват правок | `internal/gui/` (темы, контраст, HiDPI, сортировка результатов, прогресс, live-результаты) |
| Проверка | `go build ./...`, `go vet ./...`, `go test ./...` — 0 FAIL (headless-тесты GUI помечаются SKIP без дисплея) |
| Источник ID | `GUI_DESIGN_RECOMMENDATIONS.md` §3 |

## 1. Сводка

| Категория | ID |
|---|---|
| Закрыто | C-01, C-02, C-03, C-04, C-05, C-06, C-07, C-08, C-09, C-10, C-11, D-01, T-01, R-02, R-06, S-02, P-05 |
| Частично (токены готовы, применение в рендере впереди) | R-05, C-12 |
| Открыто (P0) | N-02 |
| Открыто (P1/P2) | T-02…T-06, N-01, N-03…N-09, R-01, R-03, R-07…R-12, S-01, S-03…S-06, H-01…H-09, P-01…P-04, D-02…D-06 |

Число «сырых» цветовых литералов вне токенов в `internal/gui/` (маска `#RRGGBB`, без
`theme_tokens.go`): **1** — комментарий в `theme.go` (было ≈50, C-07).

## 2. Закрытые дефекты

### Темы, цвет, контраст

| ID | Что сделано | Где | Тесты |
|---|---|---|---|
| C-01 | Пункты меню «Тема» разбираются `themeModeFromValue` (русские подписи → `ThemeMode`), выбор применяет тему и обновляет виджеты через `applyThemeToUI()` | `theme_selector.go:47,73,87`, `ui_helpers.go:105,132-134` | `theme_selector_test.go` (`TestThemeMode_Constants`, `TestApp_applyTheme_*`) |
| C-02 | «Системная» тема — `NewSystemTheme()` с `followSystem`, а не `SetTheme(nil)`; вариант берётся из `Color(name, variant)` | `theme.go:61-64`, `theme_selector.go:59` | `TestTheme_ExplicitDarkIgnoresFyneVariant`, `TestTheme_NewModernThemeForVariant` |
| C-03 | Палитра выбирается через `darkFor(variant)`: в явном режиме — по `isDark`, в `followSystem` — по `Settings().ThemeVariant()` (фактическая настройка ОС), тема следует за системой без пересоздания | `theme.go:88-102` | `TestTheme_ExplicitDarkIgnoresFyneVariant`, `TestTheme_NewModernThemeForVariant` |
| C-04 / D-01 | Мёртвый слой `ThemeManager`/`ThemeSwitcherUI` удалён (`theme_manager.go` + тесты) | удалён файл | — |
| C-05 | Foreground — не чистые `#000`/`#FFF`: `tokLightText #1D1D1F` (16.83:1), `tokDarkText #E9E9EC` (14.17:1) | `theme_tokens.go` | `TestTheme_Contrast_Light`, `TestTheme_Contrast_Dark` |
| C-06 | Появилась иерархия поверхностей: `surface`/`surfaceAlt` (тёмн. `#232328`/`#2A2A30`) вместо плоского `#202020` | `theme_tokens.go`, `theme.go` | `TestTheme_Color_DarkPaletteMatchesTokens` |
| C-07 | Цвета рендера переведены на палитру варианта темы; `chipBackground()`/`rowBackground()` читают токены | `results_charts.go:32-45`, `theme_tokens.go` | `TestTheme_Color_*PaletteMatchesTokens` |
| C-08 / C-09 | Статусные цвета и disabled-текст пересчитаны под WCAG AA (disabled 4.95:1 light / 4.88:1 dark вместо 1.55:1) | `theme_tokens.go` | `TestTheme_Contrast_Light`, `TestTheme_Contrast_Dark`, `TestTheme_Contrast_HelpTextOnBothVariants` |
| C-10 | Accent-пресеты задают `Primary/Hover/Pressed/Selection/OnPrimary`; `OnPrimary` подбирается по контрасту (`autoOnColor`) | `accent_colors.go`, `theme.go` | `TestTheme_Accent_ButtonAndOnPrimary`, `TestTheme_Accent_OnPrimaryIsContrasty`, `TestTheme_AutoOnColor` |
| C-11 | Меню «Акцент» обходит срез `accentPresetOrder` — порядок пунктов детерминирован | `ui_helpers.go:145-148` | `TestTheme_Accent_OrderCoversAllPresets` |

### Масштаб, результаты, прогресс

| ID | Что сделано | Где | Тесты |
|---|---|---|---|
| T-01 | Убран `os.Setenv("FYNE_SCALE", "1")`: авто-HiDPI и пользовательский `FYNE_SCALE` работают | `app.go:380-383` (комментарий-объяснение) | — |
| R-02 | Селектор сортировки присваивает `a.resultsSort` **и** сохраняет настройку; список опций, whitelist восстановления и ветки сравнения — один источник (`resultsSortOptions` / `normalizeResultsSortMode`) | `results_model.go:12-52`, `results_ui_sections.go:82-87`, `ui_helpers.go:274-280` | `results_sort_restore_test.go` (сортировка по всем режимам, копии, whitelist, колбэк), `gui_smoke_test.go` (`TestResultsSortSettingsRoundtrip`) |
| R-06 | Чипы/строки больше не обходят тему: палитра варианта темы вместо локальных литералов | `results_charts.go:32-45` | `TestTheme_Color_*PaletteMatchesTokens` |
| S-02 | Прогресс-бар получает долю `0…1` (`SetValue(ev.Percent)`, `Max = 1`), подпись — `scanProgressText` со счётчиком и ETA | `scan_observer.go:74`, `scan_live_results.go:80-132` | `internal/gui` (scan-обработчики) |
| R-04 | Live-результаты во время скана: `publishLiveResults` публикует снапшот по порогу времени вместо «мёртвого» экрана | `scan_live_results.go:21-78` | — |
| P-05 | Кэш pie-картинок инвалидируется при смене темы/акцента (`invalidateThemeCaches`) | `theme_selector.go:100-121` | — |

## 3. Частично

| ID | Состояние | Что осталось |
|---|---|---|
| R-05 | `portStatePalette` (Open/Closed/Filtered/Unknown, текст + подложка, оба варианта) объявлена в `theme_tokens.go:146-170` | Чипы в `results_view_render.go:292` всё ещё красятся одним `chipBackground()`, состояние порта визуально не различается |
| C-12 | Валидация может брать цвет из палитры | Размер/семантика сообщения (Caption 11 для критичных ошибок, T-03) не пересмотрены |

## 4. Открытые дефекты приоритета P0

| ID | Дефект | Где |
|---|---|---|
| N-02 | `Ctrl+4`/`Ctrl+5` заявлены как вкладки «Безопасность»/«Инвентарь», вкладок три; `selectTabShortcut` корректно выходит за границы, но сочетание молча no-op | `app_shortcuts.go:43-45`, `selectTabShortcut` |

Остальные открытые ID (P1/P2) перенесены без изменений из §3 рекомендаций — правки этой
серии на них не распространялись.

## 5. Совместимость сохранённых значений

- **Режим сортировки** (`scan.results_sort_mode`) хранит отображаемое имя пункта:
  `IP`, `Hostname`, `Тип`, `Открытые порты`. Унаследованное написание `HostName`
  приводится к `Hostname`; неизвестное значение игнорируется (текущий режим
  сохраняется), а если не распознано и текущее — откат на `IP`.
- **Значение темы** по-прежнему `light`/`dark`/`system`; меню отправляет туда
  разбор из `themeModeFromValue`, поэтому подпись пункта и запись в настройки совпадают.
- **Пресет акцента** — имя из `accentPresets`; неизвестное имя нормализуется к
  `defaultAccentPreset` (`macOS Blue`).

## 6. Как проверить

```powershell
go build ./...
go vet ./...
go test ./internal/gui/ -count=1 -run "Theme|Accent|ResultsSort|NormalizeResultsSortMode|SortedResults"
go test ./... -count=1
```

GUI-тесты, требующие драйвер Fyne, в headless-сессии помечаются `SKIP`
(`skipHeadless`); на машине с дисплеем они выполняются полностью.
