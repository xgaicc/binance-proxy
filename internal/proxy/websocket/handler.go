package websocket

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/xgaicc/binance-proxy/internal/config"
	"github.com/xgaicc/binance-proxy/internal/logging"
	"github.com/xgaicc/binance-proxy/pkg/binance"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Handler struct {
	spotWSURL    string
	futuresWSURL string
	logger       *logging.RequestLogger
}

func NewHandler(cfg *config.Config, logger *logging.RequestLogger) *Handler {
	return &Handler{
		spotWSURL:    cfg.Binance.Spot.WebSocketURL,
		futuresWSURL: cfg.Binance.Futures.WebSocketURL,
		logger:       logger,
	}
}

func (h *Handler) HandleSpotWS(w http.ResponseWriter, r *http.Request) {
	h.proxyWebSocket(w, r, h.spotWSURL, string(binance.APITypeSpot))
}

func (h *Handler) HandleFuturesWS(w http.ResponseWriter, r *http.Request) {
	h.proxyWebSocket(w, r, h.futuresWSURL, string(binance.APITypeFutures))
}

func (h *Handler) proxyWebSocket(w http.ResponseWriter, r *http.Request, targetBase, apiType string) {
	startTime := time.Now()

	// Extract client IP
	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = strings.Split(forwarded, ",")[0]
	}

	// Build target URL
	targetURL, err := url.Parse(targetBase)
	if err != nil {
		h.logger.Error("failed to parse target URL", logging.Field("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get stream path from URL
	vars := mux.Vars(r)
	streams := vars["streams"]

	// Build the WebSocket path
	if streams != "" {
		targetURL.Path = "/ws/" + streams
	} else if r.URL.Path != "" {
		// Remove the /spot or /futures prefix and use the rest
		path := r.URL.Path
		if strings.HasPrefix(path, "/spot") {
			path = strings.TrimPrefix(path, "/spot")
		} else if strings.HasPrefix(path, "/futures") {
			path = strings.TrimPrefix(path, "/futures")
		}
		if path == "" || path == "/" {
			path = "/ws"
		}
		targetURL.Path = path
	}

	// Preserve query parameters
	targetURL.RawQuery = r.URL.RawQuery

	h.logger.LogWebSocketConnect(clientIP, targetURL.Path, apiType)

	// Upgrade client connection
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("failed to upgrade client connection", logging.Field("error", err.Error()))
		return
	}
	defer clientConn.Close()

	// Connect to Binance WebSocket
	dialer := websocket.Dialer{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
	}

	// Set required headers for Binance WebSocket connection
	headers := http.Header{}
	headers.Set("Origin", "https://"+targetURL.Host)
	if apiKey := r.Header.Get(binance.APIKeyHeader); apiKey != "" {
		headers.Set(binance.APIKeyHeader, apiKey)
	}

	serverConn, _, err := dialer.Dial(targetURL.String(), headers)
	if err != nil {
		h.logger.Error("failed to connect to Binance WebSocket",
			logging.Field("error", err.Error()),
			logging.Field("target", targetURL.String()))
		return
	}
	defer serverConn.Close()

	// Bidirectional proxy
	proxy := NewConnectionProxy(clientConn, serverConn, h.logger, clientIP, apiType)
	proxy.Start()

	h.logger.LogWebSocketDisconnect(clientIP, targetURL.Path, apiType, time.Since(startTime))
}
