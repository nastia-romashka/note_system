package graph

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"myproject/internal/apperror"
	categoryclient "myproject/internal/client/category"
	"myproject/pkg/logging"
	"myproject/pkg/middleware/jwt"
)

const graphURL = "/api/graph"

type Handler struct {
	Logger          logging.Logger
	CategoryService categoryclient.CategoryService
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, graphURL, jwt.JWTMiddleware(apperror.Middleware(h.GetGraph)))
}

func (h *Handler) GetGraph(w http.ResponseWriter, r *http.Request) error {
	userUuid, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}

	graphData, err := h.CategoryService.GetGraph(r.Context(), userUuid)
	if err != nil {
		return err
	}

	data, err := json.Marshal(graphData)
	if err != nil {
		h.Logger.Error("failed to marshal graph response", "user_uuid", userUuid, "error", err)
		return fmt.Errorf("marshal graph: %w", err)
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
