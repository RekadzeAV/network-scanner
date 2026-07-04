package scanner

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	"network-scanner/internal/banner"
	"network-scanner/internal/logger"
	"network-scanner/internal/network"
	"network-scanner/internal/osdetect"
	portdb "network-scanner/internal/ports"
	"network-scanner/internal/scanner/deviceclassifier"
)

// Package scanner предоставляет основной движок сканирования сети.
//
// # Основные компоненты
//
// NetworkScanner — главный struct для сканирования сети.
//
//	NewNetworkScanner() — создает сканер с дефолтными зависимостями.
//	NewScanner() — создает сканер с внедренными зависимостями (DI).
//	Scan() — запускает сканирование сети.
//	Stop() — останавливает сканирование.
//	GetResults() — возвращает результаты сканирования.
//
// # Процесс сканирования
//
//  1. ParseNetworkRange — парсит CIDR диапазон (например 192.168.1.0/24)
//  2. Ping discovery — проверяет доступность хостов через ICMP/ports
//  3. Port scanning — сканирует TCP/UDP порты на активных хостах
//  4. MAC/Hostname — получает MAC адрес и hostname для каждого хоста
//  5. Device type — определяет тип устройства по открытым портам и MAC
//  6. SNMP probe — проверяет доступность SNMP (порт 161)
//
// # UDP сканирование
//
// UDP сканирование включает известные порты:
//
//	const knownUDPPorts = 9 // 53, 67, 68, 69, 123, 161, 162, 514, 1194
//
// Для каждого хоста запускается параллельное сканирование с ограничением
// параллельности (udpSemaphoreSize=50).
//
// # Проверка живости хоста
//
// isHostAlive проверяет доступность хоста через probe на commonPorts:
//
//	const commonHostPorts = 6 // 80, 443, 22, 135, 139, 445
//
// Использует параллельные dial-connections с таймаутом probeTimeout.
//
// # Banner grabbing
//
// При включенном grabBanners, для открытых портов из shouldGrabBannerPort
// выполняется сбор баннера с таймаутом bannerGrabTimeout.
//
// # Определение типа устройства
//
// detectDeviceType использует deviceclassifier для определения типа:
//
//   - Router/Network Device: порты 80, 443, 22, 161, 514
//   - Computer: порты 135, 139, 445, 3389
//   - Server: порты 22, 80, 443, 3306, 5432
//   - Unknown: другие комбинации
//
// # MAC адрес
//
// getMACAddress пытается получить MAC через:
//
//  1. networkProber.ResolveMAC (если внедрен)
//  2. Системную ARP таблицу (/proc/net/arp, arp -a, arp -n)
//  3. PCAP ARP request (требует root прав)
//
// # Конфигурация таймаутов
//
//	const (
//	    udpProbeTimeoutDivisor = 3    // probeTimeout = timeout / 3
//	    hostProbeTimeoutMin    = 150ms // минимальный probe timeout
//	    hostProbeTimeoutMax    = 800ms // максимальный probe timeout
//	    bannerGrabTimeoutDivisor = 2   // bannerTimeout = timeout / 2
//	    bannerGrabTimeoutMin     = 300ms
//	    bannerGrabTimeoutMax     = 2s
//	    snmpProbeTimeoutMax      = 500ms
//	)
//
// # Пример использования
//
//	import "network-scanner/internal/scanner"
//
//	func main() {
//	    ns := scanner.NewNetworkScanner("192.168.1.0/24", 2*time.Second, "1-1024", 50, false)
//	    ns.SetScanUDP(true)
//	    ns.SetGrabBanners(true)
//	    ns.SetProgressCallback(func(stage string, current, total int, msg string) {
//	        fmt.Printf("%s: %d/%d - %s\n", stage, current, total, msg)
//	    })
//	    ns.Scan()
//	    results := ns.GetResults()
//	}
//
// # Потокобезопасность
//
// NetworkScanner потокобезопасен для вызовов:
//
//   - SetScanUDP, SetScanTCPPorts, SetGrabBanners — до вызова Scan()
//   - GetResults — во время и после Scan()
//   - Stop — во время Scan() для отмены
//
//内部的 results слайс защищен sync.RWMutex.

// Result содержит результаты сканирования одного хоста
type Result struct {
	IP                string
	MAC               string
	Hostname          string
	Ports             []PortInfo
	Protocols         []string
	DeviceType        string
	DeviceVendor      string
	SNMPEnabled       bool
	IsAlive           bool
	GuessOS           string // эвристическая оценка ОС (опционально)
	GuessOSConfidence string // низкая/средняя/высокая
	GuessOSReason     string // краткое обоснование эвристики
}

// PortInfo содержит информацию о порте
type PortInfo struct {
	Port     int
	State    string // "open", "closed", "filtered"
	Protocol string // "tcp", "udp"
	Service  string
	Banner   string // сырой ответ службы (опционально)
	Version  string // краткая версия/сигнатура службы (опционально)
}

// ProgressCallback функция для передачи прогресса сканирования
type ProgressCallback func(stage string, current int, total int, message string)

// NetworkScanner выполняет сканирование сети
type NetworkScanner struct {
	network          string
	timeout          time.Duration
	portRange        string
	threads          int
	showClosed       bool
	scanTCPPorts     bool // Сканировать TCP-порты из portRange (если false — только ping/MAC/hostname)
	scanUDP          bool // Включить UDP сканирование
	grabBanners      bool // Читать баннеры с типовых TCP-портов (медленнее)
	osDetectActive   bool // Активный режим эвристик ОС (дополнительные сигнатуры)
	verbosePortLogs  bool // Подробные логи по каждому порту/пробе (шумно, медленнее)
	results          []Result
	mu               sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	progressCallback ProgressCallback
	networkProber    NetworkProber
	portScanner      PortScanner
	udpPortScanner   PortScanner
	resultPresenter  ResultPresenter
	tcpCancelBefore  int64
	tcpCancelWait    int64
	udpCancelHosts   int64
	tcpProbeTotal    int64
	tcpProbeOpen     int64
	tcpProbeClosed   int64
	udpProbeTotal    int64
	udpProbeOpen     int64
	udpProbeNoOpen   int64
	lastPingNs       int64
	lastPortscanNs   int64
	lastTotalNs      int64
}

const (
	// globalPortProbeBudget ограничивает общее число одновременных TCP probe во время этапа порт-сканирования.
	// Это предотвращает перегрузку сокетов/сети, когда одновременно сканируется много хостов.
	globalPortProbeBudget = 512
	minPerHostPortThreads = 8
	maxPerHostPortThreads = 64

	// UDP порты для сканирования
	knownUDPPorts = 9

	// Магические числа для сканирования
	udpSemaphoreSize       = 50
	udpResultBufferSize    = 9 // равно knownUDPPorts
	udpCollectTimeout      = 100 * time.Millisecond
	udpProbeTimeoutDivisor = 3

	// Таймауты для проверки живости
	hostProbeTimeoutMin = 150 * time.Millisecond
	hostProbeTimeoutMax = 800 * time.Millisecond

	// Таймауты для banner grabbing
	bannerGrabTimeoutDivisor = 2
	bannerGrabTimeoutMin     = 300 * time.Millisecond
	bannerGrabTimeoutMax     = 2 * time.Second

	// SNMP probe timeout
	snmpProbeTimeoutMax = 500 * time.Millisecond

	// SNMP UDP/TCP порт
	snmpPort = 161

	// Задержки для неблокирующих операций
	macTimeout         = 100 * time.Millisecond
	hostnameTimeout    = 100 * time.Millisecond
	arpCommandTimeout  = 3 * time.Second
	ifaceTimeout       = 3 * time.Second
	ifaceAddrTimeout   = 1 * time.Second
	arpResponseTimeout = 2 * time.Second

	// Common ports для проверки живости хоста
	commonHostPorts = 6

	// MAC OUI prefix length
	macOUIPrefixLength = 8

	// Windows ARP MAC format length
	windowsMACFormatLength = 17

	// PCAP buffer size
	pcapBufferSize = 1024
)

