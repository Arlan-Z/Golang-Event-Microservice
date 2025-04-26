package mocks

import (
	"context"

	"github.com/Arlan-Z/def-betting-api/internal/data"
	"github.com/stretchr/testify/mock"
)

type EventRepository struct {
	mock.Mock
}

func (_m *EventRepository) FindActiveEvents(ctx context.Context) ([]data.Event, error) {
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

func (_m *EventRepository) FindByID(ctx context.Context, eventID string) (*data.Event, error) {
	ret := _m.Called(ctx, eventID)

	var r0 *data.Event
	if rf, ok := ret.Get(0).(func(context.Context, string) *data.Event); ok {
		r0 = rf(ctx, eventID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*data.Event)
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

func (_m *EventRepository) UpdateResultAndStatus(ctx context.Context, eventID string, result data.Outcome) error {
	ret := _m.Called(ctx, eventID, result)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, data.Outcome) error); ok {
		r0 = rf(ctx, eventID, result)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

func (_m *EventRepository) Upsert(ctx context.Context, event *data.Event) error {
	ret := _m.Called(ctx, event)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *data.Event) error); ok {
		r0 = rf(ctx, event)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

func NewEventRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *EventRepository {
	mock := &EventRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
