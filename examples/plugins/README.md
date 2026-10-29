# Примеры плагинов Network Scanner

В этой директории находятся примеры внешних плагинов для динамической
загрузчика (L2.2, `internal/plugin`). Каждый пример — отдельный пакет `main`
с экспортированным символом `New`, который требует загрузчик для создания
экземпляра плагина.

## Структура каждого плагина

```
onlyup/                          # каталог плагина
├── main.go                      # код плагина + func New() (обязательно)
├── main_test.go                 # unit-тесты
├── plugin.json                  # манифест загрузчика
├── dist/                        # результат сборки
│   ├── plugin.json
│   └── onlyup.so (или .dll)
└── build.ps1 / build.sh         # скрипты сборки
```

### Ключевые правила

1. **`package main`** — необходимо для `-buildmode=plugin`.
2. **`func main() {}`** — обязателен (заглушка), иначе `go build ./...` упадёт.
3. **Экспортированный символ `New`** — функция-конструктор:
   ```go
   func New() (plugin.Plugin, error) {
       return &MyPlugin{}, nil
   }
   ```
4. **`plugin.json`** — манифест, который читает загрузчик L2.2.

## Доступные плагины

| Плагин | Тип | Описание |
| --- | --- | --- |
| **onlyup** | filter | Пропускает только хосты с хотя бы одним открытым портом (`state == "open"`). |
| **json-exporter** | exporter | Экспорт результатов в JSON Lines через `Export(results, path)`. |
| **alert-report** | reporter | Генерация текстового отчёта о критических портах (21, 22, 23, 80, 443, 445, 3389). |

## Сборка

### Windows

```powershell
cd examples/plugins
.\build.ps1                # собрать все плагины
.\build.ps1 -Only onlyup   # собрать один плагин
```

### Linux / macOS

```bash
cd examples/plugins
chmod +x build.sh
./build.sh                 # собрать все плагины
./build.sh onlyup          # собрать один
```

> Сборка требует **CGO-enabled** и компилятор C (MinGW-w64 на Windows,
> gcc/clang на Linux/macOS). Если CGO недоступен (`CGO_ENABLED=0`),
> сборка не будет возможна — для таких сценариев используйте встроенные
> плагины (`internal/plugin`).

### Ручная сборка

```bash
CGO_ENABLED=1 \
  go build -buildmode=plugin \
  -o dist/onlyup.so \
  onlyup
```

## Использование

### Переменная окружения

Укажите каталог с `dist/` в `NETWORK_SCANNER_PLUGINS_DIR` и запустите:

```bash
export NETWORK_SCANNER_PLUGINS_DIR=/path/to/onlyup/dist
network-scanner plugins list
```

### Встроенный плагин-каталог

Скопируйте `dist/onlyup/` рядом с `plugin.json` в `plugins/` рядом с
исполняемым файлом — загрузчик найдёт поддиректорию автоматически.

## Ограничения

- **Выгрузка невозможна.** Go runtime не поддерживает `plugin.Close()` /
  выгрузку `.so` — загрузчик сохраняет открытые библиотеки до перезапуска
  процесса. `WatchForChanges` перезагружает НОВЫЕ плагины и обновляет
  манифесты (enabled/disabled/description), но не перезаменяет код
  уже загруженного плагина.
- **Windows** требует MinGW-w64 gcc для CGO. MSVC не поддерживает
  `-buildmode=plugin`.
- **CGO_ENABLED=0**. Без включённого CGO динамическая загрузка невозможна.
- **Совместимость бинарников.** Плагин должен быть собран той же версией
  Go, что и основной бинарник, или совместимой (минорная разница допустима
  при отсутствии изменений ABI).

## Создание собственного плагина

1. Создайте каталог `myplugin/` с `package main`.
2. Реализуйте интерфейс `plugin.Plugin` (или `plugin.FilterPlugin`,
   `plugin.ExporterPlugin`, `plugin.ReporterPlugin`).
3. Добавьте экспортированную функцию `func New() (plugin.Plugin, error)`.
4. Создайте `plugin.json` с корректными полями.
5. Соберите: `go build -buildmode=plugin -o dist/myplugin.so .`.
6. Запустите: `NETWORK_SCANNER_PLUGINS_DIR=./dist network-scanner plugins list`.

Подробное описание интерфейсов — в doc-комментариях `internal/plugin/plugin.go`
(`Plugin`, `FilterPlugin`, `ExporterPlugin`, `ScannerPlugin`, `ReporterPlugin`)
и в формате манифеста — `internal/plugin/manifest.go`.
