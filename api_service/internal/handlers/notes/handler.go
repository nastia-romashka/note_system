package notes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"myproject/internal/apperror"
	noteclient "myproject/internal/client/note"
	"myproject/pkg/logging"
	"myproject/pkg/middleware/jwt"
)

const (
	notesURL = "/api/notes"
	noteURL  = "/api/notes/:uuid"
)

type note struct {
	Uuid         string   `json:"uuid,omitempty"`
	Header       string   `json:"header,omitempty"`
	Body         string   `json:"body,omitempty"`
	CreatedDate  int64    `json:"created_date,omitempty"`
	CategoryUuid string   `json:"category_uuid,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}

type Handler struct {
	Logger      logging.Logger
	NoteService noteclient.NoteService
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, notesURL, jwt.JWTMiddleware(apperror.Middleware(h.GetNotes)))
	router.HandlerFunc(http.MethodPost, notesURL, jwt.JWTMiddleware(apperror.Middleware(h.CreateNote)))
	router.HandlerFunc(http.MethodGet, noteURL, jwt.JWTMiddleware(apperror.Middleware(h.GetNoteByUuid)))
	router.HandlerFunc(http.MethodPatch, noteURL, jwt.JWTMiddleware(apperror.Middleware(h.PartiallyUpdateNote)))
	router.HandlerFunc(http.MethodDelete, noteURL, jwt.JWTMiddleware(apperror.Middleware(h.DeleteNote)))
}

func (h *Handler) GetNotes(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	categoryUuid := r.URL.Query().Get("category_uuid")
	if categoryUuid == "" {
		return apperror.BadRequestError("empty category_uuid")
	}

	notes, err := h.NoteService.GetNotesByCategory(r.Context(), categoryUuid)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(notes)
	return nil
}

func (h *Handler) CreateNote(w http.ResponseWriter, r *http.Request) error {
	var crNote noteclient.CreateNoteDTO
	if err := json.NewDecoder(r.Body).Decode(&crNote); err != nil {
		return apperror.BadRequestError("can't decode")
	}
	defer r.Body.Close()

	noteUuid, err := h.NoteService.CreateNote(r.Context(), crNote)
	if err != nil {
		return err
	}

	w.Header().Set("Location", fmt.Sprintf("%s/%s", notesURL, noteUuid))
	w.WriteHeader(http.StatusCreated)
	return nil
}

func (h *Handler) GetNoteByUuid(w http.ResponseWriter, r *http.Request) error {
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	noteUuid := params.ByName("uuid")
	if noteUuid == "" {
		return apperror.BadRequestError("empty note uuid")
	}

	noteData, err := h.NoteService.GetNote(r.Context(), noteUuid)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(noteData)
	return nil
}

func (h *Handler) PartiallyUpdateNote(w http.ResponseWriter, r *http.Request) error {
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	noteUuid := params.ByName("uuid")
	if noteUuid == "" {
		return apperror.BadRequestError("empty note uuid")
	}

	var n noteclient.UpdateNoteDTO
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		return apperror.BadRequestError("can't decode")
	}
	defer r.Body.Close()

	if err := h.NoteService.UpdateNote(r.Context(), noteUuid, n); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) DeleteNote(w http.ResponseWriter, r *http.Request) error {
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	noteUuid := params.ByName("uuid")
	if noteUuid == "" {
		return apperror.BadRequestError("empty note uuid")
	}

	if err := h.NoteService.DeleteNote(r.Context(), noteUuid); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
