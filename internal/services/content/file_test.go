package content

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"passkeeper/internal/models"
	"passkeeper/internal/repositories"
	"passkeeper/internal/token"
	"strings"
	"testing"
	"time"
)

func TestFileService_deleteFile(t *testing.T) {
	tests := []struct {
		name       string
		filePath   string
		userID     int64
		setupMocks func() string
		expectErr  bool
	}{
		{
			name:     "successful_delete",
			filePath: "test.txt",
			userID:   123,
			setupMocks: func() string {
				dir := os.TempDir() + "test"
				userDir := fmt.Sprintf("%s/123", dir)
				if err := os.MkdirAll(userDir, os.ModePerm); err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}
				filePath := fmt.Sprintf("%s/test.txt", userDir)
				if _, err := os.Create(filePath); err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				return dir
			},
			expectErr: false,
		},
		{
			name:     "file_does_not_exist",
			filePath: "fake.txt",
			userID:   456,
			setupMocks: func() string {
				return os.TempDir() + "test"
			},
			expectErr: false,
		},
		{
			name:     "deletion_error",
			filePath: "test.txt",
			userID:   789,
			setupMocks: func() string {
				dir := os.TempDir() + "test"
				userDir := fmt.Sprintf("%s/789", dir)
				if err := os.MkdirAll(userDir, 0444); err != nil {
					t.Fatalf("Failed to create restricted dir: %v", err)
				}
				return dir
			},
			expectErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootPath := tt.setupMocks()

			service := &FileService{
				filePath: rootPath,
			}
			err := service.deleteFile(tt.filePath, tt.userID)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFileService_DeleteUserFiles(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	uploadDir := os.TempDir() + "test"
	defer os.RemoveAll(uploadDir)
	var userID int64 = 1
	var forbiddenUserID int64 = 2
	fileName := "test.txt"
	tests := []struct {
		name           string
		setMocks       func() *FileService
		textId         string
		expectedStatus int
		authMiddleware func(next http.Handler) http.Handler
		filePath       string
		userID         int64
		setupFile      func()
	}{
		{
			name:   "successful_delete",
			textId: "1",
			setMocks: func() *FileService {
				repo := NewMockfileRepository(ctr)
				repo.EXPECT().DeleteByUserIDAndID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				repo.EXPECT().GetByUserIDAndId(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.FileContent{
					ID:        1,
					UserID:    userID,
					Name:      []byte("aboba"),
					FilePath:  fileName,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil).Times(1)
				handlers := &FileService{
					repository: repo,
					filePath:   uploadDir,
				}
				return handlers
			},
			expectedStatus: http.StatusOK,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					newR := r.WithContext(context.WithValue(r.Context(), token.UserKey, &models.User{ID: userID}))
					next.ServeHTTP(w, newR)
				})
			},
			filePath:  fileName,
			userID:    userID,
			setupFile: func() {},
		},
		{
			name:   "not_authorized",
			textId: "1",
			setMocks: func() *FileService {
				repo := NewMockfileRepository(ctr)
				repo.EXPECT().DeleteByUserIDAndID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(0)
				handlers := &FileService{
					repository: repo,
					filePath:   uploadDir,
				}
				return handlers
			},
			expectedStatus: http.StatusUnauthorized,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					next.ServeHTTP(w, r)
				})
			},
			setupFile: func() {},
		},
		{
			name:   "incorrect_id",
			textId: "aboba",
			setMocks: func() *FileService {
				repo := NewMockfileRepository(ctr)
				repo.EXPECT().DeleteByUserIDAndID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(0)
				handlers := &FileService{
					repository: repo,
					filePath:   uploadDir,
				}
				return handlers
			},
			expectedStatus: http.StatusBadRequest,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					newR := r.WithContext(context.WithValue(r.Context(), token.UserKey, &models.User{ID: 1}))
					next.ServeHTTP(w, newR)
				})
			},
			setupFile: func() {},
		},
		{
			name:   "delete_error",
			textId: "1",
			setMocks: func() *FileService {
				repo := NewMockfileRepository(ctr)
				repo.EXPECT().DeleteByUserIDAndID(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some error")).Times(1)
				repo.EXPECT().GetByUserIDAndId(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.FileContent{
					ID:        1,
					UserID:    userID,
					Name:      []byte("aboba"),
					FilePath:  fileName,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil).Times(1)
				handlers := &FileService{
					repository: repo,
					filePath:   uploadDir,
				}
				return handlers
			},
			expectedStatus: http.StatusInternalServerError,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					newR := r.WithContext(context.WithValue(r.Context(), token.UserKey, &models.User{ID: 1}))
					next.ServeHTTP(w, newR)
				})
			},
			setupFile: func() {},
		},
		{
			name:   "not_exist_file",
			textId: "1",
			setMocks: func() *FileService {
				repo := NewMockfileRepository(ctr)
				repo.EXPECT().DeleteByUserIDAndID(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some error")).Times(0)
				repo.EXPECT().GetByUserIDAndId(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, repositories.ErrNotExist).Times(1)
				handlers := &FileService{
					repository: repo,
					filePath:   uploadDir,
				}
				return handlers
			},
			expectedStatus: http.StatusNotFound,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					newR := r.WithContext(context.WithValue(r.Context(), token.UserKey, &models.User{ID: 1}))
					next.ServeHTTP(w, newR)
				})
			},
			setupFile: func() {},
		},
		{
			name:   "error_get_file",
			textId: "1",
			setMocks: func() *FileService {
				repo := NewMockfileRepository(ctr)
				repo.EXPECT().DeleteByUserIDAndID(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some error")).Times(0)
				repo.EXPECT().GetByUserIDAndId(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("some error")).Times(1)
				handlers := &FileService{
					repository: repo,
					filePath:   uploadDir,
				}
				return handlers
			},
			expectedStatus: http.StatusInternalServerError,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					newR := r.WithContext(context.WithValue(r.Context(), token.UserKey, &models.User{ID: 1}))
					next.ServeHTTP(w, newR)
				})
			},
			setupFile: func() {},
		},
		{
			name:   "error_delete_file",
			textId: "1",
			setMocks: func() *FileService {
				repo := NewMockfileRepository(ctr)
				repo.EXPECT().DeleteByUserIDAndID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(0)
				repo.EXPECT().GetByUserIDAndId(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.FileContent{
					ID:        1,
					UserID:    forbiddenUserID,
					Name:      []byte("aboba"),
					FilePath:  fileName,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil).Times(1)
				handlers := &FileService{
					repository: repo,
					filePath:   uploadDir,
				}
				return handlers
			},
			expectedStatus: http.StatusInternalServerError,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					newR := r.WithContext(context.WithValue(r.Context(), token.UserKey, &models.User{ID: forbiddenUserID}))
					next.ServeHTTP(w, newR)
				})
			},
			filePath: fileName,
			userID:   userID,
			setupFile: func() {
				userDir := fmt.Sprintf("%s/%d", uploadDir, forbiddenUserID)
				if err := os.MkdirAll(userDir, 0444); err != nil {
					t.Fatalf("Failed to create restricted dir: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFile()
			handlers := tt.setMocks()
			router := chi.NewRouter()
			router.Use(tt.authMiddleware)
			router.Delete("/{id}", handlers.DeleteUserFile)
			// запускаем тестовый сервер, будет выбран первый свободный порт
			srv := httptest.NewServer(router)
			// останавливаем сервер после завершения теста
			defer srv.Close()
			request := resty.New().R()
			request.Method = http.MethodDelete
			request.URL = srv.URL + "/" + tt.textId

			res, err := request.Send()
			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, tt.expectedStatus, res.StatusCode(), "unexpected response status code")
		})
	}
}

