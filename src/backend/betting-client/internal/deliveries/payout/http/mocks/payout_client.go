package mocks

import (
	"context"

	"github.com/Arlan-Z/def-betting-api/internal/data"
	"github.com/stretchr/testify/mock"
)

type PayoutClient struct {
	mock.Mock
}

func (_m *PayoutClient) NotifyPayout(ctx context.Context, notification data.PayoutNotification) error {
	ret := _m.Called(ctx, notification)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, data.PayoutNotification) error); ok {
		r0 = rf(ctx, notification)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

func NewPayoutClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *PayoutClient {
	mock := &PayoutClient{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
