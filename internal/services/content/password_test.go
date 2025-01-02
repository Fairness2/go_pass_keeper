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

func TestPasswordService_DeleteUserPasswords(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()

	tests := []struct {
		name           string
		setMocks       func() *PasswordService
		textId         string
		expectedStatus int
		authMiddleware func(next http.Handler) http.Handler
	}{
		{
			name:   "successful_delete",
			textId: "f25172cc-e7d9-404c-a52d-0353c253a422",
			setMocks: func() *PasswordService {
				repo := NewMockpasswordRepository(ctr)
				repo.EXPECT().DeleteByUserIDAndID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				handlers := &PasswordService{
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
			textId: "f25172cc-e7d9-404c-a52d-0353c253a422",
			setMocks: func() *PasswordService {
				repo := NewMockpasswordRepository(ctr)
				repo.EXPECT().DeleteByUserIDAndID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(0)
				handlers := &PasswordService{
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
			setMocks: func() *PasswordService {
				repo := NewMockpasswordRepository(ctr)
				repo.EXPECT().DeleteByUserIDAndID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(0)
				handlers := &PasswordService{
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
			textId: "f25172cc-e7d9-404c-a52d-0353c253a422",
			setMocks: func() *PasswordService {
				repo := NewMockpasswordRepository(ctr)
				repo.EXPECT().DeleteByUserIDAndID(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some error")).Times(1)
				handlers := &PasswordService{
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
			router.Delete("/{id}", handlers.DeleteUserPasswords)
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

func TestPasswordService_GetUserPasswords(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()

	tests := []struct {
		name           string
		setMocks       func() *PasswordService
		response       string
		expectedStatus int
		authMiddleware func(next http.Handler) http.Handler
	}{
		{
			name:     "successful_get",
			response: "[{\"id\":\"f25172cc-e7d9-404c-a52d-0353c253a422\",\"domen\":\"google.com\",\"username\":\"YWJvYmE=\",\"password\":\"YWJvYmE=\",\"comment\":\"aboba\"},{\"id\":\"1726ef63-756e-4dda-b669-0dcbef37a67f\",\"domen\":\"yandex.com\",\"username\":\"YWJvYmEy\",\"password\":\"YWJvYmEy\",\"comment\":\"aboba2\"}]",
			setMocks: func() *PasswordService {
				repo := NewMockpasswordRepository(ctr)
				repo.EXPECT().GetByUserID(gomock.Any(), gomock.Any()).Return([]models.PasswordWithComment{
					{
						PasswordContent: models.PasswordContent{
							ID:        "f25172cc-e7d9-404c-a52d-0353c253a422",
							UserID:    1,
							Domen:     "google.com",
							Username:  []byte("aboba"),
							Password:  []byte("aboba"),
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Comment: "aboba",
					},
					{
						PasswordContent: models.PasswordContent{
							ID:        "1726ef63-756e-4dda-b669-0dcbef37a67f",
							UserID:    1,
							Domen:     "yandex.com",
							Username:  []byte("aboba2"),
							Password:  []byte("aboba2"),
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Comment: "aboba2",
					},
				}, nil).Times(1)
				handlers := &PasswordService{
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
			setMocks: func() *PasswordService {
				repo := NewMockpasswordRepository(ctr)
				repo.EXPECT().GetByUserID(gomock.Any(), gomock.Any()).Return(nil, nil).Times(0)
				handlers := &PasswordService{
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
			setMocks: func() *PasswordService {
				repo := NewMockpasswordRepository(ctr)
				repo.EXPECT().GetByUserID(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error")).Times(1)
				handlers := &PasswordService{
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
			setMocks: func() *PasswordService {
				repo := NewMockpasswordRepository(ctr)
				repo.EXPECT().GetByUserID(gomock.Any(), gomock.Any()).Return([]models.PasswordWithComment{}, nil).Times(1)
				handlers := &PasswordService{
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
			router.Get("/", handlers.GetUserPasswords)
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

func TestPasswordService_UpdatePasswordHandler(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()

	tests := []struct {
		name           string
		setMocks       func() *PasswordService
		body           string
		expectedStatus int
		authMiddleware func(next http.Handler) http.Handler
	}{
		{
			name: "successful_update",
			body: "{\"id\":\"f25172cc-e7d9-404c-a52d-0353c253a422\",\"domen\":\"google.com\",\"username\":\"YWJvYmE=\",\"password\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *PasswordService {
				repo := NewMockpasswordRepository(ctr)
				repo.EXPECT().GetByUserIDAndId(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.PasswordContent{
					ID:        "f25172cc-e7d9-404c-a52d-0353c253a422",
					UserID:    1,
					Domen:     "google.com",
					Username:  []byte("aboba"),
					Password:  []byte("aboba"),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil).Times(1)
				repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				handlers := &PasswordService{
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
			body: "{\"id\":\"f25172cc-e7d9-404c-a52d-0353c253a422\",\"domen\":\"google.com\",\"username\":\"YWJvYmE=\",\"password\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *PasswordService {
				repo := NewMockpasswordRepository(ctr)
				handlers := &PasswordService{
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
			setMocks: func() *PasswordService {
				repo := NewMockpasswordRepository(ctr)
				handlers := &PasswordService{
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
			body: "{\"domen\":\"google.com\",\"username\":\"YWJvYmE=\",\"password\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *PasswordService {
				repo := NewMockpasswordRepository(ctr)
				handlers := &PasswordService{
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
			body: "{\"id\":\"f25172cc-e7d9-404c-a52d-0353c253a422\",\"domen\":\"google.com\",\"username\":\"YWJvYmE=\",\"password\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *PasswordService {
				repo := NewMockpasswordRepository(ctr)
				repo.EXPECT().GetByUserIDAndId(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, repositories.ErrNotExist).Times(1)
				repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(0)
				handlers := &PasswordService{
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
			body: "{\"id\":\"f25172cc-e7d9-404c-a52d-0353c253a422\",\"domen\":\"google.com\",\"username\":\"YWJvYmE=\",\"password\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *PasswordService {
				repo := NewMockpasswordRepository(ctr)
				repo.EXPECT().GetByUserIDAndId(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("some_error")).Times(1)
				repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(0)
				handlers := &PasswordService{
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
			body: "{\"id\":\"f25172cc-e7d9-404c-a52d-0353c253a422\",\"domen\":\"google.com\",\"username\":\"YWJvYmE=\",\"password\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *PasswordService {
				repo := NewMockpasswordRepository(ctr)
				repo.EXPECT().GetByUserIDAndId(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.PasswordContent{
					ID:        "f25172cc-e7d9-404c-a52d-0353c253a422",
					UserID:    1,
					Domen:     "google.com",
					Username:  []byte("aboba"),
					Password:  []byte("aboba"),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil).Times(1)
				repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some error")).Times(1)
				handlers := &PasswordService{
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
			router.Put("/", handlers.UpdatePasswordHandler)
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

func TestPasswordService_SaveTextHandler(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()

	tests := []struct {
		name           string
		setMocks       func() *PasswordService
		body           string
		expectedStatus int
		authMiddleware func(next http.Handler) http.Handler
	}{
		{
			name: "successful_create",
			body: "{\"domen\":\"google.com\",\"username\":\"YWJvYmE=\",\"password\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *PasswordService {
				repo := NewMockpasswordRepository(ctr)
				repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				handlers := &PasswordService{
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
			body: "{\"domen\":\"google.com\",\"username\":\"YWJvYmE=\",\"password\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *PasswordService {
				repo := NewMockpasswordRepository(ctr)
				handlers := &PasswordService{
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
			setMocks: func() *PasswordService {
				repo := NewMockpasswordRepository(ctr)
				handlers := &PasswordService{
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
			body: "{\"domen\":\"google.com\",\"username\":\"YWJvYmE=\",\"password\":\"YWJvYmE=\",\"comment\":\"aboba\"}",
			setMocks: func() *PasswordService {
				repo := NewMockpasswordRepository(ctr)
				repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some error")).Times(1)
				handlers := &PasswordService{
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
			router.Post("/", handlers.SavePasswordHandler)
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

func TestNewPasswordService(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	tests := []struct {
		name          string
		expectedError bool
		getExec       func() passwordRepository
	}{
		{
			name:          "successful_initialization",
			expectedError: false,
			getExec: func() passwordRepository {
				return NewMockpasswordRepository(ctr)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewPasswordService(tt.getExec())
			assert.NotNil(t, service, "NewPasswordService should not return nil")
		})
	}
}

func TestPasswordService_RegisterRoutes(t *testing.T) {
	routes := []url{
		{"/password", http.MethodPost},
		{"/password", http.MethodPut},
		{"/password", http.MethodGet},
		{"/password/aboba", http.MethodDelete},
	}
	tests := []struct {
		name           string
		middleware     func(http.Handler) http.Handler
		expectedRoutes []url
	}{
		{
			name:           "routes_without_middleware",
			middleware:     nil,
			expectedRoutes: routes,
		},
		{
			name: "routes_with_middleware",
			middleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("X-Test-Middleware", "Active")
					next.ServeHTTP(w, r)
				})
			},
			expectedRoutes: routes,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			handlers := &PasswordService{}
			router := chi.NewRouter()
			if tt.middleware != nil {
				router.Group(handlers.RegisterRoutes(tt.middleware))
			} else {
				router.Group(handlers.RegisterRoutes())
			}
			for _, r := range tt.expectedRoutes {
				res := router.Match(chi.NewRouteContext(), r.method, r.path)
				assert.True(t, res)
			}
		})
	}
}
