package handler

import (
	"encoding/json"
	"net/http"

	"github.com/manoskammas/finance-insights/apps/api/internal/service"
)

// Health serves the liveness endpoint.
type Health struct {
	System *service.System
}

// Get handles GET /health.
func (h *Health) Get(w http.ResponseWriter, r *http.Request) {
	status := h.System.Health()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(status)
}
