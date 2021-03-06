// Code generated by MockGen. DO NOT EDIT.
// Source: internal/connectiontracker/firewall.go

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockFirewall is a mock of Firewall interface.
type MockFirewall struct {
	ctrl     *gomock.Controller
	recorder *MockFirewallMockRecorder
}

// MockFirewallMockRecorder is the mock recorder for MockFirewall.
type MockFirewallMockRecorder struct {
	mock *MockFirewall
}

// NewMockFirewall creates a new mock instance.
func NewMockFirewall(ctrl *gomock.Controller) *MockFirewall {
	mock := &MockFirewall{ctrl: ctrl}
	mock.recorder = &MockFirewallMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFirewall) EXPECT() *MockFirewallMockRecorder {
	return m.recorder
}

// Block mocks base method.
func (m *MockFirewall) Block(ip string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Block", ip)
	ret0, _ := ret[0].(error)
	return ret0
}

// Block indicates an expected call of Block.
func (mr *MockFirewallMockRecorder) Block(ip interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Block", reflect.TypeOf((*MockFirewall)(nil).Block), ip)
}

// Close mocks base method.
func (m *MockFirewall) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockFirewallMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockFirewall)(nil).Close))
}

// MockipTableCoreos is a mock of ipTableCoreos interface.
type MockipTableCoreos struct {
	ctrl     *gomock.Controller
	recorder *MockipTableCoreosMockRecorder
}

// MockipTableCoreosMockRecorder is the mock recorder for MockipTableCoreos.
type MockipTableCoreosMockRecorder struct {
	mock *MockipTableCoreos
}

// NewMockipTableCoreos creates a new mock instance.
func NewMockipTableCoreos(ctrl *gomock.Controller) *MockipTableCoreos {
	mock := &MockipTableCoreos{ctrl: ctrl}
	mock.recorder = &MockipTableCoreosMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockipTableCoreos) EXPECT() *MockipTableCoreosMockRecorder {
	return m.recorder
}

// AppendUnique mocks base method.
func (m *MockipTableCoreos) AppendUnique(arg0, arg1 string, arg2 ...string) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "AppendUnique", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// AppendUnique indicates an expected call of AppendUnique.
func (mr *MockipTableCoreosMockRecorder) AppendUnique(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AppendUnique", reflect.TypeOf((*MockipTableCoreos)(nil).AppendUnique), varargs...)
}

// ChainExists mocks base method.
func (m *MockipTableCoreos) ChainExists(arg0, arg1 string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ChainExists", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ChainExists indicates an expected call of ChainExists.
func (mr *MockipTableCoreosMockRecorder) ChainExists(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ChainExists", reflect.TypeOf((*MockipTableCoreos)(nil).ChainExists), arg0, arg1)
}

// ClearAndDeleteChain mocks base method.
func (m *MockipTableCoreos) ClearAndDeleteChain(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ClearAndDeleteChain", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// ClearAndDeleteChain indicates an expected call of ClearAndDeleteChain.
func (mr *MockipTableCoreosMockRecorder) ClearAndDeleteChain(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ClearAndDeleteChain", reflect.TypeOf((*MockipTableCoreos)(nil).ClearAndDeleteChain), arg0, arg1)
}

// DeleteIfExists mocks base method.
func (m *MockipTableCoreos) DeleteIfExists(arg0, arg1 string, arg2 ...string) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteIfExists", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteIfExists indicates an expected call of DeleteIfExists.
func (mr *MockipTableCoreosMockRecorder) DeleteIfExists(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteIfExists", reflect.TypeOf((*MockipTableCoreos)(nil).DeleteIfExists), varargs...)
}

// Insert mocks base method.
func (m *MockipTableCoreos) Insert(arg0, arg1 string, arg2 int, arg3 ...string) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Insert", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Insert indicates an expected call of Insert.
func (mr *MockipTableCoreosMockRecorder) Insert(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Insert", reflect.TypeOf((*MockipTableCoreos)(nil).Insert), varargs...)
}

// NewChain mocks base method.
func (m *MockipTableCoreos) NewChain(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewChain", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// NewChain indicates an expected call of NewChain.
func (mr *MockipTableCoreosMockRecorder) NewChain(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewChain", reflect.TypeOf((*MockipTableCoreos)(nil).NewChain), arg0, arg1)
}
