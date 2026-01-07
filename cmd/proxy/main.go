package main

import (
	"flag"
	"log"

	"go.uber.org/zap"

	"github.com/xgaicc/binance-proxy/internal/config"
	"github.com/xgaicc/binance-proxy/internal/health"
	"github.com/xgaicc/binance-proxy/internal/logging"
	"github.com/xgaicc/binance-proxy/internal/proxy/rest"
	"github.com/xgaicc/binance-proxy/internal/proxy/websocket"
	"github.com/xgaicc/binance-proxy/internal/server"
)

func main() {
	configPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger, err := logging.NewLogger(&cfg.Logging)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize request logger
	reqLogger := logging.NewRequestLogger(logger, &cfg.Logging)

	// Initialize handlers
	healthHandler := health.NewHandler()

	restHandler, err := rest.NewProxyHandler(cfg, reqLogger)
	if err != nil {
		logger.Fatal("Failed to create REST proxy handler", zap.Error(err))
	}

	wsHandler := websocket.NewHandler(cfg, reqLogger)

	// Setup router
	router := rest.NewRouter(restHandler, wsHandler, healthHandler, reqLogger)

	// Create and start server
	srv := server.New(router, &cfg.Server, logger)

	logger.Info("Binance Proxy starting",
		zap.String("spot_rest", cfg.Binance.Spot.RestURL),
		zap.String("spot_ws", cfg.Binance.Spot.WebSocketURL),
		zap.String("futures_rest", cfg.Binance.Futures.RestURL),
		zap.String("futures_ws", cfg.Binance.Futures.WebSocketURL))

	if err := srv.Start(); err != nil {
		logger.Fatal("Server error", zap.Error(err))
	}
}
