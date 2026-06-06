package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"user_service/internal/apperror"
	"user_service/pkg/logging"
)

const (
	usersURL         = "/api/users"
	userURL          = "/api/users/{uuid}"
	userProfileURL   = "/api/users/{uuid}/profile"
	userActionsURL   = "/api/users/{uuid}/actions"
	createActionURL  = "/api/user-actions/{uuid}"
	authenticateURL  = "/api/users/authenticate"
	userSessionsURL  = "/api/user-sessions"
	rotateSessionURL = "/api/user-sessions/rotate"
	revokeSessionURL = "/api/user-sessions/revoke"
)

type UserService interface {
	GetOne(userUUID string) (User, error)
	GetProfile(userUUID string) (UserProfile, error)
	UpdateProfile(userUUID string, dto UpdateUserProfileDTO) error
	Create(dto CreateUserDTO) (string, error)
	Authenticate(dto AuthUserDTO) (User, error)
	CreateAction(userUUID string, dto CreateUserActionDTO) error
	GetActions(userUUID string, limit, offset int) ([]UserAction, error)
	CreateSession(dto CreateUserSessionDTO) error
	RotateSession(dto RotateUserSessionDTO) (UserSession, error)
	RevokeSession(refreshTokenHash string) error
}

type Handler struct {
	Logger      logging.Logger
	UserService UserService
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST "+usersURL, apperror.Middleware(h.CreateUser))
	mux.HandleFunc("GET "+userURL, apperror.Middleware(h.GetUser))
	mux.HandleFunc("GET "+userProfileURL, apperror.Middleware(h.GetUserProfile))
	mux.HandleFunc("PATCH "+userProfileURL, apperror.Middleware(h.UpdateUserProfile))
	mux.HandleFunc("GET "+userActionsURL, apperror.Middleware(h.GetUserActions))
	mux.HandleFunc("POST "+createActionURL, apperror.Middleware(h.CreateUserAction))
	mux.HandleFunc("POST "+authenticateURL, apperror.Middleware(h.Authenticate))
	mux.HandleFunc("POST "+userSessionsURL, apperror.Middleware(h.CreateUserSession))
	mux.HandleFunc("POST "+rotateSessionURL, apperror.Middleware(h.RotateUserSession))
	mux.HandleFunc("POST "+revokeSessionURL, apperror.Middleware(h.RevokeUserSession))
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) error {
	var dto CreateUserDTO
	defer r.Body.Close()

	if err := decodeJSONBody(r, &dto); err != nil {
		h.Logger.Warn("failed to decode create user payload", "error", err)
		return err
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
	userUUID := r.PathValue("uuid")
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

	if err := decodeJSONBody(r, &dto); err != nil {
		h.Logger.Warn("failed to decode update user profile payload", "error", err)
		return err
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

	if err := decodeJSONBody(r, &dto); err != nil {
		h.Logger.Warn("failed to decode create user action payload", "error", err)
		return err
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

	if err := decodeJSONBody(r, &dto); err != nil {
		h.Logger.Warn("failed to decode auth payload", "error", err)
		return err
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

func (h *Handler) CreateUserSession(w http.ResponseWriter, r *http.Request) error {
	var dto CreateUserSessionDTO
	defer r.Body.Close()

	if err := decodeJSONBody(r, &dto); err != nil {
		h.Logger.Warn("failed to decode create session payload", "error", err)
		return err
	}

	if err := h.UserService.CreateSession(dto); err != nil {
		return err
	}

	w.WriteHeader(http.StatusCreated)
	return nil
}

func (h *Handler) RotateUserSession(w http.ResponseWriter, r *http.Request) error {
	var dto RotateUserSessionDTO
	defer r.Body.Close()

	if err := decodeJSONBody(r, &dto); err != nil {
		h.Logger.Warn("failed to decode rotate session payload", "error", err)
		return err
	}

	session, err := h.UserService.RotateSession(dto)
	if err != nil {
		return err
	}

	data, err := json.Marshal(session)
	if err != nil {
		h.Logger.Error("failed to marshal rotate session response", "error", err)
		return fmt.Errorf("marshal rotate session: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) RevokeUserSession(w http.ResponseWriter, r *http.Request) error {
	var dto RevokeUserSessionDTO
	defer r.Body.Close()

	if err := decodeJSONBody(r, &dto); err != nil {
		h.Logger.Warn("failed to decode revoke session payload", "error", err)
		return err
	}

	if err := h.UserService.RevokeSession(dto.RefreshTokenHash); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func userUUIDFromParams(r *http.Request) string {
	return r.PathValue("uuid")
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

func decodeJSONBody(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		switch {
		case errors.Is(err, io.EOF):
			return apperror.BadRequestError("request body is required")
		default:
			return apperror.BadRequestError("invalid JSON body")
		}
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return apperror.BadRequestError("request body must contain a single JSON object")
	}

	return nil
}
