# GraphML Compatibility Check (yEd / Gephi)

Документ описывает ручную проверку совместимости `GraphML`-экспорта топологии с внешними инструментами.

## Цель

Подтвердить, что файл, сгенерированный через:

```bash
./network-scanner --topology --output-format graphml --output-file topology.graphml
```

корректно импортируется в:

- yEd Graph Editor
- Gephi

и сохраняет базовые данные узлов/связей.

## Предусловия

- Есть тестовый `topology.graphml` (желательно из сценария smoke D-track).
- Файл содержит `node`/`edge` и data-поля:
  - node: `label`, `type`
  - edge: `src_port`, `dst_port`, `source_type`, `confidence`, `evidence`

## Проверка в yEd

1. Открыть yEd.
2. `File -> Open...` и выбрать `topology.graphml`.
3. Убедиться, что граф загружается без ошибки парсинга.
4. Проверить:
   - отображаются узлы и связи;
   - у узлов видны `label`/name;
   - в свойствах связи доступны `source_type`, `confidence`, `evidence`.
5. Выполнить `Layout` (например, Organic) и убедиться, что рендер проходит без ошибок.

## Проверка в Gephi

1. Открыть Gephi.
2. `File -> Open...` и выбрать `topology.graphml`.
3. В Import Report убедиться, что нет критичных ошибок (warnings допустимы).
4. Проверить в Data Laboratory:
   - таблица узлов содержит `label`, `type`;
   - таблица рёбер содержит `src_port`, `dst_port`, `source_type`, `confidence`, `evidence`.
5. Переключиться в Overview и убедиться, что граф визуализируется.

## Критерии приемки

- [ ] yEd импортирует файл без фатальных ошибок.
- [ ] Gephi импортирует файл без фатальных ошибок.
- [ ] Количество узлов/связей после импорта соответствует исходному файлу.
- [ ] Ключевые атрибуты (`source_type`, `confidence`, `evidence`) доступны в инструментах.

## Артефакты в PR / релизе

Рекомендуется приложить:

- короткий текст "yEd/Gephi import: PASS";
- версии yEd и Gephi;
- при необходимости 1-2 скриншота (Overview/Data Laboratory).
