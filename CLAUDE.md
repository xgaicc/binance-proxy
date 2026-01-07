# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Run Commands

```bash
# Build the binary
go build -o binance-proxy ./cmd/proxy

# Run with default config (loads from configs/config.yaml)
./binance-proxy

# Run with custom config
./binance-proxy -config /path/to/config.yaml

# Docker deployment
cd deployments && docker-compose up -d
```

## Architecture Overview

This is a Binance API reverse proxy written in Go that enables trading bots to route requests through a central server with request logging.

### Request Flow

1. **Entry Point** (`cmd/proxy/main.go`): Initializes config, logger, handlers, and starts the HTTP server
2. **Router** (`internal/proxy/rest/router.go`): Gorilla mux routes requests by prefix:
   - `/spot/*` → Spot API handlers (REST + WebSocket)
   - `/futures/*` → Futures API handlers (REST + WebSocket)
   - `/health`, `/ready` → Health check endpoints (no logging middleware)

3. **REST Proxy** (`internal/proxy/rest/handler.go`): Uses `httputil.ReverseProxy` to forward requests to Binance APIs while preserving `X-MBX-APIKEY` header and query parameters (signature, timestamp)

4. **WebSocket Proxy** (`internal/proxy/websocket/`): Bidirectional proxy that:
   - Upgrades client connection
   - Connects to Binance WebSocket
   - Forwards messages in both directions via `ConnectionProxy`

### Configuration

Configuration uses Viper with this priority: environment variables > config file > defaults.

- **File**: `configs/config.yaml` or custom path via `-config` flag
- **Environment**: Variables prefixed with `PROXY_` (e.g., `PROXY_SERVER_PORT=9090`)
- **Key config sections**: `server`, `binance` (spot/futures endpoints), `logging`

### Key Packages

- `pkg/binance/endpoints.go`: Binance API constants (API types, header names)
- `internal/logging/`: Structured JSON logging with zap, API key masking
- `internal/health/`: Liveness/readiness probe handlers
