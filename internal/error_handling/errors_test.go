package error_handling

import (
	"errors"
	"net/http"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	t.Run("without wrapped error", func(t *testing.T) {
		err := NotFoundError("test not found")
		if err.Error() != "test not found" {
			t.Errorf("expected 'test not found', got '%s'", err.Error())
		}
	})

	t.Run("with wrapped error", func(t *testing.T) {
		wrapped := errors.New("db error")
		err := InternalErrorWrap("server error", wrapped)
		if err.Error() != "server error: db error" {
			t.Errorf("expected 'server error: db error', got '%s'", err.Error())
		}
	})

	t.Run("unwrap", func(t *testing.T) {
		wrapped := errors.New("db error")
		err := InternalErrorWrap("server error", wrapped)
		if !errors.Is(err, wrapped) {
			t.Error("expected errors.Is to match wrapped error")
		}
	})
}

func TestAppError_HTTPStatus(t *testing.T) {
	tests := []struct {
		name   string
		err    *AppError
		status int
	}{
		{"not found", NotFoundError("x"), http.StatusNotFound},
		{"bad request", BadRequestError("x"), http.StatusBadRequest},
		{"conflict", ConflictError("x"), http.StatusConflict},
		{"internal", InternalError("x"), http.StatusInternalServerError},
		{"method not allowed", MethodNotAllowedError(), http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.HTTPStatus != tt.status {
				t.Errorf("expected status %d, got %d", tt.status, tt.err.HTTPStatus)
			}
			if tt.err.Code == "" {
				t.Error("expected non-empty Code")
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	t.Run("builds correct APIError", func(t *testing.T) {
		err := BadRequestError("invalid input")
		if err.Code != CodeBadRequest {
			t.Errorf("expected CodeBadRequest, got %s", err.Code)
		}
		if err.Message != "invalid input" {
			t.Errorf("expected 'invalid input', got '%s'", err.Message)
		}
	})
}
