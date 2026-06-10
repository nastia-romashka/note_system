package notes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"

	"myproject/internal/apperror"
	categoryclient "myproject/internal/client/category"
	fileclient "myproject/internal/client/file"
	noteclient "myproject/internal/client/note"
	searchclient "myproject/internal/client/search"
	"myproject/internal/handlers/actionlog"
	"myproject/internal/searchsync"
	"myproject/pkg/logging"
	"myproject/pkg/middleware/jwt"
)

const (
	notesURL         = "/api/notes"
	noteURL          = "/api/notes/{uuid}"
	noteDuplicateURL = "/api/notes/{uuid}/duplicate"
	calendarURL      = "/api/calendar"
)

type note struct {
	Uuid         string                `json:"uuid,omitempty"`
	UserUuid     string                `json:"user_uuid,omitempty"`
	Header       string                `json:"header,omitempty"`
	Body         string                `json:"body,omitempty"`
	CreatedDate  int64                 `json:"created_date,omitempty"`
	UpdatedAt    int64                 `json:"updated_at,omitempty"`
	CategoryUuid string                `json:"category_uuid,omitempty"`
	Tags         []string              `json:"tags,omitempty"`
	Event        *noteclient.NoteEvent `json:"event,omitempty"`
}

type Handler struct {
	Logger          logging.Logger
	CategoryService categoryclient.CategoryService
	FileService     fileclient.FileService
	NoteService     noteclient.NoteService
	SearchService   searchclient.SearchService
	ActionRecorder  actionlog.Recorder
}

type duplicateNoteRequest struct {
	CategoryUuid string `json:"category_uuid"`
	Header       string `json:"header"`
}

type noteFile struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc(http.MethodGet+" "+notesURL, jwt.JWTMiddleware(apperror.Middleware(h.GetNotes)))
	mux.HandleFunc(http.MethodPost+" "+notesURL, jwt.JWTMiddleware(apperror.Middleware(h.CreateNote)))
	mux.HandleFunc(http.MethodPost+" "+noteDuplicateURL, jwt.JWTMiddleware(apperror.Middleware(h.DuplicateNote)))
	mux.HandleFunc(http.MethodPatch+" "+noteURL, jwt.JWTMiddleware(apperror.Middleware(h.PartiallyUpdateNote)))
	mux.HandleFunc(http.MethodDelete+" "+noteURL, jwt.JWTMiddleware(apperror.Middleware(h.DeleteNote)))
}

func (h *Handler) GetCalendarNotes(w http.ResponseWriter, r *http.Request) error {
	from, err := parseUnixQuery(r, "from")
	if err != nil {
		return err
	}
	to, err := parseUnixQuery(r, "to")
	if err != nil {
		return err
	}

	userUuid, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}

	notes, err := h.NoteService.GetCalendarNotes(r.Context(), from, to, userUuid)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(notes)
	return nil
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
	h.syncIndexedNote(r, userUuid, noteUuid, "create")
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
	noteUuid := r.PathValue("uuid")
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

func (h *Handler) DuplicateNote(w http.ResponseWriter, r *http.Request) error {
	sourceNoteUUID := r.PathValue("uuid")
	if sourceNoteUUID == "" {
		return apperror.BadRequestError("empty note uuid")
	}
	userUUID, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}

	var req duplicateNoteRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apperror.BadRequestError("can't decode")
	}
	defer r.Body.Close()

	req.Header = strings.TrimSpace(req.Header)
	req.CategoryUuid = strings.TrimSpace(req.CategoryUuid)
	if req.Header == "" {
		return apperror.BadRequestError("header is required")
	}
	if err = h.ensureCategoryBelongsToUser(r, userUUID, req.CategoryUuid); err != nil {
		return err
	}

	sourceNoteData, err := h.NoteService.GetNote(r.Context(), sourceNoteUUID, userUUID)
	if err != nil {
		return err
	}

	var sourceNote noteclient.Note
	if err = json.Unmarshal(sourceNoteData, &sourceNote); err != nil {
		h.Logger.Error("failed to decode source note for duplicate", "note_uuid", sourceNoteUUID, "error", err)
		return fmt.Errorf("decode source note: %w", err)
	}

	createDTO := noteclient.CreateNoteDTO{
		UserUuid:     userUUID,
		Header:       req.Header,
		Body:         sourceNote.Body,
		CategoryUuid: req.CategoryUuid,
		Tags:         sourceNote.Tags,
	}

	newNoteUUID, err := h.NoteService.CreateNote(r.Context(), createDTO)
	if err != nil {
		return err
	}

	h.syncCreatedNoteNode(r, userUUID, newNoteUUID, createDTO)
	h.syncIndexedNote(r, userUUID, newNoteUUID, "duplicate")

	if err = h.duplicateNoteFiles(r, userUUID, sourceNoteUUID, newNoteUUID); err != nil {
		h.Logger.Warn(
			"failed to duplicate note files, rolling back duplicated note",
			"user_uuid", userUUID,
			"source_note_uuid", sourceNoteUUID,
			"note_uuid", newNoteUUID,
			"error", err,
		)
		h.rollbackDuplicatedNote(r, userUUID, newNoteUUID)
		return err
	}

	h.ActionRecorder.Record(r, userUUID, "note.created", "note", newNoteUUID, map[string]any{
		"category_uuid":    req.CategoryUuid,
		"header":           req.Header,
		"tags_count":       len(sourceNote.Tags),
		"source_note_uuid": sourceNoteUUID,
		"duplicated_note":  true,
	})

	newNoteData, err := h.NoteService.GetNote(r.Context(), newNoteUUID, userUUID)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", fmt.Sprintf("%s/%s", notesURL, newNoteUUID))
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(newNoteData)
	return nil
}

