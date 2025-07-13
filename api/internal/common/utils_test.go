package common

import (
	"log/slog"
	"net/http"
	"testing"
)

type MockLogger struct {
	LogFunc  func(message string, level slog.Level, args ...any) error
	LogCalls []struct {
		Message  string
		Severity slog.Level
		Args     []any
	}
}

func (m *MockLogger) Log(message string, level slog.Level, args ...any) error {
	if m.LogFunc != nil {
		_ = m.LogFunc(message, level, args...)
	}

	m.LogCalls = append(m.LogCalls, struct {
		Message  string
		Severity slog.Level
		Args     []any
	}{
		Message:  message,
		Severity: level,
		Args:     args,
	})

	return nil
}

func TestRequestErrorResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		message    string
		level      slog.Level

		expectedStatusCode  int
		expectedBody        string
		expectedLogMessage  string
		expectedLogSeverity slog.Level
	}{
		{
			name:       "HTTP Bad Request",
			statusCode: http.StatusBadRequest,
			message:    "Invalid request: invalid input provided",
			level:      slog.LevelError,

			expectedStatusCode:  http.StatusBadRequest,
			expectedBody:        "{\n\t\"message\": \"Invalid request: invalid input provided\"\n}",
			expectedLogMessage:  "Invalid request: invalid input provided",
			expectedLogSeverity: slog.LevelError,
		},
		{
			name:       "Internal Server Error",
			statusCode: http.StatusInternalServerError,
			message:    "Ups! Something went wrong...",
			level:      slog.LevelError,

			expectedStatusCode:  http.StatusInternalServerError,
			expectedBody:        "{\n\t\"message\": \"Ups! Something went wrong...\"\n}",
			expectedLogMessage:  "Ups! Something went wrong...",
			expectedLogSeverity: slog.LevelError,
		},
	}

	for _, tt := range tests {
		mockLogger := &MockLogger{}
		t.Run(tt.name, func(t *testing.T) {
			response, err := RequestErrorResponse(tt.statusCode, tt.message, mockLogger)
			if err != nil {
				t.Errorf("Error calling ErrorResponse: %v", err)
			}
			if response.StatusCode != tt.expectedStatusCode {
				t.Errorf("want %d, got %d", tt.expectedStatusCode, response.StatusCode)
			}
			if response.Body != tt.expectedBody {
				t.Errorf("want %s, got %s", tt.expectedBody, response.Body)
			}
			if len(mockLogger.LogCalls) == 0 {
				t.Errorf("error: logger not called")
			}
		})
	}
}
