# Статус внедрения UX-аудита GUI

Документ — **реестр статуса** правок по мотивам [`GUI_DESIGN_RECOMMENDATIONS.md`](GUI_DESIGN_RECOMMENDATIONS.md)
(аналитический аудит, ID дефектов вида `C-01`, `R-02`, `S-02`). Здесь фиксируется только то,
что уже проверено в коде и тестах; план следующих шагов не дублирует рекомендации.

| Параметр | Значение |
|---|---|
| Ветвь | `feature/gui-design` |
| Стек | Go 1.25, `fyne.io/fyne/v2 v2.7.1` |
| Охват правок | `internal/gui/` (темы, контраст, HiDPI, сортировка результатов, прогресс, live-результаты, навигация, состояния, карта) |
| Проверка | `go build ./...`, `go vet ./...`, `go test ./...` — 0 FAIL (headless-тесты GUI помечаются SKIP без дисплея) |
| Источник ID | `GUI_DESIGN_RECOMMENDATIONS.md` §3 |

## 1. Сводка

| Категория | ID |
|---|---|
| Закрыто | C-01…C-11, C-13, C-15 · D-01…D-06 · T-01…T-06 · R-01…R-12 · S-01…S-04, S-06 · N-01…N-06, N-08, N-09 · H-01…H-03, H-06, H-09 · P-02…P-05 · P2-4…P2-8 |
| Частично | C-12, N-07, H-05, P-01, S-05, P2-2 (нет tooltip со шорткатами) |
| Открыто (P1) | H-07 (зум/панорама/mini-map карты), H-08 (иерархическая раскладка карты) |
| Открыто (P2) | C-14 = P2-3 (иконка приложения), P2-1 (карточка-секция как компонент), P2-9 (прогресс `N/M`+ETA в статус-баре), P2-10 (фокус-индикатор), P2-11 (настройки в отдельном диалоге), P2-12 (настраиваемые колонки таблицы), P2-13 (анимации) |
| Нереализуемо в Fyne v2.7.1 | H-04, S-05 (связь поле↔ошибка), P2-14 — `AriaLabel`/`Tooltip` API во фреймворке отсутствует |

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
| R-01 | Заголовок таблицы — встроенная фиксированная строка Fyne (`ShowHeaderRow=true`, `CreateHeader`/`UpdateHeader`), не скроллится вместе с данными | `results_table_view.go:141-172` | — |
| R-02 | Селектор сортировки присваивает `a.resultsSort` **и** сохраняет настройку; список опций, whitelist восстановления и ветки сравнения — один источник (`resultsSortOptions` / `normalizeResultsSortMode`) | `results_model.go:12-52`, `results_ui_sections.go:82-87`, `ui_helpers.go:274-280` | `results_sort_restore_test.go` (сортировка по всем режимам, копии, whitelist, колбэк), `gui_smoke_test.go` (`TestResultsSortSettingsRoundtrip`) |
| R-03 | Заголовок — кнопка: клик по колонке сортирует, индикатор ▲/▼ показывает текущий режим и направление | `results_table_view.go:143-172` | `results_table_sort_test.go` |
| R-04 | Live-результаты во время скана: `publishLiveResults` публикует снапшот по порогу времени, `renderScanResultsView` показывает таблицу при наличии данных (заглушка только пока хостов нет) | `scan_live_results.go:21-41`, `results_view_render.go:46-65` | `results_state_test.go` (`TestRenderScanResultsView_ScanningWithResults_RendersTable`) |
| R-05 | Чипы портов — семантический компонент: палитра состояний (Open/Closed/Filtered/Unknown), маркеры формы (●/◌/—), рамка `border-strong`, `GridWrap` переносит чипы | `theme_tokens.go:89-140`, `results_view_render.go:370-443` | — |
| R-06 | Чипы/строки больше не обходят тему: палитра варианта темы вместо локальных литералов | `results_charts.go:32-45` | `TestTheme_Color_*PaletteMatchesTokens` |
| R-09 | Счётчик строк в шапке области результатов: «Строк: показано X из Y» | `results_view.go:26-43` | `results_table_sort_test.go` |
| R-11 | `truncateStr` rune-safe (делегует `truncateString`) — кириллические баннеры не обрезаются посимвольно | `results_view_render.go:445-448` | — |
| S-01 | Пустые/ошибочные состояния — структурированный empty-state: иконка + заголовок + пояснение + кнопка действия | `results_empty_state.go`, `results_view_render.go:31-116` | `results_state_test.go` |
| S-02 | Прогресс-бар получает долю `0…1` (`SetValue(ev.Percent)`, `Max = 1`), подпись — `scanProgressText` со счётчиком и ETA | `scan_observer.go:72-74`, `scan_live_results.go:82-102` | — |
| N-02 | `Ctrl+4`/`Ctrl+5` открывают подрежимы «Безопасность»/«Инвентарь» вкладки «Сканирование» (вкладок под них нет — это подрежимы) | `app_shortcuts.go:93-96,165-181` | `shortcuts_test.go` |
| N-08 | Developer-лексика локализована: «Host Details Drawer» → «Детали хоста» (P2-7) | `results_view_render.go:161,170` | — |
| T-04 | Технические колонки IP/MAC — моноширинный шрифт (`Monospace=true`) | `results_table_view.go:103-107` | — |
| T-06 / D-04 | `minWindowWidth` применён: ширина окна не ниже 1024 px | `app.go:400-408` | `scan_ui_test.go` (`TestMinWindowWidth`) |
| D-05 | `Ctrl+Shift+I`/`Ctrl+Shift+R` зарегистрированы и показаны в диалоге F1 | `app_shortcuts.go:48-49,201-213` | — |
| P-03 | Кэш ячеек таблицы ключуется `{Col, IP}` без `Row` — устойчив к пересортировке | `results_table_view.go:20-30` | `results_table_sort_test.go` |
| P-05 | Кэш pie-картинок инвалидируется при смене темы/акцента (`invalidateThemeCaches`) | `theme_selector.go:100-121` | — |
| H-01 | Карта топологии: кликабельные узлы (`TappableNode` — кружок + подпись), hit-радиус ≥12 px | `topology_tappable_node.go`, `topology_interactive_map.go:144-165` | `topology_tappable_node_test.go` |
| H-02 | Связи карты: hit-зона через `TappableEdge` (невидимый `Rectangle` поверх `Line`) | `topology_tappable_node.go`, `topology_interactive_map.go:105-140` | `topology_tappable_node_test.go` |
| P2-7 | Русификация: `Devices/Security/Inventory` → `Устройства/Безопасность/Инвентарь`; старые значения принимаются для миграции preferences | `results_ui_sections.go:37-58` | `results_view_test.go` |
| P2-8 | Детали хоста сгруппированы: «Идентификация / Сеть / Порты» + кнопка копирования поля | `results_view_render.go:173-230` | — |

