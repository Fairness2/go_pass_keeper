package user

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"passkeeper/internal/commonerrors"
	"passkeeper/internal/config"
	"passkeeper/internal/models"
	"passkeeper/internal/payloads"
	"passkeeper/internal/repositories"
	"passkeeper/internal/token"
	"strings"
	"testing"
	"time"
)

func TestUserExists(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	dbErr := errors.New("database error")
	tests := []struct {
		name          string
		expectedError *commonerrors.RequestError
		getRepo       func() repository
	}{
		{
			name:          "user_does_not_exist",
			expectedError: nil,
			getRepo: func() repository {
				r := NewMockrepository(ctr)
				r.EXPECT().UserExists(gomock.Any(), gomock.Any()).Return(repositories.ErrNotExist).AnyTimes()
				return r
			},
		},
		{
			name:          "repository_internal_error",
			expectedError: &commonerrors.RequestError{InternalError: dbErr, HTTPStatus: http.StatusInternalServerError},
			getRepo: func() repository {
				r := NewMockrepository(ctr)
				r.EXPECT().UserExists(gomock.Any(), gomock.Any()).Return(dbErr).AnyTimes()
				return r
			},
		},
		{
			name:          "user_exists_without_errors",
			expectedError: &commonerrors.RequestError{InternalError: ErrUserAlreadyExists, HTTPStatus: http.StatusConflict},
			getRepo: func() repository {
				r := NewMockrepository(ctr)
				r.EXPECT().UserExists(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				return r
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			handlers := &Handlers{
				repository: tt.getRepo(),
			}
			ctx := context.Background()
			// Act
			err := handlers.userExists(ctx, "testuser")

			// Assert
			if tt.expectedError != nil {
				var requestError *commonerrors.RequestError
				ok := errors.As(err, &requestError)
				if !ok {
					assert.FailNow(t, "expected error of type RequestError, got %T", err)
				} else {
					assert.ErrorIsf(t, tt.expectedError.InternalError, requestError.InternalError, "expected internal error %v, got %v", tt.expectedError.InternalError, requestError.InternalError)
					assert.Equalf(t, tt.expectedError.HTTPStatus, requestError.HTTPStatus, "expected status %v, got %v", tt.expectedError.HTTPStatus, requestError.HTTPStatus)
				}
			} else if err != nil {
				assert.NoErrorf(t, err, "expected no error, got: %v", err)
			}
		})
	}
}

func TestGetUserByLogin(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	dbErr := errors.New("database error")
	expectedUser := &models.User{ID: 1, Login: "existingUser"}
	tests := []struct {
		name          string
		login         string
		expectedUser  *models.User
		expectedError *commonerrors.RequestError
		getRepo       func() repository
	}{
		{
			name:          "user_exists",
			login:         "existingUser",
			expectedUser:  expectedUser,
			expectedError: nil,
			getRepo: func() repository {
				r := NewMockrepository(ctr)
				r.EXPECT().GetUserByLogin(gomock.Any(), "existingUser").Return(expectedUser, nil).Times(1)
				return r
			},
		},
		{
			name:         "user_does_not_exist",
			login:        "nonexistentUser",
			expectedUser: nil,
			expectedError: &commonerrors.RequestError{
				InternalError: ErrLoginPasswordIncorrect,
				HTTPStatus:    http.StatusUnauthorized,
			},
			getRepo: func() repository {
				r := NewMockrepository(ctr)
				r.EXPECT().GetUserByLogin(gomock.Any(), "nonexistentUser").Return(nil, repositories.ErrNotExist).Times(1)
				return r
			},
		},
		{
			name:         "repository_internal_error",
			login:        "someUser",
			expectedUser: nil,
			expectedError: &commonerrors.RequestError{
				InternalError: dbErr,
				HTTPStatus:    http.StatusInternalServerError,
			},
			getRepo: func() repository {
				r := NewMockrepository(ctr)
				r.EXPECT().GetUserByLogin(gomock.Any(), "someUser").Return(nil, dbErr).Times(1)
				return r
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			handlers := &Handlers{
				repository: tt.getRepo(),
			}
			ctx := context.Background()

			// Act
			user, err := handlers.getUserByLogin(ctx, tt.login)

			// Assert
			assert.Equal(t, tt.expectedUser, user, "returned user should match expected user")
			if tt.expectedError != nil {
				var requestError *commonerrors.RequestError
				ok := errors.As(err, &requestError)
				if !ok {
					assert.FailNow(t, "expected error of type RequestError, got %T", err)
				} else {
					assert.ErrorIsf(t, tt.expectedError.InternalError, requestError.InternalError, "expected internal error %v, got %v", tt.expectedError.InternalError, requestError.InternalError)
					assert.Equalf(t, tt.expectedError.HTTPStatus, requestError.HTTPStatus, "expected HTTPStatus %v, got %v", tt.expectedError.HTTPStatus, requestError.HTTPStatus)
				}
			} else {
				assert.NoError(t, err, "expected no error, got one")
				assert.Equal(t, tt.expectedUser, user, "returned user should match expected user")
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	tests := []struct {
		name         string
		user         *models.User
		passwordHash string
		expectError  *commonerrors.RequestError
	}{
		{
			name: "valid_password",
			user: &models.User{
				PasswordHash: "df177f12d2e5b2977526db9b06be7f40fc41a9310f260b3d28851fb689c1da18",
			},
			passwordHash: "df177f12d2e5b2977526db9b06be7f40fc41a9310f260b3d28851fb689c1da18",
			expectError:  nil,
		},
		{
			name: "invalid_password",
			user: &models.User{
				PasswordHash: "df177f12d2e5b2977526db9b06be7f40fc41a9310f260b3d28851fb689c1da18",
			},
			passwordHash: "df177f12d2e5b2977526db9b06be7f40fc41a9310f260b3d28851fb689c1da19",
			expectError: &commonerrors.RequestError{
				InternalError: ErrLoginPasswordIncorrect,
				HTTPStatus:    http.StatusUnauthorized,
			},
		},
		{
			name:         "check_password_error",
			user:         &models.User{},
			passwordHash: "errorHash",
			expectError: &commonerrors.RequestError{
				InternalError: nil,
				HTTPStatus:    http.StatusInternalServerError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handlers{}

			// Act
			err := h.checkPassword(tt.user, tt.passwordHash)

			// Assert
			if tt.expectError != nil {
				var reqErr *commonerrors.RequestError
				assert.ErrorAs(t, err, &reqErr, "error should be of type RequestError")
				if tt.expectError.InternalError == nil {
					assert.Error(t, reqErr.InternalError)
				} else {
					assert.ErrorIsf(t, tt.expectError.InternalError, reqErr.InternalError, "expected internal error %v, got %v", tt.expectError.InternalError, reqErr.InternalError)
				}
				assert.Equal(t, tt.expectError.HTTPStatus, reqErr.HTTPStatus)
			} else {
				assert.NoError(t, err, "expected no error, but got one")
			}
		})
	}
}

func TestCreateJWTToken(t *testing.T) {
	jwtKeys := getTokens(t)

	validUser := &models.User{
		ID:    1,
		Login: "testuser",
	}

	tests := []struct {
		name          string
		user          *models.User
		tokenType     token.JWTType
		expiration    time.Duration
		mockKeys      *config.Keys
		expectedError bool
	}{
		{
			name:          "successful_token_creation",
			user:          validUser,
			tokenType:     token.JWTTypeAccess,
			expiration:    1 * time.Hour,
			mockKeys:      jwtKeys, // Assuming valid keys are set here for testing.
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			handlers := &Handlers{
				jwtKeys: tt.mockKeys,
			}

			// Act
			jwtToken, err := handlers.createJWTToken(tt.user, tt.tokenType, tt.expiration)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err, "did not expect an error")
				assert.NotEmpty(t, jwtToken, "expected token, got empty string")
			}
		})
	}
}

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name      string
		input     *payloads.Register
		expectErr bool
	}{
		{
			name: "valid_user",
			input: &payloads.Register{
				Login:    "validUser",
				Password: "validPassword123",
			},
			expectErr: false,
		},
		{
			name: "missing_password",
			input: &payloads.Register{
				Login:    "validUser",
				Password: "",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			handlers := &Handlers{
				hashKey: "veryWeakHashKey",
			}

			// Act
			user, err := handlers.createUser(tt.input.Login, tt.input.Password)

			// Assert
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.NotEmpty(t, user.PasswordHash, "password hash should be generated and non-empty")
			}
		})
	}
}

