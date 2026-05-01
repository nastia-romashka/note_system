package tags

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
	tagsURL = "/api/tags"
	tagURL  = "/api/tags/:uuid"
)

type Handler struct {
	Logger      logging.Logger
	NoteService noteclient.NoteService
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, tagsURL, jwt.JWTMiddleware(apperror.Middleware(h.GetTags)))
	router.HandlerFunc(http.MethodPost, tagsURL, jwt.JWTMiddleware(apperror.Middleware(h.CreateTag)))
	router.HandlerFunc(http.MethodDelete, tagURL, jwt.JWTMiddleware(apperror.Middleware(h.DeleteTag)))
}

func (h *Handler) GetTags(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	tagUUIDs := r.URL.Query()["id"]
	tags, err := h.NoteService.GetTags(r.Context(), tagUUIDs)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(tags)
	return nil
}

func (h *Handler) CreateTag(w http.ResponseWriter, r *http.Request) error {
	var dto noteclient.CreateTagDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		return apperror.BadRequestError("can't decode")
	}
	defer r.Body.Close()

	tagUuid, err := h.NoteService.CreateTag(r.Context(), dto)
	if err != nil {
		return err
	}

	w.Header().Set("Location", fmt.Sprintf("%s/%s", tagsURL, tagUuid))
	w.WriteHeader(http.StatusCreated)
	return nil
}

func (h *Handler) DeleteTag(w http.ResponseWriter, r *http.Request) error {
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	tagUuid := params.ByName("uuid")
	if tagUuid == "" {
		return apperror.BadRequestError("empty tag uuid")
	}

	if err := h.NoteService.DeleteTag(r.Context(), tagUuid); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
