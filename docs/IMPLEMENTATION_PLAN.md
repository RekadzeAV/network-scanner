﻿﻿﻿# 📋 Полный план реализации Network Scanner v2.0

**Дата обновления:** 2026-08-18  
**Текущая версия:** v2.0  
**Статус проекта:** 🟢 Базовый функционал завершен (100%)  
**Фокус развития:** Тесты бизнес-логики, интеграционные тесты, CI/CD

---

## 📊 Текущее состояние (2026-08-18)

| Метрика | Значение | Статус |
|---------|----------|--------|
| Базовый функционал | ✅ 100% (13/13 задач) | ✅ |
| Coverage GUI | 51.6% | ✅ Стабильно |
| Coverage Controller | 44.4% | ✅ Стабильно |
| Coverage Errors | 90.7% | ✅ |
| Coverage Core (85%+) | 23 пакета (65.7%) | ✅ |
| Пройдено тестов | 90+ файлов | ✅ |
| Интеграционных тестов | 56 (v16) | ✅ |

---

## ✅ Завершённые этапы (100%)

### Этап 1: Критические исправления
- [x] 1.1 Добавить `main.Version` и переменные сборки
- [x] 1.2 Mock-сервисы для тестирования
- [x] 1.3 Унифицировать error handling

### Этап 2: Функциональные улучшения
- [x] 2.1 REST API (POST /scan, GET /inventory)
- [x] 2.2 Экспорт отчётов (PDF/HTML)
- [x] 2.3 История сканирований (сравнение снапшотов)
- [x] 2.4 Alerting (система уведомлений)

### Этап 3: Продвинутые функции
- [x] 3.1 SNMP сбор данных
- [x] 3.2 Topology discovery (построение карты сети)
- [x] 3.3 GUI интерфейс (Fyne framework, 30+ файлов)

---

## 🟢 Текущий фокус: Тесты бизнес-логики (2026-08-18)

### GUI coverage стабилизирована на 51.6%
**Решение:** Дальнейшее увеличение GUI coverage экономически нецелесообразно (требует mocking Fyne, real dialogs). Переход к интеграционным тестам.

### Интеграционные тесты GUI (v16)
| Файл | Кол-во тестов | Покрытие |
|------|---------------|----------|
| `results_model_integration_test.go` | 12 | Pipeline: sort → filter → analytics |
| `operations_integration_test.go` | 8 | Lifecycle: run → cancel → retry → subscribers |
| `formatter_integration_test.go` | 18 | Formatting, ports, markdown escape |
| `results_view_integration_test.go` | 18 | Caching, filters, active filter count |

**Итого:** 56 новых интеграционных тестов, все стабильны (`go test -count=3`)

---

## 📋 Детализированный бэклог

См. также:
- [ROADMAP.md](ROADMAP.md) — канонический роадмап
- [DETAILED_BACKLOG_P3_STAGE2.md](DETAILED_BACKLOG_P3_STAGE2.md) — детальный бэклог задач
- [COVERAGE_STATUS.md](COVERAGE_STATUS.md) — текущее покрытие тестами
- [ ] Все ошибки логируются
- [ ] Пользователь видит понятные сообщения об ошибках

---

## 🟡 MEDIUM: Важно для релиза

### M1: Увеличить core coverage до 85%+
**Текущее:** ~75% | **Цель:** 85% | **Оценка:** 1-2 дня

**Что нужно сделать:**  
Написать тесты для критических пакетов.

**Подзадачи:**

| # | Задача | Пакет | Оценка | Статус |
|---|--------|-------|--------|--------|
| M1.1 | Тесты для `scanner/scanner.go` | UDP сканирование, проверка живости, banner grabbing | 2 часа | ⬜ |
| M1.2 | Тесты для `network/` | Парсинг CIDR, ARP сканирование, ICMP ping | 2 часа | ⬜ |
| M1.3 | Тесты для `osdetect/` | TCP fingerprinting, banner analysis | 1.5 часа | ⬜ |
| M1.4 | Тесты для `topology/` | Построение графа, SNMP сбор данных | 1.5 часа | ⬜ |

**Критерии завершения M1:**
- [ ] `go test -cover ./internal/...` показывает ≥ 85%
- [ ] Покрытие критических путей (scanner, network) > 90%

---

### M2: Добавить CI/CD pipeline
**Текущее:** Нет автоматизации | **Цель:** GitHub Actions | **Оценка:** 1 день

**Что нужно сделать:**  
Настроить автоматическую сборку, тесты, линтинг.

**Подзадачи:**

