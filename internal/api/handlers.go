package api

import (
	"encoding/json"
	"net/http"
	"time"

	"network-scanner/internal/contracts"
	"network-scanner/internal/inventory"
)

// ScanDeps зависимости для сканирования (dependency injection)
type ScanDeps struct {
	ScannerService contracts.ScannerService
	InventoryStore *inventory.Store
}

// Handler оборачивает HTTP handler с общей логикой
type Handler struct {
	config Config
	deps   ScanDeps
}

// NewHandler создаёт новый Handler
func NewHandler(config Config) *Handler {
	return &Handler{config: config}
}

// NewHandlerWithDeps создаёт новый Handler с зависимостями
func NewHandlerWithDeps(config Config, deps ScanDeps) *Handler {
	return &Handler{config: config, deps: deps}
}

// writeJSON записывает JSON ответ
func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError записывает ошибку
func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{
		"error": message,
	})
}

// handleHealth health check endpoint
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	})
}

// handleDocs возвращает Swagger OpenAPI спецификацию в YAML формате.
// Swagger-файл находится в docs/swagger.yaml и монтируется через embed.
func (h *Handler) handleDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/x-yaml")
	w.WriteHeader(http.StatusOK)
	data, err := swaggerSpecBytes()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to load swagger spec")
		return
	}
	_, _ = w.Write(data)
}
