package api

import "embed"

//go:embed swagger.yaml
var swaggerFile embed.FS

// swaggerSpecBytes возвращает содержимое swagger.yaml.
func swaggerSpecBytes() ([]byte, error) {
	return swaggerFile.ReadFile("swagger.yaml")
}
