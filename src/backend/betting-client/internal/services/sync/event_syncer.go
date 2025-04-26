package sync

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Arlan-Z/def-betting-api/internal/app/store"
	"github.com/Arlan-Z/def-betting-api/internal/data"
	eventsource "github.com/Arlan-Z/def-betting-api/internal/deliveries/eventsource/http"

	// Import usecase packages if specific exported errors need checking by type
	// betuc "github.com/Arlan-Z/def-betting-api/internal/usecases/bet"
	// eventuc "github.com/Arlan-Z/def-betting-api/internal/usecases/event"
	"go.uber.org/zap"
)

type eventFinalizerUseCase interface {
	FinalizeEvent(ctx context.Context, eventID string, actualResult data.Outcome) error
}

type betCancellerUseCase interface {
	CancelBetsForEvent(ctx context.Context, eventID string) error
}

type EventSyncer struct {
	sourceClient eventsource.EventSourceClient
	eventRepo    store.EventRepository
	eventUseCase eventFinalizerUseCase
	betUseCase   betCancellerUseCase
	syncInterval time.Duration
	logger       *zap.Logger
}

func NewEventSyncer(
	sc eventsource.EventSourceClient,
	er store.EventRepository,
	euc eventFinalizerUseCase,
	buc betCancellerUseCase,
	interval time.Duration,
	logger *zap.Logger,
) *EventSyncer {
	return &EventSyncer{
		sourceClient: sc,
		eventRepo:    er,
		eventUseCase: euc,
		betUseCase:   buc,
		syncInterval: interval,
		logger:       logger.Named("EventSyncer"),
	}
}

func (s *EventSyncer) Start(ctx context.Context) {
	s.logger.Info("Starting event synchronization worker", zap.Duration("interval", s.syncInterval))
	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	s.runSync(ctx)

	for {
		select {
		case <-ticker.C:
			s.logger.Debug("Ticker triggered event sync")
			s.runSync(ctx)
		case <-ctx.Done():
			s.logger.Info("Stopping event synchronization worker due to context cancellation")
			return
		}
	}
}

func (s *EventSyncer) runSync(ctx context.Context) {
	log := s.logger.With(zap.Time("sync_time", time.Now().UTC()))
	log.Info("Running event synchronization cycle")

	externalEvents, err := s.sourceClient.FetchActiveEvents(ctx)
	if err != nil {
		log.Error("Failed to fetch events from source API", zap.Error(err))
		return
	}
	log.Info("Fetched events from source API", zap.Int("count", len(externalEvents)))

	successCount := 0
	errorCount := 0
	finalizeAttempts := 0
	finalizeErrors := 0
	cancelAttempts := 0
	cancelErrors := 0

	for _, extEvent := range externalEvents {
		eventLog := log.With(zap.String("externalId", extEvent.APIEventID))

		internalEvent, mapErr := data.MapExternalToInternalEvent(extEvent)
		if mapErr != nil {
			eventLog.Error("Failed to map external event to internal structure", zap.Error(mapErr))
			errorCount++
			continue
		}

		shouldFinalize := false
		var finalizationResult data.Outcome
		if internalEvent.EventResult != nil && !internalEvent.IsActive {
			shouldFinalize = true
			finalizationResult = *internalEvent.EventResult
		}

		upsertErr := s.eventRepo.Upsert(ctx, &internalEvent)
		if upsertErr != nil {
			eventLog.Error("Failed to upsert event into local database", zap.Error(upsertErr))
			errorCount++
			continue
		}
		successCount++

		if shouldFinalize {
			eventLog.Info("Event detected as finalized by source API, attempting to trigger finalization", zap.String("result", string(finalizationResult)))
			finalizeAttempts++

			finalizeErr := s.eventUseCase.FinalizeEvent(ctx, internalEvent.ID, finalizationResult)

			if finalizeErr != nil {
				// TODO: Check if eventuc.ErrEventAlreadyFinalized is exported and use errors.Is
				if finalizeErr.Error() == "событие уже завершено" { // Fallback check by string
					eventLog.Info("Finalization attempt skipped: event already finalized locally.")
				} else {
					eventLog.Error("Error occurred during finalization triggered by syncer", zap.Error(finalizeErr))
					finalizeErrors++
				}
			} else {
				eventLog.Info("Finalization triggered by syncer completed successfully.")
			}
		} else if !internalEvent.IsActive && internalEvent.EventResult == nil {
			eventLog.Info("Event detected as inactive without specific result (Canceled or ended)", zap.Stringp("apiResult", extEvent.Result))
			if extEvent.Result != nil && *extEvent.Result == "Canceled" {
				eventLog.Info("Event result is 'Canceled', attempting to cancel related bets.")
				cancelAttempts++
				cancelErr := s.betUseCase.CancelBetsForEvent(ctx, internalEvent.ID)
				if cancelErr != nil && !errors.Is(cancelErr, sql.ErrNoRows) { // Ignore no rows found error
					// TODO: Check if betuc.ErrBetCancellationFailed is exported and use errors.Is
					eventLog.Error("Error occurred during bet cancellation for canceled event", zap.Error(cancelErr))
					cancelErrors++
				} else if cancelErr == nil {
					eventLog.Info("Bets cancellation process initiated successfully for canceled event.")
				} else { // It was sql.ErrNoRows
					eventLog.Info("No pending bets found to cancel for canceled event.")
				}
			}
		}
	}

	log.Info("Event synchronization cycle finished",
		zap.Int("processed", len(externalEvents)),
		zap.Int("successful_upserts", successCount),
		zap.Int("mapping/upsert_errors", errorCount),
		zap.Int("finalize_attempts", finalizeAttempts),
		zap.Int("finalize_errors", finalizeErrors),
		zap.Int("cancel_attempts", cancelAttempts),
		zap.Int("cancel_errors", cancelErrors),
	)
}
