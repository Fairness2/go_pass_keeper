// Code generated by MockGen. DO NOT EDIT.
// Source: internal/repositories/executor.go

// Package repositories is a generated GoMock package.
package content

import (
	context "context"
	sql "database/sql"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	sqlx "github.com/jmoiron/sqlx"
)

// MockSQLExecutor is a mock of SQLExecutor interface.
type MockSQLExecutor struct {
	ctrl     *gomock.Controller
	recorder *MockSQLExecutorMockRecorder
}

// MockSQLExecutorMockRecorder is the mock recorder for MockSQLExecutor.
type MockSQLExecutorMockRecorder struct {
	mock *MockSQLExecutor
}

// NewMockSQLExecutor creates a new mock instance.
func NewMockSQLExecutor(ctrl *gomock.Controller) *MockSQLExecutor {
	mock := &MockSQLExecutor{ctrl: ctrl}
	mock.recorder = &MockSQLExecutorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSQLExecutor) EXPECT() *MockSQLExecutorMockRecorder {
	return m.recorder
}

// BeginTxx mocks base method.
func (m *MockSQLExecutor) BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BeginTxx", ctx, opts)
	ret0, _ := ret[0].(*sqlx.Tx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BeginTxx indicates an expected call of BeginTxx.
func (mr *MockSQLExecutorMockRecorder) BeginTxx(ctx, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeginTxx", reflect.TypeOf((*MockSQLExecutor)(nil).BeginTxx), ctx, opts)
}

// NamedExecContext mocks base method.
func (m *MockSQLExecutor) NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NamedExecContext", ctx, query, arg)
	ret0, _ := ret[0].(sql.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NamedExecContext indicates an expected call of NamedExecContext.
func (mr *MockSQLExecutorMockRecorder) NamedExecContext(ctx, query, arg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NamedExecContext", reflect.TypeOf((*MockSQLExecutor)(nil).NamedExecContext), ctx, query, arg)
}

// PrepareNamed mocks base method.
func (m *MockSQLExecutor) PrepareNamed(query string) (*sqlx.NamedStmt, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PrepareNamed", query)
	ret0, _ := ret[0].(*sqlx.NamedStmt)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PrepareNamed indicates an expected call of PrepareNamed.
func (mr *MockSQLExecutorMockRecorder) PrepareNamed(query interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PrepareNamed", reflect.TypeOf((*MockSQLExecutor)(nil).PrepareNamed), query)
}

// QueryRowContext mocks base method.
func (m *MockSQLExecutor) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "QueryRowContext", varargs...)
	ret0, _ := ret[0].(*sql.Row)
	return ret0
}

// QueryRowContext indicates an expected call of QueryRowContext.
func (mr *MockSQLExecutorMockRecorder) QueryRowContext(ctx, query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryRowContext", reflect.TypeOf((*MockSQLExecutor)(nil).QueryRowContext), varargs...)
}

// QueryRowxContext mocks base method.
func (m *MockSQLExecutor) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "QueryRowxContext", varargs...)
	ret0, _ := ret[0].(*sqlx.Row)
	return ret0
}

// QueryRowxContext indicates an expected call of QueryRowxContext.
func (mr *MockSQLExecutorMockRecorder) QueryRowxContext(ctx, query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryRowxContext", reflect.TypeOf((*MockSQLExecutor)(nil).QueryRowxContext), varargs...)
}

// SelectContext mocks base method.
func (m *MockSQLExecutor) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, dest, query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "SelectContext", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// SelectContext indicates an expected call of SelectContext.
func (mr *MockSQLExecutorMockRecorder) SelectContext(ctx, dest, query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, dest, query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectContext", reflect.TypeOf((*MockSQLExecutor)(nil).SelectContext), varargs...)
}
