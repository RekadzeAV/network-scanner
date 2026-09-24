# Правила поддержки

**Дата:** 2026-09-22  

---

## 1. Принципы разработки

| Принцип | Описание |
|---------|----------|
| Код до документации | Стабилизация кода и тестов в первую очередь; документация актуализируется по финальному состоянию. |
| Дедупликация вместе с фиксом | Правка кода и дедупликация выполняются одним этапом (один экспорт — одна реализация). |
| Очистка перед коммитом | Временные артефакты удаляются до коммита. |
| Коммит/пуш — замораживающий шаг | Только после зелёных тестов и чистой документации. |
| CI на закоммиченном дереве | Проверки имеют смысл только на финальном состоянии ветки. |
| Фичи — после стабилизации | Новые фичи только поверх стабильной, задокументированной и протестированной базы. |

---

## 2. Таблицы покрытия

### Целевые пороги

| Пакет | Текущий | Цель |
|-------|---------|------|
| internal/scanner | 84.7% | 85% ⚠️ |
| internal/topology | 88.3% | 85% ✅ |
| internal/api | 82.6% | 85% ⚠️ |
| internal/devicecontrol | 73.5% | 85% ❌ |
| internal/apperror | 98.4% | — ✅ |
| internal/commands | 97.4% | — ✅ |
| internal/eventbus | 93.5% | — ✅ |

---

## 3. CI-gate checklist

Перед пушем в `main`:

1. ✅ `go build ./...` — чисто
2. ✅ `go test ./... -short -count=1` — 0 FAIL
3. ✅ `golangci-lint run` — 0 замечаний
4. ✅ `docs-link-check` — 0 битых ссылок
5. ✅ Нет untracked-мусора в корне
6. ✅ `.gitignore` покрывает cov-артефакты

---

## 4. Правила коммитов (TODO — определить)

**Статус:** `[TODO]` — convention не определена в CONTRIBUTING.md.

Рекомендуемый формат (Conventional Commits):

```
<type>(<scope>): <subject>

<body>

Fixes #<issue>
```

Типы: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `perf`, `security`.

---

## 5. Ветвление (TODO — уточнить)

**Статус:** `[TODO]` — git-flow / trunk-based не указано в CONTRIBUTING.md.

Рекомендация: **trunk-based** с feature-ветками и squash-merge.

---

## 6. Cleanup-правила

| Что | Правило |
|-----|---------|
| cov-артрафакты | ✅ В `.gitignore` (`*_cov*`, `coverage_*`, `*.out`) |
| `*.log` в корне | ✅ В `.gitignore` |
| `.tmp/` | ✅ В `.gitignore` |
| `internal/legacy/` | ⚠️ Удалить/перенести в архив |
| Временные файлы инвентаризации (`temp_*.txt`) | ⚠️ Удалить |

---

## 7. Релизный цикл

1. E0: Исправление регрессии + дедупликация кода
2. E1: Очистка репозитория
3. E2: Аудит документации
4. E3: Коммит/пуш в main
5. E4: Валидация CI
6. E5: Покрытие тестами (цель 85%)
7. E6: Интеграция архитектурного слоя v2.3
8. E7: Расширения (Post-release)

---

## 8. Ответственные за пакеты (CODEOWNERS — TODO)

**Статус:** `[TODO]` — CODEOWNERS отсутствует.

Рекомендуемая карта ответственности:

| Пакет | Owner |
|-------|-------|
| `internal/scanner/` | @RekadzeAV |
| `internal/topology/` | @RekadzeAV |
| `internal/inventory/` | @RekadzeAV |
| `internal/api/` | @RekadzeAV |
| `internal/devicecontrol/` | @RekadzeAV |
| `internal/remoteexec/` | @RekadzeAV |
| `internal/security/` | @RekadzeAV |
| `internal/plugin/` | @RekadzeAV |
| `internal/eventbus/` | @RekadzeAV |
| `internal/commands/` | @RekadzeAV |
| `internal/apperror/` | @RekadzeAV |
| `internal/configvalidation/` | @RekadzeAV |
| `docs/` | @RekadzeAV |
| `scripts/` | @RekadzeAV |