package token

import (
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"net/http/httptest"
	"testing"
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

			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err)
			}

			if token != tt.expected {
				t.Errorf("expected token: %v, got: %v", tt.expected, token)
			}
		})
	}
}

func TestAuthenticator_getUserIdFromToken(t *testing.T) {
	tests := []struct {
		name      string
		claims    jwt.Claims
		wantID    int64
		expectErr bool
	}{
		{
			name: "valid_id_in_claims",
			claims: jwt.MapClaims{
				"sub": "12345",
			},
			wantID:    12345,
			expectErr: false,
		},
		{
			name: "missing_sub_claim",
			claims: jwt.MapClaims{
				"role": "admin",
			},
			wantID:    0,
			expectErr: true,
		},
		{
			name: "invalid_user_id_format",
			claims: jwt.MapClaims{
				"sub": "not-a-number",
			},
			wantID:    0,
			expectErr: true,
		},
	}

	authenticator := &Authenticator{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, tt.claims)

			userID, err := authenticator.getUserIdFromToken(token)

			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err)
			}

			if userID != tt.wantID {
				t.Errorf("expected userID: %v, got: %v", tt.wantID, userID)
			}
		})
	}
}
