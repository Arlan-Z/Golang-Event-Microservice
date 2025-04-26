package event_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Arlan-Z/def-betting-api/internal/data"
	payoutmocks "github.com/Arlan-Z/def-betting-api/internal/deliveries/payout/http/mocks"
	repomocks "github.com/Arlan-Z/def-betting-api/internal/repositories/mocks"
	eventuc "github.com/Arlan-Z/def-betting-api/internal/usecases/event"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestEventUseCase_GetActiveEvents(t *testing.T) {
	mockEventRepo := repomocks.NewEventRepository(t)
	mockBetRepo := repomocks.NewBetRepository(t)
	mockPayoutClient := payoutmocks.NewPayoutClient(t)
	logger := zap.NewNop()

	uc := eventuc.NewUseCase(mockEventRepo, mockBetRepo, mockPayoutClient, logger)

	ctx := context.Background()
	expectedEvents := []data.Event{
		{ID: uuid.NewString(), EventName: "Event 1", IsActive: true, EventEndDate: time.Now().Add(time.Hour)},
	}

	mockEventRepo.On("FindActiveEvents", ctx).Return(expectedEvents, nil).Once()

	events, err := uc.GetActiveEvents(ctx)

	require.NoError(t, err)
	require.Equal(t, expectedEvents, events)

	mockEventRepo.AssertExpectations(t)
}

func TestEventUseCase_GetActiveEvents_RepoError(t *testing.T) {
	mockEventRepo := repomocks.NewEventRepository(t)
	mockBetRepo := repomocks.NewBetRepository(t)
	mockPayoutClient := payoutmocks.NewPayoutClient(t)
	logger := zap.NewNop()

	uc := eventuc.NewUseCase(mockEventRepo, mockBetRepo, mockPayoutClient, logger)

	ctx := context.Background()
	repoError := errors.New("database is down")

	mockEventRepo.On("FindActiveEvents", ctx).Return(nil, repoError).Once()

	events, err := uc.GetActiveEvents(ctx)

	require.Error(t, err)
	require.Nil(t, events)
	assert.Contains(t, err.Error(), "failed to get list of active events")

	mockEventRepo.AssertExpectations(t)
}

func TestEventUseCase_FinalizeEvent_Success_HomeWin(t *testing.T) {
	mockEventRepo := repomocks.NewEventRepository(t)
	mockBetRepo := repomocks.NewBetRepository(t)
	mockPayoutClient := payoutmocks.NewPayoutClient(t)
	logger := zap.NewNop()

	uc := eventuc.NewUseCase(mockEventRepo, mockBetRepo, mockPayoutClient, logger)

	ctx := context.Background()
	eventID := uuid.NewString()
	userID := uuid.NewString()
	betIDWin := uuid.NewString()
	betIDLoss := uuid.NewString()
	actualResult := data.HomeWin

	activeEvent := &data.Event{ID: eventID, IsActive: true, EventResult: nil}
	winningBet := data.Bet{
		ID:                    betIDWin,
		UserID:                userID,
		EventID:               eventID,
		Amount:                10.0,
		PredictedOutcome:      data.HomeWin,
		RecordedHomeWinChance: 2.0,
		Status:                data.StatusPending,
	}
	losingBet := data.Bet{
		ID:               betIDLoss,
		UserID:           uuid.NewString(),
		EventID:          eventID,
		Amount:           5.0,
		PredictedOutcome: data.AwayWin,
		Status:           data.StatusPending,
	}
	pendingBets := []data.Bet{winningBet, losingBet}
	expectedPayout := 20.0

	mockEventRepo.On("FindByID", ctx, eventID).Return(activeEvent, nil).Once()
	mockBetRepo.On("FindPendingByEventID", ctx, eventID).Return(pendingBets, nil).Once()
	mockBetRepo.On("UpdateStatusAndPayout", ctx, betIDWin, data.StatusWon, expectedPayout).Return(nil).Once()
	mockPayoutClient.On("NotifyPayout", ctx, data.PayoutNotification{UserID: userID, Amount: expectedPayout}).Return(nil).Once()
	mockBetRepo.On("UpdateStatus", ctx, betIDWin, data.StatusPaid).Return(nil).Once()
	mockBetRepo.On("UpdateStatusAndPayout", ctx, betIDLoss, data.StatusLost, 0.0).Return(nil).Once()
	mockEventRepo.On("UpdateResultAndStatus", ctx, eventID, actualResult).Return(nil).Once()

	err := uc.FinalizeEvent(ctx, eventID, actualResult)

	require.NoError(t, err)
	mockEventRepo.AssertExpectations(t)
	mockBetRepo.AssertExpectations(t)
	mockPayoutClient.AssertExpectations(t)
}

