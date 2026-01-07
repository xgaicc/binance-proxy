package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/xgaicc/binance-proxy/internal/health"
	"github.com/xgaicc/binance-proxy/internal/logging"
	"github.com/xgaicc/binance-proxy/internal/proxy/websocket"
	"github.com/xgaicc/binance-proxy/pkg/binance"
)

func NewRouter(
	restHandler *ProxyHandler,
	wsHandler *websocket.Handler,
	healthHandler *health.Handler,
	logger *logging.RequestLogger,
) *mux.Router {
	r := mux.NewRouter()

	// Health endpoints (no logging middleware)
	r.HandleFunc("/health", healthHandler.Liveness).Methods("GET")
	r.HandleFunc("/ready", healthHandler.Readiness).Methods("GET")

	// Spot API subrouter
	spotRouter := r.PathPrefix("/spot").Subrouter()
	spotRouter.Use(LoggingMiddleware(logger, string(binance.APITypeSpot)))

	// Spot WebSocket endpoints
	spotRouter.HandleFunc("/ws", wsHandler.HandleSpotWS)
	spotRouter.HandleFunc("/ws/{streams}", wsHandler.HandleSpotWS)
	spotRouter.HandleFunc("/stream", wsHandler.HandleSpotWS)

	// Spot REST API - catch all remaining paths
	spotRouter.PathPrefix("/").Handler(http.StripPrefix("/spot", restHandler.SpotHandler()))

	// Futures API subrouter
	futuresRouter := r.PathPrefix("/futures").Subrouter()
	futuresRouter.Use(LoggingMiddleware(logger, string(binance.APITypeFutures)))

	// Futures WebSocket endpoints
	futuresRouter.HandleFunc("/ws", wsHandler.HandleFuturesWS)
	futuresRouter.HandleFunc("/ws/{streams}", wsHandler.HandleFuturesWS)
	futuresRouter.HandleFunc("/stream", wsHandler.HandleFuturesWS)

	// Futures REST API - catch all remaining paths
	futuresRouter.PathPrefix("/").Handler(http.StripPrefix("/futures", restHandler.FuturesHandler()))

	return r
}
