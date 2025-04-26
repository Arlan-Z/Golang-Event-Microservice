package event

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/Arlan-Z/def-betting-api/internal/data"
	payoutclient "github.com/Arlan-Z/def-betting-api/internal/deliveries/payout/http"
	"go.uber.org/zap"
)

var (
	ErrEventNotFound             = errors.New("event not found")
	ErrEventAlreadyFinalized     = errors.New("event already finalized")
	ErrEventNotFinishedYet       = errors.New("event has not finished yet based on time")
	ErrInvalidFinalizationResult = errors.New("invalid result for event finalization")
	ErrPayoutNotificationFailed  = errors.New("failed to notify payout service")
	ErrBetUpdateFailed           = errors.New("failed to update bet status")
)

type EventRepository interface {
	FindActiveEvents(ctx context.Context) ([]data.Event, error)
	FindByID(ctx context.Context, eventID string) (*data.Event, error)
	UpdateResultAndStatus(ctx context.Context, eventID string, result data.Outcome) error
}

type BetRepository interface {
	FindPendingByEventID(ctx context.Context, eventID string) ([]data.Bet, error)
	UpdateStatusAndPayout(ctx context.Context, betID string, status data.BetStatus, payout float64) error
	UpdateStatus(ctx context.Context, betID string, status data.BetStatus) error
}

type UseCase struct {
	eventRepo    EventRepository
	betRepo      BetRepository
	payoutClient payoutclient.PayoutClient
	logger       *zap.Logger
}

func NewUseCase(er EventRepository, br BetRepository, pc payoutclient.PayoutClient, logger *zap.Logger) *UseCase {
	return &UseCase{
		eventRepo:    er,
		betRepo:      br,
		payoutClient: pc,
		logger:       logger.Named("EventUseCase"), // Added logger name
	}
}

func (uc *UseCase) GetActiveEvents(ctx context.Context) ([]data.Event, error) {
	uc.logger.Debug("Use Case: Requesting active events")

	events, err := uc.eventRepo.FindActiveEvents(ctx)
	if err != nil {
		uc.logger.Error("Error getting active events from repository", zap.Error(err))
		return nil, fmt.Errorf("failed to get list of active events")
	}

	uc.logger.Debug("Use Case: Active events retrieved", zap.Int("count", len(events)))
	return events, nil
}

