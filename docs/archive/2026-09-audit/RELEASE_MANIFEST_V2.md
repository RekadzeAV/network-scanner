# Release Manifest — Network Scanner v2.0

**Дата выпуска:** 2026-08-26  
**Версия:** 2.0.0  
**Статус:** ✅ Готов к релизу

---

## 1. Что реализовано

### Этап 1 — Сканер сети (100%)
- [x] P1: ICMP Ping, Traceroute, DNS, фильтры
- [x] P2: Wake-on-LAN, баннеры/версии, определение ОС
- [x] P3: CI на 3 ОС, интеграция, UX/perf

### Этап 2 — Анализатор + безопасность (100%)
- [x] P1: Whois (RDAP), Wi-Fi, аудит портов
- [x] P2: Device Control, Risk Signatures
- [x] P3: CVE/NVD, HTML отчёты, SSH/WinRM/WMI

### D-трек — Стабилизация (100% код)
- [x] D1: Topology hardening (confidence, sorting)
- [x] D2: Export hardening (JSON schema, GraphML, consistency)
- [x] D3: GUI UX hardening (pagination, strings, filters, analytics)

---

## 2. Операционные шаги до выпуска

### Приоритет 1: Кросс-ОС прогон
```bash
# Linux
./scripts/final-release-check.sh
./scripts/p1-closure-check.sh
./scripts/p2-closure-check.sh
./scripts/stage2-p1-closure-check.sh
./scripts/stage2-p2-closure-check.sh
./scripts/stage2-p3-closure-check.sh

# macOS (аналогичные команды)
```

### Приоритет 2: CI Sign-off
```powershell
# Windows
$env:GITHUB_TOKEN = "<token>"
make p3-close-all-win
```

### Приоритет 3: GUI приёмка
Пройти чеклист: `docs/GUI_SMOKE_CHECKLIST.md`

### Приоритет 4: External compatibility
Проверить импорт GraphML в yEd/Gephi: `docs/GRAPHML_COMPATIBILITY_CHECK.md`

---

## 3. Артефакты релиза

### Сборка
```bash
make build
# Артефакты: build/network-scanner, build/network-scanner-gui
```

### Пакеты (Linux)
```bash
make deb   # Debian package
```

### Документация
| Документ | Путь |
|----------|------|
| CHANGELOG | `CHANGELOG.md` |
| Release Notes | `docs/RELEASE_SUMMARY_UI_RESULTS.md` |
| PR Description | `docs/PR_DESCRIPTION_UI_RESULTS.md` |
| Acceptance Checklist | `docs/RELEASE_ACCEPTANCE_CHECKLIST.md` |
| Readiness Report | `docs/FINAL_RELEASE_READINESS_REPORT.md` |
| D-Track Status | `docs/D_TRACK_IMPLEMENTATION_STATUS.md` |
| GUI Smoke | `docs/GUI_SMOKE_CHECKLIST.md` |

---

## 4. Известные ограничения

1. **ICMP Ping:** не работает без прав администратора на Windows
2. **UDP сканирование:** ограничено таймаутами, не рекомендуется для крупных подсетей
3. **Wi-Fi анализ:** зависит от ОС, на macOS/Linux требует root
4. **SSH/WinRM:** требуют явного согласия пользователя (`--device-confirm`)
5. **Graphviz:** экспорт PNG/SVG требует установленной утилиты `dot`

---

## 5. Подписи

| Роль | ФИО | Дата | Статус |
|------|-----|------|--------|
| Tech Lead | ____________ | __________ | ⏳ |
| QA Lead | ____________ | __________ | ⏳ |
| Release Manager | ____________ | __________ | ⏳ |

---

*Этот документ является каноническим release manifest для Network Scanner v2.0.*
