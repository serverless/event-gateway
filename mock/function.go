// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/serverless/event-gateway/function (interfaces: Service)

// Package mock is a generated GoMock package.
package mock

import (
	gomock "github.com/golang/mock/gomock"
	function "github.com/serverless/event-gateway/function"
	reflect "reflect"
)

// MockFunctionService is a mock of Service interface
type MockFunctionService struct {
	ctrl     *gomock.Controller
	recorder *MockFunctionServiceMockRecorder
}

// MockFunctionServiceMockRecorder is the mock recorder for MockFunctionService
type MockFunctionServiceMockRecorder struct {
	mock *MockFunctionService
}

// NewMockFunctionService creates a new mock instance
func NewMockFunctionService(ctrl *gomock.Controller) *MockFunctionService {
	mock := &MockFunctionService{ctrl: ctrl}
	mock.recorder = &MockFunctionServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockFunctionService) EXPECT() *MockFunctionServiceMockRecorder {
	return m.recorder
}

// DeleteFunction mocks base method
func (m *MockFunctionService) DeleteFunction(arg0 string, arg1 function.ID) error {
	ret := m.ctrl.Call(m, "DeleteFunction", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteFunction indicates an expected call of DeleteFunction
func (mr *MockFunctionServiceMockRecorder) DeleteFunction(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteFunction", reflect.TypeOf((*MockFunctionService)(nil).DeleteFunction), arg0, arg1)
}

// GetFunction mocks base method
func (m *MockFunctionService) GetFunction(arg0 string, arg1 function.ID) (*function.Function, error) {
	ret := m.ctrl.Call(m, "GetFunction", arg0, arg1)
	ret0, _ := ret[0].(*function.Function)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFunction indicates an expected call of GetFunction
func (mr *MockFunctionServiceMockRecorder) GetFunction(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFunction", reflect.TypeOf((*MockFunctionService)(nil).GetFunction), arg0, arg1)
}

// GetFunctions mocks base method
func (m *MockFunctionService) GetFunctions(arg0 string) ([]*function.Function, error) {
	ret := m.ctrl.Call(m, "GetFunctions", arg0)
	ret0, _ := ret[0].([]*function.Function)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFunctions indicates an expected call of GetFunctions
func (mr *MockFunctionServiceMockRecorder) GetFunctions(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFunctions", reflect.TypeOf((*MockFunctionService)(nil).GetFunctions), arg0)
}

// RegisterFunction mocks base method
func (m *MockFunctionService) RegisterFunction(arg0 *function.Function) (*function.Function, error) {
	ret := m.ctrl.Call(m, "RegisterFunction", arg0)
	ret0, _ := ret[0].(*function.Function)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RegisterFunction indicates an expected call of RegisterFunction
func (mr *MockFunctionServiceMockRecorder) RegisterFunction(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterFunction", reflect.TypeOf((*MockFunctionService)(nil).RegisterFunction), arg0)
}

// UpdateFunction mocks base method
func (m *MockFunctionService) UpdateFunction(arg0 string, arg1 *function.Function) (*function.Function, error) {
	ret := m.ctrl.Call(m, "UpdateFunction", arg0, arg1)
	ret0, _ := ret[0].(*function.Function)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateFunction indicates an expected call of UpdateFunction
func (mr *MockFunctionServiceMockRecorder) UpdateFunction(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateFunction", reflect.TypeOf((*MockFunctionService)(nil).UpdateFunction), arg0, arg1)
}
