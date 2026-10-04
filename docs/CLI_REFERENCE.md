# CLI Reference — `network-scanner`

Актуальная справка по CLI-флагам (`network-scanner scan --help` и др.).

---

## Общие флаги (root commands)

| Флаг | Описание |
|------|----------|

## Подкоманды

| Подкоманда | Назначение |
|-------------|------------|
| `scan` | Сканирование сети |
| `remote-exec` | Удалённое выполнение команд |
| `device-control` | Управление устройствами (перезагрузка, статус) |
| `history` | История сканирований |

---

## `scan` — Сканирование

Базовое использование:
```bash
network-scanner scan
network-scanner scan --network 192.168.1.0/24 --ports 1-1000
```

### Флаги

#### Network (категория: `network`)

| Флаг | Сокращение | Тип | По умолчанию | Описание |
|------|------------|-----|--------------|----------|
| `--network` | `-n` | string | `""` (auto) | CIDR сеть (например, `192.168.1.0/24`) |
| `--ports` | `-p` | string | `"1-1000"` | Диапазон портов (например, `80,443,8080` или `1-1000`) |
| `--hosts-file` | | string | `""` | Файл целей вместо `--network`: target-формат (IP/CIDR/диапазон, `#`-комментарии) либо CSV/JSON (см. `--hosts-format`). **Цели из файла приоритетнее автоопределения локальной сети** |
| `--hosts-format` | | string | `""` (auto) | Формат файла целей: `auto`/`csv`/`txt`/`json`/`targets`. `auto`: `.csv`/`.json` → расширенный импортер (раскрытие CIDR, лимит хостов), иначе прежний target-формат |

#### Scan (категория: `scan`)

| Флаг | Сокращение | Тип | По умолчанию | Описание |
|------|------------|-----|--------------|----------|
| `--timeout` | `-t` | int | `2` | Таймаут в секундах |
| `--threads` | | int | `50` | Количество потоков |
| `--show-closed` | | bool | `false` | Показывать закрытые порты |
| `--udp` | `-u` | bool | `false` | Включить UDP сканирование |
| `--udp-ports` | | string | `""` | UDP-порты при `--udp`: `53,161` или `1-1024`. Пусто — типовые сервисы (53, 67, 68, 69, 123, 161, 162, 514, 1194). Нормализация: дедупликация, сортировка, диапазон 1..65535 (E7/7.2) |
| `--grab-banners` | | bool | `false` | Собирать баннеры сервисов |
| `--os-detect-active` | | bool | `false` | Активные эвристики ОС |
| `--verbose-port-logs` | | bool | `false` | Детальные логи по портам |

#### Post-scan (категория: `post-scan`)

| Флаг | Тип | По умолчанию | Описание |
|------|-----|--------------|----------|
| `--security` | bool | `false` | Запустить анализ безопасности после сканирования |
| `--topology` | bool | `false` | Построить топологию после сканирования |
| `--inventory-save` | bool | `false` | Сохранить результат в inventory |
| `--inventory-id` | string | `""` (auto) | ID снапшота для inventory |

#### SNMP (категория: `snmp`)

| Флаг | Тип | По умолчанию | Описание |
|------|-----|--------------|----------|
| `--snmp` | bool | `false` | Включить SNMP опрос устройств |
| `--snmp-community` | string | `"public"` | SNMP community string |
| `--snmp-timeout` | int | `2` | Таймаут SNMP в секундах |

#### Экспорт

| Флаг | Тип | По умолчанию | Описание |
|------|-----|--------------|----------|
| `--export-html` | bool | `false` | Экспорт результатов в HTML |
| `--export-xml` | bool | `false` | Экспорт результатов в XML |
| `--json` | bool | `false` | Вывод результатов в JSON формате |

| `--metrics` | | bool | `false` | Экспонировать метрики Prometheus на `/metrics` (E7/7.8) |
| `--metrics-addr` | | string | `127.0.0.1:9101` | Адрес эндпоинта метрик (по умолчанию — только loopback) |

Пример экспозиции метрик:

```bash
# Метрики доступны на http://127.0.0.1:9101/metrics во время сканирования
network-scanner scan --network 192.168.1.0/24 --metrics

# Явный адрес (например, для сбора Prometheus в контейнере)
network-scanner scan --network 192.168.1.0/24 --metrics --metrics-addr 0.0.0.0:9101
```