func (uc *UseCase) FinalizeEvent(ctx context.Context, eventID string, actualResult data.Outcome) error {
	// span := opentracing.StartSpan("FinalizeEventUseCase")
	// ctx = opentracing.ContextWithSpan(ctx, span)
	// defer span.Finish()

	uc.logger.Info("Use Case: Finalizing event", zap.String("eventId", eventID), zap.String("result", string(actualResult)))

	if actualResult != data.HomeWin && actualResult != data.AwayWin && actualResult != data.Draw {
		uc.logger.Warn("Attempt to finalize event with invalid result", zap.String("eventId", eventID), zap.String("result", string(actualResult)))
		return ErrInvalidFinalizationResult
	}

	event, err := uc.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		uc.logger.Error("Error retrieving event for finalization", zap.String("eventId", eventID), zap.Error(err))
		return fmt.Errorf("internal error searching for event")
	}
	if event == nil {
		uc.logger.Warn("Attempt to finalize non-existent event", zap.String("eventId", eventID))
		return ErrEventNotFound
	}
	if !event.IsActive || event.EventResult != nil {
		uc.logger.Warn("Attempt to re-finalize event", zap.String("eventId", eventID))
		return ErrEventAlreadyFinalized
	}
	// Optional check:
	// if time.Now().UTC().Before(event.EventEndDate) {
	// 	uc.logger.Warn("Attempt to finalize event before its end time", zap.String("eventId", eventID))
	// 	return ErrEventNotFinishedYet
	// }

	pendingBets, err := uc.betRepo.FindPendingByEventID(ctx, eventID)
	if err != nil {
		uc.logger.Error("Error retrieving pending bets", zap.String("eventId", eventID), zap.Error(err))
		return fmt.Errorf("internal error retrieving bets")
	}
	uc.logger.Info("Found pending bets for finalization", zap.String("eventId", eventID), zap.Int("count", len(pendingBets)))

	var finalizationErrors []error
	processedBetsCount := 0
	successfulPayouts := 0

	for _, bet := range pendingBets {
		betLogger := uc.logger.With(zap.String("betId", bet.ID), zap.String("userId", bet.UserID))
		var newStatus data.BetStatus
		var payoutAmount float64 = 0

		if bet.PredictedOutcome == actualResult {
			newStatus = data.StatusWon
			switch actualResult {
			case data.HomeWin:
				payoutAmount = bet.Amount * bet.RecordedHomeWinChance
			case data.AwayWin:
				payoutAmount = bet.Amount * bet.RecordedAwayWinChance
			case data.Draw:
				payoutAmount = bet.Amount * bet.RecordedDrawChance
			}
			payoutAmount = math.Round(payoutAmount*100) / 100

			betLogger.Info("Bet won", zap.Float64("payoutAmount", payoutAmount))
		} else {
			newStatus = data.StatusLost
			betLogger.Info("Bet lost")
		}

		err = uc.betRepo.UpdateStatusAndPayout(ctx, bet.ID, newStatus, payoutAmount)
		if err != nil {
			betLogger.Error("Error updating bet status in DB", zap.Error(err))
			finalizationErrors = append(finalizationErrors, fmt.Errorf("%w (ID: %s): %v", ErrBetUpdateFailed, bet.ID, err))
			continue
		}

		if newStatus == data.StatusWon && payoutAmount > 0 {
			notification := data.PayoutNotification{
				UserID: bet.UserID,
				Amount: payoutAmount,
			}
			err = uc.payoutClient.NotifyPayout(ctx, notification)
			if err != nil {
				betLogger.Error("Error notifying payout service", zap.Error(err))
				finalizationErrors = append(finalizationErrors, fmt.Errorf("%w (BetID: %s): %v", ErrPayoutNotificationFailed, bet.ID, err))

				errUpdate := uc.betRepo.UpdateStatus(ctx, bet.ID, data.StatusFailed)
				if errUpdate != nil {
					betLogger.Error("CRITICAL: Error updating status to Failed after payout failure", zap.Error(errUpdate))
					finalizationErrors = append(finalizationErrors, fmt.Errorf("%w (ID: %s, status -> Failed): %v", ErrBetUpdateFailed, bet.ID, errUpdate))
				}
			} else {
				betLogger.Info("Payout service notified successfully")
				errUpdate := uc.betRepo.UpdateStatus(ctx, bet.ID, data.StatusPaid)
				if errUpdate != nil {
					betLogger.Error("Error updating status to Paid after successful payout", zap.Error(errUpdate))
					finalizationErrors = append(finalizationErrors, fmt.Errorf("%w (ID: %s, status -> Paid): %v", ErrBetUpdateFailed, bet.ID, errUpdate))
				} else {
					successfulPayouts++
				}
			}
		}
		processedBetsCount++
	}

	err = uc.eventRepo.UpdateResultAndStatus(ctx, eventID, actualResult)
	if err != nil {
		uc.logger.Error("Critical error: Failed to update event status after processing bets", zap.String("eventId", eventID), zap.Error(err))
		finalizationErrors = append([]error{fmt.Errorf("failed to finalize event %s in DB: %w", eventID, err)}, finalizationErrors...)
	} else {
		uc.logger.Info("Event status successfully updated to 'finalized'", zap.String("eventId", eventID))
	}

	if len(finalizationErrors) > 0 {
		uc.logger.Error("Event finalization completed with errors", zap.String("eventId", eventID), zap.Int("processedBets", processedBetsCount), zap.Int("successfulPayouts", successfulPayouts), zap.Errors("errors", finalizationErrors))
		errMsg := fmt.Sprintf("event finalization %s completed with %d errors: ", eventID, len(finalizationErrors))
		for i, e := range finalizationErrors {
			errMsg += e.Error()
			if i < len(finalizationErrors)-1 {
				errMsg += "; "
			}
		}
		return errors.New(errMsg)
	}

	uc.logger.Info("Event finalized successfully", zap.String("eventId", eventID), zap.Int("processedBets", processedBetsCount), zap.Int("successfulPayouts", successfulPayouts))
	return nil
}
