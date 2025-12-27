package scanner

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	"network-scanner/internal/network"
)

// Result содержит результаты сканирования одного хоста
type Result struct {
	IP        string
	MAC       string
	Hostname  string
	Ports     []PortInfo
	Protocols []string
	DeviceType string
	DeviceVendor string
	IsAlive   bool
}

// PortInfo содержит информацию о порте
type PortInfo struct {
	Port     int
	State    string // "open", "closed", "filtered"
	Protocol string // "tcp", "udp"
	Service  string
}

// NetworkScanner выполняет сканирование сети
type NetworkScanner struct {
	network    string
	timeout    time.Duration
	portRange  string
	threads    int
	showClosed bool
	results    []Result
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// NewNetworkScanner создает новый сканер
func NewNetworkScanner(network string, timeout time.Duration, portRange string, threads int, showClosed bool) *NetworkScanner {
	ctx, cancel := context.WithCancel(context.Background())
	return &NetworkScanner{
		network:    network,
		timeout:    timeout,
		portRange:  portRange,
		threads:    threads,
		showClosed: showClosed,
		results:    make([]Result, 0),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Scan запускает сканирование сети
func (ns *NetworkScanner) Scan() {
	fmt.Println("Начинаю сканирование сети...")
	
	// Парсим диапазон сети
	ips, err := network.ParseNetworkRange(ns.network)
	if err != nil {
		fmt.Printf("Ошибка парсинга сети: %v\n", err)
		return
	}

	// Парсим диапазон портов
	ports, err := network.ParsePortRange(ns.portRange)
	if err != nil {
		fmt.Printf("Ошибка парсинга портов: %v\n", err)
		return
	}

	fmt.Printf("Сканирование %d хостов, порты: %d\n", len(ips), len(ports))

	// Создаем пул горутин для сканирования
	sem := make(chan struct{}, ns.threads)
	
	// Сначала проверяем доступность хостов (ping)
	fmt.Println("Проверка доступности хостов...")
	aliveIPs := make([]net.IP, 0)
	aliveMutex := sync.Mutex{}
	checkedCount := 0
	checkedMutex := sync.Mutex{}
	
	for _, ip := range ips {
		select {
		case <-ns.ctx.Done():
			return
		default:
		}

		sem <- struct{}{}
		ns.wg.Add(1)
		go func(ip net.IP) {
			defer func() { <-sem }()
			defer ns.wg.Done()

			if ns.isHostAlive(ip.String()) {
				aliveMutex.Lock()
				aliveIPs = append(aliveIPs, ip)
				aliveMutex.Unlock()
			}
			
			// Обновляем счетчик прогресса
			checkedMutex.Lock()
			checkedCount++
			if checkedCount%10 == 0 || checkedCount == len(ips) {
				fmt.Printf("\rПроверено хостов: %d/%d, найдено активных: %d", checkedCount, len(ips), len(aliveIPs))
			}
			checkedMutex.Unlock()
		}(ip)
	}
	ns.wg.Wait()
	fmt.Println() // Новая строка после прогресса

	fmt.Printf("Найдено %d активных хостов\n", len(aliveIPs))

	// Сканируем порты на активных хостах
	if len(aliveIPs) > 0 {
		fmt.Println("Сканирование портов...")
		scannedCount := 0
		scannedMutex := sync.Mutex{}
		
		for _, ip := range aliveIPs {
			select {
			case <-ns.ctx.Done():
				return
			default:
			}

			sem <- struct{}{}
			ns.wg.Add(1)
			go func(ip net.IP) {
				defer func() { <-sem }()
				defer ns.wg.Done()

				ns.scanHost(ip, ports)
				
				// Обновляем счетчик прогресса
				scannedMutex.Lock()
				scannedCount++
				fmt.Printf("\rСканирование портов: %d/%d хостов", scannedCount, len(aliveIPs))
				scannedMutex.Unlock()
			}(ip)
		}
		ns.wg.Wait()
		fmt.Println() // Новая строка после прогресса
	}

	fmt.Println("Сканирование завершено")
}

// isHostAlive проверяет, доступен ли хост
func (ns *NetworkScanner) isHostAlive(ip string) bool {
	// Используем TCP connect на несколько популярных портов
	commonPorts := []string{"80", "443", "22", "135", "139", "445"}
	
	for _, port := range commonPorts {
		select {
		case <-ns.ctx.Done():
			return false
		default:
		}
		
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, port), ns.timeout)
		if err == nil {
			conn.Close()
			return true
		}
	}
	
	// Если ни один порт не ответил, считаем хост недоступным
	// (ARP проверка слишком медленная для массового сканирования)
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
	result := Result{
		IP:        ipStr,
		MAC:       "",
		Hostname:  "",
		Ports:     make([]PortInfo, 0),
		Protocols: make([]string, 0),
		IsAlive:   true,
	}

	// Получаем MAC адрес
	mac, err := ns.getMACAddress(ip)
	if err == nil {
		result.MAC = mac
		result.DeviceVendor = getVendorFromMAC(mac)
	}

	// Получаем hostname
	hostname, err := net.LookupAddr(ipStr)
	if err == nil && len(hostname) > 0 {
		result.Hostname = hostname[0]
	}

	// Сканируем порты
	openPorts := 0
	for _, port := range ports {
		select {
		case <-ns.ctx.Done():
			return
		default:
		}

		isOpen := network.IsPortOpen(ipStr, port, ns.timeout)
		if isOpen || ns.showClosed {
			state := "open"
			if !isOpen {
				state = "closed"
			}
			
			portInfo := PortInfo{
				Port:     port,
				State:    state,
				Protocol: "tcp",
				Service:  network.GetServiceName(port),
			}
			result.Ports = append(result.Ports, portInfo)
			
			if isOpen {
				openPorts++
				// Определяем протоколы по открытым портам
				protocol := getProtocolFromPort(port)
				if protocol != "" {
					result.Protocols = appendIfNotExists(result.Protocols, protocol)
				}
			}
		}
	}

	// Определяем тип устройства
	result.DeviceType = ns.detectDeviceType(result)

	// Сохраняем результат
	ns.mu.Lock()
	ns.results = append(ns.results, result)
	ns.mu.Unlock()
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
	// Это будет реализовано через чтение системных файлов
	// Для кроссплатформенности используем упрощенный подход
	// В реальности нужно читать /proc/net/arp на Linux, arp -a на других системах
	return "", fmt.Errorf("ARP таблица недоступна")
}

// getMACViaARPRequest отправляет ARP запрос для получения MAC
func (ns *NetworkScanner) getMACViaARPRequest(ip net.IP) (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Получаем IP интерфейса
		addrs, err := iface.Addrs()
		if err != nil {
			continue
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

// detectDeviceType определяет тип устройства по открытым портам и MAC
func (ns *NetworkScanner) detectDeviceType(result Result) string {
	// Анализируем открытые порты для определения типа устройства
	ports := make(map[int]bool)
	for _, p := range result.Ports {
		ports[p.Port] = true
	}

	// Роутер/сетевое оборудование
	if ports[80] || ports[443] || ports[8080] {
		if ports[22] {
			return "Router/Network Device"
		}
	}

	// Веб-сервер
	if ports[80] || ports[443] || ports[8080] || ports[8443] {
		return "Web Server"
	}

	// База данных
	if ports[3306] || ports[5432] || ports[1433] {
		return "Database Server"
	}

	// Windows машина
	if ports[3389] || ports[445] {
		return "Windows Computer"
	}

	// Linux/Unix сервер
	if ports[22] {
		return "Linux/Unix Server"
	}

	// Принтер
	if ports[9100] || ports[515] {
		return "Printer"
	}

	// IoT устройство
	if len(result.Ports) > 0 && len(result.Ports) < 3 {
		return "IoT Device"
	}

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

func getVendorFromMAC(mac string) string {
	// Упрощенная проверка по OUI (первые 3 байта MAC)
	// В реальности нужна база данных OUI
	if len(mac) < 8 {
		return "Unknown"
	}
	
	oui := mac[:8] // Берем первые 8 символов (XX:XX:XX)
	
	// Небольшая база известных OUI
	vendors := map[string]string{
		"00:50:56": "VMware",
		"00:0c:29": "VMware",
		"00:1c:42": "Parallels",
		"08:00:27": "VirtualBox",
		"52:54:00": "QEMU",
		"00:1b:21": "Apple",
		"00:23:12": "Apple",
		"ac:de:48": "Apple",
		"b8:27:eb": "Raspberry Pi",
		"dc:a6:32": "Raspberry Pi",
	}
	
	if vendor, ok := vendors[oui]; ok {
		return vendor
	}
	
	return "Unknown"
}

func appendIfNotExists(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}

