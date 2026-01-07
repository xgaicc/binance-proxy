package rest

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/xgaicc/binance-proxy/internal/logging"
	"github.com/xgaicc/binance-proxy/pkg/binance"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           &bytes.Buffer{},
	}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	lrw.body.Write(b)
	return lrw.ResponseWriter.Write(b)
}

func LoggingMiddleware(logger *logging.RequestLogger, apiType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Capture request body
			var reqBody []byte
			if r.Body != nil {
				reqBody, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(reqBody))
			}

			// Wrap response writer to capture response
			lrw := newLoggingResponseWriter(w)

			next.ServeHTTP(lrw, r)

			// Extract client IP
			clientIP := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				clientIP = strings.Split(forwarded, ",")[0]
			}

			// Log the request/response
			logger.LogRequest(logging.RequestLog{
				Timestamp:    start,
				Duration:     time.Since(start),
				Method:       r.Method,
				Path:         r.URL.Path,
				Query:        r.URL.RawQuery,
				StatusCode:   lrw.statusCode,
				RequestBody:  string(reqBody),
				ResponseBody: lrw.body.String(),
				ClientIP:     clientIP,
				APIKey:       r.Header.Get(binance.APIKeyHeader),
				APIType:      apiType,
			})
		})
	}
}