func (h *Handler) PartiallyUpdateNote(w http.ResponseWriter, r *http.Request) error {
	noteUuid := r.PathValue("uuid")
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
	h.syncIndexedNote(r, userUuid, noteUuid, "update")
	h.ActionRecorder.Record(r, userUuid, "note.updated", "note", noteUuid, map[string]any{
		"category_uuid": n.CategoryUuid,
		"header":        n.Header,
		"tags_count":    len(n.Tags),
	})

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) DeleteNote(w http.ResponseWriter, r *http.Request) error {
	noteUuid := r.PathValue("uuid")
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
	h.syncDeletedIndex(r, userUuid, noteUuid)
	h.ActionRecorder.Record(r, userUuid, "note.deleted", "note", noteUuid, nil)

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
	createdDate, err := h.fetchNoteCreatedDate(r, userUuid, noteUuid)
	if err != nil {
		h.Logger.Warn(
			"failed to fetch note created_date for graph sync",
			"user_uuid", userUuid,
			"note_uuid", noteUuid,
			"error", err,
		)
	}

	err = h.CategoryService.CreateNoteNode(r.Context(), categoryclient.CreateGraphNoteDTO{
		Uuid:         noteUuid,
		UserUuid:     userUuid,
		CategoryUuid: noteDTO.CategoryUuid,
		Header:       noteDTO.Header,
		CreatedDate:  createdDate,
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

func (h *Handler) fetchNoteCreatedDate(r *http.Request, userUuid, noteUUID string) (int64, error) {
	noteData, err := h.NoteService.GetNote(r.Context(), noteUUID, userUuid)
	if err != nil {
		return 0, err
	}

	var note noteclient.Note
	if err = json.Unmarshal(noteData, &note); err != nil {
		return 0, fmt.Errorf("decode note for graph sync: %w", err)
	}

	return note.CreatedDate, nil
}

func (h *Handler) syncIndexedNote(r *http.Request, userUuid, noteUUID, action string) {
	noteData, err := h.NoteService.GetNote(r.Context(), noteUUID, userUuid)
	if err != nil {
		h.Logger.Warn(
			"failed to fetch note for search sync",
			"user_uuid", userUuid,
			"note_uuid", noteUUID,
			"action", action,
			"error", err,
		)
		return
	}

	var note noteclient.Note
	if err = json.Unmarshal(noteData, &note); err != nil {
		h.Logger.Warn(
			"failed to decode note for search sync",
			"user_uuid", userUuid,
			"note_uuid", noteUUID,
			"action", action,
			"error", err,
		)
		return
	}

	categoryData, err := h.CategoryService.GetUserCategories(r.Context(), userUuid)
	if err != nil {
		h.Logger.Warn(
			"failed to fetch categories for search sync",
			"user_uuid", userUuid,
			"note_uuid", noteUUID,
			"action", action,
			"error", err,
		)
		return
	}

	var categories []categoryclient.Category
	if err = json.Unmarshal(categoryData, &categories); err != nil {
		h.Logger.Warn(
			"failed to decode categories for search sync",
			"user_uuid", userUuid,
			"note_uuid", noteUUID,
			"action", action,
			"error", err,
		)
		return
	}

	tags, err := h.fetchTagsForNotes(r, userUuid, []noteclient.Note{note})
	if err != nil {
		h.Logger.Warn(
			"failed to fetch tags for search sync",
			"user_uuid", userUuid,
			"note_uuid", noteUUID,
			"action", action,
			"error", err,
		)
		return
	}

	document, err := searchsync.BuildIndexedNote(note, categories, tags)
	if err != nil {
		h.Logger.Warn(
			"failed to build indexed note document",
			"user_uuid", userUuid,
			"note_uuid", noteUUID,
			"action", action,
			"error", err,
		)
		return
	}

	if err = h.SearchService.UpsertNote(r.Context(), document); err != nil {
		h.Logger.Warn(
			"failed to sync note to search index",
			"user_uuid", userUuid,
			"note_uuid", noteUUID,
			"action", action,
			"error", err,
		)
	}
}

func (h *Handler) syncDeletedIndex(r *http.Request, userUuid, noteUUID string) {
	if err := h.SearchService.DeleteNote(r.Context(), noteUUID, userUuid); err != nil {
		h.Logger.Warn(
			"failed to delete note from search index",
			"user_uuid", userUuid,
			"note_uuid", noteUUID,
			"error", err,
		)
	}
}

func (h *Handler) duplicateNoteFiles(r *http.Request, userUUID, sourceNoteUUID, targetNoteUUID string) error {
	filesData, err := h.FileService.GetNoteFiles(r.Context(), sourceNoteUUID, userUUID)
	if err != nil {
		return err
	}

	var files []noteFile
	if err = json.Unmarshal(filesData, &files); err != nil {
		return fmt.Errorf("decode note files: %w", err)
	}

	for _, file := range files {
		response, downloadErr := h.FileService.DownloadNoteFile(r.Context(), sourceNoteUUID, file.ID, userUUID)
		if downloadErr != nil {
			return downloadErr
		}

		body := response.Body()
		if body == nil {
			return fmt.Errorf("downloaded file body is empty")
		}

		contentType := file.ContentType
		if contentType == "" {
			contentType = response.Header().Get("Content-Type")
		}

		_, _, uploadErr := h.FileService.UploadNoteFile(r.Context(), fileclient.UploadFileParams{
			NoteUUID:    targetNoteUUID,
			UserUUID:    userUUID,
			FileName:    safeFileName(file.Name, file.ID),
			ContentType: contentType,
			Size:        file.Size,
			Reader:      body,
		})
		_ = body.Close()
		if uploadErr != nil {
			return uploadErr
		}
	}

	return nil
}

func (h *Handler) rollbackDuplicatedNote(r *http.Request, userUUID, noteUUID string) {
	h.cleanupDuplicatedNoteFiles(r, userUUID, noteUUID)
	if err := h.NoteService.DeleteNote(r.Context(), noteUUID, userUUID); err != nil {
		h.Logger.Warn(
			"failed to rollback duplicated note in note service",
			"user_uuid", userUUID,
			"note_uuid", noteUUID,
			"error", err,
		)
	}
	h.syncDeletedNoteNode(r, userUUID, noteUUID)
	h.syncDeletedIndex(r, userUUID, noteUUID)
}

func (h *Handler) cleanupDuplicatedNoteFiles(r *http.Request, userUUID, noteUUID string) {
	filesData, err := h.FileService.GetNoteFiles(r.Context(), noteUUID, userUUID)
	if err != nil {
		h.Logger.Warn("failed to fetch duplicated note files for cleanup", "user_uuid", userUUID, "note_uuid", noteUUID, "error", err)
		return
	}

	var files []noteFile
	if err = json.Unmarshal(filesData, &files); err != nil {
		h.Logger.Warn("failed to decode duplicated note files for cleanup", "user_uuid", userUUID, "note_uuid", noteUUID, "error", err)
		return
	}

	for _, file := range files {
		if deleteErr := h.FileService.DeleteNoteFile(r.Context(), noteUUID, file.ID, userUUID); deleteErr != nil {
			h.Logger.Warn(
				"failed to cleanup duplicated note file",
				"user_uuid", userUUID,
				"note_uuid", noteUUID,
				"file_id", file.ID,
				"error", deleteErr,
			)
		}
	}
}

func (h *Handler) fetchTagsForNotes(r *http.Request, userUuid string, notes []noteclient.Note) ([]noteclient.Tag, error) {
	tagUUIDs := searchsync.CollectTagUUIDs(notes)
	if len(tagUUIDs) == 0 {
		return nil, nil
	}

	tagsData, err := h.NoteService.GetTags(r.Context(), tagUUIDs, userUuid)
	if err != nil {
		return nil, err
	}

	var tags []noteclient.Tag
	if err = json.Unmarshal(tagsData, &tags); err != nil {
		return nil, fmt.Errorf("decode tags response: %w", err)
	}

	return tags, nil
}

func safeFileName(name, fallback string) string {
	fileName := strings.TrimSpace(path.Base(name))
	if fileName == "" || fileName == "." || fileName == "/" {
		return fallback
	}

	return fileName
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
