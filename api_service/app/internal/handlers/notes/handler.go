package notes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"myproject/internal/apperror"
	categoryclient "myproject/internal/client/category"
	noteclient "myproject/internal/client/note"
	"myproject/internal/handlers/actionlog"
	"myproject/pkg/logging"
	"myproject/pkg/middleware/jwt"
)

const (
	notesURL    = "/api/notes"
	noteURL     = "/api/notes/:uuid"
	noteLinkURL = "/api/notes/:uuid/links/:target_uuid"
)

type note struct {
	Uuid         string   `json:"uuid,omitempty"`
	UserUuid     string   `json:"user_uuid,omitempty"`
	Header       string   `json:"header,omitempty"`
	Body         string   `json:"body,omitempty"`
	CreatedDate  int64    `json:"created_date,omitempty"`
	CategoryUuid string   `json:"category_uuid,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}

type Handler struct {
	Logger          logging.Logger
	CategoryService categoryclient.CategoryService
	NoteService     noteclient.NoteService
	ActionRecorder  actionlog.Recorder
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, notesURL, jwt.JWTMiddleware(apperror.Middleware(h.GetNotes)))
	router.HandlerFunc(http.MethodPost, notesURL, jwt.JWTMiddleware(apperror.Middleware(h.CreateNote)))
	router.HandlerFunc(http.MethodGet, noteURL, jwt.JWTMiddleware(apperror.Middleware(h.GetNoteByUuid)))
	router.HandlerFunc(http.MethodPatch, noteURL, jwt.JWTMiddleware(apperror.Middleware(h.PartiallyUpdateNote)))
	router.HandlerFunc(http.MethodDelete, noteURL, jwt.JWTMiddleware(apperror.Middleware(h.DeleteNote)))
	router.HandlerFunc(http.MethodPost, noteLinkURL, jwt.JWTMiddleware(apperror.Middleware(h.LinkNotes)))
	router.HandlerFunc(http.MethodDelete, noteLinkURL, jwt.JWTMiddleware(apperror.Middleware(h.UnlinkNotes)))
}

func (h *Handler) GetNotes(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	categoryUuid := r.URL.Query().Get("category_uuid")
	if categoryUuid == "" {
		return apperror.BadRequestError("empty category_uuid")
	}
	userUuid, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}
	if err = h.ensureCategoryBelongsToUser(r, userUuid, categoryUuid); err != nil {
		return err
	}

	notes, err := h.NoteService.GetNotesByCategory(r.Context(), categoryUuid, userUuid)
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
	userUuid, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}
	if err = h.ensureCategoryBelongsToUser(r, userUuid, crNote.CategoryUuid); err != nil {
		return err
	}
	crNote.UserUuid = userUuid

	noteUuid, err := h.NoteService.CreateNote(r.Context(), crNote)
	if err != nil {
		return err
	}
	h.syncCreatedNoteNode(r, userUuid, noteUuid, crNote)
	h.ActionRecorder.Record(r, userUuid, "note.created", "note", noteUuid, map[string]any{
		"category_uuid": crNote.CategoryUuid,
		"header":        crNote.Header,
		"tags_count":    len(crNote.Tags),
	})

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
	userUuid, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}

	noteData, err := h.NoteService.GetNote(r.Context(), noteUuid, userUuid)
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
	userUuid, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}

	var n noteclient.UpdateNoteDTO
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		return apperror.BadRequestError("can't decode")
	}
	defer r.Body.Close()
	if n.CategoryUuid != "" {
		if err = h.ensureCategoryBelongsToUser(r, userUuid, n.CategoryUuid); err != nil {
			return err
		}
	}

	if err := h.NoteService.UpdateNote(r.Context(), noteUuid, userUuid, n); err != nil {
		return err
	}
	h.syncUpdatedNoteNode(r, userUuid, noteUuid, n)
	h.ActionRecorder.Record(r, userUuid, "note.updated", "note", noteUuid, map[string]any{
		"category_uuid": n.CategoryUuid,
		"header":        n.Header,
		"tags_count":    len(n.Tags),
	})

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) DeleteNote(w http.ResponseWriter, r *http.Request) error {
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	noteUuid := params.ByName("uuid")
	if noteUuid == "" {
		return apperror.BadRequestError("empty note uuid")
	}
	userUuid, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}

	if err := h.NoteService.DeleteNote(r.Context(), noteUuid, userUuid); err != nil {
		return err
	}
	h.syncDeletedNoteNode(r, userUuid, noteUuid)
	h.ActionRecorder.Record(r, userUuid, "note.deleted", "note", noteUuid, nil)

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) LinkNotes(w http.ResponseWriter, r *http.Request) error {
	userUuid, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}

	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	sourceNoteUuid := params.ByName("uuid")
	targetNoteUuid := params.ByName("target_uuid")
	if sourceNoteUuid == "" || targetNoteUuid == "" {
		return apperror.BadRequestError("empty note uuid")
	}

	if err = h.CategoryService.LinkNotes(r.Context(), sourceNoteUuid, targetNoteUuid, userUuid); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) UnlinkNotes(w http.ResponseWriter, r *http.Request) error {
	userUuid, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}

	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	sourceNoteUuid := params.ByName("uuid")
	targetNoteUuid := params.ByName("target_uuid")
	if sourceNoteUuid == "" || targetNoteUuid == "" {
		return apperror.BadRequestError("empty note uuid")
	}

	if err = h.CategoryService.UnlinkNotes(r.Context(), sourceNoteUuid, targetNoteUuid, userUuid); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
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

func (h *Handler) ensureCategoryBelongsToUser(r *http.Request, userUuid, categoryUuid string) error {
	if categoryUuid == "" {
		return apperror.BadRequestError("category_uuid is required")
	}

	data, err := h.CategoryService.GetUserCategories(r.Context(), userUuid)
	if err != nil {
		return err
	}

	var categories []categoryclient.Category
	if err = json.Unmarshal(data, &categories); err != nil {
		h.Logger.Error("failed to decode categories response", "error", err)
		return fmt.Errorf("decode categories response: %w", err)
	}

	if !containsCategory(categories, categoryUuid) {
		return apperror.BadRequestError("category does not belong to user")
	}

	return nil
}

func containsCategory(categories []categoryclient.Category, categoryUuid string) bool {
	for _, category := range categories {
		if category.Uuid == categoryUuid {
			return true
		}
		if containsCategory(category.Children, categoryUuid) {
			return true
		}
	}

	return false
}

func (h *Handler) syncCreatedNoteNode(
	r *http.Request,
	userUuid string,
	noteUuid string,
	noteDTO noteclient.CreateNoteDTO,
) {
	err := h.CategoryService.CreateNoteNode(r.Context(), categoryclient.CreateGraphNoteDTO{
		Uuid:         noteUuid,
		UserUuid:     userUuid,
		CategoryUuid: noteDTO.CategoryUuid,
		Header:       noteDTO.Header,
	})
	if err != nil {
		h.Logger.Warn(
			"failed to sync created note node",
			"user_uuid", userUuid,
			"note_uuid", noteUuid,
			"category_uuid", noteDTO.CategoryUuid,
			"error", err,
		)
	}
}

func (h *Handler) syncUpdatedNoteNode(
	r *http.Request,
	userUuid string,
	noteUuid string,
	noteDTO noteclient.UpdateNoteDTO,
) {
	err := h.CategoryService.UpdateNoteNode(r.Context(), noteUuid, categoryclient.UpdateGraphNoteDTO{
		UserUuid:     userUuid,
		CategoryUuid: noteDTO.CategoryUuid,
		Header:       noteDTO.Header,
	})
	if err != nil {
		h.Logger.Warn(
			"failed to sync updated note node",
			"user_uuid", userUuid,
			"note_uuid", noteUuid,
			"category_uuid", noteDTO.CategoryUuid,
			"error", err,
		)
	}
}

func (h *Handler) syncDeletedNoteNode(r *http.Request, userUuid string, noteUuid string) {
	if err := h.CategoryService.DeleteNoteNode(r.Context(), noteUuid, userUuid); err != nil {
		h.Logger.Warn(
			"failed to sync deleted note node",
			"user_uuid", userUuid,
			"note_uuid", noteUuid,
			"error", err,
		)
	}
}
