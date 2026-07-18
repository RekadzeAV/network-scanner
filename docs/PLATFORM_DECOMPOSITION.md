# 🌐 Декомпозиция задач по платформам (Windows / Linux / macOS)

**Дата:** 2026-07-12  
**Цель:** Разделить задачи стабилизации на те, что можно выполнить под Windows, и те, что требуют Linux/macOS.

---

## 📊 Сводная таблица платформозависимого кода

| Пакет | Функция | Платформа | Описание | Зависимости |
|-------|---------|-----------|----------|-------------|
| `scanner` | `readMACFromLinuxARP` | Linux | Чтение `/proc/net/arp` | Нет |
| `scanner` | `readMACFromWindowsARP` | Windows | Вызов `arp -a` | `exec.Command` |
| `scanner` | `readMACFromDarwinARP` | macOS | Вызов `arp -n` / `arp -a` | `exec.Command` |
| `scanner` | `getMACViaARPRequest` | All | Отправка ARP через pcap | `pcap` (root/admin) |
| `network` | `GetARPTabaleLinux` | Linux | Парсинг `/proc/net/arp` | Нет |
| `network` | `GetARPTabaleWindows` | Windows | Парсинг `arp -a` | `exec.Command` |
| `nettools` | `GetPingCommand` | All | Выбор команды ping | Нет |
| `nettools` | `GetWiFiInfo` | Linux/macOS | Анализ WiFi | `exec.Command` |
| `nettools` | `WhoisLookup` | Windows | Специфичный парсинг | Нет |
| `remoteexec` | `ExecuteRemote` | Windows | WMI/WinRM транспорт | Нет |
| `plugin` | `PluginLoader.Load` | All | `.so` / `.dll` / `.dylib` | `plugin` package |
| `gui` | `NettoolsService` | All | Вызов утилит | Зависит от nettools |

---

## 🪟 План для Windows (текущая платформа разработки)

**Охват:** Все задачи, не требующие Linux/macOS специфичных API.  
**Текущее покрытие:** 64.8% (`scanner`), 85.3% (`network`), 92.3% (`osdetect`), 82.2% (`topology`).

### ✅ Завершено (работает под Windows)