## 3. Частично

| ID | Состояние | Что осталось |
|---|---|---|
| C-12 | Цвет валидации берётся из палитры, размер сообщения — `theme.TextSize()` (T-03) | Отдельный контрастный тон для критичных ошибок |
| N-07 | Статус-бар закреплён внизу окна (`init_ui.go`), виден на всех вкладках: слева — статус/тост, справа — профиль макета | Центральный блок: прогресс `N/M` и ETA внутри статус-бара (P2-9) |
| H-05 | Hit-зоны карты увеличены: `hitRadius` у узла, `hitWidth` у связи, курсор-рука | Общий `Padding` темы 12 px (comfortable) остаётся ниже порога 24 px для тач-устройств |
| P-01 | Опрос размера окна переведён на тикер 500 мс (было 300 мс) с адаптивным debounce рендера | Событийная модель ресайза вместо поллинга |
| S-05 | Сообщение валидации в `theme.TextSize()`, привязано к полю | Связь «поле ↔ ошибка» недоступна средствами Fyne v2.7.1 (`AriaLabel` нет в API) |

## 4. Не реализуемо в Fyne v2.7.1 (с причинами)

Пункты аудита, которые **нельзя** реализовать средствами фреймворка `fyne.io/fyne/v2 v2.7.1`.
Каждая причина проверена по исходникам модуля (`$(go env GOMODCACHE)/fyne.io/fyne/v2@v2.7.1`)
и по `go doc`, а не на слух. В таблице указано, что сделано вместо оригинального требования.

