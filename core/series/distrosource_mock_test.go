// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/core/series (interfaces: DistroSource)

// Package series is a generated GoMock package.
package series

import (
	gomock "github.com/golang/mock/gomock"
	series "github.com/juju/os/series"
	reflect "reflect"
)

// MockDistroSource is a mock of DistroSource interface
type MockDistroSource struct {
	ctrl     *gomock.Controller
	recorder *MockDistroSourceMockRecorder
}

// MockDistroSourceMockRecorder is the mock recorder for MockDistroSource
type MockDistroSourceMockRecorder struct {
	mock *MockDistroSource
}

// NewMockDistroSource creates a new mock instance
func NewMockDistroSource(ctrl *gomock.Controller) *MockDistroSource {
	mock := &MockDistroSource{ctrl: ctrl}
	mock.recorder = &MockDistroSourceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDistroSource) EXPECT() *MockDistroSourceMockRecorder {
	return m.recorder
}

// Refresh mocks base method
func (m *MockDistroSource) Refresh() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Refresh")
	ret0, _ := ret[0].(error)
	return ret0
}

// Refresh indicates an expected call of Refresh
func (mr *MockDistroSourceMockRecorder) Refresh() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Refresh", reflect.TypeOf((*MockDistroSource)(nil).Refresh))
}

// SeriesInfo mocks base method
func (m *MockDistroSource) SeriesInfo(arg0 string) (series.DistroInfoSerie, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SeriesInfo", arg0)
	ret0, _ := ret[0].(series.DistroInfoSerie)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// SeriesInfo indicates an expected call of SeriesInfo
func (mr *MockDistroSourceMockRecorder) SeriesInfo(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SeriesInfo", reflect.TypeOf((*MockDistroSource)(nil).SeriesInfo), arg0)
}