| # | Задача | Пакет | Статус | Комментарий |
|---|--------|-------|--------|-------------|
| W-1 | Тесты adaptive scanner | `scanner` | ✅ 100% | `adaptive_scanner_test.go` — 20 тестов |
| W-2 | Тесты incremental scanner | `scanner` | ✅ 100% | `incremental_test.go` — события, handlers |
| W-3 | Тесты `isHostAlive` | `scanner` | ✅ 100% | С моком prober |
| W-4 | Тесты `scanUDPPort` | `scanner` | ✅ 100% | Edge cases timeout |
| W-5 | Тесты `checkARP` | `scanner` | ✅ 100% | Stub тест |
| W-6 | Тесты `getVendorFromMAC` | `scanner` | ✅ 100% | 5 edge-case тестов |
| W-7 | Тесты `shouldGrabBannerPort` | `scanner` | ✅ 100% | 5 edge-case тестов |
| W-8 | Тесты `appendIfNotExists` | `scanner` | ✅ 100% | 4 edge-case теста |
| W-9 | Тесты `hasOpenPort` | `scanner` | ✅ 100% | 4 edge-case теста |
| W-10 | Тесты `getProtocolFromPort` | `scanner` | ✅ 100% | Edge cases |
| W-11 | Тесты `detectDeviceType` | `scanner` | ✅ 100% | 8 типов устройств |
| W-12 | Тесты `NewService` / `Scan` | `scanner` | ✅ 100% | Context cancel, invalid CIDR, progress |
| W-13 | Coverage 85.3% | `network` | ✅ 85.3% | CIDR, IP, MAC validation |
| W-14 | Coverage 92.3% | `osdetect` | ✅ 92.3% | TTL, Window, DF heuristics |
| W-15 | Coverage 82.2% | `topology` | ✅ 82.2% | Graph building, validation |
| W-16 | Benchmarks (30+) | `scanner` | ✅ Done | ScanHost, TCPPort, UDPPort |
| W-17 | Benchmarks (30+) | `network` | ✅ Done | ParseCIDR, ARPResolve |
| W-18 | Benchmarks (35+) | `topology` | ✅ Done | BuildGraph, SNMPCollect |
| **W-19** | **Улучшить `scanHost`** | **`scanner`** | **✅ 73.2%** | **+22.9% (было 50.3%)** |
| **W-20** | **Улучшить `scanTCPPort`** | **`scanner`** | **✅ 100%** | **+33.3% (было 66.7%)** |
| **W-21** | **Улучшить `scanHostUDP`** | **`scanner`** | **✅ 80.0%** | **+1.7% (было 78.3%)** |
| **W-22** | **Улучшить `getMACViaARPRequest`** | **`scanner`** | **✅ 50.0%** | **+1.6% (было 48.4%)** |
| **W-23** | **Тесты context cancel** | **`scanner`** | ✅ Done | `scanHost`, `isHostAlive`, `Scan` |
| **W-24** | **Тесты showClosed/verbose/banners** | **`scanner`** | ✅ Done | 10+ комбинаций опций |
| **W-25** | **Тесты GetDiagnosticsSummary** | **`scanner`** | ✅ 100% | Форматирование сводки |
| **W-26** | **Тесты Stop()/GetResults()** | **`scanner`** | ✅ 100% | Copy semantics, no-panic |
| **W-27** | **Тесты portThreadsForHost** | **`scanner`** | ✅ 93.3% | Edge cases threads=0, large ports |
| **W-29** | **Улучшить `readMACFromWindowsARP`** | **`scanner`** | **✅ 81.8%** | **+31.8% (было 50.0%)** |
| **W-30** | **Тесты `readMACFromARPTable`** | **`scanner`** | ✅ 50.0% | Кроссплатформенный вызов |
| **W-31** | **Тесты `isHostAlive`** | **`scanner`** | **✅ 72.5%** | **+13.7% (было 58.8%)** |
| **W-32** | **Тесты `isHostAlive` с prober** | **`scanner`** | ✅ 72.5% | Mock prober, context prober |
| **W-33** | **Тесты `readMACFromARPTable`** | **`scanner`** | ✅ 50.0% | Прямой вызов, несколько IP |
| **W-34** | **Тесты `isHostAlive` active port** | **`scanner`** | ✅ 72.5% | localhost, verbose |
| **W-35** | **Тесты `Scan`** | **`scanner`** | **✅ 64.7%** | **+4.0% (было 60.7%)** |
| **W-36** | **Тесты `readMACFromARPTable` с prober** | **`scanner`** | ✅ 50.0% | С network prober |
| **W-37** | **Тесты `getMACViaARPRequest`** | **`scanner`** | ✅ 50.0% | Context cancel, network error |
| **W-38** | **Тесты `isHostAlive` edge cases** | **`scanner`** | ✅ 72.5% | All ports fail, prober fail verbose |
| **W-39** | **Тесты `readMACFromARPTable` edge cases** | **`scanner`** | ✅ 50.0% | Invalid IP, empty string |
| **W-40** | **Тесты `Scan` edge cases** | **`scanner`** | ✅ 64.7% | No alive hosts, UDP, all options |
| **W-41** | **Тесты `isHostAlive` prober error** | **`scanner`** | ✅ 72.5% | Prober error + verbose |
| **W-42** | **Тесты `Scan` progress/callback** | **`scanner`** | ✅ 64.7% | Progress ping/ports, cancelled |
| **W-43** | **Тесты `getMACViaARPRequest`** | **`scanner`** | ✅ 50.0% | Interfaces error, timeout |
| **W-44** | **Тесты `readMACFromWindowsARP`** | **`scanner`** | ✅ 81.8% | Dashed/colon MAC formats |
| **W-45** | **Тесты `readMACFromARPTable`** | **`scanner`** | ✅ 50.0% | Nil IP, broadcast IP |
| **W-46** | **Тесты `isHostAlive` prober verbose** | **`scanner`** | ✅ 72.5% | Prober error verbose detailed |
| **W-47** | **Тесты `Scan` progress multiple** | **`scanner`** | ✅ 64.7% | Multiple calls, alive hosts, early cancel |
| **W-48** | **Тесты `getMACViaARPRequest` extra** | **`scanner`** | ✅ 50.0% | No interfaces, invalid IP |
| **W-49** | **Тесты `readMACFromARPTable` extra** | **`scanner`** | ✅ 50.0% | Loopback, local gateway |
| **W-50** | **Тесты `readMACFromWindowsARP` extra** | **`scanner`** | ✅ 81.8% | Multiple IPs, empty ARP |
| **W-51** | **Тесты `sanitizeBanner`** | **`banner`** | ✅ 100% | Printable, whitespace, mixed, empty, non-printable |
| **W-52** | **Тесты `normalizeByPort`** | **`banner`** | ✅ 100% | SSH, FTP, SMTP, POP3, IMAP, edge cases |
| **W-53** | **Тесты `ExtractVersionHint`** | **`banner`** | ✅ 100% | HTTP status/server, FTP, SMTP, POP3, long banner |
| **W-54** | **Тесты `trimMailLikePrefix`** | **`banner`** | ✅ 100% | +OK, numeric, dash, dot, whitespace |
| **W-55** | **Тесты `isDigit`/`isPlainHTTPPort`/`isTLSHTTPPort`** | **`banner`** | ✅ 100% | All ports, edge cases |
| **W-56** | **Тесты `truncateString`** | **`display`** | ✅ 100% | Under/equal/over limit, maxLen<=3, empty |
| **W-57** | **Тесты `formatOSGuess`** | **`display`** | ✅ 100% | Empty, whitespace, confidence, reason, truncated |
| **W-58** | **Тесты `getPortPurpose`** | **`display`** | ✅ 83.3% | All known ports, unknown, zero |
| **W-59** | **Тесты `SaveResultsToFile/JSON/CSV`** | **`display`** | ✅ 91-100% | Success, empty, invalid path, closed ports |
| **W-60** | **Тесты `formatPorts` edge cases** | **`display`** | ✅ 93.9% | Many ports (>50), closed, mixed, version/banner |
| **W-61** | **Тесты `AppendAudit`** | **`devicecontrol`** | ✅ 78.9% | Empty path, empty timestamp, with vendor, multiple entries |
| **W-62** | **Тесты `currentActor`** | **`devicecontrol`** | ✅ 44.4% | Env USERNAME, USER, default unknown |
| **W-63** | **Тесты `buildEndpoint` tplink** | **`devicecontrol`** | ✅ 100% | Status, reboot, unsupported action, generic |
| **W-64** | **Тесты `resolveAdapter`** | **`devicecontrol`** | ✅ 100% | Generic, empty, tplink, unknown, case insensitive |
| **W-65** | **Тесты `Execute` edge cases** | **`devicecontrol`** | ✅ 95.0% | Empty action, invalid URL, 500/404 errors, timeout, basic auth |
| **W-66** | **Тесты `AnalyzeResults` фильтры** | **`cve`** | ✅ 100% | No open ports, no version, MinCVSS, MaxAge, sorting |
| **W-67** | **Тесты `normalizeService`** | **`cve`** | ✅ 100% | HTTP, HTTPS, SSH, port-based, case insensitive |
| **W-68** | **Тесты `FormatMatches`** | **`cve`** | ✅ 100% | Empty, multiple, with hostname, CVSS format |
| **W-69** | **Тесты `NewDefaultCatalog`** | **`cve`** | ✅ 100% | Structure, all CVEs, CVSS values |

