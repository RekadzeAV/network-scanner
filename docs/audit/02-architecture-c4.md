# Архитектура проекта — C4-модель

**Дата:** 2026-09-22  
**Версия проекта:** v2.3.0  

---

## 2.1. Context (контекст)

```mermaid
graph LR
    User([Пользователь])
    CLI([CLI: network-scanner])
    GUI([GUI: Fyne])
    API([REST API: internal/api])
    Net([Сеть: SNMP, TCP, устройства])

    User -- использует --> CLI
    User -- использует --> GUI
    User -- использует --> API
    CLI -- сканирует --> Net
    GUI -- сканирует --> Net
    API -- upsert --> Inventory
    API -- orchestrates --> Scanner
```

Стороны:
- **Пользователь** — конечный пользователь (администратор/инженер сети).
- **CLI / GUI / REST API** — три клиентских интерфейса приложения.
- **Сеть** — внешняя среда (устройства, сервисы, SNMP-агенты, TCP-порты).

---

## 2.2. Container (контейнеры)

| Контейнер | Технология | Ответственный пакет | Описание |
|-----------|-----------|-------------------|----------|
| CLI утилита | Go (native) | `cmd/network-scanner` | Точка входа, CLI-интерфейс (cobra) |
| GUI приложение | Go + Fyne (CGO) | `internal/gui` | Графический интерфейс |
| REST API сервис | Go (native) | `internal/api` | HTTP API |
| SQLite Хранилище | SQLite + GORM | `internal/inventory` | Инвентарь устройств, история, findings |
| Плагин-система | Go (plugin/registry) | `internal/plugin` | Расширяемость пробы |

---

## 2.3. Component (компоненты)

```mermaid
graph TD
    CLI[cmd/network-scanner]
    GUI[internal/gui]
    API[internal/api/router.go]

    subgraph Core
        Scanner[Scanner Core<br/>internal/scanner]
        Topology[Topology Builder<br/>internal/topology]
        InventoryMgr[Inventory Manager<br/>internal/inventory]
        DeviceCtrl[Device Control<br/>internal/devicecontrol]
        RemoteExec[Remote Execution<br/>internal/remoteexec]
        SecuritySvc[Security Service<br/>internal/security]
        EventBus[Event Bus<br/>internal/eventbus]
        PluginSys[Plugin System<br/>internal/plugin]
        CmdSys[Command System<br/>internal/commands]
        AppConfig[Config Validation<br/>internal/configvalidation]
        AppErrors[AppError<br/>internal/apperror]
    end

    CLI --> Scanner
    GUI --> Scanner
    API --> Scanner
    API --> InventoryMgr
    API --> Topology

    Scanner --> Topology
    Topology --> InventoryMgr
    InventoryMgr --> DB[(SQLite)]

    Scanner --> PluginSys
    Scanner --> EventBus
    Topology --> EventBus
    InventoryMgr --> EventBus
    DeviceCtrl --> RemoteExec
    RemoteExec --> SecuritySvc
    SecuritySvc --> AppErrors
    AppConfig --> AppErrors
```

---

## 2.4. Архитектурный слой v2.3 (C1–C6)

| Модуль | Статус | Интеграция |
|--------|--------|-----------|
| `internal/apperror` | ✅ Создан, тесты | ⚠️ Самодостаточная инфраструктура, 0 внешних импортёров |
| `internal/commands` | ✅ Создан, тесты | ⚠️ Самодостаточная, 0 внешних импортёров |
| `internal/eventbus` | ✅ Создан, тесты | ⚠️ Самодостаточная, 0 внешних импортёров |
| `internal/plugin` | ✅ Создан, тесты (~60%) | ⚠️ Самодостаточная, 0 внешних импортёров |
| `internal/configvalidation` | ✅ Создан, тесты | ⚠️ Самодостаточная, 0 внешних импортёров |
| `internal/benchmark` | ✅ Создан | ⚠️ Не используется в CI |

**Решение (E6, 2026-09-19):**  
Слой v2.3 остаётся **протестированным фундаментом**. Новые фичи пишутся на `apperror`/`eventbus`/`commands` с первого коммита. Ретрофит в легаси-бизнес-пути не выполняется (высокий риск регрессий без пользовательской ценности).

---

## 2.5. Карта зависимостей

```
cmd/network-scanner → internal/scanner → internal/topology
                   → internal/inventory → internal/api
                   → internal/plugin
                   → internal/eventbus
                   → internal/commands
                   → internal/apperror
                   → internal/configvalidation
                   → internal/benchmark
                   → internal/devicecontrol
                   → internal/remoteexec
                   → internal/security
                   → internal/redact
                   → internal/contracts
```

**Циклические зависимости:** ❌ Не обнаружено.

---