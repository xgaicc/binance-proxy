package binance

const (
	// Spot API endpoints
	SpotRestURL      = "https://api.binance.com"
	SpotWebSocketURL = "wss://stream.binance.com:9443"

	// Futures API endpoints (USD-M)
	FuturesRestURL      = "https://fapi.binance.com"
	FuturesWebSocketURL = "wss://fstream.binance.com"

	// Authentication header
	APIKeyHeader = "X-MBX-APIKEY"
)

// APIType represents the type of Binance API
type APIType string

const (
	APITypeSpot    APIType = "spot"
	APITypeFutures APIType = "futures"
)
