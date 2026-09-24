package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestAuthRequiredMiddleware_HealthUnauthenticated(t *testing.T) {
	cfg := Config{AuthToken: "secret123"}
	h := NewHandler(cfg)
	r := &Router{router: mux.NewRouter(), config: cfg, handler: h}

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	r.authRequiredMiddleware(next).ServeHTTP(w, req)

	if !called {
		t.Error("next handler should be called for /health without auth")
	}
}

func TestAuthRequiredMiddleware_MissingHeader(t *testing.T) {
	cfg := Config{AuthToken: "secret123"}
	h := NewHandler(cfg)
	r := &Router{router: mux.NewRouter(), config: cfg, handler: h}

	next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		t.Error("next handler should not be called without Authorization header")
	})

	req := httptest.NewRequest("POST", "/api/v1/scan", nil)
	w := httptest.NewRecorder()

	r.authRequiredMiddleware(next).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthRequiredMiddleware_ValidToken(t *testing.T) {
	cfg := Config{AuthToken: "secret123"}
	h := NewHandler(cfg)
	r := &Router{router: mux.NewRouter(), config: cfg, handler: h}

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("POST", "/api/v1/scan", nil)
	req.Header.Set("Authorization", "Bearer secret123")
	w := httptest.NewRecorder()

	r.authRequiredMiddleware(next).ServeHTTP(w, req)

	if !called {
		t.Error("next handler should be called with valid token")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAuthRequiredMiddleware_InvalidToken(t *testing.T) {
	cfg := Config{AuthToken: "secret123"}
	h := NewHandler(cfg)
	r := &Router{router: mux.NewRouter(), config: cfg, handler: h}

	next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		t.Error("next handler should not be called with invalid token")
	})

	req := httptest.NewRequest("POST", "/api/v1/scan", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	w := httptest.NewRecorder()

	r.authRequiredMiddleware(next).ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestAuthRequiredMiddleware_AuthDisabled(t *testing.T) {
	cfg := Config{AuthToken: ""}
	h := NewHandler(cfg)
	r := &Router{router: mux.NewRouter(), config: cfg, handler: h}

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("POST", "/api/v1/scan", nil)
	w := httptest.NewRecorder()

	r.authRequiredMiddleware(next).ServeHTTP(w, req)

	if !called {
		t.Error("next handler should be called when auth is disabled")
	}
}

func TestAuthRequiredMiddleware_DocsUnauthenticated(t *testing.T) {
	cfg := Config{AuthToken: "secret123"}
	h := NewHandler(cfg)
	r := &Router{router: mux.NewRouter(), config: cfg, handler: h}

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/docs", nil)
	w := httptest.NewRecorder()

	r.authRequiredMiddleware(next).ServeHTTP(w, req)

	if !called {
		t.Error("next handler should be called for /api/docs without auth")
	}
}

func TestAuthRequiredMiddleware_InvalidHeaderFormat(t *testing.T) {
	cfg := Config{AuthToken: "secret123"}
	h := NewHandler(cfg)
	r := &Router{router: mux.NewRouter(), config: cfg, handler: h}

	next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		t.Error("next handler should not be called with invalid header format")
	})

	req := httptest.NewRequest("POST", "/api/v1/scan", nil)
	req.Header.Set("Authorization", "NotBearer token")
	w := httptest.NewRecorder()

	r.authRequiredMiddleware(next).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}