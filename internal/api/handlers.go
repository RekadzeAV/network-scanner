package api

import (
	"encoding/json"
	"net/http"
	"time"
)

// Handler оборачивает HTTP handler с общей логикой
type Handler struct {
	config Config
}

// NewHandler создаёт новый Handler
func NewHandler(config Config) *Handler {
	return &Handler{config: config}
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

// handleDocs placeholder для swagger docs
func (h *Handler) handleDocs(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"docs": "Swagger documentation will be available here",
		"endpoints": []string{
			"POST /api/v1/scan - Запустить сканирование",
			"GET /api/v1/scan/{id} - Статус сканирования",
			"GET /api/v1/results - Получить результаты",
			"GET /api/v1/inventory - Список снапшотов",
			"POST /api/v1/inventory - Сохранить снапшот",
			"GET /api/v1/inventory/{id}/diff - Сравнить снапшоты",
		},
	})
}