func TestNewFileService(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	tests := []struct {
		name          string
		expectedError bool
		getExec       func() repositories.SQLExecutor
		path          string
	}{
		{
			name:          "successful_initialization",
			expectedError: false,
			getExec: func() repositories.SQLExecutor {
				return NewMockSQLExecutor(ctr)
			},
			path: "some_path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewFileService(tt.getExec(), tt.path)
			assert.NotNil(t, service, "NewFileService should not return nil")
		})
	}
}

func TestFileService_GetUserFiles(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()

	tests := []struct {
		name           string
		setMocks       func() *FileService
		response       string
		expectedStatus int
		authMiddleware func(next http.Handler) http.Handler
	}{
		{
			name:     "successful_get",
			response: "[{\"id\":1,\"name\":\"YWJvYmE=\",\"comment\":\"aboba\"},{\"id\":2,\"name\":\"YWJvYmEy\",\"comment\":\"aboba2\"}]",
			setMocks: func() *FileService {
				repo := NewMockfileRepository(ctr)
				repo.EXPECT().GetByUserID(gomock.Any(), gomock.Any()).Return([]models.FileWithComment{
					{
						FileContent: models.FileContent{
							ID:        1,
							UserID:    1,
							Name:      []byte("aboba"),
							FilePath:  "test.txt",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Comment: "aboba",
					},
					{
						FileContent: models.FileContent{
							ID:        2,
							UserID:    1,
							Name:      []byte("aboba2"),
							FilePath:  "test2.txt",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Comment: "aboba2",
					},
				}, nil).Times(1)
				handlers := &FileService{
					repository: repo,
				}
				return handlers
			},
			expectedStatus: http.StatusOK,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					newR := r.WithContext(context.WithValue(r.Context(), token.UserKey, &models.User{ID: 1}))
					next.ServeHTTP(w, newR)
				})
			},
		},
		{
			name:     "not_authorized",
			response: "",
			setMocks: func() *FileService {
				repo := NewMockfileRepository(ctr)
				repo.EXPECT().GetByUserID(gomock.Any(), gomock.Any()).Return(nil, nil).Times(0)
				handlers := &FileService{
					repository: repo,
				}
				return handlers
			},
			expectedStatus: http.StatusUnauthorized,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					next.ServeHTTP(w, r)
				})
			},
		},
		{
			name:     "db_err",
			response: "",
			setMocks: func() *FileService {
				repo := NewMockfileRepository(ctr)
				repo.EXPECT().GetByUserID(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error")).Times(1)
				handlers := &FileService{
					repository: repo,
				}
				return handlers
			},
			expectedStatus: http.StatusInternalServerError,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					newR := r.WithContext(context.WithValue(r.Context(), token.UserKey, &models.User{ID: 1}))
					next.ServeHTTP(w, newR)
				})
			},
		},
		{
			name:     "empty_response",
			response: "[]",
			setMocks: func() *FileService {
				repo := NewMockfileRepository(ctr)
				repo.EXPECT().GetByUserID(gomock.Any(), gomock.Any()).Return([]models.FileWithComment{}, nil).Times(1)
				handlers := &FileService{
					repository: repo,
				}
				return handlers
			},
			expectedStatus: http.StatusOK,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					newR := r.WithContext(context.WithValue(r.Context(), token.UserKey, &models.User{ID: 1}))
					next.ServeHTTP(w, newR)
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers := tt.setMocks()
			router := chi.NewRouter()
			router.Use(tt.authMiddleware)
			router.Get("/", handlers.GetUserFiles)
			// запускаем тестовый сервер, будет выбран первый свободный порт
			srv := httptest.NewServer(router)
			// останавливаем сервер после завершения теста
			defer srv.Close()
			request := resty.New().R()
			request.Method = http.MethodGet
			request.URL = srv.URL

			res, err := request.Send()
			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, tt.expectedStatus, res.StatusCode(), "unexpected response status code")
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, tt.response, string(res.Body()), "unexpected response")
			}
		})
	}
}