func TestEventUseCase_FinalizeEvent_PayoutFailed(t *testing.T) {
	mockEventRepo := repomocks.NewEventRepository(t)
	mockBetRepo := repomocks.NewBetRepository(t)
	mockPayoutClient := payoutmocks.NewPayoutClient(t)
	logger := zap.NewNop()

	uc := eventuc.NewUseCase(mockEventRepo, mockBetRepo, mockPayoutClient, logger)

	ctx := context.Background()
	eventID := uuid.NewString()
	betIDWin := uuid.NewString()
	actualResult := data.HomeWin

	activeEvent := &data.Event{ID: eventID, IsActive: true, EventResult: nil}
	winningBet := data.Bet{
		ID:                    betIDWin,
		UserID:                uuid.NewString(),
		EventID:               eventID,
		Amount:                10.0,
		PredictedOutcome:      data.HomeWin,
		RecordedHomeWinChance: 2.0,
		Status:                data.StatusPending,
	}
	pendingBets := []data.Bet{winningBet}
	expectedPayout := 20.0
	payoutError := errors.New("payout service unavailable")

	mockEventRepo.On("FindByID", ctx, eventID).Return(activeEvent, nil).Once()
	mockBetRepo.On("FindPendingByEventID", ctx, eventID).Return(pendingBets, nil).Once()
	mockBetRepo.On("UpdateStatusAndPayout", ctx, betIDWin, data.StatusWon, expectedPayout).Return(nil).Once()
	mockPayoutClient.On("NotifyPayout", ctx, mock.Anything).Return(payoutError).Once()
	mockBetRepo.On("UpdateStatus", ctx, betIDWin, data.StatusFailed).Return(nil).Once()
	mockEventRepo.On("UpdateResultAndStatus", ctx, eventID, actualResult).Return(nil).Once()

	err := uc.FinalizeEvent(ctx, eventID, actualResult)

	require.Error(t, err)
	assert.Contains(t, err.Error(), eventuc.ErrPayoutNotificationFailed.Error())
	assert.Contains(t, err.Error(), payoutError.Error())

	mockEventRepo.AssertExpectations(t)
	mockBetRepo.AssertExpectations(t)
	mockPayoutClient.AssertExpectations(t)
}

func TestEventUseCase_FinalizeEvent_AlreadyFinalized(t *testing.T) {
	mockEventRepo := repomocks.NewEventRepository(t)
	mockBetRepo := repomocks.NewBetRepository(t)
	mockPayoutClient := payoutmocks.NewPayoutClient(t)
	logger := zap.NewNop()

	uc := eventuc.NewUseCase(mockEventRepo, mockBetRepo, mockPayoutClient, logger)

	ctx := context.Background()
	eventID := uuid.NewString()
	existingResult := data.AwayWin
	finalizedEvent := &data.Event{ID: eventID, IsActive: false, EventResult: &existingResult}

	mockEventRepo.On("FindByID", ctx, eventID).Return(finalizedEvent, nil).Once()

	err := uc.FinalizeEvent(ctx, eventID, data.HomeWin)

	require.Error(t, err)
	require.True(t, errors.Is(err, eventuc.ErrEventAlreadyFinalized), "Expected error ErrEventAlreadyFinalized")

	mockEventRepo.AssertExpectations(t)
	mockBetRepo.AssertNotCalled(t, "FindPendingByEventID", mock.Anything, mock.Anything)
	mockPayoutClient.AssertNotCalled(t, "NotifyPayout", mock.Anything, mock.Anything)
}

// --- New Test Cases ---

