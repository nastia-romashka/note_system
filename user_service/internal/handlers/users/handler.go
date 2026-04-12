package users

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"user_service/internal/apperror"
	"user_service/pkg/logging"
)

const (
	usersURL        = "/api/users"
	userURL         = "/api/users/:uuid"
	authenticateURL = "/api/users/authenticate"
)

type UserService interface {
	GetOne(userUUID string) (User, error)
	Create(dto CreateUserDTO) (string, error)
	Authenticate(dto AuthUserDTO) (User, error)
}

type Handler struct {
	Logger      logging.Logger
	UserService UserService
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, usersURL, apperror.Middleware(h.CreateUser))
	router.HandlerFunc(http.MethodGet, userURL, apperror.Middleware(h.GetUser))
	router.HandlerFunc(http.MethodPost, authenticateURL, apperror.Middleware(h.Authenticate))
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) error {
	var dto CreateUserDTO
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.Logger.Warn("failed to decode create user payload", "error", err)
		return apperror.BadRequestError("can't decode")
	}

	userUUID, err := h.UserService.Create(dto)
	if err != nil {
		return err
	}

	w.Header().Set("Location", fmt.Sprintf("%s/%s", usersURL, userUUID))
	w.WriteHeader(http.StatusCreated)
	return nil
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) error {
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	userUUID := params.ByName("uuid")
	if userUUID == "" {
		return apperror.BadRequestError("empty user uuid")
	}

	user, err := h.UserService.GetOne(userUUID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(user)
	if err != nil {
		h.Logger.Error("failed to marshal user response", "user_uuid", userUUID, "error", err)
		return fmt.Errorf("marshal user: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) Authenticate(w http.ResponseWriter, r *http.Request) error {
	var dto AuthUserDTO
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.Logger.Warn("failed to decode auth payload", "error", err)
		return apperror.BadRequestError("can't decode")
	}

	user, err := h.UserService.Authenticate(dto)
	if err != nil {
		return err
	}

	data, err := json.Marshal(user)
	if err != nil {
		h.Logger.Error("failed to marshal auth user response", "error", err)
		return fmt.Errorf("marshal auth user: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}
