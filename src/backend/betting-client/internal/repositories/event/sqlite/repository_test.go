package sqlite_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Arlan-Z/def-betting-api/internal/data"
	eventrepo "github.com/Arlan-Z/def-betting-api/internal/repositories/event/sqlite"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type EventRepositorySuite struct {
	suite.Suite
	db      *sqlx.DB
	repo    *eventrepo.EventRepository
	dbPath  string
	migrate *migrate.Migrate
}

func (s *EventRepositorySuite) SetupSuite() {
	tempFile, err := os.CreateTemp("", "test_events_*.db")
	require.NoError(s.T(), err)
	s.dbPath = tempFile.Name()
	tempFile.Close()

	db, err := sqlx.Open("sqlite3", s.dbPath+"?_foreign_keys=on")
	require.NoError(s.T(), err)
	s.db = db

	driver, err := sqlite3.WithInstance(db.DB, &sqlite3.Config{})
	require.NoError(s.T(), err)

	migrationsPath := "../../../../migrations" // ADJUST THIS PATH IF NEEDED
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"sqlite3", driver)
	require.NoError(s.T(), err)
	s.migrate = m

	err = s.migrate.Up()
	require.NoError(s.T(), err, "Failed to run migrations UP")

	s.repo = eventrepo.NewEventRepository(s.db)
}

func (s *EventRepositorySuite) TearDownSuite() {
	if s.migrate != nil {
		err := s.migrate.Down()
		// Use errors.Is for newer Go versions if preferred
		if err != nil && err.Error() != migrate.ErrNoChange.Error() {
			s.T().Logf("Warning: failed to run migrations DOWN: %v", err)
		}
		sourceErr, dbErr := s.migrate.Close()
		if sourceErr != nil {
			s.T().Logf("Warning: failed to close migrate source: %v", sourceErr)
		}
		if dbErr != nil {
			s.T().Logf("Warning: failed to close migrate db instance: %v", dbErr)
		}
	}

	if s.db != nil {
		err := s.db.Close()
		require.NoError(s.T(), err)
	}
	err := os.Remove(s.dbPath)
	require.NoError(s.T(), err)
}

func (s *EventRepositorySuite) BeforeTest(suiteName, testName string) {
	_, err := s.db.Exec("DELETE FROM bets;")
	require.NoError(s.T(), err)
	_, err = s.db.Exec("DELETE FROM events;")
	require.NoError(s.T(), err)
}

func TestEventRepositorySuite(t *testing.T) {
	suite.Run(t, new(EventRepositorySuite))
}

func (s *EventRepositorySuite) TestFindByID_NotFound() {
	ctx := context.Background()
	event, err := s.repo.FindByID(ctx, uuid.NewString())

	require.NoError(s.T(), err)
	require.Nil(s.T(), event, "Expected nil for non-existent event")
}

func (s *EventRepositorySuite) TestUpsertAndFindByID_Found() {
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	expectedEvent := &data.Event{
		ID:             uuid.NewString(),
		EventName:      "Test Insert Event",
		HomeTeam:       "Team A",
		AwayTeam:       "Team B",
		HomeWinChance:  1.5,
		AwayWinChance:  3.0,
		DrawChance:     2.5,
		EventStartDate: now.Add(-1 * time.Hour),
		EventEndDate:   now.Add(1 * time.Hour),
		EventResult:    nil,
		Type:           "Test",
		IsActive:       true,
	}

	err := s.repo.Upsert(ctx, expectedEvent)
	require.NoError(s.T(), err, "Upsert (insert) should not return error")

	foundEvent, err := s.repo.FindByID(ctx, expectedEvent.ID)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), foundEvent, "Event should be found after Upsert")

	require.Equal(s.T(), expectedEvent.ID, foundEvent.ID)
	require.Equal(s.T(), expectedEvent.EventName, foundEvent.EventName)
	require.Equal(s.T(), expectedEvent.HomeTeam, foundEvent.HomeTeam)
	require.Equal(s.T(), expectedEvent.AwayTeam, foundEvent.AwayTeam)
	require.Equal(s.T(), expectedEvent.HomeWinChance, foundEvent.HomeWinChance)
	require.Equal(s.T(), expectedEvent.AwayWinChance, foundEvent.AwayWinChance)
	require.Equal(s.T(), expectedEvent.DrawChance, foundEvent.DrawChance)
	require.WithinDuration(s.T(), expectedEvent.EventStartDate, foundEvent.EventStartDate, time.Second)
	require.WithinDuration(s.T(), expectedEvent.EventEndDate, foundEvent.EventEndDate, time.Second)
	require.Nil(s.T(), foundEvent.EventResult)
	require.Equal(s.T(), expectedEvent.Type, foundEvent.Type)
	require.Equal(s.T(), expectedEvent.IsActive, foundEvent.IsActive)

	expectedEvent.EventName = "Updated Test Event"
	expectedEvent.HomeWinChance = 1.8
	expectedEvent.IsActive = false
	drawResult := data.Draw
	expectedEvent.EventResult = &drawResult

	err = s.repo.Upsert(ctx, expectedEvent)
	require.NoError(s.T(), err, "Upsert (update) should not return error")

	updatedEvent, err := s.repo.FindByID(ctx, expectedEvent.ID)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), updatedEvent)

	require.Equal(s.T(), "Updated Test Event", updatedEvent.EventName)
	require.Equal(s.T(), 1.8, updatedEvent.HomeWinChance)
	require.False(s.T(), updatedEvent.IsActive)
	require.NotNil(s.T(), updatedEvent.EventResult)
	require.Equal(s.T(), data.Draw, *updatedEvent.EventResult)
}

