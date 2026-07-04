package api

import (
	"net/http"

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

func (r *Router) setupRoutes() {
	// API v1
	api := r.router.PathPrefix("/api/v1").Subrouter()

	// Middleware
	api.Use(r.handler.corsMiddleware)
	api.Use(r.handler.loggingMiddleware)
	api.Use(r.handler.rateLimitMiddleware(r.config.RateLimitPerSecond))

	// Routes
	api.HandleFunc("/scan", r.handler.handleScan).Methods("POST")
	api.HandleFunc("/scan/{id}", r.handler.handleScanStatus).Methods("GET")
	api.HandleFunc("/results", r.handler.handleResults).Methods("GET")
	api.HandleFunc("/inventory", r.handler.handleInventoryList).Methods("GET")
	api.HandleFunc("/inventory", r.handler.handleInventorySave).Methods("POST")
	api.HandleFunc("/inventory/{id}/diff", r.handler.handleInventoryDiff).Methods("GET")

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
