# Декомпозиция H1: GUI Coverage 17.8% → 60%

**Дата:** 2025-01-XX  
**Текущий статус:** 17.8%  
**Цель:** 60%  
**Приоритет:** 🔴 HIGH (blocker для v1.1.0)

---

## 📊 Текущее покрытие (17.8%)

Команда: `go test -cover ./internal/gui/...`

### Что уже покрыто (существующие тесты):
- `app_autoprofile_test.go` — `autoScanProfile()` (логика ограничений)
- `app_filter_preset_test.go` — пресеты фильтров
- `app_recommended_badge_test.go` — badge рекомендованного профиля
- `app_snmp_partial_test.go` — частичный SNMP
- `app_ui_reset_test.go` — сброс UI
- `model_test.go` — model пакет
- `formatter_simple_test.go` — formatter
- `operations_test.go` — OperationsManager
- `results_layout_test.go` — layout helpers
- `results_model_test.go` — results model
- `results_view_filters_test.go` — фильтры результатов
- `results_view_perf_test.go` — perf результатов
- `results_responsive_test.go` — responsive layout
- `save_results_test.go` — сохранение результатов
- `split_persist_test.go` — сохранение split-ов
- `topology_interactive_map_test.go` — интерактивная карта
- `benchmarks_simple_test.go` — benchmarks

### Что НЕ покрыто (основные файлы):
| Файл | Статус | Почему не покрыт |
|------|--------|------------------|
| `app.go` (~2500 строк) | ❌ 0% | `NewApp()`, `initUI()`, `setupEventHandlers()`, все методы сканирования/топологии |
| `scan_ui.go` (~200 строк) | ❌ 0% | `initScanUI()`, `buildScanControlsContainer()`, `buildResultsContainer()` |
| `results_view.go` (~600 строк) | ⚠️ Частично | `renderScanResultsView()`, `buildTableView()`, `buildCardsView()` не покрыты |
| `topology_controller.go` (~400 строк) | ❌ 0% | `buildTopology()`, `stopTopologyBuild()`, `saveTopology()` |
| `security_view.go` (~200 строк) | ❌ 0% | `buildSecurityDashboardView()` |
| `inventory_view.go` (~200 строк) | ❌ 0% | `buildInventoryDashboardView()` |
| `results_analytics_view.go` (~150 строк) | ❌ 0% | `buildResultsAnalyticsView()` |
| `results_charts.go` (~100 строк) | ❌ 0% | `buildPieChart()`, `buildBarChart()` |

---

## 🎯 Стратегия достижения 60%

### Подход:
1. **Не писать тесты на UI-инициализацию** — `NewApp()`, `initUI()` создают реальные Fyne виджеты, которые требуют GUI-окружения. Тестировать их бесполезно — они не дают логики.
2. **Фокус на бизнес-логику:** методы сканирования, обработки результатов, фильтры, топология, безопасность.
3. **Использовать mocks и stubs** для сервисов (scanner, topology, SNMP).
4. **Тестировать публичные методы** `App` и вспомогательные функции.

### Расчёт покрытия:
- Всего строк в GUI: ~4500
- Уже покрыто: ~800 строк (17.8%)
- Нужно покрыть: ~2000 строк (для достижения 60%)
- Оставшиеся ~1700 строк — UI-инициализация (не тестируется)

---

## 🔴 Блок 1: Тестирование App.go (цель: +15% покрытия)

### 1.1 Тесты для `autoScanProfile()` (УЖЕ ЕСТЬ ✅)
**Файл:** `app_autoprofile_test.go`  
**Строк покрыто:** ~50  
**Статус:** ✅ Завершено

**Что протестировано:**
- `autoScanProfile("192.168.1.0/24", "1-65535", 120)` → без изменений
- `autoScanProfile("10.0.0.0/21", "1-65535", 120)` → ports=1-1024, threads=40
- `autoScanProfile("10.0.0.0/20", "1-65535", 300)` → ports=1-512, threads=24

**Критерий завершения:** ✅ Пройдено

---

### 1.2 Тесты для `applyScanPreset()`
**Файл:** `app_test.go` (создать/добавить)  
**Строк кода:** ~80 строк в `applyScanPreset()`  
**Оценка:** 30 минут

