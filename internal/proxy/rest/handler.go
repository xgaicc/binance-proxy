package rest

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/xgaicc/binance-proxy/internal/config"
	"github.com/xgaicc/binance-proxy/internal/logging"
	"github.com/xgaicc/binance-proxy/pkg/binance"
)

type ProxyHandler struct {
	spotProxy    *httputil.ReverseProxy
	futuresProxy *httputil.ReverseProxy
	spotURL      *url.URL
	futuresURL   *url.URL
	logger       *logging.RequestLogger
}

func NewProxyHandler(cfg *config.Config, logger *logging.RequestLogger) (*ProxyHandler, error) {
	spotURL, err := url.Parse(cfg.Binance.Spot.RestURL)
	if err != nil {
		return nil, err
	}

	futuresURL, err := url.Parse(cfg.Binance.Futures.RestURL)
	if err != nil {
		return nil, err
	}

	return &ProxyHandler{
		spotProxy:    createReverseProxy(spotURL),
		futuresProxy: createReverseProxy(futuresURL),
		spotURL:      spotURL,
		futuresURL:   futuresURL,
		logger:       logger,
	}, nil
}

func createReverseProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(target)
			pr.Out.Host = target.Host

			// Preserve authentication header
			if apiKey := pr.In.Header.Get(binance.APIKeyHeader); apiKey != "" {
				pr.Out.Header.Set(binance.APIKeyHeader, apiKey)
			}

			// Preserve query parameters (including signature, timestamp, recvWindow)
			pr.Out.URL.RawQuery = pr.In.URL.RawQuery
		},
	}

	return proxy
}

func (h *ProxyHandler) SpotHandler() http.Handler {
	return h.spotProxy
}

func (h *ProxyHandler) FuturesHandler() http.Handler {
	return h.futuresProxy
}
