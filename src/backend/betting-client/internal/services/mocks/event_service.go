package mocks

import (
	"context"

	"github.com/Arlan-Z/def-betting-api/internal/data"
	"github.com/stretchr/testify/mock"
)

type EventService struct {
	mock.Mock
}

func (_m *EventService) GetActiveEvents(ctx context.Context) ([]data.Event, error) {
	ret := _m.Called(ctx)
	var r0 []data.Event
	if rf, ok := ret.Get(0).(func(context.Context) []data.Event); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]data.Event)
		}
	}
	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

func (_m *EventService) FinalizeEvent(ctx context.Context, eventID string, result data.Outcome) error {
	ret := _m.Called(ctx, eventID, result)
	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, data.Outcome) error); ok {
		r0 = rf(ctx, eventID, result)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

func NewEventService(t interface {
	mock.TestingT
	Cleanup(func())
}) *EventService {
	mock := &EventService{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
