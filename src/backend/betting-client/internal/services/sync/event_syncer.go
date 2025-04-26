package sync

import (
	"context"
	"time"

	"github.com/Arlan-Z/def-betting-api/internal/app/store"
	"github.com/Arlan-Z/def-betting-api/internal/data"
	eventsource "github.com/Arlan-Z/def-betting-api/internal/deliveries/eventsource/http"
	"go.uber.org/zap"
)

type EventSyncer struct {
	sourceClient eventsource.EventSourceClient
	eventRepo    store.EventRepository
	syncInterval time.Duration
	logger       *zap.Logger
}

func NewEventSyncer(
	sc eventsource.EventSourceClient,
	er store.EventRepository,
	interval time.Duration,
	logger *zap.Logger,
) *EventSyncer {
	return &EventSyncer{
		sourceClient: sc,
		eventRepo:    er,
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
	for _, extEvent := range externalEvents {
		internalEvent, mapErr := data.MapExternalToInternalEvent(extEvent)
		if mapErr != nil {
			log.Error("Failed to map external event to internal structure",
				zap.String("externalId", extEvent.APIEventID),
				zap.Error(mapErr),
			)
			errorCount++
			continue
		}

		upsertErr := s.eventRepo.Upsert(ctx, &internalEvent)
		if upsertErr != nil {
			log.Error("Failed to upsert event into local database",
				zap.String("eventId", internalEvent.ID),
				zap.Error(upsertErr),
			)
			errorCount++
			continue
		}
		successCount++
	}

	log.Info("Event synchronization cycle finished",
		zap.Int("processed", len(externalEvents)),
		zap.Int("successful_upserts", successCount),
		zap.Int("errors", errorCount),
	)
}
