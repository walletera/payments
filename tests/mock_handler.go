// Code generated by mockery v2.42.2. DO NOT EDIT.

package tests

import (
	context "context"

	api "github.com/walletera/payments/api"

	mock "github.com/stretchr/testify/mock"
)

// MockHandler is an autogenerated mock type for the Handler type
type MockHandler struct {
	mock.Mock
}

type MockHandler_Expecter struct {
	mock *mock.Mock
}

func (_m *MockHandler) EXPECT() *MockHandler_Expecter {
	return &MockHandler_Expecter{mock: &_m.Mock}
}

// PatchWithdrawal provides a mock function with given fields: ctx, req, params
func (_m *MockHandler) PatchWithdrawal(ctx context.Context, req *api.WithdrawalPatchBody, params api.PatchWithdrawalParams) (api.PatchWithdrawalRes, error) {
	ret := _m.Called(ctx, req, params)

	if len(ret) == 0 {
		panic("no return value specified for PatchWithdrawal")
	}

	var r0 api.PatchWithdrawalRes
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *api.WithdrawalPatchBody, api.PatchWithdrawalParams) (api.PatchWithdrawalRes, error)); ok {
		return rf(ctx, req, params)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *api.WithdrawalPatchBody, api.PatchWithdrawalParams) api.PatchWithdrawalRes); ok {
		r0 = rf(ctx, req, params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(api.PatchWithdrawalRes)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *api.WithdrawalPatchBody, api.PatchWithdrawalParams) error); ok {
		r1 = rf(ctx, req, params)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockHandler_PatchWithdrawal_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PatchWithdrawal'
type MockHandler_PatchWithdrawal_Call struct {
	*mock.Call
}

// PatchWithdrawal is a helper method to define mock.On call
//   - ctx context.Context
//   - req *api.WithdrawalPatchBody
//   - params api.PatchWithdrawalParams
func (_e *MockHandler_Expecter) PatchWithdrawal(ctx interface{}, req interface{}, params interface{}) *MockHandler_PatchWithdrawal_Call {
	return &MockHandler_PatchWithdrawal_Call{Call: _e.mock.On("PatchWithdrawal", ctx, req, params)}
}

func (_c *MockHandler_PatchWithdrawal_Call) Run(run func(ctx context.Context, req *api.WithdrawalPatchBody, params api.PatchWithdrawalParams)) *MockHandler_PatchWithdrawal_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*api.WithdrawalPatchBody), args[2].(api.PatchWithdrawalParams))
	})
	return _c
}

func (_c *MockHandler_PatchWithdrawal_Call) Return(_a0 api.PatchWithdrawalRes, _a1 error) *MockHandler_PatchWithdrawal_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockHandler_PatchWithdrawal_Call) RunAndReturn(run func(context.Context, *api.WithdrawalPatchBody, api.PatchWithdrawalParams) (api.PatchWithdrawalRes, error)) *MockHandler_PatchWithdrawal_Call {
	_c.Call.Return(run)
	return _c
}

// PostDeposit provides a mock function with given fields: ctx, req
func (_m *MockHandler) PostDeposit(ctx context.Context, req *api.DepositPostBody) (api.PostDepositRes, error) {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for PostDeposit")
	}

	var r0 api.PostDepositRes
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *api.DepositPostBody) (api.PostDepositRes, error)); ok {
		return rf(ctx, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *api.DepositPostBody) api.PostDepositRes); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(api.PostDepositRes)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *api.DepositPostBody) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockHandler_PostDeposit_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PostDeposit'
type MockHandler_PostDeposit_Call struct {
	*mock.Call
}

// PostDeposit is a helper method to define mock.On call
//   - ctx context.Context
//   - req *api.DepositPostBody
func (_e *MockHandler_Expecter) PostDeposit(ctx interface{}, req interface{}) *MockHandler_PostDeposit_Call {
	return &MockHandler_PostDeposit_Call{Call: _e.mock.On("PostDeposit", ctx, req)}
}

func (_c *MockHandler_PostDeposit_Call) Run(run func(ctx context.Context, req *api.DepositPostBody)) *MockHandler_PostDeposit_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*api.DepositPostBody))
	})
	return _c
}

func (_c *MockHandler_PostDeposit_Call) Return(_a0 api.PostDepositRes, _a1 error) *MockHandler_PostDeposit_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockHandler_PostDeposit_Call) RunAndReturn(run func(context.Context, *api.DepositPostBody) (api.PostDepositRes, error)) *MockHandler_PostDeposit_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockHandler creates a new instance of MockHandler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockHandler {
	mock := &MockHandler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
