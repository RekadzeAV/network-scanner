package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	networkRange = flag.String("range", "", "Диапазон сети для сканирования (например: 192.168.1.0/24)")
	timeout      = flag.Duration("timeout", 3*time.Second, "Таймаут для сканирования")
	portRange    = flag.String("ports", "1-1000", "Диапазон портов для сканирования (например: 1-1000 или 80,443,8080)")
	threads      = flag.Int("threads", 100, "Количество потоков для сканирования")
	showClosed   = flag.Bool("show-closed", false, "Показывать закрытые порты")
)

func main() {
	flag.Parse()

	// Определяем сеть автоматически, если не указана
	network := *networkRange
	if network == "" {
		var err error
		network, err = detectLocalNetwork()
		if err != nil {
			log.Fatalf("Не удалось определить локальную сеть: %v\nИспользуйте флаг -range для указания сети вручную", err)
		}
		fmt.Printf("Автоматически определена сеть: %s\n", network)
	}

	// Обработка сигналов для корректного завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	scanner := NewNetworkScanner(network, *timeout, *portRange, *threads)
	
	// Запускаем сканирование в отдельной горутине
	done := make(chan bool)
	go func() {
		scanner.Scan()
		done <- true
	}()

	// Ждем завершения или сигнала
	select {
	case <-sigChan:
		fmt.Println("\nПолучен сигнал прерывания, завершение сканирования...")
		scanner.Stop()
		<-done
	case <-done:
	}

	// Выводим результаты
	results := scanner.GetResults()
	displayResults(results)
	
	// Выводим аналитику
	displayAnalytics(results)
}

