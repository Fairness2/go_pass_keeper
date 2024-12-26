package repositories

import (
	"context"
	"database/sql"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"passkeeper/internal/models"
	"testing"
)

func TestNewUserRepository(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	tests := []struct {
		name    string
		db      SQLExecutor
		wantNil bool
	}{
		{
			name:    "valid_executor",
			db:      NewMockSQLExecutor(ctr),
			wantNil: false,
		},
		{
			name:    "nil_executor",
			db:      nil,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewUserRepository(tt.db)
			assert.NotNil(t, repo)
		})
	}
}

func TestUserExists(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	dbErr := errors.New("mock DB error")
	tests := []struct {
		name      string
		login     string
		setupMock func() SQLExecutor
		expectErr error
	}{
		{
			name:  "user_exists",
			login: "existing_user",
			setupMock: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				mockRow := NewMockIRow(ctr)
				mockRow.EXPECT().Scan(gomock.Any()).Return(nil).Times(1)
				mockDB.EXPECT().QueryRowContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(mockRow).Times(1)
				return mockDB
			},
			expectErr: nil,
		},
		{
			name:  "user_does_not_exist",
			login: "non_existent_user",
			setupMock: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				mockRow := NewMockIRow(ctr)
				mockRow.EXPECT().Scan(gomock.Any()).Return(sql.ErrNoRows).Times(1)
				mockDB.EXPECT().QueryRowContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(mockRow).Times(1)
				return mockDB
			},
			expectErr: ErrNotExist,
		},
		{
			name:  "database_error",
			login: "db_error_user",
			setupMock: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				mockRow := NewMockIRow(ctr)
				mockRow.EXPECT().Scan(gomock.Any()).Return(dbErr).Times(1)
				mockDB.EXPECT().QueryRowContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(mockRow).Times(1)
				return mockDB
			},
			expectErr: dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &UserRepository{db: tt.setupMock()}
			err := repo.UserExists(context.TODO(), tt.login)
			assert.ErrorIs(t, err, tt.expectErr)
		})
	}
}

func TestCreateUser(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	dbErr := errors.New("mock DB error")
	tests := []struct {
		name      string
		user      *models.User
		setupMock func() SQLExecutor
		expectErr error
	}{
		{
			name: "successful_creation",
			user: &models.User{Login: "new_user", Password: "hashed_password"},
			setupMock: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				mockStmt := NewMockINamedStmt(ctr)
				mockRow := NewMockIRow(ctr)

				mockDB.EXPECT().PrepareNamed(createUserSQL).Return(mockStmt, nil).Times(1)
				mockStmt.EXPECT().QueryRowxContext(gomock.Any(), gomock.Any()).Return(mockRow).Times(1)
				mockRow.EXPECT().Scan(gomock.Any()).DoAndReturn(func(id *int64) error {
					*id = 1 // Mock ID generation
					return nil
				}).Times(1)

				return mockDB
			},
			expectErr: nil,
		},
		{
			name: "prepare_failed",
			user: &models.User{Login: "failing_user"},
			setupMock: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				mockDB.EXPECT().PrepareNamed(createUserSQL).Return(nil, dbErr).Times(1)
				return mockDB
			},
			expectErr: dbErr,
		},
		{
			name: "query_row_error",
			user: &models.User{Login: "query_error_user"},
			setupMock: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				mockStmt := NewMockINamedStmt(ctr)
				mockRow := NewMockIRow(ctr)

				mockDB.EXPECT().PrepareNamed(createUserSQL).Return(mockStmt, nil).Times(1)
				mockStmt.EXPECT().QueryRowxContext(gomock.Any(), gomock.Any()).Return(mockRow).Times(1)
				mockRow.EXPECT().Scan(gomock.Any()).Return(sql.ErrNoRows).Times(1)
				return mockDB
			},
			expectErr: sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &UserRepository{db: tt.setupMock()}
			err := repo.CreateUser(context.TODO(), tt.user)
			assert.ErrorIs(t, err, tt.expectErr)
			if err == nil {
				assert.NotZero(t, tt.user.ID)
			}
		})
	}
}

