package profile

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"

	"myproject/internal/apperror"
	categoryclient "myproject/internal/client/category"
	fileclient "myproject/internal/client/file"
	noteclient "myproject/internal/client/note"
	userclient "myproject/internal/client/user"
	"myproject/pkg/logging"
	"myproject/pkg/middleware/jwt"
)

const (
	meURL        = "/api/me"
	meActionsURL = "/api/me/actions"
	meSummaryURL = "/api/me/summary"
)

type Handler struct {
	Logger          logging.Logger
	UserService     userclient.UserService
	CategoryService categoryclient.CategoryService
	NoteService     noteclient.NoteService
	FileService     fileclient.FileService
}

type Summary struct {
	Profile userclient.UserProfile `json:"profile"`
	Stats   SummaryStats           `json:"stats"`
}

type SummaryStats struct {
	CategoriesCount int64  `json:"categories_count"`
	NotesCount      int64  `json:"notes_count"`
	TagsCount       int64  `json:"tags_count"`
	FilesCount      int64  `json:"files_count"`
	LastActivityAt  *int64 `json:"last_activity_at"`
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, meURL, jwt.JWTMiddleware(apperror.Middleware(h.GetProfile)))
	router.HandlerFunc(http.MethodGet, meActionsURL, jwt.JWTMiddleware(apperror.Middleware(h.GetActions)))
	router.HandlerFunc(http.MethodGet, meSummaryURL, jwt.JWTMiddleware(apperror.Middleware(h.GetSummary)))
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}

	profile, err := h.UserService.GetProfile(r.Context(), userUUID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(profile)
	if err != nil {
		h.Logger.Error("failed to marshal profile response", "user_uuid", userUUID, "error", err)
		return fmt.Errorf("marshal profile: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}

	profile, err := h.UserService.GetProfile(r.Context(), userUUID)
	if err != nil {
		return err
	}

	categoryStats, err := h.CategoryService.GetStats(r.Context(), userUUID)
	if err != nil {
		return err
	}

	noteStats, err := h.NoteService.GetStats(r.Context(), userUUID)
	if err != nil {
		return err
	}

	fileStats, err := h.FileService.GetStats(r.Context(), userUUID)
	if err != nil {
		return err
	}

	lastActions, err := h.UserService.GetActions(r.Context(), userUUID, 1, 0)
	if err != nil {
		return err
	}

	var lastActivityAt *int64
	if len(lastActions) > 0 {
		value := lastActions[0].CreatedAt
		lastActivityAt = &value
	}

	summary := Summary{
		Profile: profile,
		Stats: SummaryStats{
			CategoriesCount: categoryStats.CategoriesCount,
			NotesCount:      noteStats.NotesCount,
			TagsCount:       noteStats.TagsCount,
			FilesCount:      fileStats.FilesCount,
			LastActivityAt:  lastActivityAt,
		},
	}

	data, err := json.Marshal(summary)
	if err != nil {
		h.Logger.Error("failed to marshal summary response", "user_uuid", userUUID, "error", err)
		return fmt.Errorf("marshal summary: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) GetActions(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}

	limit, err := positiveIntQuery(r, "limit", 50)
	if err != nil {
		return err
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := positiveIntQuery(r, "offset", 0)
	if err != nil {
		return err
	}

	actions, err := h.UserService.GetActions(r.Context(), userUUID, limit, offset)
	if err != nil {
		return err
	}

	data, err := json.Marshal(actions)
	if err != nil {
		h.Logger.Error("failed to marshal actions response", "user_uuid", userUUID, "error", err)
		return fmt.Errorf("marshal actions: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) userUUIDFromContext(r *http.Request) (string, error) {
	rawUserUUID := r.Context().Value("user_uuid")
	if rawUserUUID == nil {
		h.Logger.Error("there is no user_uuid in context")
		return "", apperror.MissingUserUUIDError()
	}

	userUUID, ok := rawUserUUID.(string)
	if !ok || userUUID == "" {
		h.Logger.Error("there is no user_uuid in context")
		return "", apperror.MissingUserUUIDError()
	}

	return userUUID, nil
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
