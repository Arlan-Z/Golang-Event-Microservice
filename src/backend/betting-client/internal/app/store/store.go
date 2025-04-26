package store

import (
	"context"

	"github.com/Arlan-Z/def-betting-api/internal/data"
	betrepo "github.com/Arlan-Z/def-betting-api/internal/repositories/bet/sqlite"
	eventrepo "github.com/Arlan-Z/def-betting-api/internal/repositories/event/sqlite"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type EventRepository interface {
	FindActiveEvents(ctx context.Context) ([]data.Event, error)
	FindByID(ctx context.Context, eventID string) (*data.Event, error)
	UpdateResultAndStatus(ctx context.Context, eventID string, result data.Outcome) error
}

type BetRepository interface {
	Save(ctx context.Context, bet *data.Bet) error
	FindPendingByEventID(ctx context.Context, eventID string) ([]data.Bet, error)
	UpdateStatusAndPayout(ctx context.Context, betID string, status data.BetStatus, payout float64) error
	UpdateStatus(ctx context.Context, betID string, status data.BetStatus) error
}

type Store struct {
	db     *sqlx.DB
	logger *zap.Logger
	Event  EventRepository
	Bet    BetRepository
}

func NewStore(db *sqlx.DB, logger *zap.Logger) *Store {
	log := logger.Named("RepositoryStore")
	log.Info("Initializing repository store")

	eventRepoImpl := eventrepo.NewEventRepository(db)
	betRepoImpl := betrepo.NewBetRepository(db)

	return &Store{
		db:     db,
		logger: log,
		Event:  eventRepoImpl,
		Bet:    betRepoImpl,
	}
}

func (s *Store) GetDB() *sqlx.DB {
	return s.db
}
