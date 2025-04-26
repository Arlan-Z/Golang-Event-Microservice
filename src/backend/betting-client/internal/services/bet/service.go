package bet

import (
	"context"

	"github.com/Arlan-Z/def-betting-api/internal/data" // Change path
	"go.uber.org/zap"
)

type BetUseCase interface {
	PlaceBet(ctx context.Context, req data.PlaceBetRequest) (*data.Bet, error)
	CancelBetsForEvent(ctx context.Context, eventID string) error
}

type Service interface {
	PlaceBet(ctx context.Context, req data.PlaceBetRequest) (*data.Bet, error)
}

type service struct {
	betUseCase BetUseCase
	logger     *zap.Logger
}

func NewService(uc BetUseCase, logger *zap.Logger) Service {
	return &service{
		betUseCase: uc,
		logger:     logger.Named("BetService"),
	}
}

func (s *service) PlaceBet(ctx context.Context, req data.PlaceBetRequest) (*data.Bet, error) {
	log := s.logger.With(zap.String("method", "PlaceBet"), zap.String("userId", req.UserID), zap.String("eventId", req.EventID))
	log.Info("Calling use case to place bet")

	createdBet, err := s.betUseCase.PlaceBet(ctx, req)
	if err != nil {
		log.Error("Use case returned error placing bet", zap.Error(err))
		return nil, err
	}

	log.Info("Bet placed successfully via use case", zap.String("betId", createdBet.ID))
	return createdBet, nil
}
