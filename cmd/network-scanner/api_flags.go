//go:build !gui_only

package main

import (
	"strings"
)

// apiFlags — результат парсинга API-флагов (--api).
type apiFlags struct {
	token       string
	tlsCertFile string
	tlsKeyFile  string
}

// resolveAPIFlags разбирает флаги --api-token, --api-tls-cert, --api-tls-key
// из аргументов командной строки.
func resolveAPIFlags(args []string) apiFlags {
	f := apiFlags{token: resolveAPIToken(args)}

	for i, arg := range args {
		switch {
		case strings.HasPrefix(arg, "--api-tls-cert="):
			f.tlsCertFile = strings.TrimPrefix(arg, "--api-tls-cert=")
		case arg == "--api-tls-cert" && i+1 < len(args):
			f.tlsCertFile = args[i+1]
		case strings.HasPrefix(arg, "--api-tls-key="):
			f.tlsKeyFile = strings.TrimPrefix(arg, "--api-tls-key=")
		case arg == "--api-tls-key" && i+1 < len(args):
			f.tlsKeyFile = args[i+1]
		}
	}

	return f
}