func TestEventUseCase_FinalizeEvent_ErrorEventRepoFindByID(t *testing.T) {
	mockEventRepo := repomocks.NewEventRepository(t)
	mockBetRepo := repomocks.NewBetRepository(t)
	mockPayoutClient := payoutmocks.NewPayoutClient(t)
	logger := zap.NewNop()

	uc := eventuc.NewUseCase(mockEventRepo, mockBetRepo, mockPayoutClient, logger)

	ctx := context.Background()
	eventID := uuid.NewString()
	actualResult := data.Draw
	findError := errors.New("failed to find event")

	mockEventRepo.On("FindByID", ctx, eventID).Return(nil, findError).Once()

	err := uc.FinalizeEvent(ctx, eventID, actualResult)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "internal error searching for event")

	mockEventRepo.AssertExpectations(t)
	mockBetRepo.AssertNotCalled(t, "FindPendingByEventID", mock.Anything, mock.Anything)
	mockPayoutClient.AssertNotCalled(t, "NotifyPayout", mock.Anything, mock.Anything)
	mockEventRepo.AssertNotCalled(t, "UpdateResultAndStatus", mock.Anything, mock.Anything, mock.Anything)
}

func TestEventUseCase_FinalizeEvent_ErrorBetRepoFindPending(t *testing.T) {
	mockEventRepo := repomocks.NewEventRepository(t)
	mockBetRepo := repomocks.NewBetRepository(t)
	mockPayoutClient := payoutmocks.NewPayoutClient(t)
	logger := zap.NewNop()

	uc := eventuc.NewUseCase(mockEventRepo, mockBetRepo, mockPayoutClient, logger)

	ctx := context.Background()
	eventID := uuid.NewString()
	actualResult := data.Draw
	activeEvent := &data.Event{ID: eventID, IsActive: true, EventResult: nil}
	findBetsError := errors.New("failed to find pending bets")

	mockEventRepo.On("FindByID", ctx, eventID).Return(activeEvent, nil).Once()
	mockBetRepo.On("FindPendingByEventID", ctx, eventID).Return(nil, findBetsError).Once()

	err := uc.FinalizeEvent(ctx, eventID, actualResult)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "internal error retrieving bets")

	mockEventRepo.AssertExpectations(t)
	mockBetRepo.AssertExpectations(t)
	mockPayoutClient.AssertNotCalled(t, "NotifyPayout", mock.Anything, mock.Anything)
	mockBetRepo.AssertNotCalled(t, "UpdateStatusAndPayout", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mockEventRepo.AssertNotCalled(t, "UpdateResultAndStatus", mock.Anything, mock.Anything, mock.Anything)
}

func TestEventUseCase_FinalizeEvent_ErrorEventRepoUpdateStatus(t *testing.T) {
	mockEventRepo := repomocks.NewEventRepository(t)
	mockBetRepo := repomocks.NewBetRepository(t)
	mockPayoutClient := payoutmocks.NewPayoutClient(t)
	logger := zap.NewNop()

	uc := eventuc.NewUseCase(mockEventRepo, mockBetRepo, mockPayoutClient, logger)

	ctx := context.Background()
	eventID := uuid.NewString()
	actualResult := data.Draw
	activeEvent := &data.Event{ID: eventID, IsActive: true, EventResult: nil}
	pendingBets := []data.Bet{} // No bets to process
	updateEventError := errors.New("failed to update event status")

	mockEventRepo.On("FindByID", ctx, eventID).Return(activeEvent, nil).Once()
	mockBetRepo.On("FindPendingByEventID", ctx, eventID).Return(pendingBets, nil).Once()
	mockEventRepo.On("UpdateResultAndStatus", ctx, eventID, actualResult).Return(updateEventError).Once()

	err := uc.FinalizeEvent(ctx, eventID, actualResult)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to finalize event")
	assert.Contains(t, err.Error(), updateEventError.Error())

	mockEventRepo.AssertExpectations(t)
	mockBetRepo.AssertExpectations(t)
	mockPayoutClient.AssertNotCalled(t, "NotifyPayout", mock.Anything, mock.Anything)
}