func TestCreateAndSaveUser(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()

	tests := []struct {
		name          string
		registerBody  *payloads.Register
		setupRepo     func() repository
		expectedError bool
	}{
		{
			name: "successful_user_creation",
			registerBody: &payloads.Register{
				Login:    "newUser",
				Password: "securePassword123",
			},
			setupRepo: func() repository {
				r := NewMockrepository(ctr)
				r.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				return r
			},
			expectedError: false,
		},
		{
			name: "generate_user_error",
			registerBody: &payloads.Register{
				Login:    "login",
				Password: "",
			},
			setupRepo: func() repository {
				r := NewMockrepository(ctr) // No calls expected to the repository
				return r
			},
			expectedError: true,
		},
		{
			name: "repository_save_error",
			registerBody: &payloads.Register{
				Login:    "testUser",
				Password: "securePassword123",
			},
			setupRepo: func() repository {
				r := NewMockrepository(ctr)
				r.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(errors.New("some db error")).Times(1)
				return r
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := tt.setupRepo()
			handlers := &Handlers{
				repository: repo,
				hashKey:    "veryWeakHashKey",
			}
			ctx := context.Background()
			// Act
			user, err := handlers.createAndSaveUser(ctx, tt.registerBody)

			// Assert
			if tt.expectedError {
				assert.Error(t, err, "unexpected error type")
			} else {
				assert.NotNil(t, user, "user should not be nil on success")
				assert.NoError(t, err, "unexpected error on success")
			}
		})
	}
}

