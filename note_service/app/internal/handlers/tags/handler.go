package tags

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"note_service/internal/apperror"
	"note_service/pkg/logging"
)

const (
	tagsURL = "/api/tags"
	tagURL  = "/api/tags/{uuid}"
)

type TagService interface {
	Get(tagUUIDs []string, userUUID string) ([]Tag, error)
	Create(dto CreateTagDTO) (string, error)
	Delete(tagUUID, userUUID string) error
}

type Handler struct {
	Logger     logging.Logger
	TagService TagService
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET "+tagsURL, apperror.Middleware(h.GetTags))
	mux.HandleFunc("POST "+tagsURL, apperror.Middleware(h.CreateTag))
	mux.HandleFunc("DELETE "+tagURL, apperror.Middleware(h.DeleteTag))
}

func (h *Handler) GetTags(w http.ResponseWriter, r *http.Request) error {
	tagUUIDs := parseTagIDs(r.URL.Query()["id"])
	userUUID, err := userUUIDFromQuery(r)
	if err != nil {
		return err
	}

	tags, err := h.TagService.Get(tagUUIDs, userUUID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(tags)
	if err != nil {
		h.Logger.Error("failed to marshal tags response", "error", err)
		return fmt.Errorf("marshal tags: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) CreateTag(w http.ResponseWriter, r *http.Request) error {
	var dto CreateTagDTO
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.Logger.Warn("failed to decode create tag payload", "error", err)
		return apperror.BadRequestError("can't decode")
	}

	tagUUID, err := h.TagService.Create(dto)
	if err != nil {
		return err
	}

	w.Header().Set("Location", fmt.Sprintf("%s/%s", tagsURL, tagUUID))
	w.WriteHeader(http.StatusCreated)
	return nil
}

func (h *Handler) DeleteTag(w http.ResponseWriter, r *http.Request) error {
	tagUUID := r.PathValue("uuid")
	if tagUUID == "" {
		return apperror.BadRequestError("empty tag uuid")
	}
	userUUID, err := userUUIDFromQuery(r)
	if err != nil {
		return err
	}

	if err := h.TagService.Delete(tagUUID, userUUID); err != nil {
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

func parseTagIDs(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	result := make([]string, 0, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			result = append(result, part)
		}
	}

	return result
}
