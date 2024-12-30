package repositories

import (
	"context"
	"database/sql"
	"errors"
	"github.com/golang/mock/gomock"
	"passkeeper/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCrudRepository_Create(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	errBegin := errors.New("begin error")
	expectedErr := errors.New("expected error")
	tests := []struct {
		name    string
		getExec func() SQLExecutor
		content models.TextContent
		comment models.Comment
		wantErr error
	}{
		{
			name: "successful_create",
			getExec: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				tx := NewMockITX(ctr)
				smth := NewMockIStmt(ctr)
				mockRow := NewMockIRow(ctr)

				mockDB.EXPECT().BeginTxx(gomock.Any(), gomock.Any()).Return(tx, nil).Times(1)
				tx.EXPECT().PrepareNamed(gomock.Any()).Return(smth, nil).Times(1)
				tx.EXPECT().NamedExecContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
				tx.EXPECT().Commit().Return(nil).Times(1)
				tx.EXPECT().Rollback().Return(nil).Times(1)
				smth.EXPECT().QueryRowxContext(gomock.Any(), gomock.Any()).Return(mockRow).Times(1)
				mockRow.EXPECT().Scan(gomock.Any()).DoAndReturn(func(id *string) error {
					*id = "f25172cc-e7d9-404c-a52d-0353c253a422"
					return nil
				}).Times(1)

				return mockDB
			},
			content: models.TextContent{
				UserID:    1,
				TextData:  []byte("test"),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			comment: models.Comment{
				ContentType: models.TypeText,
				Comment:     "Test comment",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "successful_update",
			getExec: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				tx := NewMockITX(ctr)

				mockDB.EXPECT().BeginTxx(gomock.Any(), gomock.Any()).Return(tx, nil).Times(1)
				tx.EXPECT().NamedExecContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(2)
				tx.EXPECT().Commit().Return(nil).Times(1)
				tx.EXPECT().Rollback().Return(nil).Times(1)

				return mockDB
			},
			content: models.TextContent{
				ID:        "f25172cc-e7d9-404c-a52d-0353c253a422",
				UserID:    1,
				TextData:  []byte("test"),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			comment: models.Comment{
				ContentID:   "f25172cc-e7d9-404c-a52d-0353c253a422",
				ContentType: models.TypeText,
				Comment:     "Test comment",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "error_on_transaction_begin",
			getExec: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				tx := NewMockITX(ctr)

				mockDB.EXPECT().BeginTxx(gomock.Any(), gomock.Any()).Return(tx, errBegin).Times(1)
				tx.EXPECT().NamedExecContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(0)
				tx.EXPECT().Commit().Return(nil).Times(0)
				tx.EXPECT().Rollback().Return(nil).Times(0)

				return mockDB
			},
			content: models.TextContent{
				ID:        "f25172cc-e7d9-404c-a52d-0353c253a422",
				UserID:    1,
				TextData:  []byte("test"),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			comment: models.Comment{
				ContentID:   "f25172cc-e7d9-404c-a52d-0353c253a422",
				ContentType: models.TypeText,
				Comment:     "Test comment",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: errBegin,
		},
		{
			name: "error_on_preparing_create_content_statement",
			getExec: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				tx := NewMockITX(ctr)
				smth := NewMockIStmt(ctr)
				mockRow := NewMockIRow(ctr)

				mockDB.EXPECT().BeginTxx(gomock.Any(), gomock.Any()).Return(tx, nil).Times(1)
				tx.EXPECT().PrepareNamed(gomock.Any()).Return(nil, expectedErr).Times(1)
				tx.EXPECT().NamedExecContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(0)
				tx.EXPECT().Commit().Return(nil).Times(0)
				tx.EXPECT().Rollback().Return(nil).Times(1)
				smth.EXPECT().QueryRowxContext(gomock.Any(), gomock.Any()).Return(mockRow).Times(0)
				mockRow.EXPECT().Scan(gomock.Any()).DoAndReturn(func(id *string) error {
					*id = "f25172cc-e7d9-404c-a52d-0353c253a422"
					return nil
				}).Times(0)

				return mockDB
			},
			content: models.TextContent{
				UserID:    1,
				TextData:  []byte("test"),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			comment: models.Comment{
				ContentType: models.TypeText,
				Comment:     "Test comment",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: expectedErr,
		},
		{
			name: "error_on_executing_create_content_statement",
			getExec: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				tx := NewMockITX(ctr)
				smth := NewMockIStmt(ctr)
				mockRow := NewMockIRow(ctr)

				mockDB.EXPECT().BeginTxx(gomock.Any(), gomock.Any()).Return(tx, nil).Times(1)
				tx.EXPECT().PrepareNamed(gomock.Any()).Return(smth, nil).Times(1)
				tx.EXPECT().NamedExecContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(0)
				tx.EXPECT().Commit().Return(nil).Times(0)
				tx.EXPECT().Rollback().Return(nil).Times(1)
				smth.EXPECT().QueryRowxContext(gomock.Any(), gomock.Any()).Return(mockRow).Times(1)
				mockRow.EXPECT().Scan(gomock.Any()).DoAndReturn(func(id *string) error {
					return expectedErr
				}).Times(1)

				return mockDB
			},
			content: models.TextContent{
				UserID:    1,
				TextData:  []byte("test"),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			comment: models.Comment{
				ContentType: models.TypeText,
				Comment:     "Test comment",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: expectedErr,
		},
		{
			name: "error_on_creating_comment",
			getExec: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				tx := NewMockITX(ctr)
				smth := NewMockIStmt(ctr)
				mockRow := NewMockIRow(ctr)

				mockDB.EXPECT().BeginTxx(gomock.Any(), gomock.Any()).Return(tx, nil).Times(1)
				tx.EXPECT().PrepareNamed(gomock.Any()).Return(smth, nil).Times(1)
				tx.EXPECT().NamedExecContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, expectedErr).Times(1)
				tx.EXPECT().Commit().Return(nil).Times(0)
				tx.EXPECT().Rollback().Return(nil).Times(1)
				smth.EXPECT().QueryRowxContext(gomock.Any(), gomock.Any()).Return(mockRow).Times(1)
				mockRow.EXPECT().Scan(gomock.Any()).DoAndReturn(func(id *string) error {
					*id = "f25172cc-e7d9-404c-a52d-0353c253a422"
					return nil
				}).Times(1)

				return mockDB
			},
			content: models.TextContent{
				UserID:    1,
				TextData:  []byte("test"),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			comment: models.Comment{
				ContentType: models.TypeText,
				Comment:     "Test comment",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: expectedErr,
		},
		{
			name: "commit_error",
			getExec: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				tx := NewMockITX(ctr)
				smth := NewMockIStmt(ctr)
				mockRow := NewMockIRow(ctr)

				mockDB.EXPECT().BeginTxx(gomock.Any(), gomock.Any()).Return(tx, nil).Times(1)
				tx.EXPECT().PrepareNamed(gomock.Any()).Return(smth, nil).Times(1)
				tx.EXPECT().NamedExecContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
				tx.EXPECT().Commit().Return(expectedErr).Times(1)
				tx.EXPECT().Rollback().Return(nil).Times(1)
				smth.EXPECT().QueryRowxContext(gomock.Any(), gomock.Any()).Return(mockRow).Times(1)
				mockRow.EXPECT().Scan(gomock.Any()).DoAndReturn(func(id *string) error {
					*id = "f25172cc-e7d9-404c-a52d-0353c253a422"
					return nil
				}).Times(1)

				return mockDB
			},
			content: models.TextContent{
				UserID:    1,
				TextData:  []byte("test"),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			comment: models.Comment{
				ContentType: models.TypeText,
				Comment:     "Test comment",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: expectedErr,
		},
		{
			name: "update_exec_error",
			getExec: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				tx := NewMockITX(ctr)

				mockDB.EXPECT().BeginTxx(gomock.Any(), gomock.Any()).Return(tx, nil).Times(1)
				tx.EXPECT().NamedExecContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, expectedErr).Times(1)
				tx.EXPECT().Commit().Return(nil).Times(0)
				tx.EXPECT().Rollback().Return(nil).Times(1)

				return mockDB
			},
			content: models.TextContent{
				ID:        "f25172cc-e7d9-404c-a52d-0353c253a422",
				UserID:    1,
				TextData:  []byte("test"),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			comment: models.Comment{
				ContentID:   "f25172cc-e7d9-404c-a52d-0353c253a422",
				ContentType: models.TypeText,
				Comment:     "Test comment",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: expectedErr,
		},
		{
			name: "update_comment_exec_error",
			getExec: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				tx := NewMockITX(ctr)

				mockDB.EXPECT().BeginTxx(gomock.Any(), gomock.Any()).Return(tx, nil).Times(1)
				first := tx.EXPECT().NamedExecContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
				tx.EXPECT().NamedExecContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, expectedErr).Times(1).After(first)
				tx.EXPECT().Commit().Return(nil).Times(0)
				tx.EXPECT().Rollback().Return(nil).Times(1)

				return mockDB
			},
			content: models.TextContent{
				ID:        "f25172cc-e7d9-404c-a52d-0353c253a422",
				UserID:    1,
				TextData:  []byte("test"),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			comment: models.Comment{
				ContentID:   "f25172cc-e7d9-404c-a52d-0353c253a422",
				ContentType: models.TypeText,
				Comment:     "Test comment",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: expectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crudRepo := &CrudRepository[models.TextContent, models.TextWithComment]{
				db:     tt.getExec(),
				sqlSet: TextSQLSet,
			}
			err := crudRepo.Create(context.Background(), tt.content, tt.comment)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCrudRepository_GetByUserID(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	want := []models.TextWithComment{
		{
			TextContent: models.TextContent{
				ID:       "f25172cc-e7d9-404c-a52d-0353c253a422",
				UserID:   1,
				TextData: []byte("test1"),
			},
			Comment: "Comment1",
		},
		{
			TextContent: models.TextContent{
				ID:       "1726ef63-756e-4dda-b669-0dcbef37a67f",
				UserID:   1,
				TextData: []byte("test2"),
			},
			Comment: "Comment2",
		},
	}
	expectedErr := errors.New("expected error")
	tests := []struct {
		name    string
		getExec func() SQLExecutor
		userID  int64
		want    []models.TextWithComment
		wantErr error
	}{
		{
			name: "successful_query",
			getExec: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				expectedResult := want
				mockDB.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, dest *[]models.TextWithComment, query string, args ...interface{}) error {
						*dest = expectedResult
						return nil
					},
				).Times(1)
				return mockDB
			},
			userID:  1,
			want:    want,
			wantErr: nil,
		},
		{
			name: "no_records_found",
			getExec: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				mockDB.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, dest *[]models.TextWithComment, query string, args ...interface{}) error {
						*dest = []models.TextWithComment{}
						return nil
					},
				).Times(1)
				return mockDB
			},
			userID:  2,
			want:    []models.TextWithComment{},
			wantErr: nil,
		},
		{
			name: "query_error",
			getExec: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				mockDB.EXPECT().SelectContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedErr).Times(1)
				return mockDB
			},
			userID:  3,
			want:    nil,
			wantErr: expectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crudRepo := &CrudRepository[models.TextContent, models.TextWithComment]{
				db:     tt.getExec(),
				sqlSet: TextSQLSet,
			}
			got, err := crudRepo.GetByUserID(context.Background(), tt.userID)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestCrudRepository_GetByUserIDAndId(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()

	expectedErr := errors.New("expected error")
	tests := []struct {
		name    string
		getExec func() SQLExecutor
		userID  int64
		id      string
		want    *models.TextContent
		wantErr error
	}{
		{
			name: "successful_query",
			getExec: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				row := NewMockIRow(ctr)
				row.EXPECT().StructScan(gomock.Any()).DoAndReturn(func(dest *models.TextContent) error {
					dest.ID = "f25172cc-e7d9-404c-a52d-0353c253a422"
					dest.UserID = 1
					dest.TextData = []byte("test1")
					return nil
				})
				row.EXPECT().Err().Return(nil)
				mockDB.EXPECT().QueryRowxContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(row)
				return mockDB
			},
			userID: 1,
			id:     "f25172cc-e7d9-404c-a52d-0353c253a422",
			want: &models.TextContent{
				ID:       "f25172cc-e7d9-404c-a52d-0353c253a422",
				UserID:   1,
				TextData: []byte("test1"),
			},
			wantErr: nil,
		},
		{
			name: "query_row_error",
			getExec: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				row := NewMockIRow(ctr)
				row.EXPECT().Err().Return(expectedErr)
				mockDB.EXPECT().QueryRowxContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(row)
				return mockDB
			},
			userID:  1,
			id:      "f25172cc-e7d9-404c-a52d-0353c253a422",
			want:    nil,
			wantErr: expectedErr,
		},
		{
			name: "no_rows_error",
			getExec: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				row := NewMockIRow(ctr)
				row.EXPECT().StructScan(gomock.Any()).Return(sql.ErrNoRows)
				row.EXPECT().Err().Return(nil)
				mockDB.EXPECT().QueryRowxContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(row)
				return mockDB
			},
			userID:  1,
			id:      "f25172cc-e7d9-404c-a52d-0353c253a422",
			want:    nil,
			wantErr: ErrNotExist,
		},
		{
			name: "struct_scan_error",
			getExec: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				row := NewMockIRow(ctr)
				row.EXPECT().StructScan(gomock.Any()).Return(expectedErr)
				row.EXPECT().Err().Return(nil)
				mockDB.EXPECT().QueryRowxContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(row)
				return mockDB
			},
			userID:  1,
			id:      "f25172cc-e7d9-404c-a52d-0353c253a422",
			want:    nil,
			wantErr: expectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crudRepo := &CrudRepository[models.TextContent, models.TextWithComment]{
				db:     tt.getExec(),
				sqlSet: TextSQLSet,
			}
			got, err := crudRepo.GetByUserIDAndId(context.Background(), tt.userID, tt.id)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestCrudRepository_DeleteByUserIDAndID(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()

	expectedErr := errors.New("expected error")
	tests := []struct {
		name    string
		getExec func() SQLExecutor
		userID  int64
		id      string
		wantErr error
	}{
		{
			name: "successful_deletion",
			getExec: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				tx := NewMockITX(ctr)

				mockDB.EXPECT().BeginTxx(gomock.Any(), gomock.Any()).Return(tx, nil).Times(1)
				tx.EXPECT().ExecContext(gomock.Any(), TextSQLSet.DeleteContentByUserIDAndID, "f25172cc-e7d9-404c-a52d-0353c253a422", int64(1)).Return(nil, nil).Times(1)
				tx.EXPECT().ExecContext(gomock.Any(), TextSQLSet.DeleteCommentByContentID, models.TypeText, "f25172cc-e7d9-404c-a52d-0353c253a422").Return(nil, nil).Times(1)
				tx.EXPECT().Commit().Return(nil).Times(1)
				tx.EXPECT().Rollback().Return(nil).Times(1)

				return mockDB
			},
			userID:  1,
			id:      "f25172cc-e7d9-404c-a52d-0353c253a422",
			wantErr: nil,
		},
		{
			name: "error_on_transaction_begin",
			getExec: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				mockDB.EXPECT().BeginTxx(gomock.Any(), gomock.Any()).Return(nil, expectedErr).Times(1)
				return mockDB
			},
			userID:  1,
			id:      "f25172cc-e7d9-404c-a52d-0353c253a422",
			wantErr: expectedErr,
		},
		{
			name: "error_on_content_deletion",
			getExec: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				tx := NewMockITX(ctr)

				mockDB.EXPECT().BeginTxx(gomock.Any(), gomock.Any()).Return(tx, nil).Times(1)
				tx.EXPECT().ExecContext(gomock.Any(), TextSQLSet.DeleteContentByUserIDAndID, "f25172cc-e7d9-404c-a52d-0353c253a422", int64(1)).Return(nil, expectedErr).Times(1)
				tx.EXPECT().Rollback().Return(nil).Times(1)
				return mockDB
			},
			userID:  1,
			id:      "f25172cc-e7d9-404c-a52d-0353c253a422",
			wantErr: expectedErr,
		},
		{
			name: "error_on_comment_deletion",
			getExec: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				tx := NewMockITX(ctr)

				mockDB.EXPECT().BeginTxx(gomock.Any(), gomock.Any()).Return(tx, nil).Times(1)
				tx.EXPECT().ExecContext(gomock.Any(), TextSQLSet.DeleteContentByUserIDAndID, "f25172cc-e7d9-404c-a52d-0353c253a422", int64(1)).Return(nil, nil).Times(1)
				tx.EXPECT().ExecContext(gomock.Any(), TextSQLSet.DeleteCommentByContentID, models.TypeText, "f25172cc-e7d9-404c-a52d-0353c253a422").Return(nil, expectedErr).Times(1)
				tx.EXPECT().Rollback().Return(nil).Times(1)
				return mockDB
			},
			userID:  1,
			id:      "f25172cc-e7d9-404c-a52d-0353c253a422",
			wantErr: expectedErr,
		},
		{
			name: "error_on_commit",
			getExec: func() SQLExecutor {
				mockDB := NewMockSQLExecutor(ctr)
				tx := NewMockITX(ctr)

				mockDB.EXPECT().BeginTxx(gomock.Any(), gomock.Any()).Return(tx, nil).Times(1)
				tx.EXPECT().ExecContext(gomock.Any(), TextSQLSet.DeleteContentByUserIDAndID, "f25172cc-e7d9-404c-a52d-0353c253a422", int64(1)).Return(nil, nil).Times(1)
				tx.EXPECT().ExecContext(gomock.Any(), TextSQLSet.DeleteCommentByContentID, models.TypeText, "f25172cc-e7d9-404c-a52d-0353c253a422").Return(nil, nil).Times(1)
				tx.EXPECT().Commit().Return(expectedErr).Times(1)
				tx.EXPECT().Rollback().Return(nil).Times(1)
				return mockDB
			},
			userID:  1,
			id:      "f25172cc-e7d9-404c-a52d-0353c253a422",
			wantErr: expectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crudRepo := &CrudRepository[models.TextContent, models.TextWithComment]{
				db:          tt.getExec(),
				sqlSet:      TextSQLSet,
				typeContent: models.TypeText,
			}
			err := crudRepo.DeleteByUserIDAndID(context.Background(), tt.userID, tt.id)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