func TestEventUseCase_FinalizeEvent_ErrorBetRepoUpdateStatusAndPayout(t *testing.T) {
	mockEventRepo := repomocks.NewEventRepository(t)
	mockBetRepo := repomocks.NewBetRepository(t)
	mockPayoutClient := payoutmocks.NewPayoutClient(t)
	logger := zap.NewNop()

	uc := eventuc.NewUseCase(mockEventRepo, mockBetRepo, mockPayoutClient, logger)

	ctx := context.Background()
	eventID := uuid.NewString()
	betIDLoss := uuid.NewString()
	actualResult := data.HomeWin
	activeEvent := &data.Event{ID: eventID, IsActive: true, EventResult: nil}
	losingBet := data.Bet{
		ID:               betIDLoss,
		UserID:           uuid.NewString(),
		EventID:          eventID,
		Amount:           5.0,
		PredictedOutcome: data.AwayWin, // Losing bet
		Status:           data.StatusPending,
	}
	pendingBets := []data.Bet{losingBet}
	updateBetError := errors.New("failed to update bet status/payout")

	mockEventRepo.On("FindByID", ctx, eventID).Return(activeEvent, nil).Once()
	mockBetRepo.On("FindPendingByEventID", ctx, eventID).Return(pendingBets, nil).Once()
	mockBetRepo.On("UpdateStatusAndPayout", ctx, betIDLoss, data.StatusLost, 0.0).Return(updateBetError).Once()
	mockEventRepo.On("UpdateResultAndStatus", ctx, eventID, actualResult).Return(nil).Once() // Event status update should still happen

	err := uc.FinalizeEvent(ctx, eventID, actualResult)

	require.Error(t, err)
	assert.Contains(t, err.Error(), eventuc.ErrBetUpdateFailed.Error())
	assert.Contains(t, err.Error(), updateBetError.Error())

	mockEventRepo.AssertExpectations(t)
	mockBetRepo.AssertExpectations(t)
	mockPayoutClient.AssertNotCalled(t, "NotifyPayout", mock.Anything, mock.Anything)
}

func TestEventUseCase_FinalizeEvent_ErrorBetRepoUpdateStatusToPaid(t *testing.T) {
	mockEventRepo := repomocks.NewEventRepository(t)
	mockBetRepo := repomocks.NewBetRepository(t)
	mockPayoutClient := payoutmocks.NewPayoutClient(t)
	logger := zap.NewNop()

	uc := eventuc.NewUseCase(mockEventRepo, mockBetRepo, mockPayoutClient, logger)

	ctx := context.Background()
	eventID := uuid.NewString()
	userID := uuid.NewString()
	betIDWin := uuid.NewString()
	actualResult := data.HomeWin
	activeEvent := &data.Event{ID: eventID, IsActive: true, EventResult: nil}
	winningBet := data.Bet{
		ID:                    betIDWin,
		UserID:                userID,
		EventID:               eventID,
		Amount:                10.0,
		PredictedOutcome:      data.HomeWin,
		RecordedHomeWinChance: 2.0,
		Status:                data.StatusPending,
	}
	pendingBets := []data.Bet{winningBet}
	expectedPayout := 20.0
	updateStatusError := errors.New("failed to update status to paid")

	mockEventRepo.On("FindByID", ctx, eventID).Return(activeEvent, nil).Once()
	mockBetRepo.On("FindPendingByEventID", ctx, eventID).Return(pendingBets, nil).Once()
	mockBetRepo.On("UpdateStatusAndPayout", ctx, betIDWin, data.StatusWon, expectedPayout).Return(nil).Once()
	mockPayoutClient.On("NotifyPayout", ctx, data.PayoutNotification{UserID: userID, Amount: expectedPayout}).Return(nil).Once()
	mockBetRepo.On("UpdateStatus", ctx, betIDWin, data.StatusPaid).Return(updateStatusError).Once() // Error here
	mockEventRepo.On("UpdateResultAndStatus", ctx, eventID, actualResult).Return(nil).Once()        // Event status update should still happen

	err := uc.FinalizeEvent(ctx, eventID, actualResult)

	require.Error(t, err)
	assert.Contains(t, err.Error(), eventuc.ErrBetUpdateFailed.Error())
	assert.Contains(t, err.Error(), updateStatusError.Error())

	mockEventRepo.AssertExpectations(t)
	mockBetRepo.AssertExpectations(t)
	mockPayoutClient.AssertExpectations(t)
}

