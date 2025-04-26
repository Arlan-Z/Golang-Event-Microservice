package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Arlan-Z/def-betting-api/internal/data"
	customvalidator "github.com/Arlan-Z/def-betting-api/internal/pkg/validator"
	"github.com/Arlan-Z/def-betting-api/internal/usecases/bet"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type BetUseCase interface {
	PlaceBet(ctx context.Context, req data.PlaceBetRequest) (*data.Bet, error)
}

type Handler struct {
	useCase BetUseCase
	logger  *zap.Logger
}

func NewHandler(uc BetUseCase, logger *zap.Logger) *Handler {
	return &Handler{
		useCase: uc,
		logger:  logger.Named("BetHandler"),
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/bets", h.PlaceBet)
}

func (h *Handler) PlaceBet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.logger.With(zap.String("operation", "PlaceBet"))
	log.Info("Received request to place a bet")

	var requestDTO data.PlaceBetRequest

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		log.Warn("Error decoding request body", zap.Error(err))
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := customvalidator.ValidateStruct(requestDTO); err != nil {
		log.Warn("Error validating request body", zap.Error(err))
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	createdBet, err := h.useCase.PlaceBet(ctx, requestDTO)

	if err != nil {
		log.Error("Error placing bet in UseCase",
			zap.String("userId", requestDTO.UserID),
			zap.String("eventId", requestDTO.EventID),
			zap.Error(err),
		)
		switch {
		case errors.Is(err, bet.ErrEventNotFound):
			http.Error(w, "Event for betting not found", http.StatusNotFound)
		case errors.Is(err, bet.ErrEventNotActive):
			http.Error(w, "Betting on this event is no longer accepted", http.StatusConflict)
		case errors.Is(err, bet.ErrSavingBetFailed):
			http.Error(w, "Failed to save bet, please try again later", http.StatusInternalServerError)
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	responseDTO := data.MapBetToDTO(*createdBet)

	log.Info("Bet placed successfully", zap.String("betId", responseDTO.ID))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(responseDTO); err != nil {
		log.Error("Error encoding JSON response", zap.Error(err))
	}
}
