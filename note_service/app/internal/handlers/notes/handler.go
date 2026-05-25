package notes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"note_service/internal/apperror"
	"note_service/pkg/logging"
)

const (
	notesURL = "/api/notes"
	noteURL  = "/api/notes/:uuid"
	statsURL = "/api/stats"
)

type NoteService interface {
	GetOne(noteUUID, userUUID string) (Note, error)
	GetByCategoryUUID(categoryUUID, userUUID string) ([]Note, error)
	GetStats(userUUID string) (NoteStats, error)
	Create(dto CreateNoteDTO) (string, error)
	Update(noteUUID, userUUID string, dto UpdateNoteDTO, tagsUpdate bool) error
	Delete(noteUUID, userUUID string) error
}

type Handler struct {
	Logger      logging.Logger
	NoteService NoteService
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, notesURL, apperror.Middleware(h.GetNotesByCategory))
	router.HandlerFunc(http.MethodPost, notesURL, apperror.Middleware(h.CreateNote))
	router.HandlerFunc(http.MethodGet, noteURL, apperror.Middleware(h.GetNote))
	router.HandlerFunc(http.MethodPatch, noteURL, apperror.Middleware(h.PartiallyUpdateNote))
	router.HandlerFunc(http.MethodDelete, noteURL, apperror.Middleware(h.DeleteNote))
	router.HandlerFunc(http.MethodGet, statsURL, apperror.Middleware(h.GetStats))
}

func (h *Handler) GetNote(w http.ResponseWriter, r *http.Request) error {
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	noteUUID := params.ByName("uuid")
	if noteUUID == "" {
		return apperror.BadRequestError("empty note uuid")
	}
	userUUID, err := userUUIDFromQuery(r)
	if err != nil {
		return err
	}

	note, err := h.NoteService.GetOne(noteUUID, userUUID)
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

	notes, err := h.NoteService.GetByCategoryUUID(categoryUUID, userUUID)
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

	stats, err := h.NoteService.GetStats(userUUID)
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

	noteUUID, err := h.NoteService.Create(dto)
	if err != nil {
		return err
	}

	w.Header().Set("Location", fmt.Sprintf("%s/%s", notesURL, noteUUID))
	w.WriteHeader(http.StatusCreated)
	return nil
}

func (h *Handler) PartiallyUpdateNote(w http.ResponseWriter, r *http.Request) error {
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	noteUUID := params.ByName("uuid")
	if noteUUID == "" {
		return apperror.BadRequestError("empty note uuid")
	}
	userUUID, err := userUUIDFromQuery(r)
	if err != nil {
		return err
	}

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

	tagsUpdate := len(dto.Tags) > 0
	if len(dto.Tags) == 0 {
		var rawBody map[string]any
		if err = json.Unmarshal(bodyBytes, &rawBody); err != nil {
			h.Logger.Warn("failed to decode patch note raw payload", "note_uuid", noteUUID, "error", err)
			return apperror.BadRequestError("can't decode")
		}
		_, tagsUpdate = rawBody["tags"]
	}

	if err = h.NoteService.Update(noteUUID, userUUID, dto, tagsUpdate); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) DeleteNote(w http.ResponseWriter, r *http.Request) error {
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	noteUUID := params.ByName("uuid")
	if noteUUID == "" {
		return apperror.BadRequestError("empty note uuid")
	}
	userUUID, err := userUUIDFromQuery(r)
	if err != nil {
		return err
	}

	if err := h.NoteService.Delete(noteUUID, userUUID); err != nil {
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
