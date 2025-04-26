package mocks

import (
	"context"

	"github.com/Arlan-Z/def-betting-api/internal/data"
	"github.com/stretchr/testify/mock"
)

type BetRepository struct {
	mock.Mock
}

func (_m *BetRepository) Save(ctx context.Context, bet *data.Bet) error {
	ret := _m.Called(ctx, bet)
	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *data.Bet) error); ok {
		r0 = rf(ctx, bet)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

func (_m *BetRepository) FindPendingByEventID(ctx context.Context, eventID string) ([]data.Bet, error) {
	ret := _m.Called(ctx, eventID)
	var r0 []data.Bet
	if rf, ok := ret.Get(0).(func(context.Context, string) []data.Bet); ok {
		r0 = rf(ctx, eventID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]data.Bet)
		}
	}
	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, eventID)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

func (_m *BetRepository) UpdateStatusAndPayout(ctx context.Context, betID string, status data.BetStatus, payout float64) error {
	ret := _m.Called(ctx, betID, status, payout)
	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, data.BetStatus, float64) error); ok {
		r0 = rf(ctx, betID, status, payout)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

func (_m *BetRepository) UpdateStatus(ctx context.Context, betID string, status data.BetStatus) error {
	ret := _m.Called(ctx, betID, status)
	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, data.BetStatus) error); ok {
		r0 = rf(ctx, betID, status)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

func NewBetRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *BetRepository {
	mock := &BetRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