// NewNetworkScanner создает новый сканер
func NewNetworkScanner(networkCIDR string, timeout time.Duration, portRange string, threads int, showClosed bool) *NetworkScanner {
	return NewScanner(
		networkCIDR,
		timeout,
		portRange,
		threads,
		showClosed,
		network.DefaultNetworkProber{Timeout: timeout},
		network.TCPPortScanner{Timeout: timeout},
		nil,
	)
}

// NewScanner создает сканер с явным внедрением зависимостей.
func NewScanner(
	networkCIDR string,
	timeout time.Duration,
	portRange string,
	threads int,
	showClosed bool,
	networkProber NetworkProber,
	portScanner PortScanner,
	resultPresenter ResultPresenter,
) *NetworkScanner {
	ctx, cancel := context.WithCancel(context.Background())
	return &NetworkScanner{
		network:          networkCIDR,
		timeout:          timeout,
		portRange:        portRange,
		threads:          threads,
		showClosed:       showClosed,
		scanTCPPorts:     true,
		scanUDP:          false, // По умолчанию UDP сканирование выключено
		grabBanners:      false,
		osDetectActive:   false,
		verbosePortLogs:  false,
		results:          make([]Result, 0),
		ctx:              ctx,
		cancel:           cancel,
		progressCallback: nil,
		networkProber:    networkProber,
		portScanner:      portScanner,
		udpPortScanner:   network.UDPPortScanner{Timeout: timeout},
		resultPresenter:  resultPresenter,
	}
}

// SetProgressCallback устанавливает callback для передачи прогресса
func (ns *NetworkScanner) SetProgressCallback(callback ProgressCallback) {
	ns.progressCallback = callback
}

// SetScanUDP включает или выключает UDP сканирование
func (ns *NetworkScanner) SetScanUDP(enable bool) {
	ns.scanUDP = enable
}

// SetScanTCPPorts включает или отключает перебор TCP-портов (при false выполняется только обнаружение хостов и сбор MAC/имени).
func (ns *NetworkScanner) SetScanTCPPorts(enable bool) {
	ns.scanTCPPorts = enable
}

// SetGrabBanners включает чтение баннеров с открытых портов (21,22,25,80,…).
func (ns *NetworkScanner) SetGrabBanners(enable bool) {
	ns.grabBanners = enable
}

// SetOSDetectActive включает расширенные (более смелые) эвристики определения ОС.
func (ns *NetworkScanner) SetOSDetectActive(enable bool) {
	ns.osDetectActive = enable
}

// SetVerbosePortLogs включает детальные логи по отдельным портам.
func (ns *NetworkScanner) SetVerbosePortLogs(enable bool) {
	ns.verbosePortLogs = enable
}

