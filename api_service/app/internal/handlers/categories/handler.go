package categories

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

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
	categoriesURL = "/api/categories"
	categoryURL   = "/api/categories/:uuid"
)

type Handler struct {
	CategoryService categoryclient.CategoryService
	FileService     fileclient.FileService
	NoteService     noteclient.NoteService
	SearchService   searchclient.SearchService
	Logger          logging.Logger
	ActionRecorder  actionlog.Recorder
}

type noteFile struct {
	ID string `json:"id"`
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, categoriesURL, jwt.JWTMiddleware(apperror.Middleware(h.GetCategories)))
	router.HandlerFunc(http.MethodPost, categoriesURL, jwt.JWTMiddleware(apperror.Middleware(h.CreateCategory)))
	router.HandlerFunc(http.MethodPatch, categoryURL, jwt.JWTMiddleware(apperror.Middleware(h.PartiallyUpdateCategory)))
	router.HandlerFunc(http.MethodDelete, categoryURL, jwt.JWTMiddleware(apperror.Middleware(h.DeleteCategory)))
}

func (h *Handler) GetCategories(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	userUuid, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}

	categories, err := h.CategoryService.GetUserCategories(r.Context(), userUuid)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(categories)

	return nil
}

func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	userUuid, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}

	var crCategory categoryclient.CreateCategoryDTO
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&crCategory); err != nil {
		return apperror.BadRequestError("can't decode")
	}
	crCategory.UserUuid = userUuid

	categoryUuid, err := h.CategoryService.CreateCategory(r.Context(), crCategory)
	if err != nil {
		return err
	}
	h.ActionRecorder.Record(r, userUuid, "category.created", "category", categoryUuid, map[string]any{
		"name":        crCategory.Name,
		"parent_uuid": crCategory.ParentUuid,
	})

	w.Header().Set("Location", fmt.Sprintf("%s/%s", categoriesURL, categoryUuid))
	w.WriteHeader(http.StatusCreated)

	return nil
}

func (h *Handler) PartiallyUpdateCategory(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	userUuid, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}

	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	categoryUuid := params.ByName("uuid")

	var categoryDTO categoryclient.UpdateCategoryDTO
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&categoryDTO); err != nil {
		return apperror.BadRequestError("can't decode")
	}

	err = h.CategoryService.UpdateCategory(r.Context(), categoryUuid, categoryDTO)
	if err != nil {
		return err
	}
	h.syncCategoryRename(r, userUuid, categoryUuid, categoryDTO)

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	userUuid, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}

	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	categoryDTO := categoryclient.DeleteCategoryDTO{
		Uuid:     params.ByName("uuid"),
		UserUuid: userUuid,
	}

	categories, err := h.fetchUserCategories(r, userUuid)
	if err != nil {
		return err
	}

	targetCategory, ok := findCategory(categories, categoryDTO.Uuid)
	if !ok {
		return apperror.APIError(http.StatusNotFound, "API-40400", "category not found", "category not found")
	}

	categoryUUIDs := collectCategoryUUIDs(targetCategory)
	notes, err := h.fetchNotesByCategories(r, userUuid, categoryUUIDs)
	if err != nil {
		return err
	}

	if err = h.deleteCategoryNotes(r, userUuid, notes); err != nil {
		return err
	}

	if err = h.CategoryService.DeleteCategory(r.Context(), categoryDTO); err != nil {
		return err
	}
	h.ActionRecorder.Record(r, userUuid, "category.deleted", "category", categoryDTO.Uuid, nil)

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

func (h *Handler) fetchUserCategories(r *http.Request, userUuid string) ([]categoryclient.Category, error) {
	categoriesData, err := h.CategoryService.GetUserCategories(r.Context(), userUuid)
	if err != nil {
		return nil, err
	}

	var categories []categoryclient.Category
	if err = json.Unmarshal(categoriesData, &categories); err != nil {
		return nil, fmt.Errorf("decode categories response: %w", err)
	}

	return categories, nil
}

