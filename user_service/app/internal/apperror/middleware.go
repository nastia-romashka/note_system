package apperror

import (
	"errors"
	"net/http"

	"user_service/pkg/logging"
)

func Middleware(next func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.GetLogger().With(
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
		)

		err := next(w, r)
		if err == nil {
			return
		}

		w.Header().Set("Content-Type", "application/json")

		var appErr *AppError
		if !errors.As(err, &appErr) {
			appErr = SystemError(err).(*AppError)
			logger.Error(
				"unhandled request error",
				"status_code", http.StatusInternalServerError,
				"code", appErr.Code,
				"message", appErr.Message,
				"developer_message", appErr.DeveloperMessage,
				"error", err,
			)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write(appErr.Marshal())
			return
		}

		if errors.Is(appErr, ErrNotFound) {
			logger.Warn(
				"request failed with not found",
				"status_code", http.StatusNotFound,
				"code", appErr.Code,
				"message", appErr.Message,
			)
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write(appErr.Marshal())
			return
		}

		if errors.Is(appErr, ErrUnauthorized) {
			logger.Warn(
				"request failed with unauthorized",
				"status_code", http.StatusUnauthorized,
				"code", appErr.Code,
				"message", appErr.Message,
			)
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write(appErr.Marshal())
			return
		}

		logger.Warn(
			"request failed with app error",
			"status_code", http.StatusBadRequest,
			"code", appErr.Code,
			"message", appErr.Message,
			"developer_message", appErr.DeveloperMessage,
		)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(appErr.Marshal())
	}
}
