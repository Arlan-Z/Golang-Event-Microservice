package main

import (
	"log"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Arlan-Z/def-betting-api/internal/app/config"
	"github.com/Arlan-Z/def-betting-api/internal/app/connections"
	"github.com/Arlan-Z/def-betting-api/internal/app/start"
	"github.com/Arlan-Z/def-betting-api/internal/app/store"

	bet_delivery "github.com/Arlan-Z/def-betting-api/internal/deliveries/bet/http"
	event_delivery "github.com/Arlan-Z/def-betting-api/internal/deliveries/event/http"
	health_delivery "github.com/Arlan-Z/def-betting-api/internal/deliveries/health/http"
	payout_client "github.com/Arlan-Z/def-betting-api/internal/deliveries/payout/http"

	bet_service "github.com/Arlan-Z/def-betting-api/internal/services/bet"
	event_service "github.com/Arlan-Z/def-betting-api/internal/services/event"

	bet_uc "github.com/Arlan-Z/def-betting-api/internal/usecases/bet"
	event_uc "github.com/Arlan-Z/def-betting-api/internal/usecases/event"

	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize zap logger: %v", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()
	sugar.Info("Logger initialized")

	cfg := config.Load()
	sugar.Info("Configuration loaded")
	sugar.Infof("Database path: %s", cfg.Database.Path)
	sugar.Infof("Payout service URL: %s", cfg.PayoutService.URL)
	sugar.Infof("Server port: %s", cfg.HTTPServer.Port)

	db, err := connections.NewSQLiteConnection(cfg.Database.Path)
	if err != nil {
		sugar.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if cerr := db.Close(); cerr != nil {
			sugar.Warnf("Warning: failed to close DB connection: %v", cerr)
		} else {
			sugar.Info("Database connection closed")
		}
	}()
	sugar.Info("Database connection established")

	repositoryStore := store.NewStore(db, logger)
	sugar.Info("Repository store initialized")

	payoutClient := payout_client.NewRestyPayoutClient(cfg.PayoutService.URL, cfg.PayoutService.Timeout, logger)
	sugar.Info("Payout service client initialized")

	eventUseCase := event_uc.NewUseCase(
		repositoryStore.Event,
		repositoryStore.Bet,
		payoutClient,
		logger,
	)
	betUseCase := bet_uc.NewUseCase(
		repositoryStore.Bet,
		repositoryStore.Event,
		logger,
	)
	sugar.Info("Use cases initialized")

	eventService := event_service.NewService(eventUseCase, logger)
	betService := bet_service.NewService(betUseCase, logger)
	sugar.Info("Services initialized")

	eventHandler := event_delivery.NewHandler(eventService, logger)
	betHandler := bet_delivery.NewHandler(betService, logger)
	healthHandler := health_delivery.NewHandler(db, logger)
	sugar.Info("HTTP handlers initialized")

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	sugar.Info("Base router middleware configured")

	r.Route("/api/v1", func(r chi.Router) {
		sugar.Info("Registering routes for /api/v1...")
		healthHandler.RegisterRoutes(r)
		eventHandler.RegisterRoutes(r)
		betHandler.RegisterRoutes(r)
	})
	sugar.Info("All routes registered")

	sugar.Info("Starting HTTP server...")
	if err := start.RunServer(r, cfg, logger); err != nil {
		sugar.Fatalf("HTTP server fatal error: %v", err)
	}

	sugar.Info("Application shut down gracefully")
}
