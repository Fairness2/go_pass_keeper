// Code generated by MockGen. DO NOT EDIT.
// Source: /Users/konstantinkuzminyh/go/pkg/mod/github.com/golang-jwt/jwt/v5@v5.2.1/claims.go

// Package token is a generated GoMock package.
package token

import (
	reflect "reflect"

	jwt "github.com/golang-jwt/jwt/v5"
	gomock "github.com/golang/mock/gomock"
)

// MockClaims is a mock of Claims interface.
type MockClaims struct {
	ctrl     *gomock.Controller
	recorder *MockClaimsMockRecorder
}

// MockClaimsMockRecorder is the mock recorder for MockClaims.
type MockClaimsMockRecorder struct {
	mock *MockClaims
}

// NewMockClaims creates a new mock instance.
func NewMockClaims(ctrl *gomock.Controller) *MockClaims {
	mock := &MockClaims{ctrl: ctrl}
	mock.recorder = &MockClaimsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClaims) EXPECT() *MockClaimsMockRecorder {
	return m.recorder
}

// GetAudience mocks base method.
func (m *MockClaims) GetAudience() (jwt.ClaimStrings, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAudience")
	ret0, _ := ret[0].(jwt.ClaimStrings)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAudience indicates an expected call of GetAudience.
func (mr *MockClaimsMockRecorder) GetAudience() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAudience", reflect.TypeOf((*MockClaims)(nil).GetAudience))
}

// GetExpirationTime mocks base method.
func (m *MockClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetExpirationTime")
	ret0, _ := ret[0].(*jwt.NumericDate)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetExpirationTime indicates an expected call of GetExpirationTime.
func (mr *MockClaimsMockRecorder) GetExpirationTime() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetExpirationTime", reflect.TypeOf((*MockClaims)(nil).GetExpirationTime))
}

// GetIssuedAt mocks base method.
func (m *MockClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetIssuedAt")
	ret0, _ := ret[0].(*jwt.NumericDate)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetIssuedAt indicates an expected call of GetIssuedAt.
func (mr *MockClaimsMockRecorder) GetIssuedAt() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetIssuedAt", reflect.TypeOf((*MockClaims)(nil).GetIssuedAt))
}

// GetIssuer mocks base method.
func (m *MockClaims) GetIssuer() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetIssuer")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetIssuer indicates an expected call of GetIssuer.
func (mr *MockClaimsMockRecorder) GetIssuer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetIssuer", reflect.TypeOf((*MockClaims)(nil).GetIssuer))
}

// GetNotBefore mocks base method.
func (m *MockClaims) GetNotBefore() (*jwt.NumericDate, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNotBefore")
	ret0, _ := ret[0].(*jwt.NumericDate)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNotBefore indicates an expected call of GetNotBefore.
func (mr *MockClaimsMockRecorder) GetNotBefore() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNotBefore", reflect.TypeOf((*MockClaims)(nil).GetNotBefore))
}

// GetSubject mocks base method.
func (m *MockClaims) GetSubject() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSubject")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSubject indicates an expected call of GetSubject.
func (mr *MockClaimsMockRecorder) GetSubject() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSubject", reflect.TypeOf((*MockClaims)(nil).GetSubject))
}