func TestGetUserByLogin(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	dbErr := errors.New("mock DB error")
	tests := []struct {
		name       string
		login      string
		setupMock  func() SQLExecutor
		expectErr  error
		expectUser *models.User
	}{
		{
			name:  "user_exists",
			login: "existing_user",
			setupMock: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				mockRow := NewMockIRow(ctr)
				mockRow.EXPECT().StructScan(gomock.Any()).DoAndReturn(func(user *models.User) error {
					*user = models.User{
						ID:           1,
						Login:        "existing_user",
						PasswordHash: "hashed_password",
					}
					return nil
				}).Times(1)
				mockDB.EXPECT().QueryRowxContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(mockRow).Times(1)
				return mockDB
			},
			expectErr: nil,
			expectUser: &models.User{
				ID:           1,
				Login:        "existing_user",
				PasswordHash: "hashed_password",
			},
		},
		{
			name:  "user_does_not_exist",
			login: "non_existent_user",
			setupMock: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				mockRow := NewMockIRow(ctr)
				mockRow.EXPECT().StructScan(gomock.Any()).Return(sql.ErrNoRows).Times(1)
				mockDB.EXPECT().QueryRowxContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(mockRow).Times(1)
				return mockDB
			},
			expectErr:  ErrNotExist,
			expectUser: nil,
		},
		{
			name:  "database_error",
			login: "db_error_user",
			setupMock: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				mockRow := NewMockIRow(ctr)
				mockRow.EXPECT().StructScan(gomock.Any()).Return(dbErr).Times(1)
				mockDB.EXPECT().QueryRowxContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(mockRow).Times(1)
				return mockDB
			},
			expectErr:  dbErr,
			expectUser: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &UserRepository{db: tt.setupMock()}
			user, err := repo.GetUserByLogin(context.TODO(), tt.login)
			assert.ErrorIs(t, err, tt.expectErr)
			assert.Equal(t, tt.expectUser, user)
		})
	}
}

func TestGetUserByID(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	dbErr := errors.New("mock DB error")
	tests := []struct {
		name       string
		id         int64
		setupMock  func() SQLExecutor
		expectErr  error
		expectUser *models.User
	}{
		{
			name: "user_exists",
			id:   1,
			setupMock: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				mockRow := NewMockIRow(ctr)
				mockRow.EXPECT().StructScan(gomock.Any()).DoAndReturn(func(user *models.User) error {
					*user = models.User{
						ID:           1,
						Login:        "existing_user",
						PasswordHash: "hashed_password",
					}
					return nil
				}).Times(1)
				mockDB.EXPECT().QueryRowxContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(mockRow).Times(1)
				return mockDB
			},
			expectErr: nil,
			expectUser: &models.User{
				ID:           1,
				Login:        "existing_user",
				PasswordHash: "hashed_password",
			},
		},
		{
			name: "user_does_not_exist",
			id:   999,
			setupMock: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				mockRow := NewMockIRow(ctr)
				mockRow.EXPECT().StructScan(gomock.Any()).Return(sql.ErrNoRows).Times(1)
				mockDB.EXPECT().QueryRowxContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(mockRow).Times(1)
				return mockDB
			},
			expectErr:  ErrNotExist,
			expectUser: nil,
		},
		{
			name: "database_error",
			id:   2,
			setupMock: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				mockRow := NewMockIRow(ctr)
				mockRow.EXPECT().StructScan(gomock.Any()).Return(dbErr).Times(1)
				mockDB.EXPECT().QueryRowxContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(mockRow).Times(1)
				return mockDB
			},
			expectErr:  dbErr,
			expectUser: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &UserRepository{db: tt.setupMock()}
			user, err := repo.GetUserByID(context.TODO(), tt.id)
			assert.ErrorIs(t, err, tt.expectErr)
			assert.Equal(t, tt.expectUser, user)
		})
	}
}
