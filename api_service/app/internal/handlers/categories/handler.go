package categories

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"myproject/internal/apperror"
	categoryclient "myproject/internal/client/category"
	"myproject/internal/handlers/actionlog"
	"myproject/pkg/logging"
	"myproject/pkg/middleware/jwt"
)

const (
	categoriesURL = "/api/categories"
	categoryURL   = "/api/categories/:uuid"
)

type Handler struct {
	CategoryService categoryclient.CategoryService
	Logger          logging.Logger
	ActionRecorder  actionlog.Recorder
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
	categoryDTO.UserUuid = userUuid

	err = h.CategoryService.UpdateCategory(r.Context(), categoryUuid, categoryDTO)
	if err != nil {
		return err
	}

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

	err = h.CategoryService.DeleteCategory(r.Context(), categoryDTO)
	if err != nil {
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
