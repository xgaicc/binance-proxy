package logging

import (
	"time"

	"go.uber.org/zap"

	"github.com/xgaicc/binance-proxy/internal/config"
)

const (
	maxRequestBodyLog  = 1000
	maxResponseBodyLog = 2000
	maxWSMessageLog    = 500
)

type RequestLog struct {
	Timestamp    time.Time
	Duration     time.Duration
	Method       string
	Path         string
	Query        string
	StatusCode   int
	RequestBody  string
	ResponseBody string
	ClientIP     string
	APIKey       string
	APIType      string
}

type RequestLogger struct {
	logger       *zap.Logger
	logRequests  bool
	logResponses bool
}

func NewRequestLogger(logger *zap.Logger, cfg *config.LoggingConfig) *RequestLogger {
	return &RequestLogger{
		logger:       logger,
		logRequests:  cfg.LogRequests,
		logResponses: cfg.LogResponses,
	}
}

func (l *RequestLogger) LogRequest(log RequestLog) {
	fields := []zap.Field{
		zap.Time("timestamp", log.Timestamp),
		zap.Duration("duration_ms", log.Duration),
		zap.String("method", log.Method),
		zap.String("path", log.Path),
		zap.Int("status_code", log.StatusCode),
		zap.String("client_ip", log.ClientIP),
		zap.String("api_type", log.APIType),
	}

	if log.Query != "" {
		fields = append(fields, zap.String("query", log.Query))
	}

	if log.APIKey != "" {
		fields = append(fields, zap.String("api_key", MaskAPIKey(log.APIKey)))
	}

	if l.logRequests && log.RequestBody != "" {
		fields = append(fields, zap.String("request_body", truncate(log.RequestBody, maxRequestBodyLog)))
	}

	if l.logResponses && log.ResponseBody != "" {
		fields = append(fields, zap.String("response_body", truncate(log.ResponseBody, maxResponseBodyLog)))
	}

	l.logger.Info("api_request", fields...)
}

func (l *RequestLogger) LogWebSocketConnect(clientIP, path, apiType string) {
	l.logger.Info("websocket_connect",
		zap.String("client_ip", clientIP),
		zap.String("path", path),
		zap.String("api_type", apiType),
		zap.Time("timestamp", time.Now()),
	)
}

func (l *RequestLogger) LogWebSocketDisconnect(clientIP, path, apiType string, duration time.Duration) {
	l.logger.Info("websocket_disconnect",
		zap.String("client_ip", clientIP),
		zap.String("path", path),
		zap.String("api_type", apiType),
		zap.Duration("duration_ms", duration),
		zap.Time("timestamp", time.Now()),
	)
}

func (l *RequestLogger) LogWebSocketMessage(direction, clientIP, apiType string, message []byte) {
	l.logger.Debug("websocket_message",
		zap.String("direction", direction),
		zap.String("client_ip", clientIP),
		zap.String("api_type", apiType),
		zap.Int("size_bytes", len(message)),
		zap.String("content", truncate(string(message), maxWSMessageLog)),
	)
}

func (l *RequestLogger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

func (l *RequestLogger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

func (l *RequestLogger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

func MaskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...[truncated]"
}
