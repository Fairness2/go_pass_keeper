// Code generated by MockGen. DO NOT EDIT.
// Source: internal/services/content/file.go

// Package content is a generated GoMock package.
package content

import (
	context "context"
	models "passkeeper/internal/models"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockfileRepository is a mock of fileRepository interface.
type MockfileRepository struct {
	ctrl     *gomock.Controller
	recorder *MockfileRepositoryMockRecorder
}

// MockfileRepositoryMockRecorder is the mock recorder for MockfileRepository.
type MockfileRepositoryMockRecorder struct {
	mock *MockfileRepository
}

// NewMockfileRepository creates a new mock instance.
func NewMockfileRepository(ctrl *gomock.Controller) *MockfileRepository {
	mock := &MockfileRepository{ctrl: ctrl}
	mock.recorder = &MockfileRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockfileRepository) EXPECT() *MockfileRepositoryMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockfileRepository) Create(ctx context.Context, content models.FileContent, comment models.Comment) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, content, comment)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockfileRepositoryMockRecorder) Create(ctx, content, comment interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockfileRepository)(nil).Create), ctx, content, comment)
}

// DeleteByUserIDAndID mocks base method.
func (m *MockfileRepository) DeleteByUserIDAndID(ctx context.Context, userID, id int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteByUserIDAndID", ctx, userID, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteByUserIDAndID indicates an expected call of DeleteByUserIDAndID.
func (mr *MockfileRepositoryMockRecorder) DeleteByUserIDAndID(ctx, userID, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteByUserIDAndID", reflect.TypeOf((*MockfileRepository)(nil).DeleteByUserIDAndID), ctx, userID, id)
}

// GetByUserID mocks base method.
func (m *MockfileRepository) GetByUserID(ctx context.Context, userID int64) ([]models.FileWithComment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByUserID", ctx, userID)
	ret0, _ := ret[0].([]models.FileWithComment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByUserID indicates an expected call of GetByUserID.
func (mr *MockfileRepositoryMockRecorder) GetByUserID(ctx, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByUserID", reflect.TypeOf((*MockfileRepository)(nil).GetByUserID), ctx, userID)
}

// GetByUserIDAndId mocks base method.
func (m *MockfileRepository) GetByUserIDAndId(ctx context.Context, userID, id int64) (*models.FileContent, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByUserIDAndId", ctx, userID, id)
	ret0, _ := ret[0].(*models.FileContent)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByUserIDAndId indicates an expected call of GetByUserIDAndId.
func (mr *MockfileRepositoryMockRecorder) GetByUserIDAndId(ctx, userID, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByUserIDAndId", reflect.TypeOf((*MockfileRepository)(nil).GetByUserIDAndId), ctx, userID, id)
}