func TestFileService_UpdateFileHandler(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()

	tests := []struct {
		name           string
		setMocks       func() *FileService
		body           string
		expectedStatus int
		authMiddleware func(next http.Handler) http.Handler
	}{
		{
			name: "successful_update",
			body: "{\"id\":1,\"name\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *FileService {
				repo := NewMockfileRepository(ctr)
				repo.EXPECT().GetByUserIDAndId(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.FileContent{
					ID:        1,
					UserID:    1,
					Name:      []byte("aboba"),
					FilePath:  "test.txt",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil).Times(1)
				repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				handlers := &FileService{
					repository: repo,
				}
				return handlers
			},
			expectedStatus: http.StatusOK,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					newR := r.WithContext(context.WithValue(r.Context(), token.UserKey, &models.User{ID: 1}))
					next.ServeHTTP(w, newR)
				})
			},
		},
		{
			name: "not_authorized",
			body: "{\"id\":1,\"name\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *FileService {
				repo := NewMockfileRepository(ctr)
				handlers := &FileService{
					repository: repo,
				}
				return handlers
			},
			expectedStatus: http.StatusUnauthorized,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					next.ServeHTTP(w, r)
				})
			},
		},
		{
			name: "incorrect_body",
			body: "aboba",
			setMocks: func() *FileService {
				repo := NewMockfileRepository(ctr)
				handlers := &FileService{
					repository: repo,
				}
				return handlers
			},
			expectedStatus: http.StatusBadRequest,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					newR := r.WithContext(context.WithValue(r.Context(), token.UserKey, &models.User{ID: 1}))
					next.ServeHTTP(w, newR)
				})
			},
		},
		{
			name: "without_id",
			body: "{\"name\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *FileService {
				repo := NewMockfileRepository(ctr)
				handlers := &FileService{
					repository: repo,
				}
				return handlers
			},
			expectedStatus: http.StatusBadRequest,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					newR := r.WithContext(context.WithValue(r.Context(), token.UserKey, &models.User{ID: 1}))
					next.ServeHTTP(w, newR)
				})
			},
		},
		{
			name: "not_exists",
			body: "{\"id\":1,\"name\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *FileService {
				repo := NewMockfileRepository(ctr)
				repo.EXPECT().GetByUserIDAndId(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, repositories.ErrNotExist).Times(1)
				repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(0)
				handlers := &FileService{
					repository: repo,
				}
				return handlers
			},
			expectedStatus: http.StatusNotFound,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					newR := r.WithContext(context.WithValue(r.Context(), token.UserKey, &models.User{ID: 1}))
					next.ServeHTTP(w, newR)
				})
			},
		},
		{
			name: "error_checking_exist",
			body: "{\"id\":1,\"name\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *FileService {
				repo := NewMockfileRepository(ctr)
				repo.EXPECT().GetByUserIDAndId(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("some_error")).Times(1)
				repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(0)
				handlers := &FileService{
					repository: repo,
				}
				return handlers
			},
			expectedStatus: http.StatusInternalServerError,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					newR := r.WithContext(context.WithValue(r.Context(), token.UserKey, &models.User{ID: 1}))
					next.ServeHTTP(w, newR)
				})
			},
		},
		{
			name: "error_update",
			body: "{\"id\":1,\"name\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *FileService {
				repo := NewMockfileRepository(ctr)
				repo.EXPECT().GetByUserIDAndId(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.FileContent{
					ID:        1,
					UserID:    1,
					Name:      []byte("aboba"),
					FilePath:  "test.txt",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil).Times(1)
				repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some error")).Times(1)
				handlers := &FileService{
					repository: repo,
				}
				return handlers
			},
			expectedStatus: http.StatusInternalServerError,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					newR := r.WithContext(context.WithValue(r.Context(), token.UserKey, &models.User{ID: 1}))
					next.ServeHTTP(w, newR)
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers := tt.setMocks()
			router := chi.NewRouter()
			router.Use(tt.authMiddleware)
			router.Put("/", handlers.UpdateFileHandler)
			// запускаем тестовый сервер, будет выбран первый свободный порт
			srv := httptest.NewServer(router)
			// останавливаем сервер после завершения теста
			defer srv.Close()
			request := resty.New().R()
			request.Method = http.MethodPut
			request.SetBody(tt.body)
			request.URL = srv.URL

			res, err := request.Send()
			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, tt.expectedStatus, res.StatusCode(), "unexpected response status code")
		})
	}
}

