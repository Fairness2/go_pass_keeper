package token

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"passkeeper/internal/config"
	"passkeeper/internal/models"
	"passkeeper/internal/repositories"
	"testing"
	"time"
)

func TestAuthenticator_getToken(t *testing.T) {
	tests := []struct {
		name         string
		authHeader   string
		expected     string
		expectErr    bool
		expectedCode int
	}{
		{
			name:         "valid_token",
			authHeader:   "Bearer valid_token",
			expected:     "valid_token",
			expectErr:    false,
			expectedCode: http.StatusOK,
		},
		{
			name:         "missing_header",
			authHeader:   "",
			expectErr:    true,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "invalid_prefix",
			authHeader:   "Token something",
			expectErr:    true,
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			authenticator := &Authenticator{}
			token, err := authenticator.getToken(req)

			if tt.expectErr {
				assert.Errorf(t, err, "expected error: %v, got: %v", tt.expectErr, err)
			} else {
				assert.NoError(t, err, "expected error: %v, got: %v", tt.expectErr, err)
				assert.Equal(t, tt.expected, token, "expected token: %v, got: %v", tt.expected, token)
			}
		})
	}
}

func TestAuthenticator_getUserIdFromToken(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	tests := []struct {
		name      string
		wantID    int64
		expectErr bool
		getToken  func() *jwt.Token
	}{
		{
			name:      "valid_id_in_claims",
			wantID:    12345,
			expectErr: false,
			getToken: func() *jwt.Token {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"sub": "12345",
				})
				return token
			},
		},
		{
			name:      "missing_sub_claim",
			wantID:    0,
			expectErr: true,
			getToken: func() *jwt.Token {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"role": "admin",
				})
				return token
			},
		},
		{
			name:      "invalid_user_id_format",
			wantID:    0,
			expectErr: true,
			getToken: func() *jwt.Token {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"sub": "not-a-number",
				})
				return token
			},
		},
		{
			name:      "error_getting_user_id_from_claims",
			wantID:    0,
			expectErr: true,
			getToken: func() *jwt.Token {
				clm := NewMockClaims(ctr)
				clm.EXPECT().GetSubject().Return("", errors.New("error"))
				token := jwt.Token{Claims: clm}
				return &token
			},
		},
	}

	authenticator := &Authenticator{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.getToken()
			userID, err := authenticator.getUserIdFromToken(token)
			if tt.expectErr {
				assert.Error(t, err, "expected error: %v, got: %v", tt.expectErr, err)
			} else {
				assert.NoError(t, err, "expected error: %v, got: %v", tt.expectErr, err)
				assert.Equal(t, tt.wantID, userID, "expected userID: %v, got: %v", tt.wantID, userID)
			}
		})
	}
}

func TestAuthenticator_getUserById(t *testing.T) {
	expectedUser := &models.User{ID: 12345}
	dbErr := errors.New("database error")
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	tests := []struct {
		name       string
		userID     int64
		expectUser *models.User
		expectErr  error
		getRep     func() UserRepository
	}{
		{
			name:       "user_exists",
			userID:     12345,
			expectUser: expectedUser,
			getRep: func() UserRepository {
				rep := NewMockUserRepository(ctr)
				rep.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).Return(expectedUser, nil)
				return rep
			},
		},
		{
			name:       "user_does_not_exist",
			userID:     67890,
			expectUser: nil,
			expectErr:  ErrUserNotExists,
			getRep: func() UserRepository {
				rep := NewMockUserRepository(ctr)
				rep.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).Return(nil, repositories.ErrNotExist)
				return rep
			},
		},
		{
			name:       "repository_error",
			userID:     54321,
			expectUser: nil,
			expectErr:  dbErr,
			getRep: func() UserRepository {
				rep := NewMockUserRepository(ctr)
				rep.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).Return(nil, dbErr)
				return rep
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authenticator := &Authenticator{repository: tt.getRep()}
			user, err := authenticator.getUserById(context.TODO(), tt.userID)
			if tt.expectErr != nil {
				assert.ErrorIs(t, err, tt.expectErr, "expected error: %v, got: %v", tt.expectErr, err)
			} else {
				assert.NoErrorf(t, err, "expected error: %v, got: %v", tt.expectErr, err)
				assert.Equalf(t, tt.expectUser, user, "expected user: %+v, got: %+v", tt.expectUser, user)
			}
		})
	}
}

