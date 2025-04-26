package start

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Arlan-Z/def-betting-api/internal/app/config"
	"go.uber.org/zap"
)

func RunServer(handler http.Handler, cfg *config.Config, logger *zap.Logger) error {
	log := logger.Named("HttpServer")

	httpServer := &http.Server{
		Addr:         ":" + cfg.HTTPServer.Port,
		Handler:      handler,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout * 2,
		IdleTimeout:  120 * time.Second,
	}

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)

	serverErrors := make(chan error, 1)

	go func() {
		log.Info("Starting HTTP server", zap.String("address", httpServer.Addr))
		err := httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- fmt.Errorf("error starting HTTP server: %w", err)
		} else {
			serverErrors <- nil
		}
	}()

	log.Info("Server goroutine started. Waiting for signal or server error.")

	select {
	case err := <-serverErrors:
		if err != nil {
			log.Error("HTTP server stopped due to error", zap.Error(err))
			return err
		}
		log.Info("HTTP server stopped (likely after Shutdown)")

	case sig := <-shutdownSignal:
		log.Info("Received OS signal", zap.String("signal", sig.String()))
		log.Info("Starting Graceful Shutdown...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Error("Error during Graceful Shutdown", zap.Error(err))
			// return fmt.Errorf("graceful shutdown error: %w", err)
		} else {
			log.Info("HTTP server shut down gracefully")
		}
	}

	return nil
}
