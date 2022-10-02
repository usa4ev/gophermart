// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/usa4ev/gophermart/internal/server (interfaces: Storage)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	orders "github.com/usa4ev/gophermart/internal/orders"
)

// MockStorage is a mock of Storage interface.
type MockStorage struct {
	ctrl     *gomock.Controller
	recorder *MockStorageMockRecorder
}

// MockStorageMockRecorder is the mock recorder for MockStorage.
type MockStorageMockRecorder struct {
	mock *MockStorage
}

// NewMockStorage creates a new mock instance.
func NewMockStorage(ctrl *gomock.Controller) *MockStorage {
	mock := &MockStorage{ctrl: ctrl}
	mock.recorder = &MockStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorage) EXPECT() *MockStorageMockRecorder {
	return m.recorder
}

// AddUser mocks base method.
func (m *MockStorage) AddUser(arg0 context.Context, arg1, arg2 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddUser", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddUser indicates an expected call of AddUser.
func (mr *MockStorageMockRecorder) AddUser(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddUser", reflect.TypeOf((*MockStorage)(nil).AddUser), arg0, arg1, arg2)
}

// GetPasswordHash mocks base method.
func (m *MockStorage) GetPasswordHash(arg0 context.Context, arg1 string) (string, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPasswordHash", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetPasswordHash indicates an expected call of GetPasswordHash.
func (mr *MockStorageMockRecorder) GetPasswordHash(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPasswordHash", reflect.TypeOf((*MockStorage)(nil).GetPasswordHash), arg0, arg1)
}

// LoadBalance mocks base method.
func (m *MockStorage) LoadBalance(arg0 context.Context, arg1 string) (float64, float64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LoadBalance", arg0, arg1)
	ret0, _ := ret[0].(float64)
	ret1, _ := ret[1].(float64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// LoadBalance indicates an expected call of LoadBalance.
func (mr *MockStorageMockRecorder) LoadBalance(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoadBalance", reflect.TypeOf((*MockStorage)(nil).LoadBalance), arg0, arg1)
}

// LoadOrders mocks base method.
func (m *MockStorage) LoadOrders(arg0 context.Context, arg1 string) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LoadOrders", arg0, arg1)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LoadOrders indicates an expected call of LoadOrders.
func (mr *MockStorageMockRecorder) LoadOrders(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoadOrders", reflect.TypeOf((*MockStorage)(nil).LoadOrders), arg0, arg1)
}

// LoadWithdrawals mocks base method.
func (m *MockStorage) LoadWithdrawals(arg0 context.Context, arg1 string) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LoadWithdrawals", arg0, arg1)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LoadWithdrawals indicates an expected call of LoadWithdrawals.
func (mr *MockStorageMockRecorder) LoadWithdrawals(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoadWithdrawals", reflect.TypeOf((*MockStorage)(nil).LoadWithdrawals), arg0, arg1)
}

// OrdersToProcess mocks base method.
func (m *MockStorage) OrdersToProcess(arg0 context.Context) (map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OrdersToProcess", arg0)
	ret0, _ := ret[0].(map[string]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OrdersToProcess indicates an expected call of OrdersToProcess.
func (mr *MockStorageMockRecorder) OrdersToProcess(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OrdersToProcess", reflect.TypeOf((*MockStorage)(nil).OrdersToProcess), arg0)
}

// StoreOrder mocks base method.
func (m *MockStorage) StoreOrder(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StoreOrder", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// StoreOrder indicates an expected call of StoreOrder.
func (mr *MockStorageMockRecorder) StoreOrder(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StoreOrder", reflect.TypeOf((*MockStorage)(nil).StoreOrder), arg0, arg1, arg2)
}

// UpdateBalances mocks base method.
func (m *MockStorage) UpdateBalances(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateBalances", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateBalances indicates an expected call of UpdateBalances.
func (mr *MockStorageMockRecorder) UpdateBalances(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateBalances", reflect.TypeOf((*MockStorage)(nil).UpdateBalances), arg0)
}

// UpdateStatuses mocks base method.
func (m *MockStorage) UpdateStatuses(arg0 context.Context, arg1 []orders.Status) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateStatuses", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateStatuses indicates an expected call of UpdateStatuses.
func (mr *MockStorageMockRecorder) UpdateStatuses(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateStatuses", reflect.TypeOf((*MockStorage)(nil).UpdateStatuses), arg0, arg1)
}

// UserExists mocks base method.
func (m *MockStorage) UserExists(arg0 context.Context, arg1 string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UserExists", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UserExists indicates an expected call of UserExists.
func (mr *MockStorageMockRecorder) UserExists(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UserExists", reflect.TypeOf((*MockStorage)(nil).UserExists), arg0, arg1)
}

// Withdraw mocks base method.
func (m *MockStorage) Withdraw(arg0 context.Context, arg1, arg2 string, arg3 float64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Withdraw", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// Withdraw indicates an expected call of Withdraw.
func (mr *MockStorageMockRecorder) Withdraw(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Withdraw", reflect.TypeOf((*MockStorage)(nil).Withdraw), arg0, arg1, arg2, arg3)
}
