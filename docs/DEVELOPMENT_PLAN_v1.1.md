# План доработок Network Scanner v1.1+

**Дата создания:** 2026-07-01  
**Версия:** 1.1.0-draft  
**Статус:** 📋 PLANNED

---

## 📊 Текущее состояние проекта

| Метрика | Значение |
|---------|----------|
| **Общий прогресс** | 100% (базовый функционал) |
| **Coverage core** | ~75% |
| **Coverage GUI** | ~25% (цель: 60%+) |
| **Пакетов кода** | 33 |
| **Пройдено тестов** | 33/33 ✅ |
| **Размер CLI** | 60.6 MB |
| **Размер GUI** | 58.5 MB |

---

## 🎯 Цель версии 1.1.0

**Стабильный релиз с улучшенным качеством кода и тестами**

---

## 🔴 Приоритет HIGH (Blocker для релиза)

### H1. Увеличить GUI coverage до 60%+

**Текущее состояние:** 25%  
**Цель:** 60%  
**Оценка:** 2-3 дня

**Задачи:**
- [ ] H1.1 Написать тесты для `app.go` (NewApp, initUI, setupEventHandlers)
  - Mock Fyne app/window для изолированных тестов
  - Тесты инициализации UI компонентов
  - Тесты сохранения/загрузки настроек
  
- [ ] H1.2 Написать тесты для `results_view.go`
  - Тесты фильтрации результатов
  - Тесты сортировки
  - Тесты отображения в разных режимах (Table/Card)
  
- [ ] H1.3 Написать тесты для `topology_controller.go`
  - Тесты построения топологии
  - Тесты SNMP сбора данных
  - Тесты экспорта (DOT/GraphML/PNG)
  
- [ ] H1.4 Написать тесты для `security_view.go`
  - Тесты аудита безопасности
  - Тесты risk signatures
  
- [ ] H1.5 Написать тесты для `inventory_view.go`
  - Тесты сравнения снапшотов
  - Тесты экспорта инвентаря

**Acceptance Criteria:**
- `go test -cover ./internal/gui/...` показывает 60%+
- Все новые тесты проходят без ошибок
- Нет regressions в существующих тестах

---

### H2. Оптимизация структуры GUI App

**Текущее состояние:** 100+ полей в структуре App  
**Цель:** Разделить на подструктуры  
**Оценка:** 1-2 дня

**Задачи:**
- [ ] H2.1 Создать подструктуры:
  ```go
  type App struct {
      Scan     *ScanController
      Results  *ResultsController
      Topology *TopologyController
      Tools    *ToolsController
      Settings *SettingsManager
      // ...
  }
  ```
  
- [ ] H2.2 Рефакторинг `app.go` — перенести методы в соответствующие контроллеры
  
- [ ] H2.3 Обновить зависимости между компонентами
  
- [ ] H2.4 Добавить interface для каждого контроллера

**Acceptance Criteria:**
- Структура App содержит не более 10 полей
- Каждый контроллер имеет свой пакет
- Нет breaking changes в публичном API

---

### H3. Улучшение обработки ошибок в GUI

**Текущее состояние:** Некоторые ошибки не обрабатываются  
**Цель:** Полная обработка ошибок  
**Оценка:** 0.5 дня

**Задачи:**
- [ ] H3.1 Добавить error handling для всех API вызовов
- [ ] H3.2 Добавить user-friendly сообщения об ошибках
- [ ] H3.3 Добавить retry logic для сетевых операций
- [ ] H3.4 Добавить logging ошибок

**Acceptance Criteria:**
- Нет panic в GUI
- Все ошибки логируются
- Пользователь видит понятные сообщения

---

## 🟡 Приоритет MEDIUM (Важно для релиза)

### M1. Увеличить coverage core до 85%+

**Текущее состояние:** ~75%  
**Цель:** 85%  
**Оценка:** 1-2 дня

**Задачи:**
- [ ] M1.1 Написать тесты для `scanner/scanner.go`
  - Тесты UDP сканирования
  - Тесты проверки живости хоста
  - Тесты banner grabbing
  
- [ ] M1.2 Написать тесты для `network/`
  - Тесты парсинга CIDR
  - Тесты ARP сканирования
  - Тесты ICMP ping
  
- [ ] M1.3 Написать тесты для `osdetect/`
  - Тесты TCP fingerprinting
  - Тесты banner analysis
  
- [ ] M1.4 Написать тесты для `topology/`
  - Тесты построения графа
  - Тесты SNMP сбора данных

**Acceptance Criteria:**
- `go test -cover ./internal/...` показывает 85%+
- Покрытие критических путей (scanner, network) > 90%

---

### M2. Добавить CI/CD pipeline

**Текущее состояние:** Нет автоматизации  
**Цель:** GitHub Actions workflow  
**Оценка:** 1 день

**Задачи:**
- [ ] M2.1 Создать `.github/workflows/go.yml`
  - Build на всех платформах (Windows, Linux, macOS)
  - Run tests
  - Calculate coverage
  - Linting (golangci-lint)
  
- [ ] M2.2 Настроить golangci-lint
  - Добавить config `.golangci.yml`
  - Настроить правила linting
  
- [ ] M2.3 Добавить code coverage badge в README
  
- [ ] M2.4 Настроить release automation
  - Create tag → trigger build
  - Upload artifacts to GitHub Releases

**Acceptance Criteria:**
- CI проходит на всех платформах
- Coverage не падает ниже порога
- Linting не показывает ошибок

