package notes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"note_service/internal/apperror"
	"note_service/pkg/logging"
)

const (
	notesURL    = "/api/notes"
	noteURL     = "/api/notes/{uuid}"
	statsURL    = "/api/stats"
	calendarURL = "/api/calendar"
)

type NoteService interface {
	GetOne(noteUUID, userUUID, workspaceID string) (Note, error)
	GetByCategoryUUID(categoryUUID, userUUID, workspaceID string) ([]Note, error)
	GetByEventRange(from, to int64, userUUID, workspaceID string) ([]Note, error)
	GetStats(userUUID, workspaceID string) (NoteStats, error)
	Create(dto CreateNoteDTO) (string, error)
	Update(noteUUID, userUUID, workspaceID string, dto UpdateNoteDTO, headerUpdate, bodyUpdate, categoryUpdate, tagsUpdate, eventUpdate bool) error
	Delete(noteUUID, userUUID, workspaceID string) error
}

type Handler struct {
	Logger      logging.Logger
	NoteService NoteService
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET "+notesURL, apperror.Middleware(h.GetNotesByCategory))
	mux.HandleFunc("GET "+calendarURL, apperror.Middleware(h.GetCalendarNotes))
	mux.HandleFunc("POST "+notesURL, apperror.Middleware(h.CreateNote))
	mux.HandleFunc("GET "+noteURL, apperror.Middleware(h.GetNote))
	mux.HandleFunc("PATCH "+noteURL, apperror.Middleware(h.PartiallyUpdateNote))
	mux.HandleFunc("DELETE "+noteURL, apperror.Middleware(h.DeleteNote))
	mux.HandleFunc("GET "+statsURL, apperror.Middleware(h.GetStats))
}

func (h *Handler) GetCalendarNotes(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := userUUIDFromQuery(r)
	if err != nil {
		return err
	}
	workspaceID := workspaceIDFromQuery(r)

	from, err := parseUnixQuery(r, "from")
	if err != nil {
		return err
	}
	to, err := parseUnixQuery(r, "to")
	if err != nil {
		return err
	}

	notes, err := h.NoteService.GetByEventRange(from, to, userUUID, workspaceID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(notes)
	if err != nil {
		h.Logger.Error("failed to marshal calendar notes response", "user_uuid", userUUID, "error", err)
		return fmt.Errorf("marshal calendar notes: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) GetNote(w http.ResponseWriter, r *http.Request) error {
	noteUUID := r.PathValue("uuid")
	if noteUUID == "" {
		return apperror.BadRequestError("empty note uuid")
	}
	userUUID, err := userUUIDFromQuery(r)
	if err != nil {
		return err
	}
	workspaceID := workspaceIDFromQuery(r)

	note, err := h.NoteService.GetOne(noteUUID, userUUID, workspaceID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(note)
	if err != nil {
		h.Logger.Error("failed to marshal note response", "note_uuid", noteUUID, "error", err)
		return fmt.Errorf("marshal note: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) GetNotesByCategory(w http.ResponseWriter, r *http.Request) error {
	categoryUUID := r.URL.Query().Get("category_uuid")
	if categoryUUID == "" {
		return apperror.BadRequestError("empty category_uuid")
	}
	userUUID, err := userUUIDFromQuery(r)
	if err != nil {
		return err
	}
	workspaceID := workspaceIDFromQuery(r)

	notes, err := h.NoteService.GetByCategoryUUID(categoryUUID, userUUID, workspaceID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(notes)
	if err != nil {
		h.Logger.Error("failed to marshal notes response", "category_uuid", categoryUUID, "error", err)
		return fmt.Errorf("marshal notes: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := userUUIDFromQuery(r)
	if err != nil {
		return err
	}
	workspaceID := workspaceIDFromQuery(r)

	stats, err := h.NoteService.GetStats(userUUID, workspaceID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(stats)
	if err != nil {
		h.Logger.Error("failed to marshal note stats response", "user_uuid", userUUID, "error", err)
		return fmt.Errorf("marshal note stats: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) CreateNote(w http.ResponseWriter, r *http.Request) error {
	var dto CreateNoteDTO
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.Logger.Warn("failed to decode create note payload", "error", err)
		return apperror.BadRequestError("can't decode")
	}
	if dto.AuthorUserUUID == "" {
		dto.AuthorUserUUID = dto.UserUuid
	}

	noteUUID, err := h.NoteService.Create(dto)
	if err != nil {
		return err
	}

	w.Header().Set("Location", fmt.Sprintf("%s/%s", notesURL, noteUUID))
	w.WriteHeader(http.StatusCreated)
	return nil
}

func (h *Handler) PartiallyUpdateNote(w http.ResponseWriter, r *http.Request) error {
	noteUUID := r.PathValue("uuid")
	if noteUUID == "" {
		return apperror.BadRequestError("empty note uuid")
	}
	userUUID, err := userUUIDFromQuery(r)
	if err != nil {
		return err
	}
	workspaceID := workspaceIDFromQuery(r)

	defer r.Body.Close()
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.Logger.Warn("failed to read patch note body", "note_uuid", noteUUID, "error", err)
		return apperror.BadRequestError("can't read request body")
	}

	var dto UpdateNoteDTO
	if err = json.Unmarshal(bodyBytes, &dto); err != nil {
		h.Logger.Warn("failed to decode patch note payload", "note_uuid", noteUUID, "error", err)
		return apperror.BadRequestError("can't decode")
	}

	var rawBody map[string]json.RawMessage
	if err = json.Unmarshal(bodyBytes, &rawBody); err != nil {
		h.Logger.Warn("failed to decode patch note raw payload", "note_uuid", noteUUID, "error", err)
		return apperror.BadRequestError("can't decode")
	}

	tagsUpdate := hasJSONField(rawBody, "tags")
	eventUpdate := hasJSONField(rawBody, "event")
	headerUpdate := hasJSONField(rawBody, "header")
	bodyUpdate := hasJSONField(rawBody, "body")
	categoryUpdate := hasJSONField(rawBody, "category_uuid")

	if err = h.NoteService.Update(
		noteUUID,
		userUUID,
		workspaceID,
		dto,
		headerUpdate,
		bodyUpdate,
		categoryUpdate,
		tagsUpdate,
		eventUpdate,
	); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) DeleteNote(w http.ResponseWriter, r *http.Request) error {
	noteUUID := r.PathValue("uuid")
	if noteUUID == "" {
		return apperror.BadRequestError("empty note uuid")
	}
	userUUID, err := userUUIDFromQuery(r)
	if err != nil {
		return err
	}
	workspaceID := workspaceIDFromQuery(r)

	if err := h.NoteService.Delete(noteUUID, userUUID, workspaceID); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func userUUIDFromQuery(r *http.Request) (string, error) {
	userUUID := r.URL.Query().Get("user_uuid")
	if userUUID == "" {
		return "", apperror.BadRequestError("user_uuid is required")
	}

	return userUUID, nil
}

func workspaceIDFromQuery(r *http.Request) string {
	return r.URL.Query().Get("workspace_id")
}

func parseUnixQuery(r *http.Request, key string) (int64, error) {
	rawValue := r.URL.Query().Get(key)
	if rawValue == "" {
		return 0, apperror.BadRequestError(key + " is required")
	}

	value, err := strconv.ParseInt(rawValue, 10, 64)
	if err != nil {
		return 0, apperror.BadRequestError("invalid " + key)
	}

	return value, nil
}

func hasJSONField(payload map[string]json.RawMessage, key string) bool {
	_, ok := payload[key]
	return ok
}