| ID | Требование аудита | Почему невозможно в v2.7.1 | Что сделано вместо |
|---|---|---|---|
| **H-04** | `AriaLabel` у интерактивных элементов и индикаторов | В Fyne нет API доступности: ни в `fyne.CanvasObject`, ни в `widget.BaseWidget`, ни в `desktop.*` нет полей/методов `Aria*` (`Select-String "Aria" widget/*.go` → 0 совпадений). Fyne не строит семантическое дерево для скринридеров — экранные адаптеры видят только OpenGL-поверхность окна | Семантика передана **визуально**: текстовые подписи у всех кнопок, маркеры формы (●/◌/— у чипов портов, ▲/▼ у сортировки, ⬡/▣ у типов), курсор-рука над узлами карты |
| **P2-14** | `AriaLabel` для интерактивных элементов (дубль H-04 из §3.10) | То же: API отсутствует во фреймворке целиком, патч невозможен без форка Fyne | То же, что для H-04 |
| **S-05** (часть) | Программная связь «поле ↔ сообщение об ошибке» (`aria-describedby`) | Требует того же API доступности. additionally: `canvas.Text` в v2.7.1 **не имеет** полей `Wrapping`/`Truncation` (проверено в `canvas/text.go`) — сообщение нельзя ни перенести, ни обрезать средствами виджета | Сообщение печатается **внутри контейнера поля** (`container.NewVBox(entry, msg)` — физическая близость вместо `describedby`), размер — `theme.TextSize()` вместо Caption 11 (T-03), текст короткий по построению `configvalidation` |
| **P2-2** (часть) | `tooltip` со шорткатом на кнопках тулбара | `Tooltip` API в v2.7.1 отсутствует: файла `widget/tooltip.go` нет, `Select-String "Tooltip" widget/*.go` → 0 совпадений (появился только в более новых версиях Fyne) | Шорткаты показаны в **диалоге F1** и в подписях пунктов меню («Сброс панелей (Ctrl+Shift+L)»); сам тулбар приведён к группам Scan / Output / View без висячего разделителя |
| **P-01** (часть) | Событийная модель ресайза вместо опроса размера окна | `fyne.Window` не предоставляет callback ресайза: в интерфейсе есть только `SetOnClosed`, `SetCloseIntercept`, `SetOnDropped` — `SetOnResized` отсутствует (`go doc fyne.io/fyne/v2.Window`). `fyne.Canvas` тоже не имеет resize-хука. Единственный источник актуального размера — опрос | Поллинг оставлен, но удешевлён: тикер **500 мс** (было 300), профиль макета пересчитывается только при **фактическом** изменении размера, перерисовка идёт через debounce `scheduleResultsRender` (S-06 — мигание убрано) |
| **P2-13** (часть) | Анимации hover/selection/появления строк у **встроенных** виджетов | `canvas.NewPositionAnimation/NewSizeAnimation/NewColorRGBAAnimation` анимируют **свои** объекты; внутрь отрисовки `widget.Button`/`widget.Table`/`widget.Card` точки расширения нет — `CreateRenderer()` позволяет заменить рендер целиком, но копия рендерера виджета недокументирована и ломается на любом патче Fyne. Плюс нет API для чтения системной настройки «уменьшить движение» | Анимации не внедрены сознательно: риск регрессий выше ценности. «Уменьшить движение» не читается, поэтому единое поведение (без движения) сейчас корректно для всех пользователей |

**Итог по разделу:** 4 пункта (H-04, P2-14, S-05-привязка, P2-2-tooltip) упираются в **отсутствие API
доступности и тултипов во Fyne v2.7.1** — это ограничение фреймворка, а не приложения. 2 пункта
(P-01-события, P2-13-анимации встроенных виджетов) — ограничение расширяемости: реализовать можно
только форком или копированием внутреннего кода Fyne, что для этого проекта нецелесообразно.

## 5. Открытые дефекты приоритета P0

Все P0-дефекты закрыты.

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
- **Состояние результатов** (`a.resultsState`) хранит одно из: `idle`, `scanning`,
  `done`, `stopped`, `timeout`. Подпись состояния (`a.resultsStateLabel`) — локализована
  (`resultsStateCaption`), технические значения констант в UI не показываются.
  Контроллер сообщает переход через `ScanActions.SetResultsState(state, found)`.

## 6. Как проверить

```powershell
go build ./...
go vet ./...
go test ./internal/gui/ -count=1 -run "Theme|Accent|ResultsSort|NormalizeResultsSortMode|SortedResults|ResultsState|SetResultsState|RenderScanResultsView"
go test ./... -count=1
```

GUI-тесты, требующие драйвер Fyne, в headless-сессии помечаются `SKIP`
(`skipHeadless`); на машине с дисплеем они выполняются полностью.
