package jwt

import (
	"context"
	"encoding/json"
	"myproject/app/internal/config"
	"myproject/app/pkg/logging"
	"net/http"
	"strings"
	"time"

	"github.com/cristalhq/jwt/v5"
)

type UserClaims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
}

func JWTMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.GetLogger()

		authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		if len(authHeader) != 2 {
			logger.Error("malformed token")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("malformed token"))
			return
		}

		logger.Debug("create jwt verifier")
		jwtToken := authHeader[1]

		key := []byte(config.GetConfig().JWT.Secret)
		verifier, err := jwt.NewVerifierHS(jwt.HS256, key)
		if err != nil {
			unauthorized(w, err)
			return
		}

		logger.Debug("parse and verify token")
		token, err := jwt.Parse([]byte(jwtToken), verifier)
		if err != nil {
			unauthorized(w, err)
			return
		}

		logger.Debug("parse user claims")
		var uc UserClaims
		err = json.Unmarshal(token.Claims(), &uc)
		if err != nil {
			unauthorized(w, err)
			return
		}

		if valid := uc.IsValidAt(time.Now()); !valid {
			logger.Error("token has been expired")
			unauthorized(w, err)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", uc.ID)
		h(w, r.WithContext(ctx))
	}
}

func unauthorized(w http.ResponseWriter, err error) {
	logging.GetLogger().Error(err)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("unauthorized"))
}