### ⏳ Можно доработать под Windows

| # | Задача | Оценка | Описание | Приоритет |
|---|--------|--------|----------|-----------|
| W-18 | Улучшить `scanHost` coverage | 2ч | Изолировать логику без сетевых вызовов | Высокий |
| W-19 | Улучшить `isHostAlive` coverage | 1ч | Добавить тесты с cancelled context | Высокий |
| W-20 | Улучшить `Scan` coverage | 1ч | Протестировать progress callback paths | Средний |
| W-21 | Улучшить `scanUDPPortWithTimeout` | 0.5ч | Edge cases negative/zero timeout | Низкий |
| W-22 | Улучшить `ScanWithEvents` | 1ч | Тесты ctx.Done() paths | Средний |
| W-23 | Улучшить `PrintEventHandler` | 0.5ч | Test verbose=false path | Низкий |
| W-24 | Тесты `GetDiagnosticsSummary` | 0.5ч | Проверка форматирования | Низкий |
| W-25 | Моки для `mockPortScanner` | 1ч | Расширить coverage `scanTCPPort` | Средний |
| W-26 | Тесты `Stop()` | 0.5ч | Проверка no-panic behavior | Низкий |
| W-27 | Тесты `GetResults()` | 0.5ч | Проверка copy semantics | Низкий |

