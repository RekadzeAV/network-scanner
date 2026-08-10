# Rollback Plan — D-трек

**Дата создания:** 2026-08-09  
**Контекст:** D1–D4 (Topology hardening, Export hardening, GUI hardening, Feature flags)

---

## 🎯 Цель

Обеспечить безопасное развертывание изменений D-трека с возможностью быстрого отката при обнаружении regressions или нестабильности.

---

## 📋 Feature Flags

Все потенциально нестабильные функции управляются через `internal/features/flags.go`:

| Flag | Описание | Default | Rollback |
|------|----------|---------|----------|
| `d1.topology.hardening` | Улучшенная дедупликация LLDP/FDB | `true` | `false` |
| `d1.topology.fallback` | Текстовый fallback без Graphviz | `true` | `false` |
| `d2.export.schema_validation` | JSON schema validation | `true` | `false` |
| `d2.export.graphml_equivalence` | JSON↔GraphML equivalence check | `false` | — |
| `d3.gui.responsive` | Responsive UI для разных экранов | `true` | `false` |
| `d3.gui.perf_budget` | Perf-budget проверка | `false` | — |
| `d4.rollback.enabled` | Включение rollback-механизма | `false` | — |

---

## 🔄 Процесс Rollback

### Level 1: Immediate (Feature Flag Toggle)
**Время:** < 1 секунды  
**Влияние:** Только новые функции отключаются, старый код работает

```go
// Пример отключения через API
features.DefaultManager.SetEnabled("d1.topology.hardening", false)
features.DefaultManager.SetEnabled("d3.gui.responsive", false)
```

**Когда использовать:**
- Обнаружена regression в дедупликации связей
- GUI лагает на больших результатах
- Любая критическая ошибка в новых функциях

### Level 2: Configuration Rollback
**Время:** < 1 минуты  
**Влияние:** Отключение через config file

```yaml
# config/features.yaml
features:
  d1_topology_hardening: false
  d1_topology_fallback: true
  d3_gui_responsive: false
```

**Когда использовать:**
- Проблема затрагивает только часть пользователей
- Нужно протестировать fix перед релизом

### Level 3: Code Rollback
**Время:** 5–15 минут  
**Влияние:** Полный откат до предыдущей стабильной версии

```bash
# Откат до последнего стабильного коммита
git revert HEAD~N..HEAD

# Или переключение на тег
git checkout v1.0.4
```

**Когда использовать:**
- Критическая ошибка в core logic
- CI падает на новых тестах
- Невозможно откатить через feature flags

### Level 4: Full System Rollback
**Время:** 30+ минут  
**Влияние:** Полный откат системы

**Когда использовать:**
- Невосстановимая потеря данных
- Критическая уязвимость безопасности
- Невозможность работы приложения

---

## 🧪 Smoke-профили для проверки

### D1 Smoke Profile
```bash
go test ./internal/topology/... -run "Dedup|Explain|Replace|SourceType"
go test ./internal/topology/... -run "Text|SaveAsText"
```

### D2 Smoke Profile
```bash
go test ./internal/topology/... -run "Schema|Equivalence|JSON|GraphML"
go test ./internal/topology/... -run "Compat|Golden|Roundtrip"
```

### D3 Smoke Profile
```bash
go test ./internal/gui/... -run "Sorted|Filter|Port|Analytics|Normalize|Perf"
```

### D4 Smoke Profile
```bash
go test ./internal/features/... -run "DefaultFlags|StatusReport|Concurrency"
```

---

## 📊 Мониторинг и Alerting

### Metrics для отслеживания
| Metric | Threshold | Alert |
|--------|-----------|-------|
| `topology_links_count` | ±20% от baseline | ⚠️ Warning |
| `topology_dedup_ratio` | < 0.8 | 🔴 Critical |
| `gui_render_time_ms` | > 100ms | ⚠️ Warning |
| `gui_render_time_ms` | > 500ms | 🔴 Critical |
| `feature_flag_disabled_count` | > 0 | ℹ️ Info |

### Logs для отслеживания
```
level=error component=topology msg="dedup failed" error="..."
level=warn component=gui msg="render slow" duration="150ms"
level=info component=features msg="flag toggled" flag="d1.topology.hardening" state=false
```

---

## ✅ Checklist перед развертыванием

- [ ] Все D-track smoke-тесты прошли в CI
- [ ] Feature flags настроены с правильными default значениями
- [ ] Rollback plan протестирован в staging
- [ ] Monitoring metrics настроены
- [ ] Alerts configured для критических метрик
- [ ] Rollback runbook обновлен
- [ ] Команда уведомлена о развертывании

---

## 📞 Contacts

| Роль | Ответственный | Contact |
|------|---------------|---------|
| Tech Lead | @maintainer | Telegram/Email |
| On-call Engineer | Rotating | PagerDuty |
| Product Owner | @product-owner | Telegram |

---

## 📝 Post-mortem Template

При срабатывании rollback:

1. **When:** Время обнаружения проблемы
2. **What:** Описание проблемы
3. **Impact:** Затронутые пользователи/функции
4. **Root Cause:** Причина проблемы
5. **Rollback Action:** Что было сделано
6. **Fix:** План исправления
7. **Prevention:** Как избежать в будущем

---

**Последнее обновление:** 2026-08-09  
**Следующая проверка:** 2026-09-01
