package config

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareIncrementHits(t *testing.T) {
	cfg := &ApiConfig{}
	handler := cfg.MiddlewareIncrementHits(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if cfg.FileserverHits.Load() != 1 {
		t.Errorf("Expected hits 1, got %d", cfg.FileserverHits.Load())
	}

	// Call again
	handler.ServeHTTP(w, req)
	if cfg.FileserverHits.Load() != 2 {
		t.Errorf("Expected hits 2, got %d", cfg.FileserverHits.Load())
	}
}
