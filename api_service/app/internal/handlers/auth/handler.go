package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"myproject/internal/apperror"
	userclient "myproject/internal/client/user"
	"myproject/internal/config"
	"myproject/pkg/cache"
	"myproject/pkg/logging"
	jwt2 "myproject/pkg/middleware/jwt"

	"github.com/cristalhq/jwt/v5"
	"github.com/julienschmidt/httprouter"
)

const (
	autURL    = "/api/auth"
	signupURL = "/api/signup"
)

type user struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type newUser struct {
	user
	Email string `json:"email"`
}

type refresh struct {
	RefreshToken string `json:"refresh_token"`
}

type Handler struct {
	Logger      logging.Logger
	RTCache     cache.Repository
	UserService userclient.UserService
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, autURL, apperror.Middleware(h.Auth))
	router.HandlerFunc(http.MethodPut, autURL, apperror.Middleware(h.Auth))
	router.HandlerFunc(http.MethodPost, signupURL, apperror.Middleware(h.Signup))
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) error {
	var nu newUser
	if err := json.NewDecoder(r.Body).Decode(&nu); err != nil {
		return apperror.BadRequestError("can't decode")
	}

	defer r.Body.Close()

	userUuid, err := h.UserService.CreateUser(r.Context(), userclient.CreateUserDTO{
		Username: nu.Username,
		Email:    nu.Email,
		Password: nu.Password,
	})
	if err != nil {
		return err
	}

	jsonBytes, errCode := h.generateAccessToken(userUuid, nu.Email)
	if errCode != 0 {
		return apperror.APIError(errCode, "API-50001", "token generation failed", "token generation failed")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201
	_, _ = w.Write(jsonBytes)
	return nil
}

func (h *Handler) Auth(w http.ResponseWriter, r *http.Request) error {
	var userUuid string
	var email string

	switch r.Method {
	case http.MethodPost:
		var u user
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			return apperror.BadRequestError("can't decode")
		}

		defer r.Body.Close()

		authUser, err := h.UserService.Authenticate(r.Context(), userclient.AuthUserDTO{
			Username: u.Username,
			Password: u.Password,
		})
		if err != nil {
			return err
		}
		userUuid = authUser.Uuid
		email = authUser.Email
	case http.MethodPut:
		var refreshTokenS refresh
		if err := json.NewDecoder(r.Body).Decode(&refreshTokenS); err != nil {
			return apperror.BadRequestError("can't decode")
		}
		defer r.Body.Close()

		userIdBytes, err := h.RTCache.Get([]byte(refreshTokenS.RefreshToken))
		if err != nil {
			return apperror.UnauthorizedError("invalid refresh token")
		}

		userUuid = string(userIdBytes)
		refreshUser, err := h.UserService.GetUser(r.Context(), userUuid)
		if err != nil {
			return err
		}
		email = refreshUser.Email
		h.RTCache.Del([]byte(refreshTokenS.RefreshToken))
	}

	jsonBytes, errCode := h.generateAccessToken(userUuid, email)
	if errCode != 0 {
		return apperror.APIError(errCode, "API-50001", "token generation failed", "token generation failed")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201
	_, _ = w.Write(jsonBytes)
	return nil
}

func (h *Handler) generateAccessToken(userUUID string, email string) ([]byte, int) {
	key := []byte(config.GetConfig().JWT.Secret)
	signer, err := jwt.NewSignerHS(jwt.HS256, key)
	if err != nil {
		return nil, 418
	}

	builder := jwt.NewBuilder(signer)

	claims := jwt2.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        userUUID,
			Audience:  []string{"users"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
		},
		Email: email,
	}

	token, err := builder.Build(claims)
	if err != nil {
		h.Logger.Error(err)
		return nil, http.StatusUnauthorized
	}

	h.Logger.Info("create refresh token")
	refreshTokenUuid := uuid.New()
	err = h.RTCache.Set([]byte(refreshTokenUuid.String()), []byte(userUUID), 0)
	if err != nil {
		h.Logger.Error(err)
		return nil, http.StatusInternalServerError
	}

	jsonBytes, err := json.Marshal(map[string]string{
		"token":         token.String(),
		"refresh_token": refreshTokenUuid.String(),
	})

	if err != nil {
		return nil, http.StatusInternalServerError
	}

	return jsonBytes, 0
}
