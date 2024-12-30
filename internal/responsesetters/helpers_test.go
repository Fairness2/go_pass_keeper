package responsesetters

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"passkeeper/internal/commonerrors"
	"passkeeper/internal/logger"
	"testing"
)

func TestSetHTTPResponse(t *testing.T) {
	cases := []struct {
		name             string
		mockStatus       int
		mockMessage      string
		expectedResponse string
	}{
		{
			name:             "ok_status",
			mockStatus:       http.StatusOK,
			mockMessage:      "OK",
			expectedResponse: "OK",
		},
		{
			name:             "not_found_status",
			mockStatus:       http.StatusNotFound,
			mockMessage:      "Not Found",
			expectedResponse: "Not Found",
		},
		{
			name:             "internal_server_error_status",
			mockStatus:       http.StatusInternalServerError,
			mockMessage:      "Internal Server Error",
			expectedResponse: "Internal Server Error",
		},
		{
			name:             "empty_message",
			mockStatus:       http.StatusInternalServerError,
			mockMessage:      "",
			expectedResponse: "",
		},
		{
			name:             "no_content_status",
			mockStatus:       http.StatusNoContent,
			mockMessage:      "Internal Server Error",
			expectedResponse: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			err := SetHTTPResponse(response, c.mockStatus, []byte(c.mockMessage))
			if err != nil {
				t.Errorf(
					"SetHTTPResponse error = %v, wantErr %v",
					err,
					false,
				)
			}
			result := response.Result()
			defer func() {
				if cErr := result.Body.Close(); cErr != nil {
					logger.Log.Warn(cErr)
				}
			}()

			if result.StatusCode != c.mockStatus {
				t.Errorf("Status is not expected. result.StatusCode = %v, want %v", result.StatusCode, c.mockStatus)
			}
			body, err := io.ReadAll(result.Body)
			if err != nil {
				t.Errorf("Read boay error is not expected. error = %v, wantErr %v", err, false)
			}
			if c.expectedResponse != string(body) {
				t.Errorf("Body is not expected. body = %v, want %v", string(body), c.expectedResponse)
			}
		})
	}
}

func TestGetErrorJSONBody(t *testing.T) {
	cases := []struct {
		name           string
		message        string
		status         int
		expectedOutput string
		expectError    bool
	}{
		{
			name:           "valid_input",
			message:        "Error message",
			status:         400,
			expectedOutput: `{"status":400,"message":"Error message"}`,
			expectError:    false,
		},
		{
			name:           "empty_message",
			message:        "",
			status:         200,
			expectedOutput: `{"status":200}`,
			expectError:    false,
		},
		{
			name:           "negative_status",
			message:        "Negative status",
			status:         -1,
			expectedOutput: `{"status":-1,"message":"Negative status"}`,
			expectError:    false,
		},
		{
			name:           "non_utf8_characters",
			message:        "\xbd\xb2",
			status:         500,
			expectedOutput: `{"status":500,"message":"\ufffd\ufffd"}`,
			expectError:    false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			output, err := GetErrorJSONBody(c.message, c.status)
			if (err != nil) != c.expectError {
				t.Errorf("Unexpected error status: got %v, expectError %v", err, c.expectError)
			}
			if string(output) != c.expectedOutput {
				t.Errorf("Unexpected output: got %s, expected %s", string(output), c.expectedOutput)
			}
		})
	}
}

func TestSetInternalError(t *testing.T) {
	cases := []struct {
		name               string
		mockError          error
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name:               "valid_error",
			mockError:          errors.New("some internal error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   "",
		},
		{
			name:               "nil_error",
			mockError:          nil,
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			SetInternalError(c.mockError, response)

			result := response.Result()
			defer func() {
				if cErr := result.Body.Close(); cErr != nil {
					logger.Log.Warn(cErr)
				}
			}()

			// Check the HTTP status code
			if result.StatusCode != c.expectedStatusCode {
				t.Errorf("Unexpected status code: got %d, want %d", result.StatusCode, c.expectedStatusCode)
			}

			// Check the response body
			body, err := io.ReadAll(result.Body)
			if err != nil {
				t.Errorf("Unexpected error reading body: %v", err)
			}
			if string(body) != c.expectedResponse {
				t.Errorf("Unexpected body: got %s, want %s", string(body), c.expectedResponse)
			}
		})
	}
}

func TestProcessRequestErrorWithBody(t *testing.T) {
	cases := []struct {
		name               string
		inputError         error
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "request_error_with_status",
			inputError: &commonerrors.RequestError{
				InternalError: errors.New("Bad Request"),
				HTTPStatus:    http.StatusBadRequest,
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"status":400,"message":"Bad Request"}`,
		},
		{
			name: "request_error_no_message",
			inputError: &commonerrors.RequestError{
				InternalError: errors.New(""),
				HTTPStatus:    http.StatusNotFound,
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   `{"status":404}`,
		},
		{
			name:               "generic_error",
			inputError:         errors.New("Something went wrong"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   `{"status":500,"message":"Something went wrong"}`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			ProcessRequestErrorWithBody(c.inputError, response)

			result := response.Result()
			defer func() {
				if cErr := result.Body.Close(); cErr != nil {
					logger.Log.Warn(cErr)
				}
			}()

			// Validate status code
			if result.StatusCode != c.expectedStatusCode {
				t.Errorf("Unexpected status code: got %d, want %d", result.StatusCode, c.expectedStatusCode)
			}

			// Validate response body
			body, err := io.ReadAll(result.Body)
			if err != nil {
				t.Errorf("Unexpected error reading body: %v", err)
			}
			if string(body) != c.expectedResponse {
				t.Errorf("Unexpected response body: got %s, want %s", string(body), c.expectedResponse)
			}
		})
	}
}

func TestProcessResponseWithStatus(t *testing.T) {
	cases := []struct {
		name               string
		message            string
		status             int
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:               "valid_status_ok",
			message:            "Success",
			status:             http.StatusOK,
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"status":200,"message":"Success"}`,
		},
		{
			name:               "valid_status_bad_request",
			message:            "Bad Request",
			status:             http.StatusBadRequest,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"status":400,"message":"Bad Request"}`,
		},
		{
			name:               "valid_status_internal_error",
			message:            "Internal Server Error",
			status:             http.StatusInternalServerError,
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       `{"status":500,"message":"Internal Server Error"}`,
		},
		{
			name:               "empty_message",
			message:            "",
			status:             http.StatusOK,
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"status":200}`,
		},
		{
			name:               "non_utf8_message",
			message:            "\xbd\xb2",
			status:             http.StatusServiceUnavailable,
			expectedStatusCode: http.StatusServiceUnavailable,
			expectedBody:       `{"status":503,"message":"\ufffd\ufffd"}`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			ProcessResponseWithStatus(c.message, c.status, response)

			result := response.Result()
			defer func() {
				if cErr := result.Body.Close(); cErr != nil {
					logger.Log.Warn(cErr)
				}
			}()

			// Validate the status code
			if result.StatusCode != c.expectedStatusCode {
				t.Errorf("Unexpected status code: got %d, want %d", result.StatusCode, c.expectedStatusCode)
			}

			// Validate the response body
			body, err := io.ReadAll(result.Body)
			if err != nil {
				t.Errorf("Unexpected error reading the body: %v", err)
			}
			if string(body) != c.expectedBody {
				t.Errorf("Unexpected response body: got %s, want %s", string(body), c.expectedBody)
			}
		})
	}
}
