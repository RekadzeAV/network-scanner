# Оценка безопасности

**Дата:** 2026-09-22  
**Модель угроз:** STRIDE  

---

## 7.1. Секреты в репо

| Источник | Статус | Комментарий |
|----------|--------|-------------|
| GITHUB_TOKEN в скриптах | ✅ Не найден в текущих файлах | `[ASSUMPTION]` |
| Параметры `--device-pass`, `--remote-exec-*` | ⚠️ CLI-флаги | Требуют маскировки в логах |

---

## 7.2. CVE-зависимости

- **govulncheck** — не запущен (нет доступа к установке)
- **Рекомендация:** Запустить `govulncheck ./...` в CI

---

## 7.3. Небезопасные паттерны

| № | Паттерн | Уровень риска | Митигирование |
|---|---------|---------------|---------------|
| 1 | TLS best-effort в remoteexec | `[RISK]` | Требует обязательного TLS-режима |
| 2 | Remote-exec без dry-run по умолчанию | `[RISK]` | Dry-run по умолчанию |
| 3 | Device-control без confirm по умолчанию | `[RISK]` | Confirm по умолчанию |
| 4 | Security report без redaction по умолчанию | `[RISK]` | Redaction по умолчанию |

---

## 7.4. Auth/authz REST API

| Компонент | Статус | Комментарий |
|-----------|--------|-------------|
| `internal/api/router.go` | ❌ без аутентификации | `[ASSUMPTION]` `[RISK]` |
| Рекомендация | P1 | Добавить аутентификацию (Bearer token / OAuth2) |

---

## 7.5. Модель угроз (STRIDE)

| Спутник | Угроза | Mitigations |
|---------|--------|-------------|
| CLI | Tampering (аргументы командной строки) | Валидация (configvalidation) |
| GUI | Information Disclosure (SNMP данные) | Redaction |
| REST API | Elevation of Privilege (без auth) | `[TODO]` Добавить auth |
| Remote-exec | Remote Code Execution | Allowlist, dry-run, confirm |
| Device-control | Tampering (настройки устройств) | Confirm, audit trail |
| Сеть | Information Disclosure (скан портов) | Маскировка в логах |
| Плагины | Elevation of Privilege | Изоляция плагинов, allowlist |

---

## 7.6. Рекомендации

| Приоритет | Задача |
|-----------|--------|
| P0 | Добавить аутентификацию к REST API |
| P0 | Dry-run по умолчанию для remote-exec |
| P0 | Confirm по умолчанию для device-control |
| P1 | Обязательный TLS для remote-exec |
| P1 | Redaction по умолчанию для security report |
| P2 | Запустить `govulncheck` в CI |
| P2 | Добавить secret scanning в CI |
| P2 | Аудит-лог для всех изменяющих операций |

---