package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/xgaicc/binance-proxy/internal/config"
)

type Server struct {
	httpServer *http.Server
	logger     *zap.Logger
	cfg        *config.ServerConfig
}

func New(handler http.Handler, cfg *config.ServerConfig, logger *zap.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.Address(),
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		},
		logger: logger,
		cfg:    cfg,
	}
}

func (s *Server) Start() error {
	// Channel to listen for shutdown signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Channel for server errors
	errCh := make(chan error, 1)

	// Start server in goroutine
	go func() {
		s.logger.Info("Starting server",
			zap.String("address", s.httpServer.Addr),
			zap.String("host", s.cfg.Host),
			zap.Int("port", s.cfg.Port))

		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("server failed: %w", err)
		}
	}()

	// Wait for shutdown signal or error
	select {
	case err := <-errCh:
		return err
	case sig := <-stop:
		s.logger.Info("Shutdown signal received", zap.String("signal", sig.String()))
	}

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()

	// Gracefully shutdown
	s.logger.Info("Shutting down server...")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Server shutdown error", zap.Error(err))
		return err
	}

	s.logger.Info("Server stopped gracefully")
	return nil
}
