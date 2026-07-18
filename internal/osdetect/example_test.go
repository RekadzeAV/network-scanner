package osdetect_test

import (
	"fmt"

	"network-scanner/internal/osdetect"
)

// Example_guessFromHostAndPorts демонстрирует эвристику определения ОС по имени хоста и портам.
func Example_guessFromHostAndPorts() {
	// Определение по имени хоста
	osName, conf, reason := osdetect.GuessFromHostAndPorts("my-windows-pc", []int{80}, false)
	fmt.Printf("%s (%s): %s\n", osName, conf, reason)

	// Определение по открытым портам (Windows)
	osName, conf, reason = osdetect.GuessFromHostAndPorts("", []int{135, 445}, false)
	fmt.Printf("%s (%s): %s\n", osName, conf, reason)

	// Определение по открытым портам (Linux/Unix — SSH + HTTPS)
	osName, conf, reason = osdetect.GuessFromHostAndPorts("server.local", []int{22, 443}, false)
	fmt.Printf("%s (%s): %s\n", osName, conf, reason)

	// Output:
	// Windows (низкая): hostname содержит win+pc/desktop
	// Windows (средняя): открыты порты 135 и 445
	// Linux/Unix или сетевое устройство (низкая): открыты порты 22 и 443
}

// Example_guessFromHostAndPorts_activeMode демонстрирует расширенные эвристики в active-режиме.
func Example_guessFromHostAndPorts_activeMode() {
	// Kubernetes cluster node
	osName, conf, reason := osdetect.GuessFromHostAndPorts("k8s-master", []int{22, 6443}, true)
	fmt.Printf("%s (%s): %s\n", osName, conf, reason)

	// macOS device
	osName, conf, reason = osdetect.GuessFromHostAndPorts("", []int{5353, 62078}, true)
	fmt.Printf("%s (%s): %s\n", osName, conf, reason)

	// Android device — hostname срабатывает раньше port-эвристики
	osName, conf, reason = osdetect.GuessFromHostAndPorts("android-phone", []int{5555, 8081}, true)
	fmt.Printf("%s (%s): %s\n", osName, conf, reason)

	// Non-detectable
	osName, conf, reason = osdetect.GuessFromHostAndPorts("unknown-device", []int{8080}, true)
	fmt.Printf("'%s' (нет данных)\n", osName)

	// Output:
	// Linux/Unix Server (средняя): active-эвристика: SSH + Kubernetes API
	// Apple iOS/macOS (высокая): active-эвристика: mDNS + Apple service ports
	// Android (средняя): hostname содержит android
	// '' (нет данных)
}