// Scan запускает сканирование сети
func (ns *NetworkScanner) Scan() {
	scanStartTime := time.Now()
	atomic.StoreInt64(&ns.tcpCancelBefore, 0)
	atomic.StoreInt64(&ns.tcpCancelWait, 0)
	atomic.StoreInt64(&ns.udpCancelHosts, 0)
	atomic.StoreInt64(&ns.tcpProbeTotal, 0)
	atomic.StoreInt64(&ns.tcpProbeOpen, 0)
	atomic.StoreInt64(&ns.tcpProbeClosed, 0)
	atomic.StoreInt64(&ns.udpProbeTotal, 0)
	atomic.StoreInt64(&ns.udpProbeOpen, 0)
	atomic.StoreInt64(&ns.udpProbeNoOpen, 0)
	atomic.StoreInt64(&ns.lastPingNs, 0)
	atomic.StoreInt64(&ns.lastPortscanNs, 0)
	atomic.StoreInt64(&ns.lastTotalNs, 0)
	fmt.Println("Начинаю сканирование сети...")
	logger.Log("Начинаю сканирование сети: %s", ns.network)
	logger.LogDebug("Параметры сканирования: сеть=%s, порты=%s, таймаут=%v, потоков=%d, showClosed=%v",
		ns.network, ns.portRange, ns.timeout, ns.threads, ns.showClosed)

	// Парсим диапазон сети
	parseStartTime := time.Now()
	ips, err := network.ParseNetworkRange(ns.network)
	if err != nil {
		logger.LogError(err, "Парсинг сети")
		fmt.Printf("Ошибка парсинга сети: %v\n", err)
		return
	}
	parseDuration := time.Since(parseStartTime)
	logger.LogDebug("Парсинг сети завершен: %d IP адресов за %v", len(ips), parseDuration)

	// Парсим диапазон портов
	var ports []int
	if ns.scanTCPPorts {
		var err error
		ports, err = network.ParsePortRange(ns.portRange)
		if err != nil {
			logger.LogError(err, "Парсинг портов")
			fmt.Printf("Ошибка парсинга портов: %v\n", err)
			return
		}
		logger.LogDebug("Парсинг портов завершен: %d портов", len(ports))
	} else {
		logger.LogDebug("TCP сканирование портов отключено")
	}

	fmt.Printf("Сканирование %d хостов, порты: %d\n", len(ips), len(ports))
	logger.Log("Сканирование %d хостов, порты: %d, таймаут: %v, потоков: %d", len(ips), len(ports), ns.timeout, ns.threads)

	// Создаем пул горутин для сканирования
	sem := make(chan struct{}, ns.threads)

	// Сначала проверяем доступность хостов (ping)
	pingStartTime := time.Now()
	fmt.Println("Проверка доступности хостов...")
	logger.Log("Начало проверки доступности хостов: %d хостов", len(ips))
	if ns.progressCallback != nil {
		ns.progressCallback("ping", 0, len(ips), "Проверка доступности хостов...")
	}
	aliveIPs := make([]net.IP, 0)
	aliveMutex := sync.Mutex{}
	checkedCount := 0
	checkedMutex := sync.Mutex{}

	cancelledDuringPing := false
	for _, ip := range ips {
		select {
		case <-ns.ctx.Done():
			cancelledDuringPing = true
			logger.LogDebug("Сканирование отменено во время проверки доступности (остановка запуска новых проверок)")
		default:
		}
		if cancelledDuringPing {
			break
		}

		sem <- struct{}{}
		ns.wg.Add(1)
		go func(ip net.IP) {
			defer func() { <-sem }()
			defer ns.wg.Done()

			hostCheckStart := time.Now()
			isAlive := ns.isHostAlive(ip.String())
			hostCheckDuration := time.Since(hostCheckStart)

			if isAlive {
				logger.LogDebug("Хост %s доступен (проверка заняла %v)", ip.String(), hostCheckDuration)
				aliveMutex.Lock()
				aliveIPs = append(aliveIPs, ip)
				aliveMutex.Unlock()
			} else {
				logger.LogDebug("Хост %s недоступен (проверка заняла %v)", ip.String(), hostCheckDuration)
			}

			// Обновляем счетчик прогресса
			checkedMutex.Lock()
			checkedCount++
			progress := checkedCount
			checkedMutex.Unlock()
			aliveMutex.Lock()
			aliveCount := len(aliveIPs)
			aliveMutex.Unlock()

			// Обновляем прогресс через callback и консоль
			if progress%10 == 0 || progress == len(ips) {
				fmt.Printf("\rПроверено хостов: %d/%d, найдено активных: %d", progress, len(ips), aliveCount)
				if ns.progressCallback != nil {
					ns.progressCallback("ping", progress, len(ips), fmt.Sprintf("Проверено хостов: %d/%d, найдено активных: %d", progress, len(ips), aliveCount))
				}
			}
		}(ip)
	}
	ns.wg.Wait()
	if cancelledDuringPing {
		logger.LogDebug("Сканирование остановлено после завершения активных проверок доступности")
		return
	}
	pingDuration := time.Since(pingStartTime)
	atomic.StoreInt64(&ns.lastPingNs, pingDuration.Nanoseconds())
	fmt.Println() // Новая строка после прогресса

	fmt.Printf("Найдено %d активных хостов\n", len(aliveIPs))
	logger.Log("Найдено активных хостов: %d из %d (проверка заняла %v)", len(aliveIPs), len(ips), pingDuration)
	// Логируем список активных хостов
	aliveIPsList := make([]string, len(aliveIPs))
	for i, ip := range aliveIPs {
		aliveIPsList[i] = ip.String()
	}
	logger.LogDebug("Список активных хостов (%d): %v", len(aliveIPs), aliveIPsList)
	if ns.progressCallback != nil {
		ns.progressCallback("ping", len(ips), len(ips), fmt.Sprintf("Найдено %d активных хостов", len(aliveIPs)))
	}

	// Сканируем порты на активных хостах
	portsScanDuration := time.Duration(0)
	if len(aliveIPs) > 0 {
		portsScanStartTime := time.Now()
		if len(ports) > 0 {
			fmt.Println("Сканирование портов...")
			logger.Log("Начало сканирования портов на %d хостах, портов на хост: %d", len(aliveIPs), len(ports))
		} else {
			fmt.Println("Сбор данных о хостах (TCP-порты не сканируются)...")
			logger.Log("Сбор данных о хостах на %d адресах без перебора TCP-портов", len(aliveIPs))
		}
		logger.LogDebug("Всего портов для сканирования: %d хостов × %d портов = %d проверок", len(aliveIPs), len(ports), len(aliveIPs)*len(ports))
		if ns.progressCallback != nil {
			ns.progressCallback("ports", 0, len(aliveIPs), "Сканирование портов...")
		}
		scannedCount := 0
		scannedMutex := sync.Mutex{}

		for _, ip := range aliveIPs {
			select {
			case <-ns.ctx.Done():
				logger.LogDebug("Сканирование отменено во время сканирования портов")
				return
			default:
			}

			sem <- struct{}{}
			ns.wg.Add(1)
			go func(ip net.IP) {
				defer func() { <-sem }()
				defer ns.wg.Done()

				hostScanStart := time.Now()
				ns.scanHost(ip, ports)
				hostScanDuration := time.Since(hostScanStart)
				logger.LogDebug("Сканирование хоста %s завершено за %v", ip.String(), hostScanDuration)

				// Обновляем счетчик прогресса
				scannedMutex.Lock()
				scannedCount++
				progress := scannedCount
				scannedMutex.Unlock()

				// Обновляем прогресс через callback и консоль (ограничиваем частоту для избежания блокировки UI)
				if progress%5 == 0 || progress == len(aliveIPs) {
					fmt.Printf("\rСканирование портов: %d/%d хостов", progress, len(aliveIPs))
					if ns.progressCallback != nil {
						// Вызываем callback в неблокирующем режиме
						select {
						case <-ns.ctx.Done():
							return
						default:
							ns.progressCallback("ports", progress, len(aliveIPs), fmt.Sprintf("Сканирование портов: %d/%d хостов", progress, len(aliveIPs)))
						}
					}
				}
			}(ip)
		}
		ns.wg.Wait()
		portsScanDuration = time.Since(portsScanStartTime)
		atomic.StoreInt64(&ns.lastPortscanNs, portsScanDuration.Nanoseconds())
		fmt.Println() // Новая строка после прогресса
		tcpCancelBefore := atomic.LoadInt64(&ns.tcpCancelBefore)
		tcpCancelWait := atomic.LoadInt64(&ns.tcpCancelWait)
		udpCancelHosts := atomic.LoadInt64(&ns.udpCancelHosts)
		if tcpCancelBefore > 0 || tcpCancelWait > 0 || udpCancelHosts > 0 {
			logger.LogDebug(
				"Агрегированная статистика отмен: TCP до запуска=%d, TCP при ожидании=%d, UDP хостов=%d",
				tcpCancelBefore,
				tcpCancelWait,
				udpCancelHosts,
			)
		}
		logger.Log("Сканирование портов завершено за %v", portsScanDuration)
	} else {
		logger.Log("Активные хосты не найдены, пропускаем сканирование портов")
	}

	totalDuration := time.Since(scanStartTime)
	atomic.StoreInt64(&ns.lastTotalNs, totalDuration.Nanoseconds())
	fmt.Println("Сканирование завершено")
	logger.Log("Сканирование завершено. Найдено устройств: %d (общее время: %v)", len(ns.results), totalDuration)
	logger.LogDebug("Статистика сканирования: хостов проверено=%d, активных хостов=%d, устройств найдено=%d",
		len(ips), len(aliveIPs), len(ns.results))
	logger.Log(
		"Диагностическая сводка: ping=%v, portscan=%v, total=%v; TCP probes total/open/closed=%d/%d/%d; UDP probes total/open/no-open=%d/%d/%d",
		pingDuration,
		portsScanDuration,
		totalDuration,
		atomic.LoadInt64(&ns.tcpProbeTotal),
		atomic.LoadInt64(&ns.tcpProbeOpen),
		atomic.LoadInt64(&ns.tcpProbeClosed),
		atomic.LoadInt64(&ns.udpProbeTotal),
		atomic.LoadInt64(&ns.udpProbeOpen),
		atomic.LoadInt64(&ns.udpProbeNoOpen),
	)
	if ns.progressCallback != nil {
		ns.progressCallback("complete", len(ns.results), len(ns.results), fmt.Sprintf("Сканирование завершено. Найдено устройств: %d", len(ns.results)))
	}
}

// GetDiagnosticsSummary returns condensed diagnostics for the last scan run.
func (ns *NetworkScanner) GetDiagnosticsSummary() string {
	pingDuration := time.Duration(atomic.LoadInt64(&ns.lastPingNs))
	portscanDuration := time.Duration(atomic.LoadInt64(&ns.lastPortscanNs))
	totalDuration := time.Duration(atomic.LoadInt64(&ns.lastTotalNs))
	return fmt.Sprintf(
		"Диагностика: ping=%v, portscan=%v, total=%v | TCP probes=%d/%d/%d (total/open/closed) | UDP probes=%d/%d/%d (total/open/no-open) | cancel TCP=%d/%d (pre/wait), UDP hosts=%d",
		pingDuration,
		portscanDuration,
		totalDuration,
		atomic.LoadInt64(&ns.tcpProbeTotal),
		atomic.LoadInt64(&ns.tcpProbeOpen),
		atomic.LoadInt64(&ns.tcpProbeClosed),
		atomic.LoadInt64(&ns.udpProbeTotal),
		atomic.LoadInt64(&ns.udpProbeOpen),
		atomic.LoadInt64(&ns.udpProbeNoOpen),
		atomic.LoadInt64(&ns.tcpCancelBefore),
		atomic.LoadInt64(&ns.tcpCancelWait),
		atomic.LoadInt64(&ns.udpCancelHosts),
	)
}