---

### M3. Улучшить документацию

**Текущее состояние:** Базовая документация  
**Цель:** Полная документация  
**Оценка:** 1-2 дня

**Задачи:**
- [ ] M3.1 Добавить godoc для всех публичных функций
  - `internal/scanner/` — документация
  - `internal/api/` — документация API
  - `internal/topology/` — документация
  
- [ ] M3.2 Создать API documentation
  - Swagger/OpenAPI spec для REST API
  - Сгенерировать HTML docs
  
- [ ] M3.3 Добавить примеры использования
  - Examples в пакетах
  - Тесты-примеры (Example functions)
  
- [ ] M3.4 Обновить README.md
  - Добавить badges (build, coverage, version)
  - Добавить секцию "Advanced Usage"
  - Добавить troubleshooting guide

**Acceptance Criteria:**
- `go doc ./...` показывает документацию для всех публичных API
- README содержит badges и примеры
- Есть examples в key packages

---

### M4. Добавить performance benchmarks

**Текущее состояние:** 10 benchmarks в GUI  
**Цель:** Benchmarks для всех критических путей  
**Оценка:** 1 день

**Задачи:**
- [ ] M4.1 Добавить benchmarks для scanner
  - Benchmark scanHost
  - Benchmark scanTCPPort
  - Benchmark scanUDPPort
  - Benchmark isHostAlive
  
- [ ] M4.2 Добавить benchmarks для network
  - Benchmark ParseNetworkRange
  - Benchmark ParsePortRange
  - Benchmark ARP scan
  
- [ ] M4.3 Добавить benchmarks для topology
  - Benchmark BuildTopology
  - Benchmark SNMP collect
  
- [ ] M4.4 Настроить benchmark comparison в CI
  - Compare с baseline
  - Fail если regression > 10%

**Acceptance Criteria:**
- Benchmarks для всех критических путей
- CI проверяет performance regression
- Baseline сохранен в репозитории

---

## 🟢 Приоритет LOW (По желанию)

### L1. Добавить UI theme customization

**Текущее состояние:** Default theme  
**Цель:** Пользовательские темы  
**Оценка:** 1-2 дня

**Задачи:**
- [ ] L1.1 Добавить выбор темы (Light/Dark/System)
- [ ] L1.2 Добавить кастомные цвета
- [ ] L1.3 Сохранение настроек темы
- [ ] L1.4 Preview темы в настройках

---

### L2. Добавить plugin system

**Текущее состояние:** Монолитная архитектура  
**Цель:** Расширяемость через плагины  
**Оценка:** 3-5 дней

**Задачи:**
- [ ] L2.1 Определить interface для плагинов
- [ ] L2.2 Создать plugin loader
- [ ] L2.3 Добавить примеры плагинов
- [ ] L2.4 Документировать создание плагинов

---

### L3. Добавить mobile support

**Текущее состояние:** Desktop only  
**Цель:** iOS/Android через Fyne  
**Оценка:** 2-3 дня

**Задачи:**
- [ ] L3.1 Настроить cross-compilation для mobile
- [ ] L3.2 Адаптировать UI для мобильных устройств
- [ ] L3.3 Добавить touch gestures
- [ ] L3.4 Тестирование на реальных устройствах

---

### L4. Добавить telemetry (опционально)

**Текущее состояние:** Нет телеметрии  
**Цель:** Аналитика использования  
**Оценка:** 1-2 дня

**Задачи:**
- [ ] L4.1 Добавить опциональный telemetry
- [ ] L4.2 Anonymous usage statistics
- [ ] L4.3 Error reporting (с consent)
- [ ] L4.4 Privacy policy и opt-out

---

## 📅 План релизов

### v1.1.0-beta (Q3 2026)

**Цель:** Стабильный релиз с улучшенным качеством

**Входящие задачи:**
- ✅ H1.1-H1.5 (GUI coverage 60%+)
- ✅ H2.1-H2.4 (Рефакторинг App)
- ✅ H3.1-H3.4 (Error handling)
- ✅ M1.1-M1.4 (Core coverage 85%+)
- ✅ M2.1-M2.4 (CI/CD)

**Критерии выхода:**
- Coverage GUI ≥ 60%
- Coverage core ≥ 85%
- Все CI checks проходят
- Нет critical bugs
- Документация обновлена

---

### v1.2.0 (Q4 2026)

**Цель:** Расширенные возможности

**Входящие задачи:**
- ✅ M3.1-M3.4 (Документация)
- ✅ M4.1-M4.4 (Benchmarks)
- ✅ L1.1-L1.4 (Theme customization)
- ✅ L2.1-L2.4 (Plugin system)

---

### v2.0.0 (Q1 2027)

**Цель:** Значительное обновление

**Входящие задачи:**
- ✅ L3.1-L3.4 (Mobile support)
- ✅ L4.1-L4.4 (Telemetry)
- Новые фичи по результатам обратной связи

---

## 📋 Метрики качества

### Code Coverage

| Пакет | Текущий | Цель v1.1 | Цель v1.2 |
|-------|---------|-----------|-----------|
| `internal/scanner` | ~80% | 90% | 95% |
| `internal/network` | ~70% | 85% | 90% |
| `internal/gui` | ~25% | 60% | 75% |
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

**План создан:** 2026-07-01  
**Следующий пересмотр:** 2026-07-08  
**Автор:** Koda AI
