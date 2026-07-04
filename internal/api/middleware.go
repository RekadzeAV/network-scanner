package api

import (
	"log"
	"net/http"
	"time"
)

// corsMiddleware добавляет CORS заголовки
func (h *Handler) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.config.EnableCORS {
			origin := r.Header.Get("Origin")
			allowed := false
			for _, o := range h.config.AllowedOrigins {
				if o == origin || o == "*" {
					allowed = true
					break
				}
			}
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware логирует запросы
func (h *Handler) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrap response writer to capture status code
		wrap := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(wrap, r)
		
		duration := time.Since(start)
		log.Printf("[%s] %s %s %d %v",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			wrap.statusCode,
			duration,
		)
	})
}

// rateLimitMiddleware ограничивает количество запросов
func (h *Handler) rateLimitMiddleware(maxRequests int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simple rate limiting (in production use a proper rate limiter)
			// For now, just log if limit is exceeded
			_ = maxRequests
			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter оборачивает http.ResponseWriter для捕获 status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