type MockErrorReader struct{}

func (m MockErrorReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("some error")
}

func TestGetBody(t *testing.T) {
	tests := []struct {
		name          string
		body          func() io.ReadCloser
		expectedError bool
		expectedBody  *payloads.Register
	}{
		{
			name: "valid_request_body",
			body: func() io.ReadCloser {
				return io.NopCloser(strings.NewReader(`{"Login": "validUser", "Password": "ValidPassword123"}`))
			},
			expectedError: false,
			expectedBody: &payloads.Register{
				Login:    "validUser",
				Password: "ValidPassword123",
			},
		},
		{
			name: "closed_request_body",
			body: func() io.ReadCloser {
				return io.NopCloser(&MockErrorReader{})
			},
			expectedError: true,
			expectedBody:  nil,
		},
		{
			name: "malformed_json",
			body: func() io.ReadCloser {
				return io.NopCloser(strings.NewReader(`{"Login": "validUser", "Password": "ValidPassword123"`))
			},
			expectedError: true,
			expectedBody:  nil,
		},
		{
			name: "validation_failure",
			body: func() io.ReadCloser {
				return io.NopCloser(strings.NewReader(`{"Login": "", "Password": "ValidPassword123"}`))
			},
			expectedError: true,
			expectedBody:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			req := &http.Request{
				Body: tt.body(),
			}
			handlers := &Handlers{}

			// Act
			var result payloads.Register
			err := handlers.getBody(req, &result)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, &result)
			}
		})
	}
}

func TestCreateTokens(t *testing.T) {
	jwtKeys := getTokens(t)

	validUser := &models.User{
		ID:    1,
		Login: "testuser",
	}
	invalidUser := &models.User{
		ID:    -1,
		Login: "testuser",
	}
	tests := []struct {
		name      string
		user      *models.User
		wantError bool
		keys      *config.Keys
	}{
		{
			name:      "successful_token_creation",
			user:      validUser,
			wantError: false,
			keys:      jwtKeys,
		},
		{
			name:      "access_token_creation_error",
			user:      invalidUser,
			wantError: true,
			keys:      jwtKeys,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers := &Handlers{
				tokenExpiration:        time.Hour,
				refreshTokenExpiration: time.Hour * 24,
				jwtKeys:                tt.keys,
			}
			// Act
			auth, err := handlers.createTokens(tt.user)

			// Assert
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, auth)
			}
		})
	}
}

type MockResp struct {
	statusCode int
	body       []byte
	headers    http.Header
}

