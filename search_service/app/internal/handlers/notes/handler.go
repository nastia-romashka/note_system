package notes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"

	"search_service/internal/apperror"
	"search_service/pkg/logging"
)

const (
	searchNotesURL = "/api/search/notes"
	indexNotesURL  = "/api/index/notes"
	indexNoteURL   = "/api/index/notes/:uuid"
	importNotesURL = "/api/index/notes/import"
)

type NoteService interface {
	Search(q, userUUID, categoryUUID string, tagUUIDs []string, page, perPage int) ([]SearchNote, error)
	Upsert(note IndexedNote) error
	UpsertMany(notes []IndexedNote) error
	Delete(noteUUID, userUUID string) error
	DeleteByUser(userUUID string) error
}

type Handler struct {
	Logger      logging.Logger
	NoteService NoteService
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, searchNotesURL, apperror.Middleware(h.SearchNotes))
	router.HandlerFunc(http.MethodPost, indexNotesURL, apperror.Middleware(h.UpsertNote))
	router.HandlerFunc(http.MethodDelete, indexNotesURL, apperror.Middleware(h.DeleteNotesByUser))
	router.HandlerFunc(http.MethodPost, importNotesURL, apperror.Middleware(h.UpsertNotes))
	router.HandlerFunc(http.MethodDelete, indexNoteURL, apperror.Middleware(h.DeleteNote))
}

func (h *Handler) SearchNotes(w http.ResponseWriter, r *http.Request) error {
	userUUID := r.URL.Query().Get("user_uuid")
	if userUUID == "" {
		return apperror.BadRequestError("user_uuid is required")
	}

	page := parsePositiveInt(r.URL.Query().Get("page"), 1)
	perPage := parsePositiveInt(r.URL.Query().Get("per_page"), 50)

	notes, err := h.NoteService.Search(
		r.URL.Query().Get("q"),
		userUUID,
		r.URL.Query().Get("category_uuid"),
		parseTagUUIDs(r.URL.Query()["tag_uuid"]),
		page,
		perPage,
	)
	if err != nil {
		return err
	}

	data, err := json.Marshal(notes)
	if err != nil {
		h.Logger.Error("failed to marshal search response", "error", err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) UpsertNote(w http.ResponseWriter, r *http.Request) error {
	var note IndexedNote
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
		return apperror.BadRequestError("can't decode")
	}

	if err := h.NoteService.Upsert(note); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) UpsertNotes(w http.ResponseWriter, r *http.Request) error {
	var notes []IndexedNote
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&notes); err != nil {
		return apperror.BadRequestError("can't decode")
	}

	if err := h.NoteService.UpsertMany(notes); err != nil {
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

	userUUID := r.URL.Query().Get("user_uuid")
	if userUUID == "" {
		return apperror.BadRequestError("user_uuid is required")
	}

	if err := h.NoteService.Delete(noteUUID, userUUID); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) DeleteNotesByUser(w http.ResponseWriter, r *http.Request) error {
	userUUID := r.URL.Query().Get("user_uuid")
	if userUUID == "" {
		return apperror.BadRequestError("user_uuid is required")
	}

	if err := h.NoteService.DeleteByUser(userUUID); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func parsePositiveInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}

	return value
}

func parseTagUUIDs(rawValues []string) []string {
	result := make([]string, 0, len(rawValues))
	for _, rawValue := range rawValues {
		for _, item := range strings.Split(rawValue, ",") {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			result = append(result, item)
		}
	}

	return result
}
