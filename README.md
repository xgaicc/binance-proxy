# Binance API Proxy

A Golang reverse proxy for trading bots to access Binance APIs through a central server with request logging for organizational visibility and monitoring.

## Features

- **REST API Proxy**: Forward requests to Binance Spot and Futures APIs
- **WebSocket Proxy**: Bidirectional proxy for market data streams
- **Pass-through Authentication**: Bots provide their own Binance API keys
- **Request Logging**: Structured JSON logs with timestamps, masked API keys
- **Health Checks**: Liveness and readiness endpoints
- **Graceful Shutdown**: Clean connection handling on termination
- **Docker Ready**: Multi-stage Dockerfile included

## Quick Start

### Build and Run

```bash
# Build
go build -o binance-proxy ./cmd/proxy

# Run
./binance-proxy

# Or with custom config
./binance-proxy -config /path/to/config.yaml
```

### Docker

```bash
# Build and run with Docker Compose
cd deployments
docker-compose up -d
```

## API Usage

### REST API

Route requests through the proxy by prefixing with `/spot` or `/futures`:

```bash
# Spot API - Get ticker price
curl "http://localhost:8080/spot/api/v3/ticker/price?symbol=BTCUSDT"

# Spot API - With authentication
curl "http://localhost:8080/spot/api/v3/account" \
  -H "X-MBX-APIKEY: your-api-key"

# Futures API - Get ticker price
curl "http://localhost:8080/futures/fapi/v1/ticker/price?symbol=BTCUSDT"

# Futures API - Place order (requires signature)
curl -X POST "http://localhost:8080/futures/fapi/v1/order" \
  -H "X-MBX-APIKEY: your-api-key" \
  -d "symbol=BTCUSDT&side=BUY&type=LIMIT&..."
```

### WebSocket

Connect to market data streams:

```bash
# Spot WebSocket - Trade stream
wscat -c "ws://localhost:8080/spot/ws/btcusdt@trade"

# Spot WebSocket - Multiple streams
wscat -c "ws://localhost:8080/spot/ws/btcusdt@trade/ethusdt@trade"

# Futures WebSocket - Aggregate trade stream
wscat -c "ws://localhost:8080/futures/ws/btcusdt@aggTrade"
```

### Health Endpoints

```bash
# Liveness check
curl http://localhost:8080/health

# Readiness check
curl http://localhost:8080/ready
```

## Configuration

Configuration is loaded from `configs/config.yaml` or via environment variables:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  readTimeout: 30s
  writeTimeout: 30s
  shutdownTimeout: 10s

binance:
  spot:
    restUrl: "https://api.binance.com"
    websocketUrl: "wss://stream.binance.com:9443"
  futures:
    restUrl: "https://fapi.binance.com"
    websocketUrl: "wss://fstream.binance.com"

logging:
  level: "info"          # debug, info, warn, error
  format: "json"         # json or console
  logRequests: true      # Log request bodies
  logResponses: true     # Log response bodies
  outputPath: "stdout"   # stdout or file path
```

### Environment Variables

Override config with environment variables prefixed with `PROXY_`:

```bash
PROXY_SERVER_PORT=9090
PROXY_LOGGING_LEVEL=debug
PROXY_BINANCE_SPOT_RESTURL=https://testnet.binance.vision
```

## Logging

All API requests are logged in structured JSON format:

```json
{
  "level": "info",
  "timestamp": "2025-01-07T10:00:00.000Z",
  "message": "api_request",
  "method": "GET",
  "path": "/api/v3/ticker/price",
  "query": "symbol=BTCUSDT",
  "status_code": 200,
  "duration_ms": 45.123,
  "client_ip": "192.168.1.100",
  "api_key": "abcd****wxyz",
  "api_type": "spot"
}
```

API keys are automatically masked in logs (showing first 4 and last 4 characters).

## Project Structure

```
binance-proxy/
├── cmd/proxy/main.go              # Application entry point
├── internal/
│   ├── config/config.go           # Configuration management
│   ├── proxy/
│   │   ├── rest/                  # REST reverse proxy
│   │   │   ├── handler.go
│   │   │   ├── router.go
│   │   │   └── middleware.go
│   │   └── websocket/             # WebSocket proxy
│   │       ├── handler.go
│   │       └── connection.go
│   ├── logging/                   # Structured logging
│   ├── health/                    # Health check endpoints
│   └── server/                    # HTTP server
├── pkg/binance/                   # Binance constants
├── configs/config.yaml            # Default configuration
└── deployments/                   # Docker files
```

## License

MIT
