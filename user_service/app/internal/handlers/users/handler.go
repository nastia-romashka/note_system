package users

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"

	"user_service/internal/apperror"
	"user_service/pkg/logging"
)

const (
	usersURL        = "/api/users"
	userURL         = "/api/users/:uuid"
	userProfileURL  = "/api/users/:uuid/profile"
	userActionsURL  = "/api/users/:uuid/actions"
	createActionURL = "/api/user-actions/:uuid"
	authenticateURL = "/api/users/authenticate"
)

type UserService interface {
	GetOne(userUUID string) (User, error)
	GetProfile(userUUID string) (UserProfile, error)
	UpdateProfile(userUUID string, dto UpdateUserProfileDTO) error
	Create(dto CreateUserDTO) (string, error)
	Authenticate(dto AuthUserDTO) (User, error)
	CreateAction(userUUID string, dto CreateUserActionDTO) error
	GetActions(userUUID string, limit, offset int) ([]UserAction, error)
}

type Handler struct {
	Logger      logging.Logger
	UserService UserService
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, usersURL, apperror.Middleware(h.CreateUser))
	router.HandlerFunc(http.MethodGet, userURL, apperror.Middleware(h.GetUser))
	router.HandlerFunc(http.MethodGet, userProfileURL, apperror.Middleware(h.GetUserProfile))
	router.HandlerFunc(http.MethodPatch, userProfileURL, apperror.Middleware(h.UpdateUserProfile))
	router.HandlerFunc(http.MethodGet, userActionsURL, apperror.Middleware(h.GetUserActions))
	router.HandlerFunc(http.MethodPost, createActionURL, apperror.Middleware(h.CreateUserAction))
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

func (h *Handler) GetUserProfile(w http.ResponseWriter, r *http.Request) error {
	userUUID := userUUIDFromParams(r)
	if userUUID == "" {
		return apperror.BadRequestError("empty user uuid")
	}

	profile, err := h.UserService.GetProfile(userUUID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(profile)
	if err != nil {
		h.Logger.Error("failed to marshal user profile response", "user_uuid", userUUID, "error", err)
		return fmt.Errorf("marshal user profile: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) error {
	userUUID := userUUIDFromParams(r)
	if userUUID == "" {
		return apperror.BadRequestError("empty user uuid")
	}

	var dto UpdateUserProfileDTO
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.Logger.Warn("failed to decode update user profile payload", "error", err)
		return apperror.BadRequestError("can't decode")
	}

	if err := h.UserService.UpdateProfile(userUUID, dto); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) GetUserActions(w http.ResponseWriter, r *http.Request) error {
	userUUID := userUUIDFromParams(r)
	if userUUID == "" {
		return apperror.BadRequestError("empty user uuid")
	}

	limit, err := positiveIntQuery(r, "limit", 50)
	if err != nil {
		return err
	}
	offset, err := positiveIntQuery(r, "offset", 0)
	if err != nil {
		return err
	}

	actions, err := h.UserService.GetActions(userUUID, limit, offset)
	if err != nil {
		return err
	}

	data, err := json.Marshal(actions)
	if err != nil {
		h.Logger.Error("failed to marshal user actions response", "user_uuid", userUUID, "error", err)
		return fmt.Errorf("marshal user actions: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) CreateUserAction(w http.ResponseWriter, r *http.Request) error {
	userUUID := userUUIDFromParams(r)
	if userUUID == "" {
		return apperror.BadRequestError("empty user uuid")
	}

	var dto CreateUserActionDTO
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.Logger.Warn("failed to decode create user action payload", "error", err)
		return apperror.BadRequestError("can't decode")
	}

	if err := h.UserService.CreateAction(userUUID, dto); err != nil {
		return err
	}

	w.WriteHeader(http.StatusCreated)
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

func userUUIDFromParams(r *http.Request) string {
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	return params.ByName("uuid")
}

func positiveIntQuery(r *http.Request, name string, defaultValue int) (int, error) {
	rawValue := r.URL.Query().Get(name)
	if rawValue == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(rawValue)
	if err != nil || value < 0 {
		return 0, apperror.BadRequestError(fmt.Sprintf("invalid %s", name))
	}

	return value, nil
}
