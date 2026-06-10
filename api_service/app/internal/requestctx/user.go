package requestctx

import (
	"net/http"

	"myproject/internal/apperror"
)

func UserUUID(r *http.Request) (string, error) {
	rawUserUUID := r.Context().Value("user_uuid")
	if rawUserUUID == nil {
		return "", apperror.MissingUserUUIDError()
	}

	userUUID, ok := rawUserUUID.(string)
	if !ok || userUUID == "" {
		return "", apperror.MissingUserUUIDError()
	}

	return userUUID, nil
}
