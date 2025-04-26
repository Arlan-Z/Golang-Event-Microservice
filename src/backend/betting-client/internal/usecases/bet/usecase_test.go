package bet_test

import (
	"context"
	"database/sql" // Added for sql.ErrNoRows
	"errors"
	"fmt" // Added for error message check
	"testing"
	"time"

	"github.com/Arlan-Z/def-betting-api/internal/data"
	repomocks "github.com/Arlan-Z/def-betting-api/internal/repositories/mocks"
	betuc "github.com/Arlan-Z/def-betting-api/internal/usecases/bet"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBetUseCase_PlaceBet_Success(t *testing.T) {
	mockBetRepo := repomocks.NewBetRepository(t)
	mockEventRepo := repomocks.NewEventRepository(t)
	logger := zap.NewNop()
	uc := betuc.NewUseCase(mockBetRepo, mockEventRepo, logger)

	ctx := context.Background()
	userID := uuid.NewString()
	eventID := uuid.NewString()
	now := time.Now()
	amount := 50.0
	predictedOutcome := data.Draw

	// FIX: EventStartDate must be in the future for the bet to be accepted
	activeEvent := &data.Event{
		ID:             eventID,
		IsActive:       true,
		EventStartDate: now.Add(time.Minute), // Starts in the future
		EventEndDate:   now.Add(time.Hour),   // Ends further in the future
		HomeWinChance:  1.8,
		AwayWinChance:  3.2,
		DrawChance:     2.9,
	}

	req := data.PlaceBetRequest{
		UserID:           userID,
		EventID:          eventID,
		Amount:           amount,
		PredictedOutcome: predictedOutcome,
	}

	mockEventRepo.On("FindByID", ctx, eventID).Return(activeEvent, nil).Once()

	var capturedBet *data.Bet
	mockBetRepo.On("Save", ctx, mock.MatchedBy(func(bet *data.Bet) bool {
		capturedBet = bet
		return bet.UserID == userID &&
			bet.EventID == eventID &&
			bet.Amount == amount &&
			bet.PredictedOutcome == predictedOutcome &&
			bet.Status == data.StatusPending &&
			bet.RecordedHomeWinChance == activeEvent.HomeWinChance &&
			bet.RecordedAwayWinChance == activeEvent.AwayWinChance &&
			bet.RecordedDrawChance == activeEvent.DrawChance &&
			!bet.PlacedAt.IsZero() &&
			bet.ID != ""
	})).Return(nil).Once()

	createdBet, err := uc.PlaceBet(ctx, req)

	// Now this require.NoError should pass
	require.NoError(t, err)
	require.NotNil(t, createdBet)
	require.Equal(t, capturedBet, createdBet)
	assert.Equal(t, userID, createdBet.UserID)
	assert.Equal(t, eventID, createdBet.EventID)
	assert.Equal(t, amount, createdBet.Amount)
	assert.Equal(t, predictedOutcome, createdBet.PredictedOutcome)
	assert.Equal(t, data.StatusPending, createdBet.Status)
	assert.Equal(t, activeEvent.HomeWinChance, createdBet.RecordedHomeWinChance)
	assert.Equal(t, activeEvent.AwayWinChance, createdBet.RecordedAwayWinChance)
	assert.Equal(t, activeEvent.DrawChance, createdBet.RecordedDrawChance)
	assert.NotZero(t, createdBet.ID)
	assert.NotZero(t, createdBet.PlacedAt)

	mockBetRepo.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
}

func TestBetUseCase_PlaceBet_EventNotFound(t *testing.T) {
	mockBetRepo := repomocks.NewBetRepository(t)
	mockEventRepo := repomocks.NewEventRepository(t)
	logger := zap.NewNop()
	uc := betuc.NewUseCase(mockBetRepo, mockEventRepo, logger)

	ctx := context.Background()
	eventID := uuid.NewString()
	req := data.PlaceBetRequest{EventID: eventID, UserID: uuid.NewString(), Amount: 10, PredictedOutcome: data.HomeWin}

	mockEventRepo.On("FindByID", ctx, eventID).Return(nil, nil).Once()

	createdBet, err := uc.PlaceBet(ctx, req)

	require.Error(t, err)
	require.Nil(t, createdBet)
	assert.True(t, errors.Is(err, betuc.ErrEventNotFound), "Expected error ErrEventNotFound")

	mockEventRepo.AssertExpectations(t)
	mockBetRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

func TestBetUseCase_PlaceBet_EventNotActive_Ended(t *testing.T) {
	mockBetRepo := repomocks.NewBetRepository(t)
	mockEventRepo := repomocks.NewEventRepository(t)
	logger := zap.NewNop()
	uc := betuc.NewUseCase(mockBetRepo, mockEventRepo, logger)

	ctx := context.Background()
	eventID := uuid.NewString()
	req := data.PlaceBetRequest{EventID: eventID, UserID: uuid.NewString(), Amount: 10, PredictedOutcome: data.HomeWin}

	endedEvent := &data.Event{
		ID:             eventID,
		IsActive:       true,
		EventStartDate: time.Now().Add(-2 * time.Hour), // Ensure start date is also in the past
		EventEndDate:   time.Now().Add(-time.Hour),
	}
	mockEventRepo.On("FindByID", ctx, eventID).Return(endedEvent, nil).Once()

	createdBet, err := uc.PlaceBet(ctx, req)

	require.Error(t, err)
	require.Nil(t, createdBet)
	assert.True(t, errors.Is(err, betuc.ErrEventNotActive), "Expected error ErrEventNotActive")

	mockEventRepo.AssertExpectations(t)
	mockBetRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

func TestBetUseCase_PlaceBet_EventNotActive_ExplicitFlag(t *testing.T) {
	mockBetRepo := repomocks.NewBetRepository(t)
	mockEventRepo := repomocks.NewEventRepository(t)
	logger := zap.NewNop()
	uc := betuc.NewUseCase(mockBetRepo, mockEventRepo, logger)

	ctx := context.Background()
	eventID := uuid.NewString()
	req := data.PlaceBetRequest{EventID: eventID, UserID: uuid.NewString(), Amount: 10, PredictedOutcome: data.HomeWin}

	inactiveEvent := &data.Event{
		ID:             eventID,
		IsActive:       false,                       // Explicitly inactive
		EventStartDate: time.Now().Add(time.Minute), // Dates don't matter if IsActive is false
		EventEndDate:   time.Now().Add(time.Hour),
	}
	mockEventRepo.On("FindByID", ctx, eventID).Return(inactiveEvent, nil).Once()

	createdBet, err := uc.PlaceBet(ctx, req)

	require.Error(t, err)
	require.Nil(t, createdBet)
	assert.True(t, errors.Is(err, betuc.ErrEventNotActive), "Expected error ErrEventNotActive")

	mockEventRepo.AssertExpectations(t)
	mockBetRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

func TestBetUseCase_PlaceBet_EventNotActive_Started(t *testing.T) {
	mockBetRepo := repomocks.NewBetRepository(t)
	mockEventRepo := repomocks.NewEventRepository(t)
	logger := zap.NewNop()
	uc := betuc.NewUseCase(mockBetRepo, mockEventRepo, logger)

	ctx := context.Background()
	eventID := uuid.NewString()
	req := data.PlaceBetRequest{EventID: eventID, UserID: uuid.NewString(), Amount: 10, PredictedOutcome: data.HomeWin}

	startedEvent := &data.Event{
		ID:             eventID,
		IsActive:       true,
		EventStartDate: time.Now().Add(-time.Minute), // Started a minute ago
		EventEndDate:   time.Now().Add(time.Hour),
	}
	mockEventRepo.On("FindByID", ctx, eventID).Return(startedEvent, nil).Once()

	createdBet, err := uc.PlaceBet(ctx, req)

	require.Error(t, err)
	require.Nil(t, createdBet)
	assert.True(t, errors.Is(err, betuc.ErrEventNotActive), "Expected error ErrEventNotActive because event started")

	mockEventRepo.AssertExpectations(t)
	mockBetRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

func TestBetUseCase_PlaceBet_SaveError(t *testing.T) {
	mockBetRepo := repomocks.NewBetRepository(t)
	mockEventRepo := repomocks.NewEventRepository(t)
	logger := zap.NewNop()
	uc := betuc.NewUseCase(mockBetRepo, mockEventRepo, logger)

	ctx := context.Background()
	eventID := uuid.NewString()
	now := time.Now()
	req := data.PlaceBetRequest{EventID: eventID, UserID: uuid.NewString(), Amount: 10, PredictedOutcome: data.HomeWin}

	// FIX: EventStartDate must be in the future
	activeEvent := &data.Event{
		ID:             eventID,
		IsActive:       true,
		EventStartDate: now.Add(time.Minute), // Starts in the future
		EventEndDate:   now.Add(time.Hour),
		HomeWinChance:  2.0, AwayWinChance: 3.0, DrawChance: 2.5,
	}
	saveError := errors.New("db connection lost")

	mockEventRepo.On("FindByID", ctx, eventID).Return(activeEvent, nil).Once()
	mockBetRepo.On("Save", ctx, mock.Anything).Return(saveError).Once() // Error on save

	createdBet, err := uc.PlaceBet(ctx, req)

	require.Error(t, err)
	require.Nil(t, createdBet)
	// Now this assertion should pass
	assert.True(t, errors.Is(err, betuc.ErrSavingBetFailed), "Expected error ErrSavingBetFailed")

	mockEventRepo.AssertExpectations(t)
	mockBetRepo.AssertExpectations(t)
}

func TestBetUseCase_PlaceBet_RecordsCorrectOdds(t *testing.T) {
	mockBetRepo := repomocks.NewBetRepository(t)
	mockEventRepo := repomocks.NewEventRepository(t)
	logger := zap.NewNop()
	uc := betuc.NewUseCase(mockBetRepo, mockEventRepo, logger)

	ctx := context.Background()
	eventID := uuid.NewString()
	now := time.Now()

	// FIX: EventStartDate must be in the future
	eventOdds1 := &data.Event{
		ID: eventID, IsActive: true, EventStartDate: now.Add(time.Minute), EventEndDate: now.Add(time.Hour),
		HomeWinChance: 1.5, AwayWinChance: 4.0, DrawChance: 3.0,
	}
	// FIX: EventStartDate must be in the future (can be the same as eventOdds1 for this test)
	eventOdds2 := &data.Event{
		ID: eventID, IsActive: true, EventStartDate: now.Add(time.Minute), EventEndDate: now.Add(time.Hour),
		HomeWinChance: 1.6, AwayWinChance: 3.8, DrawChance: 2.9, // Odds changed
	}

	req1 := data.PlaceBetRequest{EventID: eventID, UserID: uuid.NewString(), Amount: 10, PredictedOutcome: data.HomeWin}
	req2 := data.PlaceBetRequest{EventID: eventID, UserID: uuid.NewString(), Amount: 20, PredictedOutcome: data.AwayWin}

	// Bet 1 with odds 1
	mockEventRepo.On("FindByID", ctx, eventID).Return(eventOdds1, nil).Once()
	mockBetRepo.On("Save", ctx, mock.MatchedBy(func(b *data.Bet) bool {
		return b.RecordedHomeWinChance == 1.5 // Check recorded odds 1
	})).Return(nil).Once()
	_, err := uc.PlaceBet(ctx, req1)
	require.NoError(t, err) // Should pass now

	// Bet 2 with odds 2
	mockEventRepo.On("FindByID", ctx, eventID).Return(eventOdds2, nil).Once()
	mockBetRepo.On("Save", ctx, mock.MatchedBy(func(b *data.Bet) bool {
		return b.RecordedAwayWinChance == 3.8 // Check recorded odds 2
	})).Return(nil).Once()
	_, err = uc.PlaceBet(ctx, req2)
	require.NoError(t, err) // Should pass now

	mockEventRepo.AssertExpectations(t)
	mockBetRepo.AssertExpectations(t)
}

func TestBetUseCase_CancelBetsForEvent_Success(t *testing.T) {
	mockBetRepo := repomocks.NewBetRepository(t)
	mockEventRepo := repomocks.NewEventRepository(t)
	logger := zap.NewNop()
	uc := betuc.NewUseCase(mockBetRepo, mockEventRepo, logger)

	ctx := context.Background()
	eventID := uuid.NewString()
	betID1 := uuid.NewString()
	betID2 := uuid.NewString()

	pendingBets := []data.Bet{
		{ID: betID1, EventID: eventID, Status: data.StatusPending},
		{ID: betID2, EventID: eventID, Status: data.StatusPending},
	}

	mockBetRepo.On("FindPendingByEventID", ctx, eventID).Return(pendingBets, nil).Once()
	mockBetRepo.On("UpdateStatus", ctx, betID1, data.StatusCanceled).Return(nil).Once()
	mockBetRepo.On("UpdateStatus", ctx, betID2, data.StatusCanceled).Return(nil).Once()

	err := uc.CancelBetsForEvent(ctx, eventID)

	require.NoError(t, err)
	mockBetRepo.AssertExpectations(t)
}

func TestBetUseCase_CancelBetsForEvent_NoPendingBets(t *testing.T) {
	mockBetRepo := repomocks.NewBetRepository(t)
	mockEventRepo := repomocks.NewEventRepository(t)
	logger := zap.NewNop()
	uc := betuc.NewUseCase(mockBetRepo, mockEventRepo, logger)

	ctx := context.Background()
	eventID := uuid.NewString()
	pendingBets := []data.Bet{}

	mockBetRepo.On("FindPendingByEventID", ctx, eventID).Return(pendingBets, nil).Once()

	err := uc.CancelBetsForEvent(ctx, eventID)

	require.NoError(t, err)
	mockBetRepo.AssertExpectations(t)
	mockBetRepo.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything, mock.Anything)
}

func TestBetUseCase_CancelBetsForEvent_FindPendingErrorNoRows(t *testing.T) {
	mockBetRepo := repomocks.NewBetRepository(t)
	mockEventRepo := repomocks.NewEventRepository(t)
	logger := zap.NewNop()
	uc := betuc.NewUseCase(mockBetRepo, mockEventRepo, logger)

	ctx := context.Background()
	eventID := uuid.NewString()

	mockBetRepo.On("FindPendingByEventID", ctx, eventID).Return(nil, sql.ErrNoRows).Once()

	err := uc.CancelBetsForEvent(ctx, eventID)

	require.NoError(t, err) // Should not return error on sql.ErrNoRows
	mockBetRepo.AssertExpectations(t)
	mockBetRepo.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything, mock.Anything)
}

func TestBetUseCase_CancelBetsForEvent_UpdateError(t *testing.T) {
	mockBetRepo := repomocks.NewBetRepository(t)
	mockEventRepo := repomocks.NewEventRepository(t)
	logger := zap.NewNop()
	uc := betuc.NewUseCase(mockBetRepo, mockEventRepo, logger)

	ctx := context.Background()
	eventID := uuid.NewString()
	betID1 := uuid.NewString()
	betID2 := uuid.NewString()
	updateError := errors.New("failed to update status")

	pendingBets := []data.Bet{
		{ID: betID1, EventID: eventID, Status: data.StatusPending},
		{ID: betID2, EventID: eventID, Status: data.StatusPending},
	}

	mockBetRepo.On("FindPendingByEventID", ctx, eventID).Return(pendingBets, nil).Once()
	mockBetRepo.On("UpdateStatus", ctx, betID1, data.StatusCanceled).Return(nil).Once()
	mockBetRepo.On("UpdateStatus", ctx, betID2, data.StatusCanceled).Return(updateError).Once()

	err := uc.CancelBetsForEvent(ctx, eventID)

	require.Error(t, err)
	// FIX: Check the base error and optionally the count
	assert.True(t, errors.Is(err, betuc.ErrBetCancellationFailed))
	// Check the formatted string which includes the count
	assert.Contains(t, err.Error(), fmt.Sprintf("%d bets failed to cancel", 1))

	mockBetRepo.AssertExpectations(t)
}
