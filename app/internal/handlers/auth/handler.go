package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"myproject/app/internal/config"
	"myproject/app/pkg/cache"
	"myproject/app/pkg/logging"
	jwt2 "myproject/app/pkg/middleware/jwt"

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
	Logger  logging.Logger
	RTCache cache.Repository
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, autURL, h.Auth)
	router.HandlerFunc(http.MethodPut, autURL, h.Auth)
	router.HandlerFunc(http.MethodPost, signupURL, h.Signup)
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var nu newUser
	if err := json.NewDecoder(r.Body).Decode(&nu); err != nil {
		h.Logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	//TODO validtor username and password
	//TODO create user using UserService
	jsonBytes, errCode := h.generateAccessToken(w)
	if errCode != 0 {
		w.WriteHeader(errCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201
	w.Write(jsonBytes)
}

func (h *Handler) Auth(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var u user
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			h.Logger.Fatal(err)
		}

		defer r.Body.Close()
		// TODO client to UserService and get user by username and password
		// for now sub check
		if u.Username != "me" || u.Password != "pass" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	case http.MethodPut:
		var refreshTokenS refresh
		if err := json.NewDecoder(r.Body).Decode(&refreshTokenS); err != nil {
			h.Logger.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		userIdBytes, err := h.RTCache.Get([]byte(refreshTokenS.RefreshToken))
		h.Logger.Info("refresh token user_id: %", userIdBytes)
		if err != nil {
			h.Logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		h.RTCache.Del([]byte(refreshTokenS.RefreshToken))
	}

	jsonBytes, errCode := h.generateAccessToken(w)
	if errCode != 0 {
		w.WriteHeader(errCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201
	w.Write(jsonBytes)
}

func (h *Handler) generateAccessToken(w http.ResponseWriter) ([]byte, int) {
	key := []byte(config.GetConfig().JWT.Secret)
	signer, err := jwt.NewSignerHS(jwt.HS256, key)
	if err != nil {
		return nil, 418
	}

	builder := jwt.NewBuilder(signer)

	//TODO insert real user data
	claims := jwt2.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "uuid_here",
			Audience:  []string{"users"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * 60)),
		},
		Email: "email@will.be.here",
	}

	token, err := builder.Build(claims)
	if err != nil {
		h.Logger.Error(err)
		return nil, http.StatusUnauthorized
	}

	h.Logger.Info("create refresh token")
	refreshTokenUuid := uuid.New()
	err = h.RTCache.Set([]byte(refreshTokenUuid.String()), []byte("user_uuid"), 0)
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
