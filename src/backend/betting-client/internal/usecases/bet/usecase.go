package bet

import (
	"context"
	"errors"
	"fmt"
	"time"

	"database/sql"

	"github.com/Arlan-Z/def-betting-api/internal/data"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrEventNotFound         = errors.New("event for betting not found")
	ErrEventNotActive        = errors.New("event is inactive or has already ended, betting is not possible")
	ErrSavingBetFailed       = errors.New("failed to save bet")
	ErrBetCancellationFailed = errors.New("couldn't cancel one or more bets")
)

type EventRepository interface {
	FindByID(ctx context.Context, eventID string) (*data.Event, error)
}

type BetRepository interface {
	Save(ctx context.Context, bet *data.Bet) error
	FindPendingByEventID(ctx context.Context, eventID string) ([]data.Bet, error)
	UpdateStatus(ctx context.Context, betID string, status data.BetStatus) error
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

func (uc *UseCase) CancelBetsForEvent(ctx context.Context, eventID string) error {
	log := uc.logger.With(zap.String("eventId", eventID), zap.String("operation", "CancelBetsForEvent"))
	log.Info("Use Case: Attempt to cancel bets for an event")

	pendingBets, err := uc.betRepo.FindPendingByEventID(ctx, eventID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info("No pending bids were found to cancel")
			return nil
		}
		log.Error("Error searching for pending bids to cancel", zap.Error(err))
		return fmt.Errorf("internal error when searching for bids to cancel")
	}

	if len(pendingBets) == 0 {
		log.Info("No pending bids were found to cancel")
		return nil
	}

	log.Info("Found pending bids to cancel", zap.Int("count", len(pendingBets)))

	var cancellationErrors []error
	canceledCount := 0
	for _, bet := range pendingBets {
		err := uc.betRepo.UpdateStatus(ctx, bet.ID, data.StatusCanceled)
		if err != nil {
			log.Error("Error updating the status of the Cancelled bet", zap.String("betId", bet.ID), zap.Error(err))
			cancellationErrors = append(cancellationErrors, fmt.Errorf("bet %s: %w", bet.ID, err))
		} else {
			canceledCount++
		}
	}

	if len(cancellationErrors) > 0 {
		log.Error("Errors occurred during the cancellation of bets", zap.Int("total", len(pendingBets)), zap.Int("canceled", canceledCount), zap.Int("errors", len(cancellationErrors)))
		return fmt.Errorf("%w: %d bets failed to cancel", ErrBetCancellationFailed, len(cancellationErrors))
	}

	log.Info("All pending bids have been successfully cancelled", zap.Int("count", canceledCount))
	return nil
}