func (m *MockResp) Header() http.Header {
	return m.headers
}

func (m *MockResp) Write(bytes []byte) (int, error) {
	m.body = bytes
	return len(bytes), nil
}

func (m *MockResp) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

func TestSetResponse(t *testing.T) {
	tests := []struct {
		name           string
		payload        payloads.Authorization
		expectedBody   []byte
		expectedHeader string
		expectedCode   int
		expectError    bool
	}{
		{
			name: "valid_response",
			payload: payloads.Authorization{
				Token:   "test-token",
				Refresh: "refresh-token",
			},
			expectedBody:   []byte(`{"token":"test-token","refresh":"refresh-token"}`),
			expectedHeader: "Bearer test-token",
			expectedCode:   http.StatusOK,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handlers{}
			resp := &MockResp{headers: make(http.Header)}
			// Act
			h.setResponse(tt.payload, resp)

			// Assert
			assert.Equal(t, tt.expectedCode, resp.statusCode)
			if !tt.expectError {
				assert.Equal(t, tt.expectedCode, resp.statusCode)
				assert.Equal(t, tt.expectedBody, resp.body)
				assert.Equal(t, tt.expectedHeader, resp.headers.Get("Authorization"))
			}
		})
	}
}

func getTokens(t *testing.T) *config.Keys {
	pkey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(`-----BEGIN RSA PRIVATE KEY-----
MIICWwIBAAKBgF8sjeK3PDamGn0icENKSjwWpuWjmPrKEoVWXIO7os5iM1CZOn7h
qC4OgRKArfNC2BVa2zvVcrzRxRzFobyM6fblMbRzgE//5ct6YtpkWUgWEOZfXZ/X
FY5AUlngBKZtU2MS/CUX+PFXICUIVTDoCL6ngwNqTQj/dCin6E75Q1c/AgMBAAEC
gYAnYTsQHPs4LYB2WIKVBS80L7c8+3U4B9aj/zjmdQQHW1CaP9yZVWuOKwgzDLVt
GzJnm6Fs34PLJwzlO80RREkmEynnTYNVJejCLwuyT1oEGV6rFsql2HcIZ073NCxN
WakwL6Ay7QHH5S+hJDHCuxAx7kKoqiIXRcvbcwpRAnE5kQJBAKISE5uw1ejUocnE
ad7M2PVTz36ZS9d/3glpRQiQ2exeFRtcsq1J6O7G5OK62UMv3tcHyjf2suY0fPAA
jlPt/jcCQQCWVTylcs1Q319VRecJxSiCPjj97AA2VO1gcgzCWQ7mTp+N8QIegrD/
ZvvHqSLt79CexWrnOI6SvPuMf+8fwas5AkAlw76L7cW6bimQ4VKmFueLKs9TuZbB
jUsIuF3cpBwThsy2RoBf/rPnR7M33cAYdsQfKPKG3dZL6/kc15RSnEc7AkAgXTdS
MxXqjDw84nCr1Ms0xuqEF/Ovvrbf5Y3DpWKkyFZnO3SGVwJ96ZDY2hvP96oFFGFA
aBehlZfeFojHYG1ZAkEAnfwWAoPmvHxDaakOMsZg9PVVHIMhJ3Uck7lU5HKofHhq
rW4FGtaAhyoIZ2DQgctfe+PMcflOzkzkg9Cpqax7Cg==
-----END RSA PRIVATE KEY-----`))
	if err != nil {
		t.Fatal(err)
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(`-----BEGIN PUBLIC KEY-----
MIGeMA0GCSqGSIb3DQEBAQUAA4GMADCBiAKBgF8sjeK3PDamGn0icENKSjwWpuWj
mPrKEoVWXIO7os5iM1CZOn7hqC4OgRKArfNC2BVa2zvVcrzRxRzFobyM6fblMbRz
gE//5ct6YtpkWUgWEOZfXZ/XFY5AUlngBKZtU2MS/CUX+PFXICUIVTDoCL6ngwNq
TQj/dCin6E75Q1c/AgMBAAE=
-----END PUBLIC KEY-----`))
	if err != nil {
		t.Fatal(err)
	}

	return &config.Keys{
		Private: pkey,
		Public:  pubKey,
	}
}

