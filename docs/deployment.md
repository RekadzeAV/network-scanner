# Deployment

## Обзор

Развёртывание зависит от окружения; рекомендуемые точки входа:

- `make build`
- `make test`
- `make deploy`

## Развёртывание на Linux: systemd (E7/7.6)

Юниты лежат в `config/systemd/`:

| Файл | Назначение |
|------|------------|
| `network-scanner-scan.service` | разовое сканирование (`Type=oneshot`) |
| `network-scanner-scan.timer` | периодический запуск (по умолчанию раз в 6 часов) |

Установка:

```bash
make install            # бинарники в /usr/local/bin
make install-systemd    # юнит + таймер, каталоги /var/lib и /var/log

# ARP/ICMP требуют capability на бинарнике (полный root не нужен):
sudo setcap cap_net_raw,cap_net_admin+ep /usr/local/bin/network-scanner
```

Использование:

```bash
systemctl list-timers network-scanner-scan.timer   # расписание
sudo systemctl start network-scanner-scan.service  # разовый запуск
journalctl -u network-scanner-scan.service -f      # журнал
systemctl edit network-scanner-scan.timer          # свой интервал
```

Модель безопасности сервиса: выделенный пользователь `network-scanner`,
`NoNewPrivileges`, `ProtectSystem=strict`, `ProtectHome`, `PrivateTmp`,
`CapabilityBoundingSet`/`AmbientCapabilities` только для `CAP_NET_RAW`/`CAP_NET_ADMIN`.

### Проверка конфигурации

```bash
make units-check        # Linux/macOS (bash)
make units-check-win    # Windows (PowerShell)
```

Скрипт (`scripts/verify-units.sh` / `.ps1`) валидирует: обязательные секции и
ключи юнитов, согласованность `timer.Unit` → существующий `.service`, отсутствие
`ExecReload` (у приложения нет обработчика SIGHUP), неприменение `User=root`,
наличие hardening-ключей и `ReadWritePaths`, корректность desktop-файла
(`Exec`/`TryExec`, отсутствие неподдерживаемой схемы `network://`, наличие иконки).

## Развёртывание в Docker (E7/7.6)

```bash
docker compose build
docker compose run --rm network-scanner scan --network 192.168.1.0/24 --ports 1-1000
docker compose run --rm network-scanner scan --udp --udp-ports 53,161

# Метрики Prometheus (E7/7.8)
docker compose run --rm -p 127.0.0.1:9101:9101 network-scanner \
  scan --network 192.168.1.0/24 --metrics --metrics-addr 0.0.0.0:9101
```

Особенности:

- `network_mode: host` обязателен для ARP/ICMP (broadcast не проходит через
  bridge-сеть); требуется Linux — на Docker Desktop для macOS/Windows используйте
  обычный режим с пробросом портов;
- контейнер работает от непривилегированного пользователя с минимальными
  capabilities `NET_ADMIN`/`NET_RAW` и `no-new-privileges`;
- `.dockerignore` исключает `.git`, артефакты сборки, логи, coverage и
  потенциальные секреты (`.env`) из контекста сборки;
- `BUILD_TIME`/`GIT_COMMIT`/`VERSION` переопределяются переменными окружения.

Образ проверяется в CI: job `Docker image build` собирает его и выполняет
`docker run --rm network-scanner:ci --version`.

## Rollback

Если релиз нужно откатить:

1. Определить последний стабильный тег (`git tag --sort=-v:refname`).
2. Пересобрать артефакты из стабильного тега.
3. Проверить smoke-проверками:
   - `./scripts/integration-check.sh` (Linux/macOS)
   - `.\scripts\integration-check.ps1` (Windows)
   - `make units-check` (конфигурация systemd/desktop)
4. Зафиксировать причину и последствия отката в release notes/changelog.
