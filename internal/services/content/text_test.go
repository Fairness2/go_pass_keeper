package content

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"passkeeper/internal/models"
	"passkeeper/internal/repositories"
	"passkeeper/internal/token"
	"testing"
	"time"
)

func TestTextService_DeleteUserText(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()

	tests := []struct {
		name           string
		setMocks       func() *TextService
		textId         string
		expectedStatus int
		authMiddleware func(next http.Handler) http.Handler
	}{
		{
			name:   "successful_delete",
			textId: "1",
			setMocks: func() *TextService {
				repo := NewMocktextRepository(ctr)
				repo.EXPECT().DeleteByUserIDAndID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				handlers := &TextService{
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
			name:   "not_authorized",
			textId: "1",
			setMocks: func() *TextService {
				repo := NewMocktextRepository(ctr)
				repo.EXPECT().DeleteByUserIDAndID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(0)
				handlers := &TextService{
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
			name:   "incorrect_id",
			textId: "aboba",
			setMocks: func() *TextService {
				repo := NewMocktextRepository(ctr)
				repo.EXPECT().DeleteByUserIDAndID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(0)
				handlers := &TextService{
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
			name:   "delete_error",
			textId: "1",
			setMocks: func() *TextService {
				repo := NewMocktextRepository(ctr)
				repo.EXPECT().DeleteByUserIDAndID(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some error")).Times(1)
				handlers := &TextService{
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
			router.Delete("/{id}", handlers.DeleteUserText)
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

func TestTextService_GetUserTexts(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()

	tests := []struct {
		name           string
		setMocks       func() *TextService
		response       string
		expectedStatus int
		authMiddleware func(next http.Handler) http.Handler
	}{
		{
			name:     "successful_get",
			response: "[{\"id\":1,\"text_data\":\"YWJvYmE=\",\"comment\":\"aboba\"},{\"id\":2,\"text_data\":\"YWJvYmEy\",\"comment\":\"aboba2\"}]",
			setMocks: func() *TextService {
				repo := NewMocktextRepository(ctr)
				repo.EXPECT().GetByUserID(gomock.Any(), gomock.Any()).Return([]models.TextWithComment{
					{
						TextContent: models.TextContent{
							ID:        1,
							UserID:    1,
							TextData:  []byte("aboba"),
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Comment: "aboba",
					},
					{
						TextContent: models.TextContent{
							ID:        2,
							UserID:    1,
							TextData:  []byte("aboba2"),
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Comment: "aboba2",
					},
				}, nil).Times(1)
				handlers := &TextService{
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
			setMocks: func() *TextService {
				repo := NewMocktextRepository(ctr)
				repo.EXPECT().GetByUserID(gomock.Any(), gomock.Any()).Return(nil, nil).Times(0)
				handlers := &TextService{
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
			setMocks: func() *TextService {
				repo := NewMocktextRepository(ctr)
				repo.EXPECT().GetByUserID(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error")).Times(1)
				handlers := &TextService{
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
			setMocks: func() *TextService {
				repo := NewMocktextRepository(ctr)
				repo.EXPECT().GetByUserID(gomock.Any(), gomock.Any()).Return([]models.TextWithComment{}, nil).Times(1)
				handlers := &TextService{
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
			router.Get("/", handlers.GetUserTexts)
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

func TestTextService_UpdateTextHandler(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()

	tests := []struct {
		name           string
		setMocks       func() *TextService
		body           string
		expectedStatus int
		authMiddleware func(next http.Handler) http.Handler
	}{
		{
			name: "successful_update",
			body: "{\"id\":1,\"text_data\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *TextService {
				repo := NewMocktextRepository(ctr)
				repo.EXPECT().GetByUserIDAndId(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.TextContent{
					ID:        1,
					UserID:    1,
					TextData:  []byte("aboba"),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil).Times(1)
				repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				handlers := &TextService{
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
			body: "{\"id\":1,\"text_data\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *TextService {
				repo := NewMocktextRepository(ctr)
				handlers := &TextService{
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
			setMocks: func() *TextService {
				repo := NewMocktextRepository(ctr)
				handlers := &TextService{
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
			body: "{\"text_data\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *TextService {
				repo := NewMocktextRepository(ctr)
				handlers := &TextService{
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
			body: "{\"id\":1,\"text_data\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *TextService {
				repo := NewMocktextRepository(ctr)
				repo.EXPECT().GetByUserIDAndId(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, repositories.ErrNotExist).Times(1)
				repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(0)
				handlers := &TextService{
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
			body: "{\"id\":1,\"text_data\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *TextService {
				repo := NewMocktextRepository(ctr)
				repo.EXPECT().GetByUserIDAndId(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("some_error")).Times(1)
				repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(0)
				handlers := &TextService{
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
			body: "{\"id\":1,\"text_data\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *TextService {
				repo := NewMocktextRepository(ctr)
				repo.EXPECT().GetByUserIDAndId(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.TextContent{
					ID:        1,
					UserID:    1,
					TextData:  []byte("aboba"),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil).Times(1)
				repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some error")).Times(1)
				handlers := &TextService{
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
			router.Put("/", handlers.UpdateTextHandler)
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

func TestTextService_SaveTextHandler(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()

	tests := []struct {
		name           string
		setMocks       func() *TextService
		body           string
		expectedStatus int
		authMiddleware func(next http.Handler) http.Handler
	}{
		{
			name: "successful_create",
			body: "{\"text_data\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *TextService {
				repo := NewMocktextRepository(ctr)
				repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				handlers := &TextService{
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
			body: "{\"text_data\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *TextService {
				repo := NewMocktextRepository(ctr)
				handlers := &TextService{
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
			setMocks: func() *TextService {
				repo := NewMocktextRepository(ctr)
				handlers := &TextService{
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
			name: "with_id",
			body: "{\"id\":1,\"text_data\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *TextService {
				repo := NewMocktextRepository(ctr)
				handlers := &TextService{
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
			name: "error_update",
			body: "{\"text_data\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *TextService {
				repo := NewMocktextRepository(ctr)
				repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some error")).Times(1)
				handlers := &TextService{
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
			router.Post("/", handlers.SaveTextHandler)
			// запускаем тестовый сервер, будет выбран первый свободный порт
			srv := httptest.NewServer(router)
			// останавливаем сервер после завершения теста
			defer srv.Close()
			request := resty.New().R()
			request.Method = http.MethodPost
			request.SetBody(tt.body)
			request.URL = srv.URL

			res, err := request.Send()
			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, tt.expectedStatus, res.StatusCode(), "unexpected response status code")
		})
	}
}

func TestNewTextService(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	tests := []struct {
		name          string
		expectedError bool
		getExec       func() repositories.SQLExecutor
	}{
		{
			name:          "successful_initialization",
			expectedError: false,
			getExec: func() repositories.SQLExecutor {
				return NewMockSQLExecutor(ctr)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewTextService(tt.getExec())
			assert.NotNil(t, service, "NewTextService should not return nil")
		})
	}
}