func TestEventUseCase_FinalizeEvent_ErrorBetRepoUpdateStatusToFailed(t *testing.T) {
	mockEventRepo := repomocks.NewEventRepository(t)
	mockBetRepo := repomocks.NewBetRepository(t)
	mockPayoutClient := payoutmocks.NewPayoutClient(t)
	logger := zap.NewNop()

	uc := eventuc.NewUseCase(mockEventRepo, mockBetRepo, mockPayoutClient, logger)

	ctx := context.Background()
	eventID := uuid.NewString()
	betIDWin := uuid.NewString()
	actualResult := data.HomeWin
	activeEvent := &data.Event{ID: eventID, IsActive: true, EventResult: nil}
	winningBet := data.Bet{
		ID:                    betIDWin,
		UserID:                uuid.NewString(),
		EventID:               eventID,
		Amount:                10.0,
		PredictedOutcome:      data.HomeWin,
		RecordedHomeWinChance: 2.0,
		Status:                data.StatusPending,
	}
	pendingBets := []data.Bet{winningBet}
	expectedPayout := 20.0
	payoutError := errors.New("payout service unavailable")
	updateStatusError := errors.New("failed to update status to failed") // Error here

	mockEventRepo.On("FindByID", ctx, eventID).Return(activeEvent, nil).Once()
	mockBetRepo.On("FindPendingByEventID", ctx, eventID).Return(pendingBets, nil).Once()
	mockBetRepo.On("UpdateStatusAndPayout", ctx, betIDWin, data.StatusWon, expectedPayout).Return(nil).Once()
	mockPayoutClient.On("NotifyPayout", ctx, mock.Anything).Return(payoutError).Once()
	mockBetRepo.On("UpdateStatus", ctx, betIDWin, data.StatusFailed).Return(updateStatusError).Once() // Error here
	mockEventRepo.On("UpdateResultAndStatus", ctx, eventID, actualResult).Return(nil).Once()          // Event status update should still happen

	err := uc.FinalizeEvent(ctx, eventID, actualResult)

	require.Error(t, err)
	// Should contain both the payout error and the status update error
	assert.Contains(t, err.Error(), eventuc.ErrPayoutNotificationFailed.Error())
	assert.Contains(t, err.Error(), eventuc.ErrBetUpdateFailed.Error())
	assert.Contains(t, err.Error(), payoutError.Error())
	assert.Contains(t, err.Error(), updateStatusError.Error())

	mockEventRepo.AssertExpectations(t)
	mockBetRepo.AssertExpectations(t)
	mockPayoutClient.AssertExpectations(t)
}

func TestEventUseCase_FinalizeEvent_NoPendingBets(t *testing.T) {
	mockEventRepo := repomocks.NewEventRepository(t)
	mockBetRepo := repomocks.NewBetRepository(t)
	mockPayoutClient := payoutmocks.NewPayoutClient(t)
	logger := zap.NewNop()

	uc := eventuc.NewUseCase(mockEventRepo, mockBetRepo, mockPayoutClient, logger)

	ctx := context.Background()
	eventID := uuid.NewString()
	actualResult := data.Draw
	activeEvent := &data.Event{ID: eventID, IsActive: true, EventResult: nil}
	pendingBets := []data.Bet{} // Empty slice

	mockEventRepo.On("FindByID", ctx, eventID).Return(activeEvent, nil).Once()
	mockBetRepo.On("FindPendingByEventID", ctx, eventID).Return(pendingBets, nil).Once()
	mockEventRepo.On("UpdateResultAndStatus", ctx, eventID, actualResult).Return(nil).Once()

	err := uc.FinalizeEvent(ctx, eventID, actualResult)

	require.NoError(t, err)
	mockEventRepo.AssertExpectations(t)
	mockBetRepo.AssertExpectations(t)
	mockPayoutClient.AssertNotCalled(t, "NotifyPayout", mock.Anything, mock.Anything)
	mockBetRepo.AssertNotCalled(t, "UpdateStatusAndPayout", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mockBetRepo.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything, mock.Anything)
}

// OMG THIS IS SO LONG