Метрики обновляются по событиям шины (E6): `scan.started` → `network_scanner_scan_total`
и `network_scanner_scan_active`; `scan.completed` → `network_scanner_scan_hosts_total`,
`network_scanner_scan_open_ports_total`, `network_scanner_scan_duration_seconds`;
`scan.failed` → `network_scanner_scan_failures_total`.

---

## `remote-exec` — Удалённое выполнение

```bash
network-scanner remote-exec --transport ssh --target 192.168.1.10 --user admin --command "uptime"
```

| Флаг | Сокращение | Тип | По умолчанию | Описание |
|------|------------|-----|--------------|----------|
| `--transport` | `-t` | string | `""` | Транспорт: `ssh`/`wmi`/`winrm` |
| `--target` | `-T` | string | `""` | Целевой хост/IP |
| `--user` | `-u` | string | `""` | Пользователь |
| `--pass` | `-p` | string | `""` | Пароль |
| `--command` | `-c` | string | `""` | Команда для выполнения |
| `--allow-hosts` | | string | `""` | Список разрешённых хостов (CSV) |
| `--allow-commands` | | string | `""` | Список разрешённых команд (CSV) |
| `--policy-file` | | string | `""` | Файл политики |
| `--policy-strict` | | bool | `false` | Строгая политика |
| `--consent` | | string | `"I_UNDERSTAND"` | Подтверждение операции |
| `--dry-run` | | bool | `false` | Проверить политику без выполнения |
| `--require-tls` | | bool | `false` | Строгий TLS-канал: `ssh` — `StrictHostKeyChecking=yes`; `winrm` — `-usessl` по `https`; для `wmi` не поддерживается (ошибка). Синоним: `--strict-tls` |
| `--timeout` | | int | `15` | Таймаут в секундах |
| `--audit-log` | | string | `""` | Путь к JSONL-журналу операций. **Если не задан — журнал ведётся по умолчанию** в пользовательском конфиг-каталоге (`<UserConfigDir>/network-scanner/remote-exec.log`), никогда в CWD (E7/7.9) |

Пример строгого режима (E7/7.10):

```bash
# SSH: строгая проверка host key
network-scanner remote-exec --transport ssh --target 192.168.1.10 --command "uptime" --require-tls

# WinRM: шифрованный канал (winrs -usessl по https/5986)
network-scanner remote-exec --transport winrm --target host1 --command "hostname" --require-tls --execute --consent I_UNDERSTAND
```

---

## `device-control` — Управление устройствами

```bash
network-scanner device-control --action reboot --target http://192.168.1.1 --confirm I_UNDERSTAND
```

| Флаг | Сокращение | Тип | По умолчанию | Описание |
|------|------------|-----|--------------|----------|
| `--action` | `-a` | string | `""` | Действие: `status`/`reboot` |
| `--target` | `-T` | string | `""` | HTTP(S) endpoint устройства |
| `--vendor` | | string | `"generic-http"` | Провайдер: `generic-http`/`tp-link-http` |
| `--user` | `-u` | string | `""` | Username |
| `--pass` | `-p` | string | `""` | Password |
| `--confirm` | | string | `""` | Подтверждение reboot: `I_UNDERSTAND` |
| `--timeout` | | int | `10` | Таймаут в секундах |
| `--audit-log` | | string | `""` | Путь к JSONL-журналу. **Если не задан — журнал ведётся по умолчанию** в `<UserConfigDir>/network-scanner/device-actions.log` (никогда в CWD, E7/7.9) |

---

## `history` — История

| Флаг | Описание |
|------|----------|
| *(см. исходники)* | Просмотр истории сканирований и сравнение снапшотов |

---

## Примеры

```bash
# Автоматическое определение сети
network-scanner

# Сканирование сети с указанными портами
network-scanner --network 192.168.1.0/24 --ports 80,443,8080

# Сканирование с таймаутом 5с и 100 потоками
network-scanner --timeout 5 --threads 100

# UDP-сканирование с сбором баннеров
network-scanner --udp --grab-banners

# С топологией и security-анализом
network-scanner --topology --security --snmp

# Экспорт в HTML
network-scanner --export-html

# JSON-вывод (для конвейеров)
network-scanner --json
```