| # | Задача | Описание | Оценка | Статус |
|---|--------|----------|--------|--------|
| M2.1 | Создать `.github/workflows/go.yml` | Build на Windows/Linux/macOS, tests, coverage, linting | 2 часа | ⬜ |
| M2.2 | Настроить golangci-lint | Добавить config `.golangci.yml`, настроить правила | 1 час | ⬜ |
| M2.3 | Добавить badge в README | Build status, coverage, version | 30 мин | ⬜ |
| M2.4 | Настроить release automation | Create tag → trigger build → upload artifacts | 2 часа | ⬜ |

**Критерии завершения M2:**
- [ ] CI проходит на всех платформах
- [ ] Coverage не падает ниже 85%
- [ ] Linting не показывает ошибок

---

### M3: Улучшить документацию
**Текущее:** Базовая | **Цель:** Полная | **Оценка:** 1-2 дня

**Что нужно сделать:**  
Добавить godoc, Swagger, примеры, обновить README.

**Подзадачи:**

| # | Задача | Описание | Оценка | Статус |
|---|--------|----------|--------|--------|
| M3.1 | Добавить godoc | Документация для `internal/scanner/`, `internal/api/`, `internal/topology/` | 2 часа | ⬜ |
| M3.2 | Создать API documentation | Swagger/OpenAPI spec, сгенерировать HTML docs | 2 часа | ⬜ |
| M3.3 | Добавить примеры использования | Examples в пакетах, тесты-примеры (Example functions) | 1 час | ⬜ |
| M3.4 | Обновить README.md | Badges, "Advanced Usage", troubleshooting guide | 1 час | ⬜ |

**Критерии завершения M3:**
- [ ] `go doc ./...` показывает документацию для всех публичных API
- [ ] README содержит badges и примеры
- [ ] Есть examples в key packages

---

### M4: Добавить performance benchmarks
**Текущее:** 10 benchmarks | **Цель:** Для всех критических путей | **Оценка:** 1 день

**Что нужно сделать:**  
Добавить бенчмарки и проверку регрессии в CI.

**Подзадачи:**

| # | Задача | Описание | Оценка | Статус |
|---|--------|----------|--------|--------|
| M4.1 | Benchmarks для scanner | scanHost, scanTCPPort, scanUDPPort, isHostAlive | 1.5 часа | ⬜ |
| M4.2 | Benchmarks для network | ParseNetworkRange, ParsePortRange, ARP scan | 1 час | ⬜ |
| M4.3 | Benchmarks для topology | BuildTopology, SNMP collect | 1 час | ⬜ |
| M4.4 | Benchmark comparison в CI | Compare с baseline, fail если regression > 10% | 1.5 часа | ⬜ |

**Критерии завершения M4:**
- [ ] Benchmarks для всех критических путей
- [ ] CI проверяет performance regression
- [ ] Baseline сохранен в репозитории

---

## 🟢 LOW: По желанию

### L1: Добавить UI theme customization
**Текущее:** Default theme | **Цель:** Пользовательские темы | **Оценка:** 1-2 дня

**Подзадачи:**

| # | Задача | Описание | Оценка | Статус |
|---|--------|----------|--------|--------|
| L1.1 | Выбор темы | Light/Dark/System | 1 час | ⬜ |
| L1.2 | Кастомные цвета | Пользовательские цвета для элементов | 1 час | ⬜ |
| L1.3 | Сохранение настроек темы | Persist theme in settings | 30 мин | ⬜ |
| L1.4 | Preview темы | Preview темы в настройках | 1 час | ⬜ |

---

### L2: Добавить plugin system
**Текущее:** Монолитная архитектура | **Цель:** Расширяемость через плагины | **Оценка:** 3-5 дней

**Подзадачи:**

| # | Задача | Описание | Оценка | Статус |
|---|--------|----------|--------|--------|
| L2.1 | Определить interface для плагинов | Plugin interface с lifecycle | 2 часа | ⬜ |
| L2.2 | Создать plugin loader | Загрузка плагинов из директории | 2 часа | ⬜ |
| L2.3 | Добавить примеры плагинов | 2-3 примера (filter, export, alert) | 3 часа | ⬜ |
| L2.4 | Документировать создание плагинов | Guide + examples | 1 час | ⬜ |

---

### L3: Добавить mobile support
**Текущее:** Desktop only | **Цель:** iOS/Android через Fyne | **Оценка:** 2-3 дня

**Подзадачи:**

| # | Задача | Описание | Оценка | Статус |
|---|--------|----------|--------|--------|
| L3.1 | Cross-compilation для mobile | Настройка для iOS/Android | 1 час | ⬜ |
| L3.2 | Адаптировать UI | Responsive design для мобильных | 2 часа | ⬜ |
| L3.3 | Добавить touch gestures | Swipe, pinch, long press | 2 часа | ⬜ |
| L3.4 | Тестирование на устройствах | Реальные тесты на iOS/Android | 1 день | ⬜ |

---

