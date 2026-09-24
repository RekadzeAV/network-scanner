# Инвентаризация проекта — network-scanner

**Дата:** 2026-09-22  
**Версия проекта:** v2.3.0 (in progress)  

---

## 1.1. Дерево файлов (сокращённо)

```
network-scanner/
├── cmd/network-scanner/{main.go, cmd/*.go}
├── internal/{api, apperror, benchmark, builder, commands, configvalidation,
│   contracts, devicecontrol, eventbus, inventory, plugin, redact, remoteexec,
│   scanner, security, topology, network, banner, display, presenter, gui,
│   snmpcollector, comparator, alerting, report, osdetect, ports, cache,
│   batch, profiler, wol, nettools, diff, errors, logger, mock, legacy}/
├── scripts/{p1,p2,p3}-closure-check.{ps1,sh}
├── scripts/smoke-*.{ps1,sh}, docs-link-check.{ps1,sh}, build*.{sh,bat}
├── docs/{README,ARCHITECTURE,GUI,IMPLEMENTATION_PLAN,INSTALL,PROJECT_STRUCTURE,
│   QUICKSTART-macOS,QUICKSTART_WINDOWS_BUILD,ROADMAP,TECHNICAL,USER_GUIDE,
│   BUILD_STRUCTURE,LOGGING,deployment,GRAPHML_COMPATIBILITY_CHECK,
│   GUI_SMOKE_CHECKLIST,swagger.yaml,PROJECT_PROMPT}.md
├── docs/adr/0001-project-governance.md
├── docs/archive/{2026-01,2026-04,2026-08,2026-09-audit,2026-09-15-docs-sync}/
├── .github/workflows/{ci,go,release}.yml
├── config/, inventory/
├── Makefile, Dockerfile, go.mod, go.sum
├── CHANGELOG.md, CONTRIBUTING.md, README.md, LICENSE
└── QUICKSTART_WINDOWS_BUILD.md
```

### Мусор / потенциальные артефакты

| Путь | Описание | Статус |
|------|----------|--------|
| `internal/legacy/` (14 файлов) | cov-артефакты, GUI_STRUCTURE.txt, launch-gui.sh, LOG-*.txt, README.md | ⚠️ Удалить / перенести в архив |
| `temp_all_files.txt`, `temp_tree_*.txt` | Временные файлы инвентаризации | ⚠️ Удалить / `.gitignore` |
| `AUDIT_REPORT.md` | Итоговый отчёт аудита | ✅ Создан |

---

## 1.2. Стек и зависимости

| Компонент | Технология |
|-----------|-----------|
| Язык | Go 1.24+ (CI: Go 1.25) |
| GUI | `fyne.io/fyne/v2` (CGO) |
| GraphQL | `github.com/graph-gophers/graphql-go` |
| ORM / SQLite | `gorm.io/gorm` + `gorm.io/driver/sqlite` |
| Статистика | `github.com/montanaflynn/stats` |
| SNMP | `github.com/miekg/snmp` |
| TTY-детект | `github.com/mattn/go-isatty` |
| CLI | `spf13/cobra` (indirect) |
| Линтер | golangci-lint v1.64.8 |
| Pre-commit | golangci-lint, gosec, gofumpt |
| Релиз | GoReleaser |

---

## 1.3. Таблица документов

| Документ | Статус | Комментарий |
|----------|--------|-------------|
| README.md | ✅ | v2.3.0 |
| ARCHITECTURE.md | ⚠️ | Не v2.3 |
| TECHNICAL.md | ⚠️ | Частично |
| ROADMAP.md | ⚠️ | Дублирует IMPLEMENTATION_PLAN.md |
| IMPLEMENTATION_PLAN.md | ⚠️ | Дублирует ROADMAP.md |
| FINAL_PLAN_2026-09-15.md | ✅ | Финальный план |
| PROJECT_STRUCTURE.md | ✅ | — |
| CHANGELOG.md (root + docs) | ⚠️ | Дубликат |
| USER_GUIDE.md | ✅ | — |
| GUI.md | ⚠️ | Требует обновления |
| INSTALL.md | ✅ | — |
| QUICKSTART-macOS.md | ✅ | — |
| QUICKSTART_WINDOWS_BUILD.md | ⚠️ | Дублирует INSTALL |
| BUILD_STRUCTURE.md | ✅ | — |
| LOGGING.md | ⚠️ | Устаревшая модель |
| deployment.md | ⚠️ | Общие рекомендации |
| swagger.yaml | ⚠️ | Требует regenerate |
| docs/adr/0001 | ✅ | Governance |

---

## 1.4. Модульная карта internal/

Подтверждено на диске (35+): `api`, `apperror`, `benchmark`, `builder`, `commands`,
`configvalidation`, `contracts`, `devicecontrol`, `eventbus`, `inventory`,
`plugin`, `redact`, `remoteexec`, `scanner`, `security`, `topology`, `network`,
`banner`, `display`, `presenter`, `gui`, `snmpcollector`, `comparator`,
`alerting`, `report`, `osdetect`, `ports`, `cache`, `batch`, `profiler`,
`wol`, `nettools`, `diff`, `errors`, `logger`, `mock`, `legacy`.

---

## 1.5. Модель данных (SQLite)

```
devices:           [id, ip, mac, hostname, vendor, os, ports_json, last_seen, status]
scan_history:      [id, timestamp, command, args, output_ref]
security_findings: [id, device_id, severity, type, description, remediated]
```