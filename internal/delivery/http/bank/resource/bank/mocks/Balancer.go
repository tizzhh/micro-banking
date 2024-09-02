// Code generated by mockery v2.45.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// Balancer is an autogenerated mock type for the Balancer type
type Balancer struct {
	mock.Mock
}

// Deposit provides a mock function with given fields: ctx, email, amount
func (_m *Balancer) Deposit(ctx context.Context, email string, amount float32) (float32, error) {
	ret := _m.Called(ctx, email, amount)

	if len(ret) == 0 {
		panic("no return value specified for Deposit")
	}

	var r0 float32
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, float32) (float32, error)); ok {
		return rf(ctx, email, amount)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, float32) float32); ok {
		r0 = rf(ctx, email, amount)
	} else {
		r0 = ret.Get(0).(float32)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, float32) error); ok {
		r1 = rf(ctx, email, amount)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Withdraw provides a mock function with given fields: ctx, email, amount
func (_m *Balancer) Withdraw(ctx context.Context, email string, amount float32) (float32, error) {
	ret := _m.Called(ctx, email, amount)

	if len(ret) == 0 {
		panic("no return value specified for Withdraw")
	}

	var r0 float32
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, float32) (float32, error)); ok {
		return rf(ctx, email, amount)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, float32) float32); ok {
		r0 = rf(ctx, email, amount)
	} else {
		r0 = ret.Get(0).(float32)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, float32) error); ok {
		r1 = rf(ctx, email, amount)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewBalancer creates a new instance of Balancer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBalancer(t interface {
	mock.TestingT
	Cleanup(func())
}) *Balancer {
	mock := &Balancer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
