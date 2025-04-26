package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Arlan-Z/def-betting-api/internal/data"
	customvalidator "github.com/Arlan-Z/def-betting-api/internal/pkg/validator"
	"github.com/Arlan-Z/def-betting-api/internal/usecases/event"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type EventUseCase interface {
	GetActiveEvents(ctx context.Context) ([]data.Event, error)
	FinalizeEvent(ctx context.Context, eventID string, actualResult data.Outcome) error
}

type Handler struct {
	useCase EventUseCase
	logger  *zap.Logger
}

func NewHandler(uc EventUseCase, logger *zap.Logger) *Handler {
	return &Handler{
		useCase: uc,
		logger:  logger.Named("EventHandler"),
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/events", h.GetActiveEvents)
	r.Post("/events/{eventID}/finalize", h.FinalizeEvent)
}

func (h *Handler) GetActiveEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.logger.With(zap.String("operation", "GetActiveEvents"))
	log.Debug("Received request for active events")

	events, err := h.useCase.GetActiveEvents(ctx)
	if err != nil {
		log.Error("Error getting active events from UseCase", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	dtos := data.MapEventsToDTOs(events)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(dtos); err != nil {
		log.Error("Error encoding JSON response", zap.Error(err))
	}
	log.Debug("Successful response", zap.Int("eventCount", len(dtos)))
}

func (h *Handler) FinalizeEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	eventID := chi.URLParam(r, "eventID")
	log := h.logger.With(zap.String("operation", "FinalizeEvent"), zap.String("eventId", eventID))
	log.Info("Received request to finalize event")

	if eventID == "" {
		log.Warn("eventID not specified in request")
		http.Error(w, "Event ID not specified", http.StatusBadRequest)
		return
	}

	var requestBody struct {
		Result data.Outcome `json:"result" validate:"required,oneof=HomeWin AwayWin Draw"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		log.Warn("Error decoding request body", zap.Error(err))
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := customvalidator.ValidateStruct(requestBody); err != nil {
		log.Warn("Error validating request body", zap.Error(err))
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	err := h.useCase.FinalizeEvent(ctx, eventID, requestBody.Result)

	if err != nil {
		log.Error("Error finalizing event in UseCase", zap.Error(err))
		switch {
		case errors.Is(err, event.ErrEventNotFound):
			http.Error(w, "Event not found", http.StatusNotFound)
		case errors.Is(err, event.ErrEventAlreadyFinalized):
			http.Error(w, "Event already finalized", http.StatusConflict)
		case errors.Is(err, event.ErrInvalidFinalizationResult):
			http.Error(w, "Invalid result for finalization", http.StatusBadRequest)
		case errors.Is(err, event.ErrBetUpdateFailed), errors.Is(err, event.ErrPayoutNotificationFailed):
			http.Error(w, "Internal server error during bet or payout processing", http.StatusInternalServerError)
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	log.Info("Event successfully queued for finalization or finalized")
	w.WriteHeader(http.StatusOK)
	// fmt.Fprintf(w, "Finalization process for event %s started successfully", eventID)
}