type MockErrorReader struct{}

func (m MockErrorReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("some error")
}

func TestFileService_saveFile(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func() (io.Reader, string, func())
		expectErr  bool
		perm       os.FileMode
	}{
		{
			name: "successful_save",
			setupMocks: func() (io.Reader, string, func()) {
				tmpDir := os.TempDir() + "test"
				os.MkdirAll(tmpDir, os.ModePerm)
				cleanup := func() {
					os.RemoveAll(tmpDir)
				}
				return io.NopCloser(strings.NewReader("test content")), tmpDir, cleanup
			},
			expectErr: false,
			perm:      os.ModePerm,
		},
		{
			name: "directory_creation_error",
			setupMocks: func() (io.Reader, string, func()) {
				invalidDir := "/unwritable-directory"
				cleanup := func() {}
				return io.NopCloser(strings.NewReader("test content")), invalidDir, cleanup
			},
			expectErr: true,
			perm:      os.ModePerm,
		},
		{
			name: "file_creation_error",
			setupMocks: func() (io.Reader, string, func()) {
				tmpDir := os.TempDir() + "test"
				os.MkdirAll(tmpDir, os.ModePerm)
				cleanup := func() {
					os.RemoveAll(tmpDir)
				}
				userDir := fmt.Sprintf("%s/1", tmpDir)
				if err := os.MkdirAll(userDir, 0444); err != nil {
					t.Fatalf("Failed to create restricted dir: %v", err)
				}
				os.Chmod(tmpDir, 0444) // Make directory unwritable
				return io.NopCloser(strings.NewReader("test content")), tmpDir, func() {
					os.Chmod(tmpDir, 0755) // Revert permissions
					cleanup()
				}
			},
			expectErr: true,
			perm:      0755,
		},
		{
			name: "file_copy_error",
			setupMocks: func() (io.Reader, string, func()) {
				tmpDir := os.TempDir() + "test"
				os.MkdirAll(tmpDir, os.ModePerm)
				cleanup := func() {
					os.RemoveAll(tmpDir)
				}
				brokenReader := &MockErrorReader{} // Emulate a nil reader
				return brokenReader, tmpDir, cleanup
			},
			expectErr: true,
			perm:      os.ModePerm,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileReader, path, cleanup := tt.setupMocks()
			defer cleanup()
			service := &FileService{
				filePath:    path,
				permissions: tt.perm,
			}
			result, err := service.saveFile(fileReader, 1)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestFileService_getSaveFileBody(t *testing.T) {
	tmpDir := os.TempDir() + "test"
	os.MkdirAll(tmpDir, os.ModePerm)
	defer os.RemoveAll(tmpDir)
	tests := []struct {
		name            string
		setupRequest    func() *resty.Request
		expectedName    []byte
		expectedComment string
		expectedError   bool
		formSize        int64
		dir             string
	}{
		{
			name: "valid_request",
			setupRequest: func() *resty.Request {
				filePath := fmt.Sprintf("%s/testFile", tmpDir)
				file, err := os.Create(filePath)
				file.Write([]byte("test contenttest contenttest contenttest contenttest contenttest contenttest contenttest contenttest content"))
				file.Close()
				if err != nil {
					t.Fatalf("Failed to create file: %v", err)
				}
				req := resty.New().R().
					SetFile("file", filePath).
					SetMultipartFormData(map[string]string{
						"name":    "aboba",
						"comment": "aboba",
					})
				req.Header.Set("Content-Type", "multipart/form-data")
				return req
			},
			expectedName:    []byte("aboba"),
			expectedComment: "aboba",
			expectedError:   false,
			formSize:        10 << 20,
			dir:             tmpDir,
		},
		{
			name: "not_file_request",
			setupRequest: func() *resty.Request {
				req := resty.New().R().
					SetMultipartFormData(map[string]string{
						"name":    "aboba",
						"comment": "aboba",
					})
				req.Header.Set("Content-Type", "multipart/form-data")
				return req
			},
			expectedName:    []byte("aboba"),
			expectedComment: "aboba",
			expectedError:   true,
			formSize:        512,
			dir:             tmpDir,
		},
		{
			name: "request_without_name",
			setupRequest: func() *resty.Request {
				filePath := fmt.Sprintf("%s/testFile", tmpDir)
				file, err := os.Create(filePath)
				file.Write([]byte("test contenttest contenttest contenttest contenttest contenttest contenttest contenttest contenttest content"))
				file.Close()
				if err != nil {
					t.Fatalf("Failed to create file: %v", err)
				}
				req := resty.New().R().
					SetFile("file", filePath).
					SetMultipartFormData(map[string]string{
						"comment": "aboba",
					})
				req.Header.Set("Content-Type", "multipart/form-data")
				return req
			},
			expectedName:    []byte("aboba"),
			expectedComment: "aboba",
			expectedError:   true,
			formSize:        10 << 20,
			dir:             tmpDir,
		},
		{
			name: "without_comment",
			setupRequest: func() *resty.Request {
				filePath := fmt.Sprintf("%s/testFile", tmpDir)
				file, err := os.Create(filePath)
				file.Write([]byte("test contenttest contenttest contenttest contenttest contenttest contenttest contenttest contenttest content"))
				file.Close()
				if err != nil {
					t.Fatalf("Failed to create file: %v", err)
				}
				req := resty.New().R().
					SetFile("file", filePath).
					SetMultipartFormData(map[string]string{
						"name": "aboba",
					})
				req.Header.Set("Content-Type", "multipart/form-data")
				return req
			},
			expectedName:    []byte("aboba"),
			expectedComment: "",
			expectedError:   false,
			formSize:        10 << 20,
			dir:             tmpDir,
		},
		{
			name: "error_creating_file",
			setupRequest: func() *resty.Request {
				filePath := fmt.Sprintf("%s/testFile", tmpDir)
				file, err := os.Create(filePath)
				file.Write([]byte("test contenttest contenttest contenttest contenttest contenttest contenttest contenttest contenttest content"))
				file.Close()
				if err != nil {
					t.Fatalf("Failed to create file: %v", err)
				}
				req := resty.New().R().
					SetFile("file", filePath).
					SetMultipartFormData(map[string]string{
						"name":    "aboba",
						"comment": "aboba",
					})
				req.Header.Set("Content-Type", "multipart/form-data")
				return req
			},
			expectedName:    []byte("aboba"),
			expectedComment: "aboba",
			expectedError:   true,
			formSize:        10 << 20,
			dir:             "/unwritable-directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &FileService{
				filePath:    tt.dir,
				maxFormSize: tt.formSize,
				permissions: os.ModePerm,
			}
			router := chi.NewRouter()
			router.Post("/", func(writer http.ResponseWriter, request *http.Request) {
				file, comment, err := s.getSaveFileBody(request, 1)
				if tt.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectedName, file.Name)
					assert.Equal(t, tt.expectedComment, comment.Comment)
				}
			})
			// запускаем тестовый сервер, будет выбран первый свободный порт
			srv := httptest.NewServer(router)
			// останавливаем сервер после завершения теста
			defer srv.Close()

			req := tt.setupRequest()
			_, err := req.Post(srv.URL)
			assert.NoError(t, err)
		})
	}
}