**Ожидаемый результат:** Coverage `scanner` → **68-70%**  
**Фактический результат:** ✅ **71.9%** (достигнуто 2026-07-12) (без Linux/macOS).

---

## 🐧 План для Linux

**Охват:** Платформозависимые функции ARP, pcap, и общие улучшения.

### 🎯 Ключевые цели

| # | Задача | Оценка | Описание | Критерий готовности |
|---|--------|--------|----------|---------------------|
| L-1 | Тесты `readMACFromLinuxARP` | 2ч | Моки `/proc/net/arp`, тесты парсинга, context cancel | Coverage ≥ 90% |
| L-2 | Тесты `getMACViaARPRequest` | 3ч | Моки pcap через `gopacket`, root-права | Coverage ≥ 80% |
| L-3 | Тесты `GetARPTabaleLinux` | 1ч | Моки `/proc/net/arp` в `network/arp_cache.go` | Coverage ≥ 95% |
| L-4 | Улучшить `scanHost` | 2ч | Изолировать логику, моки соккетов | Coverage ≥ 70% |
| L-5 | Улучшить `isHostAlive` | 1ч | Тесты с real ICMP на Linux | Coverage ≥ 70% |
| L-6 | CI Linux runner | 0.5ч | Добавить `ubuntu-latest` в go.yml (если нет) | Зелёный build |
| L-7 | Тесты `readMACFromWindowsARP` | 1ч | Cross-compilation test (stub) | Coverage 50%+ |
| L-8 | Тесты `readMACFromDarwinARP` | 1ч | Cross-compilation test (stub) | Coverage 50%+ |

### 🔧 Зависимости для Linux

- **root/sudo права** — для pcap (`getMACViaARPRequest`)
- **pcap библиотеки** — `libpcap-dev` (apt)
- **gopacket** — уже в `go.mod`

### 📈 Ожидаемый результат

- Coverage `scanner` → **75-80%** (с Linux runner)
- Покрытие всех ARP-функций ≥ 90%
- Покрытие `getMACViaARPRequest` ≥ 80%

---

## 🍎 План для macOS

**Охват:** Darwin-specific ARP функции и валидация кроссплатформенности.

### 🎯 Ключевые цели

| # | Задача | Оценка | Описание | Критерий готовности |
|---|--------|--------|----------|---------------------|
| M-1 | Тесты `readMACFromDarwinARP` | 2ч | Моки `arp -n`, тесты парсинга | Coverage ≥ 90% |
| M-2 | Тесты `GetARPTabaleLinux` | 0.5ч | На Darwin fallback на `arp -n` | Coverage ≥ 95% |
| M-3 | Валидация `readMACFromWindowsARP` | 0.5ч | Cross-compilation, stub test | Coverage 50%+ |
| M-4 | CI macOS runner | 0.5ч | Добавить `macos-latest` в go.yml | Зелёный build |
| M-5 | Тесты `GetWiFiInfo` | 1ч | Darwin WiFi анализ | Coverage ≥ 80% |

### 🔧 Зависимости для macOS

