# Release Notes: Network Scanner v2.2

**Дата релиза:** 2026-08-29  
**Версия:** 2.2.0  
**Статус:** ✅ Готов к релизу

---

## Сводка

| Метрика | Значение |
|---------|----------|
| Версия | 2.2.0 |
| Задач выполнено | 13/13 (100%) |
| Тесты | ✅ Все проходят |
| Race detector | ✅ Без предупреждений |
| Сборка | ✅ Успешна |

---

## Изменения v2.2

### Критические исправления (TASK-028)

- **TASK-028.1:** Исправлен `batch_test.go` — обновлена проверка SNMPResponse
- **TASK-028.2:** Снят `t.Skip()` из `TestIntegrationScanStatus_MultipleScans`
- **TASK-028.3:** Добавлена поддержка `InsecureTLS` в `devicecontrol.Execute()`
- **TASK-028.4..028.7:** Исправлены тесты плагинов, API, GUI controller

### ARP-резолвер (TASK-029)

- Реализован `ARPCache` в `DefaultNetworkProber`
- Добавлен `ResolveMAC()` с использованием ARP-кэша
- Обновлён `NewDefaultNetworkProber()` для dependency injection
- Исправлены type mismatches в `internal/scanner/`

### Динамическая загрузка плагинов (TASK-030)

- Реализована загрузка через `plugin.Open()` и `plugin.Lookup()`
- Поддержка `.so`/`.dll`/`.dylib`
- `LoadBuiltin()` загружает встроенные плагины без CGO

### GUI улучшения (TASK-031..TASK-033)

- **TASK-031:** Pinch-zoom с валидацией (0.5x — 3.0x)
- **TASK-032:** Контекстное меню через `widget.NewPopUpMenu()` (Обновить/Настройки/Экспорт)
- **TASK-033:** Прокрутка через `ScrollToTop()`/`ScrollToBottom()`

### Проверка прав (TASK-034)

- OS-specific check через build tags: `linux`, `windows`, `darwin`
- `FormatPermissionReport()` — человеко-читаемый отчёт

### Документация (TASK-035..TASK-038)

- **TASK-035:** Ссылка на `ARCHITECTURE.md` в `zaclikivaniya.md`
- **TASK-036:** Обновлены комментарии в тестах (`stub` → `mock`)
- **TASK-037:** Проверен актуальность `GUI_SMOKE_CHECKLIST.md`
- **TASK-038:** Полностью переписан `INSTALL.md` для всех платформ

---

## Technical Details

### Фикс race condition в generateScanID()

```go
// Добавлен atomic counter для уникальности ID
var scanIDCounter uint64

func generateScanID() string {
    scanIDMu.Lock()
    defer scanIDMu.Unlock()
    counter := atomic.AddUint64(&scanIDCounter, 1)
    return fmt.Sprintf("scan-%d-%d", time.Now().UnixNano(), counter)
}
```

### Фикс parallel test access

```go
// Добавлен mutex для защиты scanStoreInstance
var testMu sync.Mutex

func TestIntegrationScanStatus_MultipleScans(t *testing.T) {
    testMu.Lock()
    defer testMu.Unlock()
    // ...
}
```

---

## Тестирование

```bash
# Все тесты
go test ./... -count=1

# С race detector
go test ./... -race -count=1

# Сборка
go build ./...
```

**Результат:** ✅ Все тесты проходят, race detector не выявляет проблем

---

## Known Limitations

- Cross-OS tests (Linux/macOS) требуют прогона в целевой среде
- CI evidence требует `GITHUB_TOKEN` для успешного run

---

## Release Artifacts

- `network-scanner` — CLI binary
- `network-scanner-gui` — GUI binary
- Кроссплатформенные сборки в `build/release/`

---

**Версия документа:** 2.2.0  
**Последнее обновление:** 2026-08-29