func TestFileService_SaveFileHandler(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	tmpDir := os.TempDir() + "test"
	os.MkdirAll(tmpDir, os.ModePerm)
	defer os.RemoveAll(tmpDir)
	tests := []struct {
		name           string
		setupRequest   func() *resty.Request
		formSize       int64
		dir            string
		getRepo        func() fileRepository
		expectedStatus int
		authMiddleware func(next http.Handler) http.Handler
	}{
		{
			name: "valid_request",
			setupRequest: func() *resty.Request {
				filePath := fmt.Sprintf("%s/testFile", tmpDir)
				file, err := os.Create(filePath)
				file.Write([]byte("test contenttest contenttest contenttest contenttest contenttest contenttest contenttest contenttest content"))
				file.Close()
				if err != nil {
					t.Fatalf("Failed to create file: %v", err)
				}
				req := resty.New().R().
					SetFile("file", filePath).
					SetMultipartFormData(map[string]string{
						"name":    "aboba",
						"comment": "aboba",
					})
				req.Header.Set("Content-Type", "multipart/form-data")
				return req
			},
			formSize: 10 << 20,
			dir:      tmpDir,
			getRepo: func() fileRepository {
				repo := NewMockfileRepository(ctr)
				repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				return repo
			},
			expectedStatus: http.StatusOK,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					newR := r.WithContext(context.WithValue(r.Context(), token.UserKey, &models.User{ID: 1}))
					next.ServeHTTP(w, newR)
				})
			},
		},
		{
			name: "not_file_request",
			setupRequest: func() *resty.Request {
				req := resty.New().R().
					SetMultipartFormData(map[string]string{
						"name":    "aboba",
						"comment": "aboba",
					})
				req.Header.Set("Content-Type", "multipart/form-data")
				return req
			},
			formSize: 512,
			dir:      tmpDir,
			getRepo: func() fileRepository {
				repo := NewMockfileRepository(ctr)
				repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(0)
				return repo
			},
			expectedStatus: http.StatusBadRequest,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					newR := r.WithContext(context.WithValue(r.Context(), token.UserKey, &models.User{ID: 1}))
					next.ServeHTTP(w, newR)
				})
			},
		},
		{
			name: "error_create",
			setupRequest: func() *resty.Request {
				filePath := fmt.Sprintf("%s/testFile", tmpDir)
				file, err := os.Create(filePath)
				file.Write([]byte("test contenttest contenttest contenttest contenttest contenttest contenttest contenttest contenttest content"))
				file.Close()
				if err != nil {
					t.Fatalf("Failed to create file: %v", err)
				}
				req := resty.New().R().
					SetFile("file", filePath).
					SetMultipartFormData(map[string]string{
						"name":    "aboba",
						"comment": "aboba",
					})
				req.Header.Set("Content-Type", "multipart/form-data")
				return req
			},
			formSize: 10 << 20,
			dir:      tmpDir,
			getRepo: func() fileRepository {
				repo := NewMockfileRepository(ctr)
				repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some_error")).Times(1)
				return repo
			},
			expectedStatus: http.StatusInternalServerError,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					newR := r.WithContext(context.WithValue(r.Context(), token.UserKey, &models.User{ID: 1}))
					next.ServeHTTP(w, newR)
				})
			},
		},
		{
			name: "not_authorized",
			setupRequest: func() *resty.Request {
				filePath := fmt.Sprintf("%s/testFile", tmpDir)
				file, err := os.Create(filePath)
				file.Write([]byte("test contenttest contenttest contenttest contenttest contenttest contenttest contenttest contenttest content"))
				file.Close()
				if err != nil {
					t.Fatalf("Failed to create file: %v", err)
				}
				req := resty.New().R().
					SetFile("file", filePath).
					SetMultipartFormData(map[string]string{
						"name":    "aboba",
						"comment": "aboba",
					})
				req.Header.Set("Content-Type", "multipart/form-data")
				return req
			},
			formSize: 10 << 20,
			dir:      tmpDir,
			getRepo: func() fileRepository {
				repo := NewMockfileRepository(ctr)
				repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some_error")).Times(0)
				return repo
			},
			expectedStatus: http.StatusUnauthorized,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					next.ServeHTTP(w, r)
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &FileService{
				filePath:    tt.dir,
				maxFormSize: tt.formSize,
				permissions: os.ModePerm,
				repository:  tt.getRepo(),
			}
			router := chi.NewRouter()
			router.Use(tt.authMiddleware)
			router.Post("/", s.SaveFileHandler)
			// запускаем тестовый сервер, будет выбран первый свободный порт
			srv := httptest.NewServer(router)
			// останавливаем сервер после завершения теста
			defer srv.Close()

			req := tt.setupRequest()
			res, err := req.Post(srv.URL)
			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, tt.expectedStatus, res.StatusCode(), "unexpected response status code")
		})
	}
}

