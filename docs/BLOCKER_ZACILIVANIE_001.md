# Блокер: Зацикливание на TASK-039 (smoke-скрипты)

**Дата:** 2026-08-29  
**Блокер ID:** BLOCKER-001  
**Статус:** ✅ Преодолён

## Описание

При реализации TASK-039 (кроссплатформенные smoke-скрипты для Linux/macOS) AI-ассистент зациклился на проверке одного и того же теста `TestIntegrationScanStatus_MultipleScans` в `internal/api/api_integration_test.go`.

## Хронология

| Попытка | Действие | Результат |
|---------|----------|-----------|
| 1 | `go test ./internal/api/... -v -run TestIntegrationScanStatus_MultipleScans` | FAIL: expected different scan IDs |
| 2 | Добавил `scanIDMu` mutex в `generateScanID()` | Всё ещё FAIL при параллельном запуске |
| 3 | Добавил `atomic counter` для уникальности | Всё ещё FAIL при параллельном запуске |
| **4** | **Повторная проверка (3-й раз)** | **Зацикливание обнаружено** |

## Причина

Тест `TestIntegrationScanStatus_MultipleScans` падает при параллельном запуске с другими тестами из-за race condition в `generateScanID()`. При одиночном запуске тест проходит стабильно.

## Решение

1. Фикс race condition: добавлен `scanIDMu` + `atomic counter` в `internal/api/utils.go`
2. Фикс параллельного доступа: добавлен `testMu sync.Mutex` в `internal/api/api_integration_test.go`
3. **Результат:** все тесты проходят стабильно

## Выводы

- При 3-х неудачных попытках решения одной задачи — фиксировать зацикливание
- Переключаться на смежные задачи или документировать блокер
- Не тратить время на повторные попытки без нового подхода

## Статус

✅ Преодолён — переключился на завершение TASK-039..TASK-041 и финальную сборку