func TestRegistrationHandler(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()

	jwtKeys := getTokens(t)

	tests := []struct {
		name           string
		setMocks       func() *Handlers
		requestBody    string
		expectedStatus int
	}{
		{
			name:        "successful_registration",
			requestBody: `{"Login":"newuser","Password":"Secure123"}`,
			setMocks: func() *Handlers {
				repo := NewMockrepository(ctr)
				repo.EXPECT().UserExists(gomock.Any(), gomock.Any()).Return(repositories.ErrNotExist).Times(1)
				repo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil).Times(1).
					Do(func(ctx context.Context, user *models.User) {
						user.ID = 1
					})

				handlers := &Handlers{
					repository:             repo,
					jwtKeys:                jwtKeys,
					tokenExpiration:        1 * time.Hour,
					refreshTokenExpiration: 1 * time.Hour * 24,
					hashKey:                "weakHashKey",
				}
				return handlers
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "user_already_exists",
			requestBody: `{"Login":"existinguser","Password":"Secure123"}`,
			setMocks: func() *Handlers {
				repo := NewMockrepository(ctr)
				repo.EXPECT().UserExists(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				handlers := &Handlers{
					repository: repo,
					jwtKeys:    jwtKeys,
				}
				return handlers
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:        "invalid_request_body",
			requestBody: ``,
			setMocks: func() *Handlers {
				handlers := &Handlers{}
				return handlers
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "user_creation_error",
			requestBody: `{"Login":"newuser","Password":"Secure123"}`,
			setMocks: func() *Handlers {
				repo := NewMockrepository(ctr)
				repo.EXPECT().UserExists(gomock.Any(), gomock.Any()).Return(repositories.ErrNotExist).Times(1)
				repo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(errors.New("some error")).Times(1)

				handlers := &Handlers{
					repository: repo,
					jwtKeys:    &config.Keys{},
					hashKey:    "weakHashKey",
				}
				return handlers
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "token_creation_error",
			requestBody: `{"Login":"newuser","Password":"Secure123"}`,
			setMocks: func() *Handlers {
				repo := NewMockrepository(ctr)
				repo.EXPECT().UserExists(gomock.Any(), gomock.Any()).Return(repositories.ErrNotExist).Times(1)
				repo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil).AnyTimes().
					Do(func(ctx context.Context, user *models.User) {
						user.ID = 0
					})

				handlers := &Handlers{
					repository: repo,
					jwtKeys:    jwtKeys,
					hashKey:    "weakHashKey",
				}
				return handlers
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers := tt.setMocks()
			router := chi.NewRouter()
			router.Post("/", handlers.RegistrationHandler)
			// запускаем тестовый сервер, будет выбран первый свободный порт
			srv := httptest.NewServer(router)
			// останавливаем сервер после завершения теста
			defer srv.Close()
			request := resty.New().R()
			request.Method = http.MethodPost
			request.Body = tt.requestBody
			request.URL = srv.URL

			res, err := request.Send()
			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, tt.expectedStatus, res.StatusCode(), "unexpected response status code")
		})
	}
}

func TestLoginHandler(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()

	jwtKeys := getTokens(t)

	tests := []struct {
		name           string
		setMocks       func() *Handlers
		requestBody    string
		expectedStatus int
	}{
		{
			name:        "successful_login",
			requestBody: `{"Login":"newuser","Password":"qwerty123"}`,
			setMocks: func() *Handlers {
				repo := NewMockrepository(ctr)
				repo.EXPECT().GetUserByLogin(gomock.Any(), gomock.Any()).Return(&models.User{
					ID:           1,
					Login:        "newuser",
					Password:     "qwerty123",
					PasswordHash: "d63f17da1a2ffa672d670eef0108bb54338b17c4e69f984eb378d2d3f23d2552",
				}, nil).Times(1)

				handlers := &Handlers{
					repository:             repo,
					jwtKeys:                jwtKeys,
					tokenExpiration:        1 * time.Hour,
					refreshTokenExpiration: 1 * time.Hour * 24,
					hashKey:                "weakHashKey",
				}
				return handlers
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "user_cant_create",
			requestBody: `{"Login":"existinguser","Password":"qwerty123"}`,
			setMocks: func() *Handlers {
				repo := NewMockrepository(ctr)

				handlers := &Handlers{
					repository: repo,
					jwtKeys:    jwtKeys,
				}
				return handlers
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "invalid_request_body",
			requestBody: ``,
			setMocks: func() *Handlers {
				handlers := &Handlers{}
				return handlers
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "user_not_exist",
			requestBody: `{"Login":"newuser","Password":"Secure123"}`,
			setMocks: func() *Handlers {
				repo := NewMockrepository(ctr)
				repo.EXPECT().GetUserByLogin(gomock.Any(), gomock.Any()).Return(nil, repositories.ErrNotExist).Times(1)

				handlers := &Handlers{
					repository: repo,
					jwtKeys:    &config.Keys{},
					hashKey:    "weakHashKey",
				}
				return handlers
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "token_creation_error",
			requestBody: `{"Login":"newuser","Password":"qwerty123"}`,
			setMocks: func() *Handlers {
				repo := NewMockrepository(ctr)
				repo.EXPECT().GetUserByLogin(gomock.Any(), gomock.Any()).Return(&models.User{
					ID:           0,
					Login:        "newuser",
					Password:     "qwerty123",
					PasswordHash: "d63f17da1a2ffa672d670eef0108bb54338b17c4e69f984eb378d2d3f23d2552",
				}, nil).Times(1)

				handlers := &Handlers{
					repository: repo,
					jwtKeys:    jwtKeys,
					hashKey:    "weakHashKey",
				}
				return handlers
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "password_mismatch",
			requestBody: `{"Login":"newuser","Password":"incorrectPassword"}`,
			setMocks: func() *Handlers {
				repo := NewMockrepository(ctr)
				repo.EXPECT().GetUserByLogin(gomock.Any(), gomock.Any()).Return(&models.User{
					ID:           1,
					Login:        "newuser",
					Password:     "qwerty123",
					PasswordHash: "df177f12d2e5b2977526db9b06be7f40fc41a9310f260b3d28851fb689c1da18",
				}, nil).Times(1)

				handlers := &Handlers{
					repository: repo,
					jwtKeys:    jwtKeys,
					hashKey:    "weakHashKey",
				}
				return handlers
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers := tt.setMocks()
			router := chi.NewRouter()
			router.Post("/", handlers.LoginHandler)
			// запускаем тестовый сервер, будет выбран первый свободный порт
			srv := httptest.NewServer(router)
			// останавливаем сервер после завершения теста
			defer srv.Close()
			request := resty.New().R()
			request.Method = http.MethodPost
			request.Body = tt.requestBody
			request.URL = srv.URL

			res, err := request.Send()
			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, tt.expectedStatus, res.StatusCode(), "unexpected response status code")
		})
	}
}

func TestNewHandlers(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	tests := []struct {
		name          string
		expectedError bool
		getConf       func() HandlerConfig
	}{
		{
			name:          "successful_initialization",
			expectedError: false,
			getConf: func() HandlerConfig {
				return HandlerConfig{Repository: NewMockrepository(ctr)}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewHandlers(tt.getConf())
			assert.NotNil(t, service, "NewTextService should not return nil")
		})
	}
}

func TestRegisterRoutes(t *testing.T) {
	tests := []struct {
		name           string
		middleware     func(http.Handler) http.Handler
		expectedRoutes map[string]string
	}{
		{
			name:           "routes_without_middleware",
			middleware:     nil,
			expectedRoutes: map[string]string{"/register": http.MethodPost, "/login": http.MethodPost},
		},
		{
			name: "routes_with_middleware",
			middleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("X-Test-Middleware", "Active")
					next.ServeHTTP(w, r)
				})
			},
			expectedRoutes: map[string]string{"/register": http.MethodPost, "/login": http.MethodPost},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			handlers := &Handlers{}
			router := chi.NewRouter()
			if tt.middleware != nil {
				router.Group(handlers.RegisterRoutes(tt.middleware))
			} else {
				router.Group(handlers.RegisterRoutes())
			}
			for route, method := range tt.expectedRoutes {
				res := router.Match(chi.NewRouteContext(), method, route)
				assert.True(t, res)
			}
		})
	}
}
