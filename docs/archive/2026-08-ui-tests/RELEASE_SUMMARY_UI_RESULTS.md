# Release Summary: UI Results View

Краткий итог доработок UI блока результатов сканирования для релиза.

## Что добавлено

- Два полноценных режима отображения результатов:
  - `Таблица` (HostName, IP, MAC, Порты)
  - `Карточки` (карточки устройств + аналитика)
- Подрежимы вкладки `Сканирование`:
  - `Devices` — основной режим устройств (таблица/карточки + детали)
  - `Security` — security dashboard на текущем отфильтрованном скоупе
- Порты в UI отображаются в виде чипов с переносом строк.
- Для таблицы доступна горизонтальная прокрутка на узких окнах.
- Для карточек реализована адаптивная сетка (на узком экране 1 колонка).
- Аналитика:
  - в табличном режиме — markdown summary по протоколам/типам устройств,
  - в карточном режиме — 2 круговые диаграммы с легендой и процентами.
- `Host Details Drawer`:
  - отображение выбранного хоста (`Host/IP/MAC/Type/Vendor/OS/SNMP/Open ports`);
  - быстрые действия: `Ping`, `Traceroute`, `DNS`, `Whois`, `Wake-on-LAN`.
- `Operations Center` во вкладке `Инструменты`:
  - история последних операций;
  - выбор операции из списка;
  - действия `Retry`/`Cancel` в зависимости от статуса операции.
- `Security Dashboard`:
  - агрегированная сводка findings (`audit + risk signatures`);
  - таблица findings (`source/severity/host/title`);
  - экспорт HTML отчета через кнопку `Export security report (HTML)`.
- Во вкладке `Сканирование` добавлен управляемый `Автопрофиль`:
  - авто-коррекция `ports/threads` для крупных подсетей;
  - переключатель включения/выключения;
  - визуальные индикаторы состояния (`ВКЛ/ВЫКЛ`) в панели параметров и в верхней зоне результатов;
  - кнопка с пояснением логики автокоррекции.

## Управление отображением

- Переключение режима: `Таблица` / `Карточки`.
- Переключение подрежима: `Devices` / `Security`.
- Сортировка: `IP` / `HostName`.
- Лимит отображаемых чипов портов: `12` / `24` / `48`.
- Фильтрация:
  - строковый фильтр (HostName/IP/MAC/тип),
  - быстрые фильтры по типам (`Network Device`, `Computer`, `Server`, `Unknown`),
  - флаг `Только с открытыми портами`,
  - фильтр по `CIDR`,
  - фильтр по состоянию портов (`all/open/closed/filtered`).
- UX-полиш:
  - индикатор количества активных фильтров,
  - кнопка `Очистить` для строки поиска,
  - кнопка `Сбросить фильтры` для полного сброса.

## Персистентность

Сохраняются и восстанавливаются между запусками:
- режим отображения,
- подрежим результатов (`Devices`/`Security`),
- сортировка,
- лимит чипов,
- строка фильтра,
- быстрые фильтры по типам,
- флаг `Только с открытыми портами`,
- фильтр `CIDR`,
- фильтр состояния портов,
- параметры инструментов (включая `Audit min severity`),
- рекомендованный профиль и класс бейджа (`small/medium/large/very-large`).

## Экспорт результатов

- `Сохранить результаты` теперь экспортирует **текущее представление**:
  - с учетом активных фильтров,
  - с учетом выбранной сортировки.
- Если после фильтрации список пустой, экспорт не выполняется.
- Для security-представления добавлен отдельный экспорт HTML-отчета (`Security Dashboard`).

## Технические улучшения

- Рефакторинг UI-логики результатов:
  - `internal/gui/results_view.go`
  - `internal/gui/results_analytics_view.go`
  - `internal/gui/security_view.go`
  - `internal/gui/scan_controller.go`
  - `internal/gui/topology_controller.go`
  - `internal/gui/operations.go`
  - `internal/gui/results_charts.go`
  - `internal/gui/results_model.go`
- Добавлен кэш ресурсов диаграмм для ускорения повторной отрисовки.
- Фильтрация и сортировка консолидированы на общем пайплайне (`results_model` + UI-специфичные условия).
- Добавлены unit-тесты operations runtime (`internal/gui/operations_test.go`).
- Добавлены/обновлены тесты для модели сортировки/фильтрации/нормализации.
- Стабилизирован вывод `Ping/Traceroute` в GUI-инструментах:
  - raw output переведен на совместимый markdown-формат;
  - устранено некорректное отображение текста в `RichText`;
  - добавлен RU-парсинг статистики Windows `ping`.
- Повышена предсказуемость производительности:
  - ускорена проверка доступности хостов;
  - добавлен адаптивный лимит параллельных порт-проверок.

## Рекомендуемые сценарии приёмки

См. [GUI_SMOKE_CHECKLIST.md](GUI_SMOKE_CHECKLIST.md) (включая проверку сохранения текущего представления).

## D-Track evidence block (если релиз включает topology/export hardening)

Рекомендуемый формат:

```md
## D-Track Evidence

- Smoke: PASS (`smoke-d-track-topology-export`)
- Export consistency (`json` vs `graphml`): PASS
- Graphviz fallback (`png/svg` -> `json`): PASS
- yEd import: PASS
- Gephi import: PASS
- CI run: <link>
```

Расширенный шаблон: [D_TRACK_EVIDENCE_TEMPLATE.md](D_TRACK_EVIDENCE_TEMPLATE.md).

Короткий ready-to-paste вариант: [D_TRACK_EVIDENCE_PR_SNIPPET.md](D_TRACK_EVIDENCE_PR_SNIPPET.md).

## D-Track Evidence (current snapshot)

### Smoke / CI

- [x] Windows smoke: `.\scripts\smoke-d-track-topology-export.ps1` PASS
- [ ] Linux/macOS smoke: `./scripts/smoke-d-track-topology-export.sh` (pending)
- [ ] CI job `D-Track Smoke (Topology Export)` (pending)
- [ ] CI URL: pending

### Export Consistency (`json` vs `graphml`)

- [x] Node-set equivalence: PASS
- [x] Edge-set equivalence: PASS
- [x] GraphML metadata keys present: `source_type`, `confidence`, `evidence`

### Graphviz Fallback

- [x] Without `dot`: command does not fail, fallback JSON is generated, diagnostic message is present
- [ ] With `dot`: direct `png/svg` generation (pending)

### External Compatibility

- [ ] yEd import PASS (pending)
- [ ] Gephi import PASS (pending)

### Residual Risk

- Pending: external imports (yEd/Gephi), CI confirmation, and `dot`-available validation.