func (s *EventRepositorySuite) TestFindActiveEvents() {
	ctx := context.Background()
	now := time.Now().UTC()

	activeFutureEvent := &data.Event{ID: uuid.NewString(), EventName: "Active Future", EventEndDate: now.Add(1 * time.Hour), IsActive: true}
	activePastEvent := &data.Event{ID: uuid.NewString(), EventName: "Active Past", EventEndDate: now.Add(-1 * time.Hour), IsActive: true}
	inactiveFutureEvent := &data.Event{ID: uuid.NewString(), EventName: "Inactive Future", EventEndDate: now.Add(1 * time.Hour), IsActive: false}

	err := s.repo.Upsert(ctx, activeFutureEvent)
	require.NoError(s.T(), err)
	err = s.repo.Upsert(ctx, activePastEvent)
	require.NoError(s.T(), err)
	err = s.repo.Upsert(ctx, inactiveFutureEvent)
	require.NoError(s.T(), err)

	activeEvents, err := s.repo.FindActiveEvents(ctx)
	require.NoError(s.T(), err)

	require.Len(s.T(), activeEvents, 1, "Should find only one active future event")
	require.Equal(s.T(), activeFutureEvent.ID, activeEvents[0].ID)
}

func (s *EventRepositorySuite) TestFindActiveEvents_NoneFound() {
	ctx := context.Background()
	now := time.Now().UTC()

	activePastEvent := &data.Event{ID: uuid.NewString(), EventName: "Active Past", EventEndDate: now.Add(-1 * time.Hour), IsActive: true}
	inactiveFutureEvent := &data.Event{ID: uuid.NewString(), EventName: "Inactive Future", EventEndDate: now.Add(1 * time.Hour), IsActive: false}

	err := s.repo.Upsert(ctx, activePastEvent)
	require.NoError(s.T(), err)
	err = s.repo.Upsert(ctx, inactiveFutureEvent)
	require.NoError(s.T(), err)

	activeEvents, err := s.repo.FindActiveEvents(ctx)
	require.NoError(s.T(), err)

	require.Empty(s.T(), activeEvents, "Should find no active events")
}

func (s *EventRepositorySuite) TestUpdateResultAndStatus() {
	ctx := context.Background()
	activeEvent := &data.Event{ID: uuid.NewString(), EventName: "To Finalize", EventEndDate: time.Now().Add(1 * time.Hour), IsActive: true}
	err := s.repo.Upsert(ctx, activeEvent)
	require.NoError(s.T(), err)

	result := data.HomeWin
	err = s.repo.UpdateResultAndStatus(ctx, activeEvent.ID, result)
	require.NoError(s.T(), err)

	updatedEvent, err := s.repo.FindByID(ctx, activeEvent.ID)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), updatedEvent)
	require.False(s.T(), updatedEvent.IsActive, "Event should become inactive")
	require.NotNil(s.T(), updatedEvent.EventResult, "Result should be set")
	require.Equal(s.T(), result, *updatedEvent.EventResult)

	err = s.repo.UpdateResultAndStatus(ctx, activeEvent.ID, data.AwayWin)
	require.NoError(s.T(), err)

	finalEvent, err := s.repo.FindByID(ctx, activeEvent.ID)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), finalEvent)
	require.Equal(s.T(), result, *finalEvent.EventResult)
}
