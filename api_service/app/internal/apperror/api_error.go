package apperror

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type AppError struct {
	StatusCode       int
	ErrorCode        string
	Message          string
	DeveloperMessage string
}

func (e *AppError) Error() string {
	return fmt.Sprintf(
		"api error: status=%d code=%s message=%s developer_message=%s",
		e.StatusCode,
		e.ErrorCode,
		e.Message,
		e.DeveloperMessage,
	)
}

func APIError(statusCode int, errorCode string, message string, developerMessage string) error {
	return &AppError{
		StatusCode:       statusCode,
		ErrorCode:        errorCode,
		Message:          message,
		DeveloperMessage: developerMessage,
	}
}

func BadRequestError(message string) error {
	if message == "" {
		message = "bad request"
	}

	return &AppError{
		StatusCode:       http.StatusBadRequest,
		ErrorCode:        "API-40000",
		Message:          message,
		DeveloperMessage: message,
	}
}

func UnauthorizedError(message string) error {
	if message == "" {
		message = "unauthorized"
	}

	return &AppError{
		StatusCode:       http.StatusUnauthorized,
		ErrorCode:        "API-40100",
		Message:          message,
		DeveloperMessage: message,
	}
}

func MissingUserUUIDError() error {
	return &AppError{
		StatusCode:       http.StatusTeapot,
		ErrorCode:        "API-41800",
		Message:          "there is no user_uuid in context",
		DeveloperMessage: "there is no user_uuid in context",
	}
}

func Middleware(next func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := next(w, r)
		if err == nil {
			return
		}

		appErr, ok := err.(*AppError)
		if !ok {
			w.WriteHeader(http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(appErr.StatusCode)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"code":              appErr.ErrorCode,
			"message":           appErr.Message,
			"developer_message": appErr.DeveloperMessage,
		})
	}
}
