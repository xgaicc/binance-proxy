package health

import (
	"encoding/json"
	"net/http"
	"time"
)

type Handler struct {
	startTime time.Time
}

func NewHandler() *Handler {
	return &Handler{
		startTime: time.Now(),
	}
}

type HealthResponse struct {
	Status    string `json:"status"`
	Uptime    string `json:"uptime"`
	Timestamp string `json:"timestamp"`
}

func (h *Handler) Liveness(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status:    "ok",
		Uptime:    time.Since(h.startTime).Round(time.Second).String(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Readiness(w http.ResponseWriter, r *http.Request) {
	// Could add checks for Binance API connectivity here
	h.Liveness(w, r)
}