func TestAuthenticator_Middleware(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()

	tests := []struct {
		name            string
		authHeader      string
		setupGenerator  func() Generator
		setupRepository func() UserRepository
		expectedStatus  int
	}{
		{
			name:       "valid_token_and_user",
			authHeader: "Bearer valid_token",
			setupGenerator: func() Generator {
				gen := NewMockGenerator(ctr)
				gen.EXPECT().Parse("valid_token").Return(&jwt.Token{
					Claims: jwt.MapClaims{"sub": "12345"},
				}, nil).AnyTimes()
				return gen
			},
			setupRepository: func() UserRepository {
				rep := NewMockUserRepository(ctr)
				rep.EXPECT().GetUserByID(gomock.Any(), int64(12345)).Return(&models.User{ID: 12345}, nil).AnyTimes()
				return rep
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "invalid_subject_token_",
			authHeader: "Bearer valid_token",
			setupGenerator: func() Generator {
				gen := NewMockGenerator(ctr)
				gen.EXPECT().Parse("valid_token").Return(&jwt.Token{
					Claims: jwt.MapClaims{"sub": "invalid"},
				}, nil).AnyTimes()
				return gen
			},
			setupRepository: func() UserRepository {
				rep := NewMockUserRepository(ctr)
				rep.EXPECT().GetUserByID(gomock.Any(), int64(12345)).Return(&models.User{ID: 12345}, nil).AnyTimes()
				return rep
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing_authorization_header",
			authHeader: "",
			setupGenerator: func() Generator {
				return NewMockGenerator(ctr)
			},
			setupRepository: func() UserRepository {
				return NewMockUserRepository(ctr)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid_token",
			authHeader: "Bearer invalidauth",
			setupGenerator: func() Generator {
				gen := NewMockGenerator(ctr)
				gen.EXPECT().Parse("invalidauth").Return(nil, errors.New("invalid token")).AnyTimes()
				return gen
			},
			setupRepository: func() UserRepository {
				return NewMockUserRepository(ctr)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "user_not_found",
			authHeader: "Bearer valid_token",
			setupGenerator: func() Generator {
				gen := NewMockGenerator(ctr)
				gen.EXPECT().Parse("valid_token").Return(&jwt.Token{
					Claims: jwt.MapClaims{"sub": "12345"},
				}, nil).AnyTimes()
				return gen
			},
			setupRepository: func() UserRepository {
				rep := NewMockUserRepository(ctr)
				rep.EXPECT().GetUserByID(gomock.Any(), int64(12345)).Return(nil, repositories.ErrNotExist).AnyTimes()
				return rep
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "database_error",
			authHeader: "Bearer valid_token",
			setupGenerator: func() Generator {
				gen := NewMockGenerator(ctr)
				gen.EXPECT().Parse("valid_token").Return(&jwt.Token{
					Claims: jwt.MapClaims{"sub": "12345"},
				}, nil).AnyTimes()
				return gen
			},
			setupRepository: func() UserRepository {
				rep := NewMockUserRepository(ctr)
				rep.EXPECT().GetUserByID(gomock.Any(), int64(12345)).Return(nil, errors.New("database error")).AnyTimes()
				return rep
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authenticator := &Authenticator{
				generator:  tt.setupGenerator(),
				repository: tt.setupRepository(),
			}
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			middleware := authenticator.Middleware(nextHandler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rr := httptest.NewRecorder()

			middleware.ServeHTTP(rr, req)
			assert.Equal(t, tt.expectedStatus, rr.Code, "expected status: %v, got: %v", tt.expectedStatus, rr.Code)
		})
	}
}

func TestNewAuthenticator(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	tests := []struct {
		name            string
		dbPool          repositories.SQLExecutor
		jwtKeys         *config.Keys
		tokenExpiration time.Duration
		expectErr       bool
	}{
		{
			name:            "valid_inputs",
			dbPool:          NewMockSQLExecutor(ctr), // Mock database pool
			jwtKeys:         &config.Keys{Private: nil, Public: nil},
			tokenExpiration: time.Hour,
			expectErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authenticator := NewAuthenticator(tt.dbPool, tt.jwtKeys, tt.tokenExpiration)
			assert.NotNil(t, authenticator, "expected authenticator, got: %v", authenticator)
		})
	}
}