### L4: Добавить telemetry (опционально)
**Текущее:** Нет телеметрии | **Цель:** Аналитика использования | **Оценка:** 1-2 дня

**Подзадачи:**

| # | Задача | Описание | Оценка | Статус |
|---|--------|----------|--------|--------|
| L4.1 | Добавить опциональный telemetry | Anonymous usage statistics | 1 час | ⬜ |
| L4.2 | Privacy policy | Opt-out механизм | 30 мин | ⬜ |
| L4.3 | Error reporting | Отправка ошибок с consent | 1 час | ⬜ |
| L4.4 | Documentation | Описание telemetry в README | 30 мин | ⬜ |

---

## 📅 План релизов

### v1.1.0-beta (Q3 2026)
**Цель:** Стабильный релиз с улучшенным качеством

**Входящие задачи:**
- H1.1-H1.14 (GUI coverage 60%+)
- H2.1-H2.4 (Рефакторинг App)
- H3.1-H3.4 (Error handling)
- M1.1-M1.4 (Core coverage 85%+)
- M2.1-M2.4 (CI/CD)

**Критерии выхода:**
- [ ] Coverage GUI ≥ 60%
- [ ] Coverage core ≥ 85%
- [ ] Все CI checks проходят
- [ ] Нет critical bugs
- [ ] Документация обновлена

---

### v1.2.0 (Q4 2026)
**Цель:** Расширенные возможности

**Входящие задачи:**
- M3.1-M3.4 (Документация)
- M4.1-M4.4 (Benchmarks)
- L1.1-L1.4 (Theme customization)
- L2.1-L2.4 (Plugin system)

---

### v2.0.0 (Q1 2027)
**Цель:** Значительное обновление

**Входящие задачи:**
- L3.1-L3.4 (Mobile support)
- L4.1-L4.4 (Telemetry)
- Новые фичи по результатам обратной связи

---

## 📋 Метрики качества

### Code Coverage

| Пакет | Текущий | Цель v1.1 | Цель v1.2 |
|-------|---------|-----------|-----------|
| `internal/scanner` | ~80% | 90% | 95% |
| `internal/network` | ~70% | 85% | 90% |
| `internal/gui` | ~17.8% | 60% | 75% |
| `internal/api` | ~60% | 80% | 90% |
| **Общий** | ~75% | **85%** | **90%** |

### Performance

| Метрика | Текущий | Цель v1.1 |
|---------|---------|-----------|
| Scan 254 hosts | ~30s | < 25s |
| GUI startup | ~2s | < 1.5s |
| Memory usage | ~150MB | < 120MB |

### Code Quality

| Метрика | Текущий | Цель v1.1 |
|---------|---------|-----------|
| golangci-lint warnings | 0 | 0 |
| Duplicate code | < 5% | < 3% |
| Cyclomatic complexity | < 15 | < 10 |

---

## 🔄 Процесс разработки

### 1. Planning (Еженедельно)
- Review backlog
- Prioritize tasks
- Estimate effort
- Assign to sprints

### 2. Development (2-week sprints)
- Develop features
- Write tests
- Update documentation
- Run benchmarks

### 3. Review
- Code review
- Test coverage check
- Performance check
- Documentation check

### 4. Release
- Update version
- Update CHANGELOG
- Build artifacts
- Publish release

---

## 📝 Notes

### Зависимости между задачами

```
H1 (GUI tests) → M1 (Core tests) → M2 (CI/CD)
                      ↓
                M3 (Documentation)
                      ↓
                M4 (Benchmarks)
                      ↓
              H2 (Refactor App)
```

### Риски

| Риск | Вероятность | Влияние | Митигация |
|------|-------------|---------|-----------|
| Сложность рефакторинга App | Medium | High | Поэтапный рефакторинг |
| Падение coverage | Medium | Medium | Автоматические проверки в CI |
| Performance regression | Low | High | Benchmarks в CI |
| Breaking changes | Low | High | Semantic versioning |

### Зависимости от внешних факторов
- Fyne framework updates
- Go version updates
- OS updates (Windows DPI handling)

---

## ✅ Checklist перед релизом v1.1.0

- [ ] Все задачи HIGH выполнены
- [ ] Coverage GUI ≥ 60%
- [ ] Coverage core ≥ 85%
- [ ] CI/CD настроен
- [ ] Все тесты проходят
- [ ] Benchmarks показывают улучшение/стабильность
- [ ] Документация обновлена
- [ ] CHANGELOG.md обновлен
- [ ] Release notes написаны
- [ ] Smoke tests пройдены
- [ ] Performance tested
- [ ] Security audit (если нужно)
- [ ] User acceptance testing

---

**План создан:** 2025-01-XX  
**Следующий пересмотр:** 2026-01-XX  
**Автор:** Koda AI