**Что тестировать:**
```go
func TestApplyScanPreset_Quick(t *testing.T) {
    // 1. Создать app
    // 2. Вызвать app.applyScanPreset("quick")
    // 3. Проверить:
    //    - portRangeEntry.Text == "22,80,443,445,3389"
    //    - timeoutEntry.Text == "1"
    //    - threadsEntry.Text == "120"
    //    - scanUDPCheck.Checked == false
    //    - scanBannersCheck.Checked == false
    //    - scanOSActiveCheck.Checked == false
}

func TestApplyScanPreset_Deep(t *testing.T) {
    // 1. Создать app
    // 2. Вызвать app.applyScanPreset("deep")
    // 3. Проверить:
    //    - portRangeEntry.Text == "1-2000"
    //    - timeoutEntry.Text == "3"
    //    - threadsEntry.Text == "40"
    //    - scanUDPCheck.Checked == true
    //    - scanBannersCheck.Checked == true
    //    - scanOSActiveCheck.Checked == true
}

func TestApplyScanPreset_Balanced(t *testing.T) {
    // 1. Создать app
    // 2. Вызвать app.applyScanPreset("balanced")
    // 3. Проверить промежуточные значения
}
```

**Критерий завершения:**
- [ ] 3 теста проходят
- [ ] Покрытие `applyScanPreset()` = 100%

---

### 1.3 Тесты для `applyRecommendedScanProfile()`
**Файл:** `app_test.go` (добавить)  
**Строк кода:** ~120 строк  
**Оценка:** 45 минут

**Что тестировать:**
```go
func TestApplyRecommendedScanProfile_NoNetwork(t *testing.T) {
    // 1. Создать app с пустым networkEntry
    // 2. Вызвать app.applyRecommendedScanProfile()
    // 3. Проверить что не паникует и badge обновлён
}

func TestApplyRecommendedScanProfile_SmallSubnet(t *testing.T) {
    // 1. Создать app с networkEntry = "192.168.1.0/24"
    // 2. Вызвать app.applyRecommendedScanProfile()
    // 3. Проверить что параметры установлены в "умеренные"
}

func TestApplyRecommendedScanProfile_LargeSubnet(t *testing.T) {
    // 1. Создать app с networkEntry = "10.0.0.0/16"
    // 2. Вызвать app.applyRecommendedScanProfile()
    // 3. Проверить что параметры установлены в "консервативные"
}
```

**Критерий завершения:**
- [ ] 3 теста проходят
- [ ] Покрытие `applyRecommendedScanProfile()` = 100%

---

### 1.4 Тесты для `loadScanSettings()` и `saveScanSettings()`
**Файл:** `app_test.go` (добавить)  
**Строк кода:** ~100 строк  
**Оценка:** 30 минут

**Что тестировать:**
```go
func TestSaveAndLoadScanSettings(t *testing.T) {
    // 1. Создать app
    // 2. Изменить все поля (network, portRange, timeout, threads, UDP, banners, OS, verbose)
    // 3. Вызвать app.saveScanSettings()
    // 4. Создать новый app (app2)
    // 5. Проверить что app2 загрузил сохранённые значения
}

func TestLoadScanSettings_Defaults(t *testing.T) {
    // 1. Создать app без сохранённых настроек
    // 2. Вызвать app.loadScanSettings()
    // 3. Проверить что установлены значения по умолчанию
}
```

**Критерий завершения:**
- [ ] 2 теста проходят
- [ ] Покрытие `loadScanSettings()` и `saveScanSettings()` = 100%

---

### 1.5 Тесты для `setPortRangeControlsEnabled()`
**Файл:** `app_test.go` (добавить)  
**Строк кода:** ~20 строк  
**Оценка:** 10 минут

**Что тестировать:**
```go
func TestSetPortRangeControlsEnabled_True(t *testing.T) {
    // 1. Создать app
    // 2. Вызвать app.setPortRangeControlsEnabled(true)
    // 3. Проверить что portWellKnownBtn, portRegisteredBtn, portDynamicBtn enabled
}

func TestSetPortRangeControlsEnabled_False(t *testing.T) {
    // 1. Создать app
    // 2. Вызвать app.setPortRangeControlsEnabled(false)
    // 3. Проверить что кнопки disabled
}
```

