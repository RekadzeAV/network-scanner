package api

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// Router создаёт и настраивает маршрутизатор
type Router struct {
	router  *mux.Router
	config  Config
	handler *Handler
}

// NewRouter создаёт новый Router
func NewRouter(config Config) *Router {
	h := NewHandler(config)
	r := &Router{
		router:  mux.NewRouter(),
		config:  config,
		handler: h,
	}
	r.setupRoutes()
	return r
}

// GetRouter возвращает готовый router
func (r *Router) GetRouter() http.Handler {
	return r.router
}

// authRequiredMiddleware проверяет Bearer-токен аутентификации.
// Health и docs endpoints остаются публичными.
func (r *Router) authRequiredMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Публичные эндпоинты не требуют аутентификации
		if strings.HasPrefix(req.URL.Path, "/health") ||
			strings.HasPrefix(req.URL.Path, "/api/docs") {
			next.ServeHTTP(w, req)
			return
		}

		token := r.config.AuthToken
		if token == "" {
			// Auth отключена (development / tests)
			next.ServeHTTP(w, req)
			return
		}

		authHeader := req.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "Invalid Authorization header format; expected 'Bearer <token>'", http.StatusUnauthorized)
			return
		}

		if parts[1] != token {
			http.Error(w, "Invalid or unauthorized token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, req)
	})
}

func (r *Router) setupRoutes() {
	// API v1
	api := r.router.PathPrefix("/api/v1").Subrouter()

	// Middleware (Auth — первым)
	api.Use(r.authRequiredMiddleware)
	api.Use(r.handler.corsMiddleware)
	api.Use(r.handler.loggingMiddleware)
	api.Use(r.handler.rateLimitMiddleware(r.config.RateLimitPerSecond))

	// Routes
	api.HandleFunc("/scan", r.handler.handleScan).Methods("POST")
	api.HandleFunc("/scan/{id}", r.handler.handleScanStatus).Methods("GET")
	api.HandleFunc("/results", r.handler.handleResults).Methods("GET")
	api.HandleFunc("/inventory", r.handler.handleInventoryList).Methods("GET")
	api.HandleFunc("/inventory", r.handler.handleInventorySave).Methods("POST")
	// Diff двух снапшотов: канонический маршрут с двумя path-параметрами.
	api.HandleFunc("/inventory/{id_a}/diff/{id_b}", r.handler.handleInventoryDiff).Methods("GET")
	// Обратная совместимость: id_b можно передать query-параметром ?id_b=...
	api.HandleFunc("/inventory/{id_a}/diff", r.handler.handleInventoryDiff).Methods("GET")

	// History
	api.HandleFunc("/history", r.handler.historyHandler).Methods("GET")
	api.HandleFunc("/history/compare/{id_a}/{id_b}", r.handler.compareHandler).Methods("GET")

	// Alerts
	api.HandleFunc("/alerts", r.handler.alertsHandler).Methods("GET")
	api.HandleFunc("/alerts/check", r.handler.checkAlertsHandler).Methods("POST")
	api.HandleFunc("/alerts/clear", r.handler.clearAlertsHandler).Methods("DELETE")
	api.HandleFunc("/alerts/trigger/{id_a}/{id_b}", r.handler.triggerAlertHandler).Methods("POST")

	// SNMP
	api.HandleFunc("/snmp/collect", r.handler.snmpCollectHandler).Methods("POST")

	// Topology
	api.HandleFunc("/topology/build", r.handler.topologyBuildHandler).Methods("POST")
	api.HandleFunc("/topology/export/{format}", r.handler.topologyExportHandler).Methods("POST")
	api.HandleFunc("/topology/dot", r.handler.topologyDOTHandler).Methods("POST")
	api.HandleFunc("/topology/stats", r.handler.topologyStatsHandler).Methods("POST")

	// Health check
	r.router.HandleFunc("/health", r.handler.handleHealth).Methods("GET")

	// Swagger docs (placeholder)
	r.router.HandleFunc("/api/docs", r.handler.handleDocs).Methods("GET")
}
