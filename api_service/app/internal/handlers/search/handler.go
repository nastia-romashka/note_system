package search

import (
	"encoding/json"
	"fmt"
	"net/http"

	"myproject/internal/apperror"
	categoryclient "myproject/internal/client/category"
	noteclient "myproject/internal/client/note"
	searchclient "myproject/internal/client/search"
	"myproject/internal/searchsync"
	"myproject/pkg/logging"
	"myproject/pkg/middleware/jwt"

	"github.com/julienschmidt/httprouter"
)

const (
	searchNotesURL   = "/api/search/notes"
	searchReindexURL = "/api/search/reindex"
)

type Handler struct {
	Logger          logging.Logger
	SearchService   searchclient.SearchService
	CategoryService categoryclient.CategoryService
	NoteService     noteclient.NoteService
}

type reindexResponse struct {
	IndexedNotes      int      `json:"indexed_notes"`
	ScannedCategories int      `json:"scanned_categories"`
	CategoryUUIDs     []string `json:"category_uuids"`
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, searchNotesURL, jwt.JWTMiddleware(apperror.Middleware(h.SearchNotes)))
	router.HandlerFunc(http.MethodPost, searchReindexURL, jwt.JWTMiddleware(apperror.Middleware(h.ReindexNotes)))
}

func (h *Handler) SearchNotes(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}

	notes, err := h.SearchService.SearchNotes(
		r.Context(),
		r.URL.Query().Get("q"),
		userUUID,
		r.URL.Query().Get("category_uuid"),
		r.URL.Query()["tag_uuid"],
	)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(notes)
	return nil
}

func (h *Handler) ReindexNotes(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}

	categories, categoryUUIDs, err := h.fetchCategories(r, userUUID)
	if err != nil {
		return err
	}

	notes, err := h.fetchNotesByCategories(r, userUUID, categoryUUIDs)
	if err != nil {
		return err
	}

	tags, err := h.fetchTagsForNotes(r, userUUID, notes)
	if err != nil {
		return err
	}

	documents, err := searchsync.BuildIndexedNotes(notes, categories, tags)
	if err != nil {
		h.Logger.Error("failed to build indexed notes for full reindex", "user_uuid", userUUID, "error", err)
		return fmt.Errorf("build indexed notes for reindex: %w", err)
	}

	if err = h.SearchService.DeleteNotesByUser(r.Context(), userUUID); err != nil {
		h.Logger.Error("failed to clear search index before reindex", "user_uuid", userUUID, "error", err)
		return fmt.Errorf("clear search index before reindex: %w", err)
	}

	if err = h.SearchService.UpsertNotes(r.Context(), documents); err != nil {
		h.Logger.Error("failed to import reindexed notes into search", "user_uuid", userUUID, "error", err)
		return fmt.Errorf("import reindexed notes: %w", err)
	}

	response := reindexResponse{
		IndexedNotes:      len(documents),
		ScannedCategories: len(categoryUUIDs),
		CategoryUUIDs:     categoryUUIDs,
	}

	data, err := json.Marshal(response)
	if err != nil {
		h.Logger.Error("failed to marshal reindex response", "user_uuid", userUUID, "error", err)
		return fmt.Errorf("marshal reindex response: %w", err)
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

func (h *Handler) fetchCategories(r *http.Request, userUUID string) ([]categoryclient.Category, []string, error) {
	categoriesData, err := h.CategoryService.GetUserCategories(r.Context(), userUUID)
	if err != nil {
		return nil, nil, err
	}

	var categories []categoryclient.Category
	if err = json.Unmarshal(categoriesData, &categories); err != nil {
		return nil, nil, fmt.Errorf("decode categories response: %w", err)
	}

	return categories, flattenCategoryUUIDs(categories), nil
}

func (h *Handler) fetchNotesByCategories(r *http.Request, userUUID string, categoryUUIDs []string) ([]noteclient.Note, error) {
	notesByID := make(map[string]noteclient.Note)

	for _, categoryUUID := range categoryUUIDs {
		notesData, err := h.NoteService.GetNotesByCategory(r.Context(), categoryUUID, userUUID)
		if err != nil {
			return nil, fmt.Errorf("fetch notes for category %s: %w", categoryUUID, err)
		}

		var categoryNotes []noteclient.Note
		if err = json.Unmarshal(notesData, &categoryNotes); err != nil {
			return nil, fmt.Errorf("decode notes response for category %s: %w", categoryUUID, err)
		}

		for _, note := range categoryNotes {
			notesByID[note.Uuid] = note
		}
	}

	result := make([]noteclient.Note, 0, len(notesByID))
	for _, note := range notesByID {
		result = append(result, note)
	}

	return result, nil
}

func (h *Handler) fetchTagsForNotes(r *http.Request, userUUID string, notes []noteclient.Note) ([]noteclient.Tag, error) {
	tagUUIDs := searchsync.CollectTagUUIDs(notes)
	if len(tagUUIDs) == 0 {
		return nil, nil
	}

	tagsData, err := h.NoteService.GetTags(r.Context(), tagUUIDs, userUUID)
	if err != nil {
		return nil, err
	}

	var tags []noteclient.Tag
	if err = json.Unmarshal(tagsData, &tags); err != nil {
		return nil, fmt.Errorf("decode tags response: %w", err)
	}

	return tags, nil
}

func flattenCategoryUUIDs(categories []categoryclient.Category) []string {
	result := make([]string, 0)
	for _, category := range categories {
		result = append(result, category.Uuid)
		result = append(result, flattenCategoryUUIDs(category.Children)...)
	}

	return result
}
