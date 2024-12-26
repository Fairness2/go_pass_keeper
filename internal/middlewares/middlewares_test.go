package middlewares

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONHeaders(t *testing.T) {
	tests := []struct {
		name           string
		nextHandler    http.Handler
		expectedHeader string
	}{
		{
			name: "default_content_type_set",
			nextHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// no specific action in the next handler
			}),
			expectedHeader: "application/json",
		},
		{
			name: "overriding_content_type_allowed",
			nextHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
			}),
			expectedHeader: "text/html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := JSONHeaders(tt.nextHandler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			resultHeader := rec.Header().Get("Content-Type")
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, tt.expectedHeader, resultHeader)
		})
	}
}