func TestFileService_DownloadFileHandler(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	tmpDir := os.TempDir() + "test_download"
	os.MkdirAll(tmpDir, os.ModePerm)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name           string
		setupMocks     func() *FileService
		textId         string
		expectedStatus int
		authMiddleware func(next http.Handler) http.Handler
		setupFile      func() (string, func())
	}{
		{
			name:   "successful_download",
			textId: "1",
			setupMocks: func() *FileService {
				repo := NewMockfileRepository(ctr)
				repo.EXPECT().GetByUserIDAndId(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.FileContent{
					ID:        1,
					UserID:    1,
					Name:      []byte("testfile"),
					FilePath:  "testfile.txt",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil).Times(1)
				return &FileService{repository: repo, filePath: tmpDir}
			},
			expectedStatus: http.StatusOK,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					newR := r.WithContext(context.WithValue(r.Context(), token.UserKey, &models.User{ID: 1}))
					next.ServeHTTP(w, newR)
				})
			},
			setupFile: func() (string, func()) {
				filePath := fmt.Sprintf("%s/1/testfile.txt", tmpDir)
				os.MkdirAll(fmt.Sprintf("%s/1", tmpDir), os.ModePerm)
				file, err := os.Create(filePath)
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				file.Write([]byte("test content"))
				file.Close()
				return filePath, func() { os.RemoveAll(filePath) }
			},
		},
		{
			name:   "unauthorized_user",
			textId: "1",
			setupMocks: func() *FileService {
				repo := NewMockfileRepository(ctr)
				return &FileService{repository: repo, filePath: tmpDir}
			},
			expectedStatus: http.StatusUnauthorized,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					next.ServeHTTP(w, r)
				})
			},
			setupFile: func() (string, func()) {
				return "", func() {}
			},
		},
		{
			name:   "invalid_file_id",
			textId: "invalid",
			setupMocks: func() *FileService {
				repo := NewMockfileRepository(ctr)
				return &FileService{repository: repo, filePath: tmpDir}
			},
			expectedStatus: http.StatusBadRequest,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					newR := r.WithContext(context.WithValue(r.Context(), token.UserKey, &models.User{ID: 1}))
					next.ServeHTTP(w, newR)
				})
			},
			setupFile: func() (string, func()) {
				return "", func() {}
			},
		},
		{
			name:   "file_not_found",
			textId: "1",
			setupMocks: func() *FileService {
				repo := NewMockfileRepository(ctr)
				repo.EXPECT().GetByUserIDAndId(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, repositories.ErrNotExist).Times(1)
				return &FileService{repository: repo, filePath: tmpDir}
			},
			expectedStatus: http.StatusNotFound,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					newR := r.WithContext(context.WithValue(r.Context(), token.UserKey, &models.User{ID: 1}))
					next.ServeHTTP(w, newR)
				})
			},
			setupFile: func() (string, func()) {
				return "", func() {}
			},
		},
		{
			name:   "internal_server_error",
			textId: "1",
			setupMocks: func() *FileService {
				repo := NewMockfileRepository(ctr)
				repo.EXPECT().GetByUserIDAndId(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("some error")).Times(1)
				return &FileService{repository: repo, filePath: tmpDir}
			},
			expectedStatus: http.StatusInternalServerError,
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					newR := r.WithContext(context.WithValue(r.Context(), token.UserKey, &models.User{ID: 1}))
					next.ServeHTTP(w, newR)
				})
			},
			setupFile: func() (string, func()) {
				return "", func() {}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupMocks()
			router := chi.NewRouter()
			router.Use(tt.authMiddleware)
			router.Get("/{id}", service.DownloadFileHandler)

			testSrv := httptest.NewServer(router)
			defer testSrv.Close()

			// Setup test file if necessary
			filePath, cleanup := tt.setupFile()
			if cleanup != nil {
				defer cleanup()
			}

			request := resty.New().R()
			request.Method = http.MethodGet
			request.URL = testSrv.URL + "/" + tt.textId

			res, err := request.Send()
			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, tt.expectedStatus, res.StatusCode(), "unexpected response status code")

			if tt.expectedStatus == http.StatusOK {
				content, _ := os.ReadFile(filePath)
				assert.Equal(t, string(content), string(res.Body()), "file content does not match")
			}
		})
	}
}
