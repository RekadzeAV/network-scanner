//go:build !gui_only && !unix

package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	"network-scanner/cmd/network-scanner/cmd"
	"network-scanner/internal/api"
)

// Build information - set via -ldflags during build
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Включаем Per-Monitor DPI awareness для корректного отображения на Windows
	SetProcessDPIAwareness()

	// Передаём build-информацию в CLI-слой (для подкоманды `version`).
	cmd.SetBuildInfo(Version, BuildTime, GitCommit)

	// Check for --version flag first
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("network-scanner version %s\n", Version)
		fmt.Printf("Build time: %s\n", BuildTime)
		fmt.Printf("Git commit: %s\n", GitCommit)
		os.Exit(0)
	}

	// Check for --api flag
	if len(os.Args) > 1 && os.Args[1] == "--api" {
		cfg := api.DefaultConfig()
		flags := resolveAPIFlags(os.Args)
		cfg.AuthToken = flags.token
		cfg.TLSCertFile = flags.tlsCertFile
		cfg.TLSKeyFile = flags.tlsKeyFile

		router := api.NewRouter(cfg)
		addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		fmt.Printf("Starting REST API server on %s\n", addr)
		if cfg.AuthToken == "" {
			fmt.Println("WARNING: API authentication disabled (set --api-token or NETWORK_SCANNER_API_TOKEN)")
		} else {
			fmt.Println("API authentication: enabled (Bearer token)")
		}

		// Таймауты обязательны (gosec G114): защищают от slowloris-подобного
		// исчерпания соединений; ReadHeaderTimeout ниже порога медленных клиентов.
		srv := &http.Server{
			Addr:              addr,
			Handler:           router.GetRouter(),
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      60 * time.Second,
			IdleTimeout:       120 * time.Second,
		}

		if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
			// P4-1: HTTPS с TLS 1.2+ (строгий минимум для защиты токена и payload).
			fmt.Println("TLS: enabled (minimum version 1.2)")
			srv.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
			if err := srv.ListenAndServeTLS(cfg.TLSCertFile, cfg.TLSKeyFile); err != nil {
				fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
				os.Exit(1)
			}
		} else {
			// HTTP: предупреждение о том, что токен и данные идут открытым текстом.
			fmt.Println("WARNING: API served over plain HTTP — token and data are not encrypted")
			if err := srv.ListenAndServe(); err != nil {
				fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
				os.Exit(1)
			}
		}
		return
	}

	// Справка и команды обслуживаются cobra. Для обратной совместимости
	// поддерживается вызов без подкоманды с флагами сканирования:
	//   network-scanner --network 192.168.1.0/24 --ports 1-1000
	// Такая форма используется smoke/closure-скриптами проекта и документацией.
	if len(os.Args) > 1 && startsWithDash(os.Args[1]) {
		cmd.ExecuteLegacyScan(os.Args[1:])
		return
	}

	cmd.ExecuteCLI()
}

// startsWithDash — является ли аргумент флагом (одиночный или двойной дефис).
func startsWithDash(s string) bool {
	return len(s) > 0 && s[0] == '-'
}
