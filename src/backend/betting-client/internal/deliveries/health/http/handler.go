package http

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Handler struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewHandler(db *sqlx.DB, logger *zap.Logger) *Handler {
	return &Handler{
		db:     db,
		logger: logger.Named("HealthHandler"),
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/healthz", h.Healthz)
	r.Get("/readyz", h.Readyz)
}

func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	// fmt.Fprintln(w, "OK")
	h.logger.Debug("Healthz probe successful")
}

func (h *Handler) Readyz(w http.ResponseWriter, r *http.Request) {
	log := h.logger.With(zap.String("operation", "Readyz"))

	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	var dbConn *sql.DB = h.db.DB
	err := dbConn.PingContext(ctx)
	if err != nil {
		log.Error("Readiness probe failed: database ping error", zap.Error(err))
		http.Error(w, "Service Unavailable: Database connection failed", http.StatusServiceUnavailable)
		return
	}

	log.Debug("Readyz probe successful")
	w.WriteHeader(http.StatusOK)
	// fmt.Fprintln(w, "OK")
}