// isHostAlive проверяет, доступен ли хост
func (ns *NetworkScanner) isHostAlive(ip string) bool {
	if ns.networkProber != nil {
		if contextAwareProber, ok := ns.networkProber.(ContextNetworkProber); ok {
			isAlive, err := contextAwareProber.PingContext(ip, ns.ctx.Done())
			if err == nil {
				return isAlive
			}
			logger.LogDebug("Падение до встроенного пинга для %s из-за ошибки context prober: %v", ip, err)
		}
		isAlive, err := ns.networkProber.Ping(ip)
		if err == nil {
			return isAlive
		}
		logger.LogDebug("Падение до встроенного пинга для %s из-за ошибки prober: %v", ip, err)
	}

	// Быстрая проверка живости: запускаем probe по нескольким портам параллельно
	// и завершаем проверку сразу после первого успешного подключения.
	commonPorts := []string{"80", "443", "22", "135", "139", "445"}
	if len(commonPorts) != commonHostPorts {
		logger.LogDebug("commonHostPorts=%d, но commonPorts имеет %d элементов — рассинхрон", commonHostPorts, len(commonPorts))
	}
	if ns.verbosePortLogs {
		logger.LogDebug("Проверка доступности хоста %s через порты: %v", ip, commonPorts)
	}

	probeTimeout := ns.timeout / udpProbeTimeoutDivisor
	if probeTimeout < hostProbeTimeoutMin {
		probeTimeout = hostProbeTimeoutMin
	}
	if probeTimeout > hostProbeTimeoutMax {
		probeTimeout = hostProbeTimeoutMax
	}

	ctx, cancel := context.WithCancel(ns.ctx)
	defer cancel()
	results := make(chan bool, len(commonPorts))

	for _, port := range commonPorts {
		go func(port string) {
			select {
			case <-ctx.Done():
				results <- false
				return
			default:
			}

			portCheckStart := time.Now()
			dialer := &net.Dialer{Timeout: probeTimeout}
			conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(ip, port))
			portCheckDuration := time.Since(portCheckStart)

			if err == nil {
				if conn != nil {
					conn.Close()
				}
				if ns.verbosePortLogs {
					logger.LogDebug("Хост %s доступен через порт %s (проверка заняла %v)", ip, port, portCheckDuration)
				}
				results <- true
				cancel()
				return
			}
			if ns.verbosePortLogs {
				logger.LogDebug("Хост %s не отвечает на порт %s: %v (проверка заняла %v)", ip, port, err, portCheckDuration)
			}
			results <- false
		}(port)
	}

	for i := 0; i < len(commonPorts); i++ {
		select {
		case <-ns.ctx.Done():
			logger.LogDebug("Проверка хоста %s отменена", ip)
			return false
		case ok := <-results:
			if ok {
				return true
			}
		}
	}

	logger.LogDebug("Хост %s недоступен (ни один из проверенных портов не ответил)", ip)
	return false
}

// scanTCPPort checks a single TCP port using injected scanner when available.
func (ns *NetworkScanner) scanTCPPort(ip string, port int) bool {
	if ns.portScanner != nil {
		isOpen, err := ns.portScanner.ScanPort(ip, port, "tcp")
		if err == nil {
			return isOpen
		}
		logger.LogDebug("PortScanner вернул ошибку для %s:%d, fallback на IsPortOpen: %v", ip, port, err)
	}
	return network.IsPortOpen(ip, port, ns.timeout)
}

// scanUDPPort checks a single UDP port using injected UDP scanner when available.
func (ns *NetworkScanner) scanUDPPort(ip string, port int) bool {
	return ns.scanUDPPortWithTimeout(ip, port, ns.timeout)
}

func (ns *NetworkScanner) scanUDPPortWithTimeout(ip string, port int, timeout time.Duration) bool {
	if ns.udpPortScanner != nil && timeout == ns.timeout {
		isOpen, err := ns.udpPortScanner.ScanPort(ip, port, "udp")
		if err == nil {
			return isOpen
		}
		logger.LogDebug("UDP PortScanner вернул ошибку для %s:%d, fallback на IsUDPPortOpen: %v", ip, port, err)
	}
	if timeout <= 0 {
		timeout = ns.timeout
	}
	return network.IsUDPPortOpen(ip, port, timeout)
}

// checkARP проверяет наличие хоста через ARP (не используется в быстром сканировании)
func (ns *NetworkScanner) checkARP(ip string) bool {
	// Эта функция оставлена для будущих улучшений
	// ARP запросы слишком медленные для массового сканирования
	return false
}