**Критерий завершения:**
- [ ] 2 теста проходят
- [ ] Покрытие = 100%

---

### 1.6 Тесты для `refreshAutoProfileStateLabel()`
**Файл:** `app_test.go` (добавить)  
**Строк кода:** ~30 строк  
**Оценка:** 15 минут

**Что тестировать:**
```go
func TestRefreshAutoProfileStateLabel_Checked(t *testing.T) {
    // 1. Создать app
    // 2. Установить autoProfileCheck.Checked = true
    // 3. Вызвать app.refreshAutoProfileStateLabel()
    // 4. Проверить что autoProfileStateText показывает "Включён"
}

func TestRefreshAutoProfileStateLabel_Unchecked(t *testing.T) {
    // 1. Создать app
    // 2. Установить autoProfileCheck.Checked = false
    // 3. Вызвать app.refreshAutoProfileStateLabel()
    // 4. Проверить что autoProfileStateText показывает "Выключен"
}
```

**Критерий завершения:**
- [ ] 2 теста проходят
- [ ] Покрытие = 100%

---

## 🔴 Блок 2: Тестирование результатов и фильтров (цель: +10% покрытия)

### 2.1 Тесты для `renderScanResultsView()` — состояние Idle/Scanning
**Файл:** `results_view_test.go` (создать)  
**Строк кода:** ~100 строк  
**Оценка:** 1 час

**Что тестировать:**
```go
func TestRenderScanResultsView_Idle(t *testing.T) {
    // 1. Создать app
    // 2. Установить resultsState = "idle"
    // 3. Вызвать app.renderScanResultsView()
    // 4. Проверить что resultsBody содержит label "Результаты сканирования появятся здесь..."
}

func TestRenderScanResultsView_Scanning(t *testing.T) {
    // 1. Создать app
    // 2. Установить resultsState = "scanning"
    // 3. Вызвать app.renderScanResultsView()
    // 4. Проверить что resultsBody содержит label "Сканирование..."
}

func TestRenderScanResultsView_NoResults(t *testing.T) {
    // 1. Создать app
    // 2. Установить resultsState = "done", scanResults = []
    // 3. Вызвать app.renderScanResultsView()
    // 4. Проверить что resultsBody содержит label "Результаты не найдены."
}
```

**Критерий завершения:**
- [ ] 3 теста проходят
- [ ] Покрытие веток idle/scanning/no results = 100%

---

### 2.2 Тесты для `filteredSortedResults()` — кэш
**Файл:** `results_view_test.go` (создать)  
**Строк кода:** ~80 строк  
**Оценка:** 45 минут

**Что тестировать:**
```go
func TestFilteredSortedResults_CacheHit(t *testing.T) {
    // 1. Создать app с 10 результатами
    // 2. Вызвать filteredSortedResults() дважды
    // 3. Проверить что второй раз используется кэш (по таймингу или инт)
}

func TestFilteredSortedResults_CacheInvalidation(t *testing.T) {
    // 1. Создать app, заполнить кэш
    // 2. Изменить scanResultsVersion
    // 3. Вызвать filteredSortedResults()
    // 4. Проверить что кэш инвалидирован
}
```

**Критерий завершения:**
- [ ] 2 теста проходят
- [ ] Покрытие кэша = 100%

---

### 2.3 Тесты для `buildTableView()` и `buildCardsView()`
**Файл:** `results_view_test.go` (добавить)  
**Строк кода:** ~150 строк  
**Оценка:** 1.5 часа

**Что тестировать:**
```go
func TestBuildTableView_ReturnsWidget(t *testing.T) {
    // 1. Создать app с 5 результатами
    // 2. Установить resultsMode = "Таблица"
    // 3. Вызвать app.filteredSortedResults()
    // 4. Вызвать app.buildTableView(filtered)
    // 5. Проверить что возвращён widget.Table
}

func TestBuildCardsView_ReturnsList(t *testing.T) {
    // 1. Создать app с 5 результатами
    // 2. Установить resultsMode = "Карточки"
    // 3. Вызвать app.filteredSortedResults()
    // 4. Вызвать app.buildCardsView(filtered)
    // 5. Проверить что возвращён widget.List
}
```

**Критерий завершения:**
- [ ] 2 теста проходят
- [ ] Покрытие = 100%

---

