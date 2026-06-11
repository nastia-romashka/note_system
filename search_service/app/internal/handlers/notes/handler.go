package notes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"search_service/internal/apperror"
	"search_service/pkg/logging"
)

const (
	searchNotesURL = "/api/search/notes"
	indexNotesURL  = "/api/index/notes"
	indexNoteURL   = "/api/index/notes/{uuid}"
	importNotesURL = "/api/index/notes/import"
)

type NoteService interface {
	Search(q, workspaceID, categoryUUID string, tagUUIDs []string, page, perPage int) ([]SearchNote, error)
	Upsert(note IndexedNote) error
	UpsertMany(notes []IndexedNote) error
	Delete(noteUUID string) error
	DeleteByWorkspace(workspaceID string) error
}

type Handler struct {
	Logger      logging.Logger
	NoteService NoteService
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET "+searchNotesURL, apperror.Middleware(h.SearchNotes))
	mux.HandleFunc("POST "+indexNotesURL, apperror.Middleware(h.UpsertNote))
	mux.HandleFunc("DELETE "+indexNotesURL, apperror.Middleware(h.DeleteNotesByUser))
	mux.HandleFunc("POST "+importNotesURL, apperror.Middleware(h.UpsertNotes))
	mux.HandleFunc("DELETE "+indexNoteURL, apperror.Middleware(h.DeleteNote))
}

func (h *Handler) SearchNotes(w http.ResponseWriter, r *http.Request) error {
	workspaceID := r.URL.Query().Get("workspace_id")
	if workspaceID == "" {
		return apperror.BadRequestError("workspace_id is required")
	}

	page := parsePositiveInt(r.URL.Query().Get("page"), 1)
	perPage := parsePositiveInt(r.URL.Query().Get("per_page"), 50)

	notes, err := h.NoteService.Search(
		r.URL.Query().Get("q"),
		workspaceID,
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
	noteUUID := r.PathValue("uuid")
	if noteUUID == "" {
		return apperror.BadRequestError("empty note uuid")
	}

	if err := h.NoteService.Delete(noteUUID); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) DeleteNotesByUser(w http.ResponseWriter, r *http.Request) error {
	workspaceID := r.URL.Query().Get("workspace_id")
	if workspaceID == "" {
		return apperror.BadRequestError("workspace_id is required")
	}

	if err := h.NoteService.DeleteByWorkspace(workspaceID); err != nil {
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
