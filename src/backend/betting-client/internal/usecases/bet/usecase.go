package bet

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Arlan-Z/def-betting-api/internal/data" // Change path
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrEventNotFound   = errors.New("event for betting not found")
	ErrEventNotActive  = errors.New("event is inactive or has already ended, betting is not possible")
	ErrSavingBetFailed = errors.New("failed to save bet")
)

type EventRepository interface {
	FindByID(ctx context.Context, eventID string) (*data.Event, error)
}

type BetRepository interface {
	Save(ctx context.Context, bet *data.Bet) error
}

type UseCase struct {
	betRepo   BetRepository
	eventRepo EventRepository
	logger    *zap.Logger
}

func NewUseCase(br BetRepository, er EventRepository, logger *zap.Logger) *UseCase {
	return &UseCase{
		betRepo:   br,
		eventRepo: er,
		logger:    logger.Named("BetUseCase"), // Added logger name
	}
}

func (uc *UseCase) PlaceBet(ctx context.Context, req data.PlaceBetRequest) (*data.Bet, error) {
	// span := opentracing.StartSpan("PlaceBetUseCase")
	// ctx = opentracing.ContextWithSpan(ctx, span)
	// defer span.Finish()
	log := uc.logger.With(zap.String("userId", req.UserID), zap.String("eventId", req.EventID))
	log.Info("Use Case: Attempting to place bet")

	event, err := uc.eventRepo.FindByID(ctx, req.EventID)
	if err != nil {
		log.Error("Error retrieving event for bet", zap.Error(err))
		return nil, fmt.Errorf("internal error checking event")
	}
	if event == nil {
		log.Warn("Event for betting not found")
		return nil, ErrEventNotFound
	}

	now := time.Now().UTC()
	if !event.IsActive || now.After(event.EventEndDate) || now.After(event.EventStartDate) {
		log.Warn("Attempt to bet on inactive or started/finished event",
			zap.Bool("isActive", event.IsActive),
			zap.Time("eventStart", event.EventStartDate),
			zap.Time("eventEnd", event.EventEndDate),
		)
		return nil, ErrEventNotActive
	}

	newBet := &data.Bet{
		ID:                    uuid.NewString(),
		UserID:                req.UserID,
		EventID:               req.EventID,
		Amount:                req.Amount,
		PredictedOutcome:      req.PredictedOutcome,
		RecordedHomeWinChance: event.HomeWinChance,
		RecordedAwayWinChance: event.AwayWinChance,
		RecordedDrawChance:    event.DrawChance,
		PlacedAt:              now,
		Status:                data.StatusPending,
		PayoutAmount:          0,
	}

	err = uc.betRepo.Save(ctx, newBet)
	if err != nil {
		log.Error("Error saving bet in repository", zap.Error(err))
		return nil, ErrSavingBetFailed
	}

	log.Info("Bet placed successfully", zap.String("betId", newBet.ID))
	return newBet, nil
}