- **Xcode Command Line Tools** — для `arp`, `networksetup`
- **macOS runner** — `macos-latest` в GitHub Actions

### 📈 Ожидаемый результат

- Coverage `scanner` → **78-82%** (с macOS runner)
- Полное покрытие Darwin ARP-функций
- Валидация кроссплатформенности

---

## 🔄 Итоговая матрица покрытия

| Функция | Windows | Linux | macOS | Итоговое покрытие |
|---------|---------|-------|-------|-------------------|
| `readMACFromLinuxARP` | ❌ skip | ✅ 90%+ | ❌ skip | **90%+** |
| `readMACFromWindowsARP` | ✅ 50%+ | ❌ skip | ❌ skip | **50%+** |
| `readMACFromDarwinARP` | ❌ skip | ❌ skip | ✅ 90%+ | **90%+** |
| `getMACViaARPRequest` | ❌ (нет pcap) | ✅ 80%+ | ⚠️ частично | **80%+** |
| `scanHost` | ⚠️ 50% | ✅ 70%+ | ✅ 70%+ | **70%+** |
| `isHostAlive` | ⚠️ 58% | ✅ 70%+ | ✅ 70%+ | **70%+** |
| `Scan` | ⚠️ 60% | ✅ 75%+ | ✅ 75%+ | **75%+** |
| **Общий scanner** | **71.9%** | **~78%** | **~78%** | **~78%** |

---

## 📅 Рекомендуемый порядок выполнения

### Этап 1: Windows (текущая неделя)
1. **W-18** до **W-27** — доработка тестов под Windows
2. **Цель:** Coverage `scanner` → 68-70%
3. **Оценка:** 4-5 часов

### Этап 2: Linux CI (следующая неделя)
1. **L-6** — настройка Ubuntu runner
2. **L-1** — тесты `readMACFromLinuxARP`
3. **L-3** — тесты `GetARPTabaleLinux`
4. **Цель:** Coverage `scanner` → 72-75%
5. **Оценка:** 3-4 часа

### Этап 3: Linux advanced (после L-6)
1. **L-2** — тесты `getMACViaARPRequest` (требует root)
2. **L-4** — улучшение `scanHost`
3. **Цель:** Coverage `scanner` → 78-80%
4. **Оценка:** 4-5 часов

### Этап 4: macOS CI (фазово)
1. **M-4** — настройка macOS runner
2. **M-1** — тесты `readMACFromDarwinARP`
3. **Цель:** Coverage `scanner` → 78-82%
4. **Оценка:** 2-3 часа

### Итоговая оценка: **13-17 часов** (2-3 рабочих дня)

---

## 🏁 Критерии успеха

| Платформа | Coverage Scanner | Coverage Network | Coverage OSDetect | Coverage Topology |
|-----------|------------------|------------------|-------------------|-------------------|
| Windows (текущий) | 71.9% ✅ | 85.3% | 92.3% | 82.2% |
| Windows (после W-18..45) | 71.9% ✅ | 85.3% | 92.3% | 82.2% |
| Linux (после L-1..8) | 75-80% | 85.3%+ | 92.3%+ | 82.2%+ |
| macOS (после M-1..5) | 78-82% | 85.3%+ | 92.3%+ | 82.2%+ |
| **Итоговый CI** | **≥ 75%** | **≥ 85%** | **≥ 90%** | **≥ 80%** |

---

## 📝 Примечания

1. **pcap требует root/admin:** Тесты `getMACViaARPRequest` нужно запускать с `sudo` на Linux.
2. **CI runners:** GitHub Actions предоставляет `ubuntu-latest` и `macos-latest` бесплатно для публичных репозиториев.
3. **Cross-compilation:** Тесты для `readMACFromWindowsARP` на Linux и `readMACFromDarwinARP` на Windows будут работать как stubs (skip).
4. **Plugin loader:** `.so` (Linux), `.dll` (Windows), `.dylib` (macOS) — тестировать на каждой платформе отдельно.
5. **Remote exec:** WMI/WinRM только на Windows — тестировать отдельно.