### 2.4 Тесты для `buildHostDetailsDrawer()`
**Файл:** `results_view_test.go` (добавить)  
**Строк кода:** ~60 строк  
**Оценка:** 30 минут

**Что тестировать:**
```go
func TestBuildHostDetailsDrawer_NoSelection(t *testing.T) {
    // 1. Создать app с результатами
    // 2. Не выбирать хост (selectedHostIP = "")
    // 3. Вызвать app.buildHostDetailsDrawer(filtered)
    // 4. Проверить что возвращён Card с "Нет данных"
}

func TestBuildHostDetailsDrawer_WithSelection(t *testing.T) {
    // 1. Создать app с результатами
    // 2. Выбрать хост (selectedHostIP = "192.168.1.1")
    // 3. Вызвать app.buildHostDetailsDrawer(filtered)
    // 4. Проверить что возвращён Card с markdown данными хоста
}
```

**Критерий завершения:**
- [ ] 2 теста проходят
- [ ] Покрытие = 100%

---

## 🔴 Блок 3: Тестирование топологии (цель: +10% покрытия)

### 3.1 Тесты для `buildTopology()`
**Файл:** `topology_controller_test.go` (создать)  
**Строк кода:** ~200 строк  
**Оценка:** 2 часа

**Что тестировать:**
```go
func TestBuildTopology_NoResults(t *testing.T) {
    // 1. Создать app без результатов сканирования
    // 2. Вызвать app.buildTopology()
    // 3. Проверить что topologyStatus показывает ошибку
}

func TestBuildTopology_WithResults(t *testing.T) {
    // 1. Создать app с результатами сканирования
    // 2. Mock scanner для topology
    // 3. Вызвать app.buildTopology()
    // 4. Проверить что lastTopology не nil
}

func TestBuildTopology_WithSNMP(t *testing.T) {
    // 1. Создать app с результатами и SNMP community
    // 2. Mock SNMP collector
    // 3. Вызвать app.buildTopology()
    // 4. Проверить что SNMP данные добавлены в топологию
}
```

**Критерий завершения:**
- [ ] 3 теста проходят
- [ ] Покрытие `buildTopology()` = 100%

---

### 3.2 Тесты для `stopTopologyBuild()` и `saveTopology()`
**Файл:** `topology_controller_test.go` (добавить)  
**Строк кода:** ~50 строк  
**Оценка:** 20 минут

**Что тестировать:**
```go
func TestStopTopologyBuild(t *testing.T) {
    // 1. Создать app с запущенной топологией
    // 2. Вызвать app.stopTopologyBuild()
    // 3. Проверить что topologyCancel вызван
}

func TestSaveTopology(t *testing.T) {
    // 1. Создать app с lastTopology
    // 2. Вызвать app.saveTopology()
    // 3. Проверить что файл создан (или ошибка если нет lastTopology)
}
```

**Критерий завершения:**
- [ ] 2 теста проходят
- [ ] Покрытие = 100%

---

## 🔴 Блок 4: Тестирование безопасности и инвентаризации (цель: +5% покрытия)

### 4.1 Тесты для `buildSecurityDashboardView()`
**Файл:** `security_view_test.go` (создать)  
**Строк кода:** ~100 строк  
**Оценка:** 1 час

**Что тестировать:**
```go
func TestBuildSecurityDashboardView_Empty(t *testing.T) {
    // 1. Создать app с пустыми результатами
    // 2. Вызвать app.buildSecurityDashboardView([])
    // 3. Проверить что возвращён пустой dashboard
}

func TestBuildSecurityDashboardView_WithResults(t *testing.T) {
    // 1. Создать app с результатами
    // 2. Вызвать app.buildSecurityDashboardView(results)
    // 3. Проверить что возвращён dashboard с security info
}
```

**Критерий завершения:**
- [ ] 2 теста проходят
- [ ] Покрытие = 100%

---

### 4.2 Тесты для `buildInventoryDashboardView()`
**Файл:** `inventory_view_test.go` (создать)  
**Строк кода:** ~100 строк  
**Оценка:** 1 час

