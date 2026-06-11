package profile

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
	Profile        userclient.UserProfile `json:"profile"`
	Stats          SummaryStats           `json:"stats"`
	UpcomingEvents []noteclient.Note      `json:"upcoming_events"`
}

type SummaryStats struct {
	CategoriesCount int64  `json:"categories_count"`
	NotesCount      int64  `json:"notes_count"`
	TagsCount       int64  `json:"tags_count"`
	FilesCount      int64  `json:"files_count"`
	LastActivityAt  *int64 `json:"last_activity_at"`
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc(http.MethodPatch+" "+meURL, jwt.JWTMiddleware(apperror.Middleware(h.UpdateProfile)))
	mux.HandleFunc(http.MethodGet+" "+meSummaryURL, jwt.JWTMiddleware(apperror.Middleware(h.GetSummary)))
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}

	var dto userclient.UpdateUserProfileDTO
	defer r.Body.Close()

	if err = json.NewDecoder(r.Body).Decode(&dto); err != nil {
		return apperror.BadRequestError("can't decode")
	}

	if err = h.UserService.UpdateProfile(r.Context(), userUUID, dto); err != nil {
		return err
	}

	metadata := map[string]any{}
	if dto.Username != "" {
		metadata["username"] = dto.Username
	}
	if dto.Email != "" {
		metadata["email"] = dto.Email
	}
	if dto.NewPassword != "" {
		metadata["password_changed"] = true
	}
	_ = h.UserService.CreateAction(r.Context(), userUUID, userclient.CreateUserActionDTO{
		Action:     "profile.updated",
		EntityType: "user",
		EntityId:   userUUID,
		Status:     "success",
		Metadata:   metadata,
	})

	w.WriteHeader(http.StatusNoContent)
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

	personalWorkspace, err := h.UserService.GetPersonalWorkspace(r.Context(), userUUID)
	if err != nil {
		return err
	}

	categoryStats, err := h.CategoryService.GetStats(r.Context(), personalWorkspace.Uuid)
	if err != nil {
		return err
	}

	noteStats, err := h.NoteService.GetStats(r.Context(), userUUID, personalWorkspace.Uuid)
	if err != nil {
		return err
	}

	fileStats, err := h.FileService.GetStats(r.Context(), userUUID, personalWorkspace.Uuid)
	if err != nil {
		return err
	}

	upcomingEvents, err := h.fetchUpcomingEvents(r, userUUID, personalWorkspace.Uuid)
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
		UpcomingEvents: upcomingEvents,
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

func (h *Handler) fetchUpcomingEvents(r *http.Request, userUUID, workspaceID string) ([]noteclient.Note, error) {
	now := time.Now().Unix()
	const upcomingWindow = int64(30 * 24 * 60 * 60)

	notesData, err := h.NoteService.GetCalendarNotes(r.Context(), now, now+upcomingWindow, userUUID, workspaceID)
	if err != nil {
		return nil, err
	}

	var notes []noteclient.Note
	if err = json.Unmarshal(notesData, &notes); err != nil {
		return nil, fmt.Errorf("decode upcoming events: %w", err)
	}

	if len(notes) > 5 {
		notes = notes[:5]
	}

	return notes, nil
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