func (h *Handler) fetchNotesByCategories(r *http.Request, userUuid string, categoryUUIDs []string) ([]noteclient.Note, error) {
	notesByUUID := make(map[string]noteclient.Note)

	for _, categoryUUID := range categoryUUIDs {
		notesData, err := h.NoteService.GetNotesByCategory(r.Context(), categoryUUID, userUuid)
		if err != nil {
			return nil, err
		}

		var categoryNotes []noteclient.Note
		if err = json.Unmarshal(notesData, &categoryNotes); err != nil {
			return nil, fmt.Errorf("decode notes response for category %s: %w", categoryUUID, err)
		}

		for _, note := range categoryNotes {
			if note.Uuid == "" {
				continue
			}
			notesByUUID[note.Uuid] = note
		}
	}

	notes := make([]noteclient.Note, 0, len(notesByUUID))
	for _, note := range notesByUUID {
		notes = append(notes, note)
	}

	return notes, nil
}

func (h *Handler) deleteCategoryNotes(r *http.Request, userUuid string, notes []noteclient.Note) error {
	for _, note := range notes {
		filesData, err := h.FileService.GetNoteFiles(r.Context(), note.Uuid, userUuid)
		if err != nil {
			return err
		}

		var files []noteFile
		if err = json.Unmarshal(filesData, &files); err != nil {
			return fmt.Errorf("decode files response for note %s: %w", note.Uuid, err)
		}

		for _, file := range files {
			if file.ID == "" {
				continue
			}

			if err = h.FileService.DeleteNoteFile(r.Context(), note.Uuid, file.ID, userUuid); err != nil {
				return err
			}
		}

		if err = h.NoteService.DeleteNote(r.Context(), note.Uuid, userUuid); err != nil {
			return err
		}

		if err = h.SearchService.DeleteNote(r.Context(), note.Uuid, userUuid); err != nil {
			return err
		}
	}

	return nil
}

func findCategory(categories []categoryclient.Category, targetUUID string) (categoryclient.Category, bool) {
	for _, category := range categories {
		if category.Uuid == targetUUID {
			return category, true
		}

		if child, ok := findCategory(category.Children, targetUUID); ok {
			return child, true
		}
	}

	return categoryclient.Category{}, false
}

func collectCategoryUUIDs(category categoryclient.Category) []string {
	uuids := []string{category.Uuid}
	for _, child := range category.Children {
		uuids = append(uuids, collectCategoryUUIDs(child)...)
	}

	return uuids
}

func (h *Handler) syncCategoryRename(r *http.Request, userUuid, categoryUUID string, dto categoryclient.UpdateCategoryDTO) {
	if dto.Name == "" {
		return
	}

	notesData, err := h.NoteService.GetNotesByCategory(r.Context(), categoryUUID, userUuid)
	if err != nil {
		h.Logger.Warn(
			"failed to fetch notes for category search reindex",
			"user_uuid", userUuid,
			"category_uuid", categoryUUID,
			"error", err,
		)
		return
	}

	var notes []noteclient.Note
	if err = json.Unmarshal(notesData, &notes); err != nil {
		h.Logger.Warn(
			"failed to decode notes for category search reindex",
			"user_uuid", userUuid,
			"category_uuid", categoryUUID,
			"error", err,
		)
		return
	}
	if len(notes) == 0 {
		return
	}

	categories, err := h.fetchUserCategories(r, userUuid)
	if err != nil {
		h.Logger.Warn(
			"failed to fetch categories for category search reindex",
			"user_uuid", userUuid,
			"category_uuid", categoryUUID,
			"error", err,
		)
		return
	}

	tags, err := h.fetchTagsForNotes(r, userUuid, notes)
	if err != nil {
		h.Logger.Warn(
			"failed to fetch tags for category search reindex",
			"user_uuid", userUuid,
			"category_uuid", categoryUUID,
			"error", err,
		)
		return
	}

	documents, err := searchsync.BuildIndexedNotes(notes, categories, tags)
	if err != nil {
		h.Logger.Warn(
			"failed to build indexed notes for category search reindex",
			"user_uuid", userUuid,
			"category_uuid", categoryUUID,
			"error", err,
		)
		return
	}

	if err = h.SearchService.UpsertNotes(r.Context(), documents); err != nil {
		h.Logger.Warn(
			"failed to bulk sync category notes to search index",
			"user_uuid", userUuid,
			"category_uuid", categoryUUID,
			"error", err,
		)
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