**Что тестировать:**
```go
func TestBuildInventoryDashboardView_NoSnapshots(t *testing.T) {
    // 1. Создать app без снапшотов
    // 2. Вызвать app.buildInventoryDashboardView()
    // 3. Проверить что возвращён dashboard с сообщением "Нет снапшотов"
}

func TestBuildInventoryDashboardView_WithSnapshots(t *testing.T) {
    // 1. Создать app с снапшотами
    // 2. Вызвать app.buildInventoryDashboardView()
    // 3. Проверить что возвращён dashboard с таблицей снапшотов
}
```

**Критерий завершения:**
- [ ] 2 теста проходят
- [ ] Покрытие = 100%

---

## 📊 Итоговый расчёт

| Блок | Задача | Ожидание | Статус |
|------|--------|----------|--------|
| 1.1 | autoScanProfile | ✅ +2% | ✅ Завершено |
| 1.2 | applyScanPreset | +3% | ⬜ Не начато |
| 1.3 | applyRecommendedScanProfile | +3% | ⬜ Не начато |
| 1.4 | load/save settings | +2% | ⬜ Не начато |
| 1.5 | setPortRangeControlsEnabled | +1% | ⬜ Не начато |
| 1.6 | refreshAutoProfileStateLabel | +1% | ⬜ Не начато |
| 2.1 | renderScanResultsView (idle/scanning) | +3% | ⬜ Не начато |
| 2.2 | filteredSortedResults (кэш) | +2% | ⬜ Не начато |
| 2.3 | buildTableView/buildCardsView | +3% | ⬜ Не начато |
| 2.4 | buildHostDetailsDrawer | +2% | ⬜ Не начато |
| 3.1 | buildTopology | +4% | ⬜ Не начато |
| 3.2 | stopTopologyBuild/saveTopology | +1% | ⬜ Не начато |
| 4.1 | buildSecurityDashboardView | +2% | ⬜ Не начато |
| 4.2 | buildInventoryDashboardView | +2% | ⬜ Не начато |
| **Итого** | | **+29%** | |

**Текущее:** 17.8%  
**После H1:** ~47.8%  
**Добавить ещё ~13%:** H1.7 (results_analytics_view, results_charts, topology_interactive_map)

---

## 📋 Порядок выполнения

### Спринт 1 (1-2 дня):
1. ✅ Блок 1.1 (autoScanProfile) — УЖЕ ЕСТЬ
2. ⬜ Блок 1.2 (applyScanPreset)
3. ⬜ Блок 1.3 (applyRecommendedScanProfile)
4. ⬜ Блок 1.4 (load/save settings)

### Спринт 2 (1-2 дня):
5. ⬜ Блок 1.5-1.6 (helpers)
6. ⬜ Блок 2.1-2.2 (renderScanResultsView, кэш)
7. ⬜ Блок 2.3-2.4 (tableView, cardsView, hostDetails)

### Спринт 3 (1 день):
8. ⬜ Блок 3.1-3.2 (topology)
9. ⬜ Блок 4.1-4.2 (security, inventory)

### Спринт 4 (0.5 дня):
10. ⬜ Добивка: results_analytics_view, results_charts
11. ⬜ Финальная проверка: `go test -cover ./internal/gui/...`

---

## ⚠️ Важные заметки

1. **Не тестировать UI-инициализацию:** `NewApp()`, `initUI()`, `buildScanTabContent()` — создают реальные Fyne виджеты, которые требуют GUI-окружения. Их тестирование сложно и не даёт покрытия бизнес-логики.

2. **Использовать mocks:** Для `scanner.Scanner`, `topology.TopologyService`, `snmpcollector.SNMPService` создавать простые mock-структуры.

3. **Не тестировать goroutines:** Методы которые запускают горутины (`startScan()`, `buildTopology()`) тестировать сложно. Тестировать только синхронные части.

4. **Фокус на чистые функции:** `autoScanProfile()`, `filteredSortedResults()`, `passesCIDRFilter()`, `activeFilterCount()` — это то, что даёт максимальное покрытие при минимальных усилиях.

---

## ✅ Критерии выхода из H1

- [ ] `go test -cover ./internal/gui/...` показывает ≥ 60%
- [ ] Все новые тесты проходят без ошибок
- [ ] Нет regressions в существующих тестах
- [ ] Покрытие критических путей (фильтры, кэш, топология) = 100%

---

**План создан:** 2025-01-XX  
**Автор:** Koda AI  
**Статус:** 📋 PLANNED
