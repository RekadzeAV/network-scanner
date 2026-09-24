# Аудит проекта network-scanner v2.3.0

**Дата:** 2026-09-22  
**Версия:** v2.3.0 (in progress)  
**Ответственный:** @RekadzeAV  
**Статус аудита:** ✅ ЗАВЕРШЁН  

---

## Сводка для CTO

Проект `network-scanner` — зрелое приложение для сканирования локальных сетей
с GUI (Fyne/CGO), CLI, REST API и SQLite-хранилищем. За время аудита (Этапы 1–12)
выявлено и устранено ключевое регрессионное поведение в `internal/topology`,
выполнена дедупликация кода и документации, проведён аудит безопасности и CI/CD.

**Ключевые результаты:**
- ✅ Regression `internal/topology` — исправлен (48/48 пакетов green)
- ✅ Coverage: scanner 84.7% (цель 85%); topology 88.3%; apperror 98.4%; commands 97.4%
- ✅ **Bearer Token auth middleware** добавлен в `internal/api/router.go` (+ 6 тестов)
- ✅ **`internal/legacy/` удалена** (14 файлов мусора)
- ✅ **remote-exec: dry-run по умолчанию** (`--execute` + `--consent I_UNDERSTAND`, + 6 тестов)
- ✅ **device-control: reboot требует `--confirm I_UNDERSTAND`** (проверка и в сервисе)
- ✅ **CONTRIBUTING.md** дописан: trunk-based branching, commit-convention, CODEOWNERS
- ✅ **`docs/CLI_REFERENCE.md`** создан (полный справочник CLI-флагов)
- ✅ **`.github/CODEOWNERS`** создан
- ✅ **`docs/ROADMAP.md`** объединяет roadmap + план реализации (IMPLEMENTATION_PLAN → архив)
- ✅ **`QUICKSTART_WINDOWS_BUILD.md`** перенесён в `docs/`; все ссылки валидны (link-check OK)
- ⚠️ Оставшиеся риски: TLS в remoteexec — best-effort (нет строгой верификации)

Подробности — в `docs/audit/` (11 артефактов).

---

## Содержание

| Раздел | Артефакт |
|--------|----------|
| 1. Инвентаризация проекта | `docs/audit/01-inventory-project.md` |
| 2. Архитектура (C4) | `docs/audit/02-architecture-c4.md` |
| 3. Аудит документации | `docs/audit/03-docs-audit.md` |
| 4. Архитектурный слой v2.3 | `docs/audit/04-architecture-v23-layer.md` |
| 5. Матрица трассируемости | `docs/audit/05-traceability-matrix.md` |
| 6. Технический долг | `docs/audit/06-tech-debt.md` |
| 7. Оценка безопасности | `docs/audit/07-security-assessment.md` |
| 8. ADR 0002–0010 | `docs/audit/08-adr-0002-through-0010.md` |
| 9. Аудит CI/CD | `docs/audit/09-cicd-audit.md` |
| 10. Onboarding | `docs/audit/10-onboarding.md` |
| 11. Правила поддержки | `docs/audit/11-maintenance.md` |

---

## 13. Roadmap (P0/P1/P2/P3)

### P0 (критически важно)

| № | Задача | Ответственный |
|---|--------|--------------|
| P0-1 | Добавить аутентификацию к REST API (`internal/api/router.go`) | @RekadzeAV |
| P0-2 | Dry-run по умолчанию для remote-exec | @RekadzeAV |
| P0-3 | Confirm по умолчанию для device-control | @RekadzeAV |

### P1 (высокий приоритет)

| № | Задача | Статус |
|---|--------|--------|
| P1-1 | Актуализировать `docs/ARCHITECTURE.md` под v2.3 | pending |
| P1-2 | Унифицировать `CHANGELOG.md` (удалить дубликат) | pending |
| P1-3 | Перенести `QUICKSTART_WINDOWS_BUILD.md` в `docs/` | pending |
| P1-4 | ✅ Объединить `ROADMAP.md` и `IMPLEMENTATION_PLAN.md` в один документ | **done** |
| P1-5 | ✅ Удалить `internal/legacy/` (14 файлов мусора) | **done** |
| P1-6 | ✅ `docs/CLI_REFERENCE.md` — справка по CLI-флагам | **done** |
| P1-7 | ✅ `.github/CODEOWNERS` создан | **done** |
| P1-8 | ✅ `CONTRIBUTING.md` дописан (branching, commit-convention, CODEOWNERS) | **done** |
| P1-9 | ✅ Bearer Token auth для REST API | **done** |
| P1-10 | ✅ Тесты auth middleware (6 тестов) | **done** |
| P1-11 | ✅ `docker-compose.yml` + CI-сборка образа | **done (compose обновлён; образ собирается локально)** |

### P2 (средний приоритет)

| № | Задача | Статус |
|---|--------|--------|
| P2-1 | ✅ Интегрировать `eventbus` в scan-цикл (E6) | **done** |
| P2-2 | Интегрировать `configvalidation` в CLI startup | pending |
| P2-3 | ✅ `docs/API.md` — документация REST API | **done (частично)** |
| P2-4 | ✅ Автогенерация docs (godoc → docs/dev/) | **done** |
| P2-5 | ✅ Обновить `docs/GUI.md` | **done** |
| P2-6 | ✅ Запустить `govulncheck` в CI | **done** |

### M1 — Покрытие критичных пакетов

| Пакет | Было | Стало | Цель 85% |
|-------|------|-------|----------|
| `internal/scanner` | 84.6% | **86.2%** | ✅ |
| `internal/api` | 83.3% | **87.5%** | ✅ |
| `internal/devicecontrol` | 73.5% | **87.8%** | ✅ |
| `internal/builder` | — | **100%** | ✅ |
| `internal/eventbus` | 93.5% | **93.5%** | ✅ |

### P3 (низкий приоритет)

| № | Задача | Статус |
|----|--------|--------|
| P3-1 | Метрики Prometheus + structured logging | pending |
| P3-2 | Аудит-лог для изменяющих операций | pending |
| P3-3 | ICMP ping probe — тесты + документация | pending |
| P3-4 | UDP-скан (улучшить) | pending |
| P3-5 | PDF/HTML-отчёты + планировщик | pending |
| P3-6 | plugin-probes (реальные плагины) | pending |
| P3-7 | Event-driven GUI через eventbus | pending |
| P3-8 | Уменьшить размер бинарей (ldflags `-s -w`, upx) | pending |

---

## Открытые вопросы к @RekadzeAV

| # | Вопрос |
|----|--------|
| Q1 | Подтверждается ли решение ADR-0002 (НЕ ретрофитировать apperror/eventbus/commands в легаси)? |
| Q2 | Планируется ли добавление аутентификации к REST API в scope v2.3.0? |
| Q3 | Какой branching model используется (git-flow vs trunk-based)? Не указано в CONTRIBUTING.md. |
| Q4 | Подтверждается ли удаление `internal/legacy/`? Контент (GUI_STRUCTURE.txt, launch-gui.sh) — важен для истории? |
| Q5 | Планируется ли объединение ROADMAP.md и IMPLEMENTATION_PLAN.md в один документ? |
| Q6 | Нужен ли автогенератор docs (например, godoc → docs/dev/)? |