// scanHost сканирует один хост
func (ns *NetworkScanner) scanHost(ip net.IP, ports []int) {
	ipStr := ip.String()
	logger.LogDebug("Сканирование хоста: %s, портов: %d", ipStr, len(ports))
	result := Result{
		IP:        ipStr,
		MAC:       "",
		Hostname:  "",
		Ports:     make([]PortInfo, 0),
		Protocols: make([]string, 0),
		IsAlive:   true,
	}

	// Получаем MAC адрес и hostname асинхронно, не блокируя сканирование портов
	// Запускаем в фоне и собираем результаты после сканирования портов
	macChan := make(chan string, 1)
	macErrChan := make(chan error, 1)
	hostnameChan := make(chan []string, 1)
	hostnameErrChan := make(chan error, 1)

	// Запускаем получение MAC в фоне
	go func() {
		macStartTime := time.Now()
		logger.LogDebug("Начало получения MAC адреса для хоста %s", ipStr)
		mac, err := ns.getMACAddress(ip)
		macDuration := time.Since(macStartTime)
		if err != nil {
			logger.LogDebug("Не удалось получить MAC адрес для %s: %v (заняло %v)", ipStr, err, macDuration)
			macErrChan <- err
			return
		}
		logger.LogDebug("MAC адрес для %s получен: %s (заняло %v)", ipStr, mac, macDuration)
		macChan <- mac
	}()

	// Запускаем получение hostname в фоне
	go func() {
		hostnameStartTime := time.Now()
		logger.LogDebug("Начало получения hostname для хоста %s", ipStr)
		hostname, err := net.LookupAddr(ipStr)
		hostnameDuration := time.Since(hostnameStartTime)
		if err != nil {
			logger.LogDebug("Не удалось получить hostname для %s: %v (заняло %v)", ipStr, err, hostnameDuration)
			hostnameErrChan <- err
			return
		}
		if len(hostname) > 0 {
			logger.LogDebug("Hostname для %s получен: %v (заняло %v)", ipStr, hostname, hostnameDuration)
		}
		hostnameChan <- hostname
	}()

	// Сканируем порты параллельно, но с динамическим ограничением.
	// Ранее здесь был фиксированный лимит 100 на хост, что при большом количестве
	// параллельных хостов раздувало общее число соединений и вызывало просадки.
	portThreads := ns.portThreadsForHost(len(ports))
	portSem := make(chan struct{}, portThreads)
	portResults := make(chan PortInfo, len(ports))
	portWg := sync.WaitGroup{}

	cancelledBeforeLaunch := false
	// Запускаем параллельное сканирование портов
	for _, port := range ports {
		// Если контекст отменен, не запускаем новые горутины
		if ns.ctx.Err() != nil {
			if !cancelledBeforeLaunch {
				atomic.AddInt64(&ns.tcpCancelBefore, 1)
				cancelledBeforeLaunch = true
			}
			break
		}

		portSem <- struct{}{}
		portWg.Add(1)
		go func(p int) {
			defer func() { <-portSem }()
			defer portWg.Done()

			// Проверяем контекст перед проверкой порта
			select {
			case <-ns.ctx.Done():
				return
			default:
			}

			portCheckStart := time.Now()
			atomic.AddInt64(&ns.tcpProbeTotal, 1)
			isOpen := ns.scanTCPPort(ipStr, p)
			portCheckDuration := time.Since(portCheckStart)
			if isOpen {
				atomic.AddInt64(&ns.tcpProbeOpen, 1)
			} else {
				atomic.AddInt64(&ns.tcpProbeClosed, 1)
			}

			if ns.verbosePortLogs {
				if isOpen {
					logger.LogDebug("Хост %s: порт %d/%s открыт (проверка заняла %v)", ipStr, p, "tcp", portCheckDuration)
				} else if ns.showClosed {
					logger.LogDebug("Хост %s: порт %d/%s закрыт (проверка заняла %v)", ipStr, p, "tcp", portCheckDuration)
				}
			}

			if isOpen || ns.showClosed {
				state := "open"
				if !isOpen {
					state = "closed"
				}

				portInfo := PortInfo{
					Port:     p,
					State:    state,
					Protocol: "tcp",
					Service:  network.GetServiceName(p),
				}
				if isOpen && ns.grabBanners && shouldGrabBannerPort(p) {
					bt := ns.timeout / bannerGrabTimeoutDivisor
					if bt < bannerGrabTimeoutMin {
						bt = bannerGrabTimeoutMin
					}
					if bt > bannerGrabTimeoutMax {
						bt = bannerGrabTimeoutMax
					}
					if b, err := banner.GrabTCP(ipStr, p, bt); err == nil && strings.TrimSpace(b) != "" {
						portInfo.Banner = b
						portInfo.Version = banner.ExtractVersionHint(p, b)
					} else {
						portInfo.Banner = "нет ответа"
						portInfo.Version = ""
					}
				}

				// Отправляем результат в канал
				select {
				case portResults <- portInfo:
				case <-ns.ctx.Done():
					return
				}

				if isOpen && ns.verbosePortLogs {
					logger.LogDebug("Хост %s: найден открытый порт %d (%s)", ipStr, p, portInfo.Service)
				}
			}
		}(port)
	}

	// Ждем завершения всех проверок портов в отдельной горутине
	portDone := make(chan struct{})
	go func() {
		portWg.Wait()
		close(portResults)
		close(portDone)
	}()

	// Собираем результаты портов
	openPorts := 0
	portsCollected := false
	cancelledWhileCollecting := false
	for !portsCollected {
		select {
		case portInfo, ok := <-portResults:
			if !ok {
				// Канал закрыт, все результаты собраны
				portsCollected = true
				break
			}
			result.Ports = append(result.Ports, portInfo)

			if portInfo.State == "open" {
				openPorts++
				// Определяем протоколы по открытым портам
				protocol := getProtocolFromPort(portInfo.Port)
				if protocol != "" {
					result.Protocols = appendIfNotExists(result.Protocols, protocol)
					logger.LogDebug("Хост %s: определен протокол %s по порту %d", ipStr, protocol, portInfo.Port)
				}
			}
		case <-ns.ctx.Done():
			// Отмена сканирования - ждем завершения горутин и собираем уже полученные результаты
			if !cancelledWhileCollecting {
				atomic.AddInt64(&ns.tcpCancelWait, 1)
				cancelledWhileCollecting = true
			}
			<-portDone
			// Собираем оставшиеся результаты
			for portInfo := range portResults {
				result.Ports = append(result.Ports, portInfo)
				if portInfo.State == "open" {
					openPorts++
					protocol := getProtocolFromPort(portInfo.Port)
					if protocol != "" {
						result.Protocols = appendIfNotExists(result.Protocols, protocol)
					}
				}
			}
			portsCollected = true
		}
	}

	// UDP сканирование (если включено)
	if ns.scanUDP {
		ns.scanHostUDP(ipStr, &result)
	}

	// Собираем результаты MAC и hostname (неблокирующе)
	select {
	case mac := <-macChan:
		result.MAC = mac
		result.DeviceVendor = getVendorFromMAC(mac)
		logger.LogDebug("Хост %s: MAC адрес установлен: %s, производитель: %s", ipStr, mac, result.DeviceVendor)
	case <-macErrChan:
		logger.LogDebug("Хост %s: MAC адрес не получен (ошибка или таймаут)", ipStr)
		// Игнорируем ошибку получения MAC
	case <-time.After(macTimeout):
		logger.LogDebug("Хост %s: MAC адрес не получен (таймаут ожидания)", ipStr)
		// Не ждем долго, если MAC еще не получен
	default:
		logger.LogDebug("Хост %s: MAC адрес еще не готов", ipStr)
		// Продолжаем без MAC, если он еще не готов
	}

	select {
	case hostname := <-hostnameChan:
		if len(hostname) > 0 {
			result.Hostname = hostname[0]
			logger.LogDebug("Хост %s: hostname установлен: %s", ipStr, hostname[0])
		}
	case <-hostnameErrChan:
		logger.LogDebug("Хост %s: hostname не получен (ошибка DNS)", ipStr)
		// Игнорируем ошибку DNS
	case <-time.After(hostnameTimeout):
		logger.LogDebug("Хост %s: hostname не получен (таймаут ожидания)", ipStr)
		// Не ждем долго, если hostname еще не получен
	default:
		logger.LogDebug("Хост %s: hostname еще не готов", ipStr)
		// Продолжаем без hostname, если он еще не готов
	}

	// Определяем тип устройства
	result.DeviceType = ns.detectDeviceType(result)

	openTCPPorts := make([]int, 0)
	for _, p := range result.Ports {
		if p.State == "open" && p.Protocol == "tcp" {
			openTCPPorts = append(openTCPPorts, p.Port)
		}
	}
	if osName, conf, reason := osdetect.GuessFromHostAndPorts(result.Hostname, openTCPPorts, ns.osDetectActive); osName != "" {
		result.GuessOS = osName
		result.GuessOSConfidence = conf
		result.GuessOSReason = reason
	}
	// SNMP определяем по уже собранным данным; активный probe используем только при необходимости
	// и с коротким таймаутом, чтобы не замедлять массовое сканирование.
	result.SNMPEnabled = hasOpenPort(result.Ports, snmpPort, "udp") || hasOpenPort(result.Ports, snmpPort, "tcp")
	if !result.SNMPEnabled {
		snmpProbeTimeout := ns.timeout
		if snmpProbeTimeout > snmpProbeTimeoutMax {
			snmpProbeTimeout = snmpProbeTimeoutMax
		}
		result.SNMPEnabled = ns.scanUDPPortWithTimeout(ipStr, snmpPort, snmpProbeTimeout)
	}
	logger.LogDebug("Хост %s: определен тип устройства: %s", ipStr, result.DeviceType)

	// Сохраняем результат
	ns.mu.Lock()
	ns.results = append(ns.results, result)
	ns.mu.Unlock()

	logger.LogDebug("Хост %s: найдено открытых портов: %d", ipStr, openPorts)
}

func (ns *NetworkScanner) portThreadsForHost(portCount int) int {
	if portCount <= 0 {
		return 0
	}

	hostWorkers := ns.threads
	if hostWorkers <= 0 {
		hostWorkers = 1
	}

	perHost := globalPortProbeBudget / hostWorkers
	if perHost < minPerHostPortThreads {
		perHost = minPerHostPortThreads
	}
	if perHost > maxPerHostPortThreads {
		perHost = maxPerHostPortThreads
	}
	if perHost > portCount {
		perHost = portCount
	}
	if perHost < 1 {
		perHost = 1
	}
	return perHost
}

// scanHostUDP сканирует известные UDP порты для указанного хоста
func (ns *NetworkScanner) scanHostUDP(ipStr string, result *Result) {
	logger.LogDebug("Начинаю UDP сканирование для хоста %s", ipStr)
	defer logger.LogDebug("UDP сканирование для хоста %s завершено", ipStr)

	udpPorts := []int{53, 67, 68, 69, 123, 161, 162, 514, 1194}
	udpSem := make(chan struct{}, udpSemaphoreSize)
	udpWg := sync.WaitGroup{}
	udpResults := make(chan PortInfo, udpResultBufferSize)
	udpDone := make(chan struct{})

	udpScanCancelled := false
udpPortLoop:
	for _, udpPort := range udpPorts {
		select {
		case <-ns.ctx.Done():
			atomic.AddInt64(&ns.udpCancelHosts, 1)
			udpScanCancelled = true
			break udpPortLoop
		default:
		}
		if udpScanCancelled {
			break udpPortLoop
		}

		udpSem <- struct{}{}
		udpWg.Add(1)
		go func(p int) {
			defer func() { <-udpSem }()
			defer udpWg.Done()

			select {
			case <-ns.ctx.Done():
				return
			default:
			}

			udpCheckStart := time.Now()
			atomic.AddInt64(&ns.udpProbeTotal, 1)
			isOpen := ns.scanUDPPort(ipStr, p)
			udpCheckDuration := time.Since(udpCheckStart)

			if isOpen {
				atomic.AddInt64(&ns.udpProbeOpen, 1)
				if ns.verbosePortLogs {
					logger.LogDebug("Хост %s: UDP порт %d открыт (проверка заняла %v)", ipStr, p, udpCheckDuration)
				}
				portInfo := PortInfo{
					Port:     p,
					State:    "open",
					Protocol: "udp",
					Service:  network.GetServiceName(p),
				}
				select {
				case udpResults <- portInfo:
				case <-ns.ctx.Done():
					return
				}
			} else if ns.showClosed {
				atomic.AddInt64(&ns.udpProbeNoOpen, 1)
				if ns.verbosePortLogs {
					logger.LogDebug("Хост %s: UDP порт %d закрыт/фильтруется (проверка заняла %v)", ipStr, p, udpCheckDuration)
				}
				portInfo := PortInfo{
					Port:     p,
					State:    "filtered",
					Protocol: "udp",
					Service:  network.GetServiceName(p),
				}
				select {
				case udpResults <- portInfo:
				case <-ns.ctx.Done():
					return
				}
			} else {
				atomic.AddInt64(&ns.udpProbeNoOpen, 1)
			}
		}(udpPort)
	}

	// Ждем завершения UDP сканирования
	go func() {
		udpWg.Wait()
		close(udpResults)
		close(udpDone)
	}()

	// Собираем UDP результаты
	udpCollected := false
	for !udpCollected {
		select {
		case udpPortInfo, ok := <-udpResults:
			if !ok {
				udpCollected = true
				break
			}
			result.Ports = append(result.Ports, udpPortInfo)
			if udpPortInfo.State == "open" {
				result.Protocols = appendIfNotExists(result.Protocols, getProtocolFromPort(udpPortInfo.Port))
				logger.LogDebug("Хост %s: определен протокол %s по UDP порту %d", ipStr, getProtocolFromPort(udpPortInfo.Port), udpPortInfo.Port)
			}
		case <-ns.ctx.Done():
			<-udpDone
			for udpPortInfo := range udpResults {
				result.Ports = append(result.Ports, udpPortInfo)
			}
			udpCollected = true
		case <-time.After(udpCollectTimeout):
			// Небольшая задержка для сбора результатов
		}
	}
}

// getMACAddress получает MAC адрес через ARP
func (ns *NetworkScanner) getMACAddress(ip net.IP) (string, error) {
	if ip == nil || ip.To4() == nil {
		return "", fmt.Errorf("MAC адрес через ARP доступен только для IPv4")
	}

	if ns.networkProber != nil {
		if hwAddr, err := ns.networkProber.ResolveMAC(ip.String()); err == nil && hwAddr != nil {
			return hwAddr.String(), nil
		}
	}

	// Сначала пытаемся прочитать из ARP таблицы системы (если доступно)
	mac, err := ns.readMACFromARPTable(ip)
	if err == nil {
		return mac, nil
	}

	// Если не получилось, пытаемся отправить ARP запрос через pcap
	// Это требует root прав на некоторых системах
	return ns.getMACViaARPRequest(ip)
}

// readMACFromARPTable читает MAC из системной ARP таблицы
func (ns *NetworkScanner) readMACFromARPTable(ip net.IP) (string, error) {
	ipStr := ip.String()

	switch runtime.GOOS {
	case "linux":
		return ns.readMACFromLinuxARP(ipStr)
	case "windows":
		return ns.readMACFromWindowsARP(ipStr)
	case "darwin":
		return ns.readMACFromDarwinARP(ipStr)
	default:
		return "", fmt.Errorf("платформа %s не поддерживается", runtime.GOOS)
	}
}

// readMACFromLinuxARP читает MAC адрес из /proc/net/arp на Linux
func (ns *NetworkScanner) readMACFromLinuxARP(ipStr string) (string, error) {
	file, err := os.Open("/proc/net/arp")
	if err != nil {
		return "", fmt.Errorf("не удалось открыть /proc/net/arp: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Пропускаем заголовок
	if scanner.Scan() {
		_ = scanner.Text()
	}

	for scanner.Scan() {
		// Проверяем контекст для возможности отмены
		select {
		case <-ns.ctx.Done():
			return "", fmt.Errorf("сканирование отменено")
		default:
		}

		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		// Формат: IP address HW type Flags HW address Mask Device
		// fields[0] = IP, fields[3] = HW address (MAC)
		if fields[0] == ipStr {
			mac := fields[3]
			// Проверяем, что это валидный MAC адрес (не "00:00:00:00:00:00")
			if mac != "00:00:00:00:00:00" && mac != "<incomplete>" {
				return mac, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("ошибка чтения /proc/net/arp: %v", err)
	}

	return "", fmt.Errorf("MAC адрес для %s не найден в ARP таблице", ipStr)
}

// readMACFromWindowsARP читает MAC адрес через команду arp -a на Windows
func (ns *NetworkScanner) readMACFromWindowsARP(ipStr string) (string, error) {
	// Создаем контекст с таймаутом для избежания зависания в Windows
	ctx, cancel := context.WithTimeout(ns.ctx, arpCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "arp", "-a", ipStr)
	output, err := cmd.Output()
	if err != nil {
		// Проверяем, не был ли это таймаут
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("таймаут выполнения arp -a")
		}
		return "", fmt.Errorf("ошибка выполнения arp -a: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// Формат Windows: "  192.168.1.1          00-11-22-33-44-55     dynamic"
		// Ищем строку с нужным IP
		if strings.Contains(line, ipStr) {
			fields := strings.Fields(line)
			for i, field := range fields {
				// Ищем поле, которое выглядит как MAC адрес (XX-XX-XX-XX-XX-XX)
				if strings.Contains(field, "-") && len(field) == 17 {
					// Конвертируем формат Windows (XX-XX-XX-XX-XX-XX) в стандартный (XX:XX:XX:XX:XX:XX)
					mac := strings.ReplaceAll(field, "-", ":")
					// Проверяем, что это не пустой MAC
					if mac != "00:00:00:00:00:00" {
						return mac, nil
					}
				}
				// Также проверяем формат с двоеточиями
				if strings.Contains(field, ":") && len(field) == 17 && i > 0 {
					// Проверяем, что предыдущее поле - это IP
					if i > 0 && fields[i-1] == ipStr {
						if field != "00:00:00:00:00:00" {
							return field, nil
						}
					}
				}
			}
		}
	}

	return "", fmt.Errorf("MAC адрес для %s не найден в ARP таблице", ipStr)
}

// readMACFromDarwinARP читает MAC адрес через команду arp -a на macOS
func (ns *NetworkScanner) readMACFromDarwinARP(ipStr string) (string, error) {
	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(ns.ctx, arpCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "arp", "-n", ipStr)
	output, err := cmd.Output()
	if err != nil {
		// Проверяем, не был ли это таймаут
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("таймаут выполнения arp -n")
		}
		return "", fmt.Errorf("ошибка выполнения arp -n: %v", err)
	}

	// Формат macOS: "? (192.168.1.1) at aa:bb:cc:dd:ee:ff on en0 ifscope [ethernet]"
	// или просто: "192.168.1.1 (192.168.1.1) at aa:bb:cc:dd:ee:ff on en0"
	outputStr := string(output)
	if strings.Contains(outputStr, "at ") {
		// Ищем MAC адрес после "at "
		parts := strings.Split(outputStr, "at ")
		if len(parts) > 1 {
			// Берем часть после "at " и извлекаем MAC
			macPart := strings.Fields(parts[1])[0]
			// Проверяем формат MAC адреса
			if strings.Contains(macPart, ":") && len(macPart) == 17 {
				if macPart != "00:00:00:00:00:00" && macPart != "(incomplete)" {
					return macPart, nil
				}
			}
		}
	}

	// Альтернативный способ: парсим весь вывод arp -a
	cmd = exec.CommandContext(ctx, "arp", "-a")
	output, err = cmd.Output()
	if err != nil {
		// Проверяем, не был ли это таймаут
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("таймаут выполнения arp -a")
		}
		return "", fmt.Errorf("ошибка выполнения arp -a: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, ipStr) && strings.Contains(line, "at ") {
			parts := strings.Split(line, "at ")
			if len(parts) > 1 {
				macPart := strings.Fields(parts[1])[0]
				if strings.Contains(macPart, ":") && len(macPart) == 17 {
					if macPart != "00:00:00:00:00:00" && macPart != "(incomplete)" {
						return macPart, nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("MAC адрес для %s не найден в ARP таблице", ipStr)
}

// getMACViaARPRequest отправляет ARP запрос для получения MAC
func (ns *NetworkScanner) getMACViaARPRequest(ip net.IP) (string, error) {
	// Получаем интерфейсы с таймаутом (избегаем зависания в Windows)
	interfacesChan := make(chan []net.Interface, 1)
	errChan := make(chan error, 1)
	go func() {
		interfaces, err := net.Interfaces()
		if err != nil {
			errChan <- err
			return
		}
		interfacesChan <- interfaces
	}()

	var interfaces []net.Interface
	select {
	case interfaces = <-interfacesChan:
		// Успешно получили интерфейсы
	case err := <-errChan:
		return "", err
	case <-time.After(ifaceTimeout):
		return "", fmt.Errorf("таймаут получения сетевых интерфейсов")
	case <-ns.ctx.Done():
		return "", fmt.Errorf("сканирование отменено")
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Получаем IP интерфейса с таймаутом (избегаем зависания в Windows)
		addrsChan := make(chan []net.Addr, 1)
		addrErrChan := make(chan error, 1)
		go func() {
			addrs, err := iface.Addrs()
			if err != nil {
				addrErrChan <- err
				return
			}
			addrsChan <- addrs
		}()

		var addrs []net.Addr
		select {
		case addrs = <-addrsChan:
			// Успешно получили адреса
		case <-addrErrChan:
			continue
		case <-time.After(ifaceAddrTimeout):
			// Таймаут для получения адресов интерфейса, пропускаем этот интерфейс
			continue
		case <-ns.ctx.Done():
			return "", fmt.Errorf("сканирование отменено")
		}

		var localIP net.IP
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				localIP = ipnet.IP
				break
			}
		}

		if localIP == nil {
			continue
		}

		// Пытаемся открыть интерфейс (может требовать root прав)
		handle, err := pcap.OpenLive(iface.Name, pcapBufferSize, true, pcap.BlockForever)
		if err != nil {
			// Если не получилось (нет прав), пропускаем
			continue
		}
		defer handle.Close()

		// Создаем ARP запрос
		eth := layers.Ethernet{
			SrcMAC:       iface.HardwareAddr,
			DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			EthernetType: layers.EthernetTypeARP,
		}

		arp := layers.ARP{
			AddrType:          layers.LinkTypeEthernet,
			Protocol:          layers.EthernetTypeIPv4,
			HwAddressSize:     6,
			ProtAddressSize:   4,
			Operation:         layers.ARPRequest,
			SourceHwAddress:   []byte(iface.HardwareAddr),
			SourceProtAddress: []byte(localIP.To4()),
			DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
			DstProtAddress:    []byte(ip.To4()),
		}

		buf := gopacket.NewSerializeBuffer()
		opts := gopacket.SerializeOptions{}
		if err := gopacket.SerializeLayers(buf, opts, &eth, &arp); err != nil {
			continue
		}

		// Отправляем пакет
		if err := handle.WritePacketData(buf.Bytes()); err != nil {
			continue
		}

		// Ждем ответ
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		timeout := time.After(arpResponseTimeout)
		for {
			select {
			case packet := <-packetSource.Packets():
				arpLayer := packet.Layer(layers.LayerTypeARP)
				if arpLayer != nil {
					arpResp, _ := arpLayer.(*layers.ARP)
					if arpResp != nil && arpResp.Operation == layers.ARPReply {
						srcIP := net.IP(arpResp.SourceProtAddress)
						if srcIP.Equal(ip) {
							return net.HardwareAddr(arpResp.SourceHwAddress).String(), nil
						}
					}
				}
			case <-timeout:
				return "", fmt.Errorf("таймаут ARP запроса")
			case <-ns.ctx.Done():
				return "", fmt.Errorf("сканирование отменено")
			}
		}
	}

	return "", fmt.Errorf("MAC адрес не найден")
}

// detectDeviceType определяет тип устройства по открытым портам, MAC и hostname
// Использует улучшенную эвристику с учетом производителя и комбинаций портов
func (ns *NetworkScanner) detectDeviceType(result Result) string {
	ports := make([]deviceclassifier.Port, 0, len(result.Ports))
	for _, p := range result.Ports {
		ports = append(ports, deviceclassifier.Port{
			Port:     p.Port,
			State:    p.State,
			Protocol: p.Protocol,
		})
	}
	return deviceclassifier.Classify(deviceclassifier.Input{
		Ports:        ports,
		DeviceVendor: result.DeviceVendor,
		Hostname:     result.Hostname,
	})
}

// Stop останавливает сканирование
func (ns *NetworkScanner) Stop() {
	ns.cancel()
	ns.wg.Wait()
}

// GetResults возвращает результаты сканирования
func (ns *NetworkScanner) GetResults() []Result {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	return ns.results
}

// Вспомогательные функции

// getProtocolFromPort определяет протокол по номеру порта (для списка «протоколы» на хосте).
func getProtocolFromPort(port int) string {
	return portdb.ProtocolLabel(port)
}

// getVendorFromMAC определяет производителя устройства по MAC адресу
// Использует проверку по OUI (первые 3 байта MAC адреса)
// Расширенная база популярных производителей
func getVendorFromMAC(mac string) string {
	if len(mac) < 8 {
		return "Unknown"
	}

	oui := mac[:8] // Берем первые 8 символов (XX:XX:XX)

	// Расширенная база известных OUI производителей
	vendors := map[string]string{
		// Виртуализация
		"00:50:56": "VMware",
		"00:0c:29": "VMware",
		"00:1c:42": "Parallels",
		"08:00:27": "VirtualBox",
		"52:54:00": "QEMU",
		"00:15:5d": "Microsoft Hyper-V",
		"00:03:ff": "Microsoft Hyper-V",

		// Apple
		"00:1b:21": "Apple",
		"00:23:12": "Apple",
		"00:25:00": "Apple",
		"00:25:4b": "Apple",
		"00:26:08": "Apple",
		"00:26:4a": "Apple",
		"00:26:bb": "Apple",
		"ac:de:48": "Apple",
		"a4:c1:38": "Apple",
		"a8:60:b6": "Apple",
		"c0:25:e9": "Apple",
		"d0:03:4b": "Apple",
		"e0:ac:cb": "Apple",
		"f0:db:e2": "Apple",
		"f4:f1:5a": "Apple",
		"f8:1e:df": "Apple",

		// Raspberry Pi
		"b8:27:eb": "Raspberry Pi",
		"dc:a6:32": "Raspberry Pi",
		"e4:5f:01": "Raspberry Pi",

		// Сетевые производители
		"00:1e:13": "Cisco",
		"00:1e:79": "Cisco",
		"00:26:ca": "Cisco",
		"00:50:f2": "Cisco",
		"00:90:0c": "Cisco",
		"00:90:21": "Cisco",
		"00:90:2b": "Cisco",
		"00:90:7f": "Cisco",
		"00:a0:40": "Cisco",
		"00:c0:4f": "Cisco",
		"00:e0:1e": "Cisco",
		"00:e0:f7": "Cisco",
		"00:e0:fe": "Cisco",
		"00:21:70": "Netgear",
		"00:24:b2": "Netgear",
		"00:09:5b": "Netgear",
		"00:1f:33": "Netgear",
		"00:0f:b5": "Belkin",
		"00:17:3f": "Belkin",
		"00:1e:c2": "Belkin",
		"00:22:3f": "Belkin",
		"00:1d:7e": "D-Link",
		"00:21:91": "D-Link",
		"00:24:01": "D-Link",
		"00:26:5a": "D-Link",
		"00:1b:11": "TP-Link",
		"00:27:19": "TP-Link",
		"00:50:fc": "TP-Link",
		"00:0c:41": "TP-Link",
		"00:1f:3a": "TP-Link",
		"00:21:6a": "TP-Link",
		"00:23:cd": "TP-Link",
		"00:25:86": "TP-Link",
		"00:27:22": "TP-Link",
		"00:0d:0b": "ASUS",
		"00:1d:60": "ASUS",
		"00:22:15": "ASUS",
		"00:24:8c": "ASUS",
		"00:26:18": "ASUS",
		"00:1e:8c": "ASUS",
		"00:11:2f": "Linksys",
		"00:13:10": "Linksys",
		"00:14:bf": "Linksys",
		"00:18:39": "Linksys",
		"00:1a:70": "Linksys",
		"00:1c:df": "Linksys",
		"00:21:29": "Linksys",
		"00:23:69": "Linksys",
		"00:25:9c": "Linksys",

		// Производители компьютеров
		"00:1e:68": "Dell",
		"00:14:22": "Dell",
		"00:0b:db": "Dell",
		"00:0d:56": "Dell",
		"00:1a:a0": "Dell",
		"00:1c:23": "Dell",
		"00:1e:c9": "Dell",
		"00:23:ae": "Dell",
		"00:0a:95": "HP",
		"00:0b:cd": "HP",
		"00:0e:7f": "HP",
		"00:11:85": "HP",
		"00:14:38": "HP",
		"00:17:a4": "HP",
		"00:1e:0b": "HP",
		"00:1f:29": "HP",
		"00:21:5a": "HP",
		"00:23:24": "HP",
		"00:25:b3": "HP",
		"00:26:55": "HP",
		"00:27:0e": "HP",
		"00:30:48": "HP",
		"00:50:8b": "HP",
		"00:21:cc": "Lenovo",
		"00:23:7d": "Lenovo",
		"00:25:64": "Lenovo",
		"00:1f:16": "Samsung",
		"00:23:39": "Samsung",
		"00:24:90": "Samsung",
		"00:26:5d": "Samsung",
		"00:15:99": "Samsung",
		"00:16:6c": "Samsung",
		"00:18:af": "Samsung",
		"00:1b:98": "Samsung",
		"00:1d:25": "Samsung",
		"00:1e:7d": "Samsung",
		"00:21:4c": "Samsung",
		"00:23:6c": "Samsung",
		"00:25:66": "Samsung",
		"00:26:e2": "Samsung",
		"00:13:a9": "Sony",
		"00:16:fe": "Sony",
		"00:19:c5": "Sony",
		"00:1a:80": "Sony",
		"00:1d:0d": "Sony",
		"00:1f:e4": "Sony",
		"00:21:9e": "Sony",
		"00:24:21": "Sony",
		"00:26:4c": "Sony",

		// Мобильные устройства
		"00:46:4b": "Huawei",
		"00:46:65": "Huawei",
		"00:46:cf": "Huawei",
		"00:25:9e": "Huawei",
		"00:26:43": "Huawei",
		"00:9e:c8": "Xiaomi",
		"28:e3:1f": "Xiaomi",
		"34:ce:00": "Xiaomi",
		"50:8f:4c": "Xiaomi",
		"64:09:80": "Xiaomi",
		"ac:c1:ee": "Xiaomi",
		"f0:b4:29": "Xiaomi",
		"fc:64:ba": "Xiaomi",
	}

	if vendor, ok := vendors[oui]; ok {
		return vendor
	}

	return "Unknown"
}

// appendIfNotExists добавляет элемент в слайс, если его там еще нет
func appendIfNotExists(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}

func hasOpenPort(ports []PortInfo, targetPort int, protocol string) bool {
	for _, p := range ports {
		if p.Port == targetPort && p.State == "open" {
			if protocol == "" || strings.EqualFold(p.Protocol, protocol) {
				return true
			}
		}
	}
	return false
}

func shouldGrabBannerPort(port int) bool {
	switch port {
	case 21, 22, 25, 110, 143, 587, 993, 995, 80, 443, 8080, 8443:
		return true
	default:
		return false
	}
}
