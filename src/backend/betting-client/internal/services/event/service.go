package event

import (
	"context"

	"github.com/Arlan-Z/def-betting-api/internal/data" // Change path
	"go.uber.org/zap"
	// We would need to import the usecase package to reference its errors, e.g.:
	// "github.com/Arlan-Z/def-betting-api/internal/usecases/event"
)

type EventUseCase interface {
	GetActiveEvents(ctx context.Context) ([]data.Event, error)
	FinalizeEvent(ctx context.Context, eventID string, actualResult data.Outcome) error
}

type Service interface {
	GetActiveEvents(ctx context.Context) ([]data.Event, error)
	FinalizeEvent(ctx context.Context, eventID string, result data.Outcome) error
}

type service struct {
	eventUseCase EventUseCase
	logger       *zap.Logger
}

func NewService(uc EventUseCase, logger *zap.Logger) Service {
	return &service{
		eventUseCase: uc,
		logger:       logger.Named("EventService"),
	}
}

func (s *service) GetActiveEvents(ctx context.Context) ([]data.Event, error) {
	log := s.logger.With(zap.String("method", "GetActiveEvents"))
	log.Debug("Calling use case to get active events")

	events, err := s.eventUseCase.GetActiveEvents(ctx)
	if err != nil {
		log.Warn("Use case returned error getting active events", zap.Error(err))
		return nil, err
	}

	log.Debug("Successfully retrieved active events from use case", zap.Int("count", len(events)))
	return events, nil
}

func (s *service) FinalizeEvent(ctx context.Context, eventID string, result data.Outcome) error {
	log := s.logger.With(zap.String("method", "FinalizeEvent"), zap.String("eventId", eventID))
	log.Info("Calling use case to finalize event")

	err := s.eventUseCase.FinalizeEvent(ctx, eventID, result)
	if err != nil {
		log.Error("Use case returned error finalizing event", zap.Error(err))
		return err
	}

	log.Info("Use case successfully processed event finalization request")
	return nil
}

// Example of referencing an error from the use case package:
// if errors.Is(err, event.ErrEventNotFound) { ... }
