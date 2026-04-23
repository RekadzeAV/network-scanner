# Manual Sign-off Template

Шаблон для быстрого закрытия оставшихся ручных пунктов перед финальным release sign-off.

## Context

- Reviewer:
- Date:
- Environment (OS/arch):
- Build under test (commit/tag):

## GUI Smoke (`docs/GUI_SMOKE_CHECKLIST.md`)

- [ ] Пройден полностью
- Notes:
  - Сканирование/остановка/состояния:
  - Режимы `Таблица`/`Карточки`:
  - Автопрофиль (`ВКЛ/ВЫКЛ`, индикаторы, инфо-диалог):
  - Персистентность настроек:
  - Экспорт текущего представления:
  - Инструменты (`Ping`/`Traceroute` raw output, Windows ping summary):

## Topology Manual Checks

- [ ] Вкладка `Топология` открывается корректно
- [ ] Построение топологии без ошибок
- [ ] Превью PNG/масштабирование/сохранение работают
- Notes:
  - SNMP summary:
  - Build duration / responsiveness:
  - Preview/save behavior:

## GraphML Compatibility (`docs/GRAPHML_COMPATIBILITY_CHECK.md`)

- [ ] yEd import verified
- [ ] Gephi import verified
- [ ] Results captured in compatibility checklist/doc
- Notes:
  - yEd:
  - Gephi:

## CI Evidence

- [ ] `CI` workflow green (`Lint`, `Test`, `Build and Smoke`)
- [ ] Windows/Linux/macOS confirmed
- [ ] CI URL recorded in `docs/P3_CLOSURE_CHECKLIST.md`
- CI run URL:

## Final Decision

- [ ] Ready for final sign-off
- [ ] Defer (follow-up required)
- Residual risks / follow-ups:

