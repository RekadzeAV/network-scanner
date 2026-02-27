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
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	"network-scanner/internal/logger"
	"network-scanner/internal/network"
)

// Result содержит результаты сканирования одного хоста
type Result struct {
	IP           string
	MAC          string
	Hostname     string
	Ports        []PortInfo
	Protocols    []string
	DeviceType   string
	DeviceVendor string
	IsAlive      bool
}

// PortInfo содержит информацию о порте
type PortInfo struct {
	Port     int
	State    string // "open", "closed", "filtered"
	Protocol string // "tcp", "udp"
	Service  string
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
	scanUDP          bool // Включить UDP сканирование
	results          []Result
	mu               sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	progressCallback ProgressCallback
}

// NewNetworkScanner создает новый сканер
func NewNetworkScanner(network string, timeout time.Duration, portRange string, threads int, showClosed bool) *NetworkScanner {
	ctx, cancel := context.WithCancel(context.Background())
	return &NetworkScanner{
		network:          network,
		timeout:          timeout,
		portRange:        portRange,
		threads:          threads,
		showClosed:       showClosed,
		scanUDP:          false, // По умолчанию UDP сканирование выключено
		results:          make([]Result, 0),
		ctx:              ctx,
		cancel:           cancel,
		progressCallback: nil,
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

// Scan запускает сканирование сети
func (ns *NetworkScanner) Scan() {
	scanStartTime := time.Now()
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
	ports, err := network.ParsePortRange(ns.portRange)
	if err != nil {
		logger.LogError(err, "Парсинг портов")
		fmt.Printf("Ошибка парсинга портов: %v\n", err)
		return
	}
	logger.LogDebug("Парсинг портов завершен: %d портов", len(ports))

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

	for _, ip := range ips {
		select {
		case <-ns.ctx.Done():
			logger.LogDebug("Сканирование отменено во время проверки доступности")
			return
		default:
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
			aliveCount := len(aliveIPs)
			checkedMutex.Unlock()

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
	pingDuration := time.Since(pingStartTime)
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
	if len(aliveIPs) > 0 {
		portsScanStartTime := time.Now()
		fmt.Println("Сканирование портов...")
		logger.Log("Начало сканирования портов на %d хостах, портов на хост: %d", len(aliveIPs), len(ports))
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
		portsScanDuration := time.Since(portsScanStartTime)
		fmt.Println() // Новая строка после прогресса
		logger.Log("Сканирование портов завершено за %v", portsScanDuration)
	} else {
		logger.Log("Активные хосты не найдены, пропускаем сканирование портов")
	}

	totalDuration := time.Since(scanStartTime)
	fmt.Println("Сканирование завершено")
	logger.Log("Сканирование завершено. Найдено устройств: %d (общее время: %v)", len(ns.results), totalDuration)
	logger.LogDebug("Статистика сканирования: хостов проверено=%d, активных хостов=%d, устройств найдено=%d",
		len(ips), len(aliveIPs), len(ns.results))
	if ns.progressCallback != nil {
		ns.progressCallback("complete", len(ns.results), len(ns.results), fmt.Sprintf("Сканирование завершено. Найдено устройств: %d", len(ns.results)))
	}
}

// isHostAlive проверяет, доступен ли хост
func (ns *NetworkScanner) isHostAlive(ip string) bool {
	// Используем TCP connect на несколько популярных портов
	commonPorts := []string{"80", "443", "22", "135", "139", "445"}
	logger.LogDebug("Проверка доступности хоста %s через порты: %v", ip, commonPorts)

	for _, port := range commonPorts {
		select {
		case <-ns.ctx.Done():
			logger.LogDebug("Проверка хоста %s отменена", ip)
			return false
		default:
		}

		portCheckStart := time.Now()
		// Используем Dialer с явным таймаутом для лучшей работы в Windows
		dialer := &net.Dialer{
			Timeout: ns.timeout,
		}
		conn, err := dialer.Dial("tcp", net.JoinHostPort(ip, port))
		portCheckDuration := time.Since(portCheckStart)

		if err == nil {
			if conn != nil {
				conn.Close()
			}
			logger.LogDebug("Хост %s доступен через порт %s (проверка заняла %v)", ip, port, portCheckDuration)
			return true
		} else {
			logger.LogDebug("Хост %s не отвечает на порт %s: %v (проверка заняла %v)", ip, port, err, portCheckDuration)
		}
	}

	// Если ни один порт не ответил, считаем хост недоступным
	// (ARP проверка слишком медленная для массового сканирования)
	logger.LogDebug("Хост %s недоступен (ни один из проверенных портов не ответил)", ip)
	return false
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

	// Сканируем порты параллельно
	// Используем пул горутин для ограничения одновременных проверок портов на хост
	// Это предотвращает перегрузку сети и системы
	portThreads := 100 // Количество одновременных проверок портов на один хост
	if portThreads > len(ports) {
		portThreads = len(ports)
	}
	portSem := make(chan struct{}, portThreads)
	portResults := make(chan PortInfo, len(ports))
	portWg := sync.WaitGroup{}

	// Запускаем параллельное сканирование портов
	for _, port := range ports {
		// Проверяем контекст перед запуском новой горутины
		select {
		case <-ns.ctx.Done():
			logger.LogDebug("Сканирование портов хоста %s отменено перед запуском проверок", ipStr)
			// Прерываем запуск новых горутин, но продолжаем собирать результаты уже запущенных
		default:
		}

		// Если контекст отменен, не запускаем новые горутины
		if ns.ctx.Err() != nil {
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
			isOpen := network.IsPortOpen(ipStr, p, ns.timeout)
			portCheckDuration := time.Since(portCheckStart)

			if isOpen {
				logger.LogDebug("Хост %s: порт %d/%s открыт (проверка заняла %v)", ipStr, p, "tcp", portCheckDuration)
			} else if ns.showClosed {
				logger.LogDebug("Хост %s: порт %d/%s закрыт (проверка заняла %v)", ipStr, p, "tcp", portCheckDuration)
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

				// Отправляем результат в канал
				select {
				case portResults <- portInfo:
				case <-ns.ctx.Done():
					return
				}

				if isOpen {
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
			logger.LogDebug("Сканирование портов хоста %s отменено, ожидание завершения...", ipStr)
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
		logger.LogDebug("Начинаю UDP сканирование для хоста %s", ipStr)
		udpPorts := []int{53, 67, 68, 69, 123, 161, 162, 514, 1194} // Известные UDP порты
		udpSem := make(chan struct{}, 50)                           // Ограничиваем параллельность UDP сканирования
		udpWg := sync.WaitGroup{}
		udpResults := make(chan PortInfo, len(udpPorts))
		udpDone := make(chan struct{})

		udpScanCancelled := false
	udpPortLoop:
		for _, udpPort := range udpPorts {
			// Проверяем, не отменено ли сканирование
			select {
			case <-ns.ctx.Done():
				logger.LogDebug("UDP сканирование хоста %s отменено", ipStr)
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
				isOpen := network.IsUDPPortOpen(ipStr, p, ns.timeout)
				udpCheckDuration := time.Since(udpCheckStart)

				if isOpen {
					logger.LogDebug("Хост %s: UDP порт %d открыт (проверка заняла %v)", ipStr, p, udpCheckDuration)
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
					logger.LogDebug("Хост %s: UDP порт %d закрыт/фильтруется (проверка заняла %v)", ipStr, p, udpCheckDuration)
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
					openPorts++
					protocol := getProtocolFromPort(udpPortInfo.Port)
					if protocol != "" {
						result.Protocols = appendIfNotExists(result.Protocols, protocol)
						logger.LogDebug("Хост %s: определен протокол %s по UDP порту %d", ipStr, protocol, udpPortInfo.Port)
					}
				}
			case <-ns.ctx.Done():
				<-udpDone
				for udpPortInfo := range udpResults {
					result.Ports = append(result.Ports, udpPortInfo)
					if udpPortInfo.State == "open" {
						openPorts++
						protocol := getProtocolFromPort(udpPortInfo.Port)
						if protocol != "" {
							result.Protocols = appendIfNotExists(result.Protocols, protocol)
						}
					}
				}
				udpCollected = true
			case <-time.After(100 * time.Millisecond):
				// Небольшая задержка для сбора результатов
			}
		}
		logger.LogDebug("UDP сканирование для хоста %s завершено", ipStr)
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
	case <-time.After(100 * time.Millisecond):
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
	case <-time.After(100 * time.Millisecond):
		logger.LogDebug("Хост %s: hostname не получен (таймаут ожидания)", ipStr)
		// Не ждем долго, если hostname еще не получен
	default:
		logger.LogDebug("Хост %s: hostname еще не готов", ipStr)
		// Продолжаем без hostname, если он еще не готов
	}

	// Определяем тип устройства
	result.DeviceType = ns.detectDeviceType(result)
	logger.LogDebug("Хост %s: определен тип устройства: %s", ipStr, result.DeviceType)

	// Сохраняем результат
	ns.mu.Lock()
	ns.results = append(ns.results, result)
	ns.mu.Unlock()

	logger.LogDebug("Хост %s: найдено открытых портов: %d", ipStr, openPorts)
}

// getMACAddress получает MAC адрес через ARP
func (ns *NetworkScanner) getMACAddress(ip net.IP) (string, error) {
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
	ctx, cancel := context.WithTimeout(ns.ctx, 3*time.Second)
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
	ctx, cancel := context.WithTimeout(ns.ctx, 3*time.Second)
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
	case <-time.After(3 * time.Second):
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
		case <-time.After(1 * time.Second):
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
		handle, err := pcap.OpenLive(iface.Name, 1024, true, pcap.BlockForever)
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
		timeout := time.After(2 * time.Second)
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
	// Анализируем открытые порты для определения типа устройства
	ports := make(map[int]bool)
	for _, p := range result.Ports {
		ports[p.Port] = true
	}

	// Определяем производителя для более точной классификации
	vendor := strings.ToLower(result.DeviceVendor)
	isNetworkVendor := strings.Contains(vendor, "cisco") || strings.Contains(vendor, "netgear") ||
		strings.Contains(vendor, "d-link") || strings.Contains(vendor, "tp-link") ||
		strings.Contains(vendor, "linksys") || strings.Contains(vendor, "asus") ||
		strings.Contains(vendor, "belkin")
	isRaspberryPi := strings.Contains(vendor, "raspberry")
	isVM := strings.Contains(vendor, "vmware") || strings.Contains(vendor, "virtualbox") ||
		strings.Contains(vendor, "qemu") || strings.Contains(vendor, "hyper-v") ||
		strings.Contains(vendor, "parallels")

	// Анализ hostname для дополнительной информации
	hostname := strings.ToLower(result.Hostname)
	isRouterHostname := strings.Contains(hostname, "router") || strings.Contains(hostname, "gateway") ||
		strings.Contains(hostname, "ap") || strings.Contains(hostname, "accesspoint") ||
		strings.Contains(hostname, "wifi") || strings.Contains(hostname, "wlan")

	// Принтер (высокий приоритет - специфичные порты)
	if ports[9100] || ports[515] || ports[631] || ports[161] {
		return "Printer"
	}

	// База данных (высокий приоритет - специфичные порты)
	if ports[3306] || ports[5432] || ports[1433] || ports[27017] || ports[6379] {
		return "Database Server"
	}

	// Роутер/сетевое оборудование
	// Комбинация веб-портов + SSH + сетевой производитель или router в hostname
	if (ports[80] || ports[443] || ports[8080]) && ports[22] {
		if isNetworkVendor || isRouterHostname {
			return "Router/Network Device"
		}
		// Если есть веб + SSH, но нет явных признаков сервера - вероятно роутер
		if !ports[3306] && !ports[5432] && !ports[1433] {
			return "Router/Network Device"
		}
	}

	// Сетевое оборудование по производителю и портам
	if isNetworkVendor && (ports[80] || ports[443] || ports[8080] || ports[22] || ports[23]) {
		return "Router/Network Device"
	}

	// Виртуальная машина
	if isVM {
		if ports[3306] || ports[5432] || ports[1433] {
			return "Virtual Machine (Database)"
		}
		if ports[80] || ports[443] || ports[8080] {
			return "Virtual Machine (Web Server)"
		}
		if ports[22] {
			return "Virtual Machine (Linux Server)"
		}
		return "Virtual Machine"
	}

	// Raspberry Pi
	if isRaspberryPi {
		if ports[22] {
			return "Raspberry Pi (Linux Server)"
		}
		if ports[80] || ports[443] {
			return "Raspberry Pi (Web Server)"
		}
		return "Raspberry Pi"
	}

	// Windows машина (RDP, SMB, NetBIOS)
	if ports[3389] || (ports[445] && ports[135]) || ports[139] {
		return "Windows Computer"
	}

	// Linux/Unix сервер (SSH + другие серверные порты)
	if ports[22] {
		if ports[80] || ports[443] || ports[8080] {
			return "Linux/Unix Server (Web)"
		}
		if ports[3306] || ports[5432] {
			return "Linux/Unix Server (Database)"
		}
		return "Linux/Unix Server"
	}

	// Веб-сервер (HTTP/HTTPS без SSH)
	if ports[80] || ports[443] || ports[8080] || ports[8443] {
		// Если только веб-порты и нет других серверных портов
		if !ports[22] && !ports[3306] && !ports[5432] {
			return "Web Server"
		}
	}

	// Медиа-сервер (DLNA, Plex, etc.)
	if ports[32400] || ports[8200] || ports[5000] || ports[1900] {
		return "Media Server"
	}

	// Игровая консоль или устройство
	if ports[3074] || ports[9308] || ports[3658] {
		return "Gaming Console"
	}

	// IoT устройство (мало портов, специфичные комбинации)
	if len(result.Ports) > 0 && len(result.Ports) < 3 {
		// Проверяем на типичные IoT порты
		if ports[1883] || ports[5683] || ports[8080] {
			return "IoT Device"
		}
		// Если только один нестандартный порт
		if len(result.Ports) == 1 {
			port := result.Ports[0].Port
			if port > 1024 && port < 65535 && !ports[80] && !ports[443] {
				return "IoT Device"
			}
		}
	}

	// Сетевое хранилище (NAS)
	if ports[2049] || ports[111] || (ports[445] && !ports[3389]) {
		return "Network Storage (NAS)"
	}

	// Если ничего не подошло
	return "Unknown Device"
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

// getProtocolFromPort определяет протокол по номеру порта
func getProtocolFromPort(port int) string {
	protocols := map[int]string{
		21:   "FTP",
		22:   "SSH",
		23:   "Telnet",
		25:   "SMTP",
		53:   "DNS",
		80:   "HTTP",
		110:  "POP3",
		143:  "IMAP",
		443:  "HTTPS",
		445:  "SMB",
		3306: "MySQL",
		3389: "RDP",
		5432: "PostgreSQL",
		5900: "VNC",
		8080: "HTTP",
		8443: "HTTPS",
	}
	return protocols[port]
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
