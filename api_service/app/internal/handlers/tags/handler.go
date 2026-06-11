package tags

import (
	"encoding/json"
	"fmt"
	"net/http"

	"myproject/internal/apperror"
	noteclient "myproject/internal/client/note"
	userclient "myproject/internal/client/user"
	"myproject/internal/handlers/actionlog"
	"myproject/internal/requestctx"
	"myproject/pkg/logging"
	"myproject/pkg/middleware/jwt"
	workspacemw "myproject/pkg/middleware/workspace"
)

const (
	tagsURL = "/api/tags"
	tagURL  = "/api/tags/{uuid}"
)

type Handler struct {
	Logger           logging.Logger
	NoteService      noteclient.NoteService
	WorkspaceService userclient.UserService
	ActionRecorder   actionlog.Recorder
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc(http.MethodPost+" "+tagsURL, jwt.JWTMiddleware(workspacemw.Middleware(h.WorkspaceService, apperror.Middleware(h.CreateTag))))
	mux.HandleFunc(http.MethodDelete+" "+tagURL, jwt.JWTMiddleware(workspacemw.Middleware(h.WorkspaceService, apperror.Middleware(h.DeleteTag))))
}

func (h *Handler) CreateTag(w http.ResponseWriter, r *http.Request) error {
	var dto noteclient.CreateTagDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		return apperror.BadRequestError("can't decode")
	}
	defer r.Body.Close()
	userUuid, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}
	workspaceID, err := h.workspaceIDFromContext(r)
	if err != nil {
		return err
	}
	dto.UserUuid = userUuid
	dto.WorkspaceID = workspaceID

	tagUuid, err := h.NoteService.CreateTag(r.Context(), dto)
	if err != nil {
		return err
	}
	h.ActionRecorder.Record(r, userUuid, "tag.created", "tag", tagUuid, map[string]any{
		"name": dto.Name,
	})

	w.Header().Set("Location", fmt.Sprintf("%s/%s", tagsURL, tagUuid))
	w.WriteHeader(http.StatusCreated)
	return nil
}

func (h *Handler) DeleteTag(w http.ResponseWriter, r *http.Request) error {
	tagUuid := r.PathValue("uuid")
	if tagUuid == "" {
		return apperror.BadRequestError("empty tag uuid")
	}
	userUuid, err := h.userUUIDFromContext(r)
	if err != nil {
		return err
	}
	workspaceID, err := h.workspaceIDFromContext(r)
	if err != nil {
		return err
	}

	if err := h.NoteService.DeleteTag(r.Context(), tagUuid, userUuid, workspaceID); err != nil {
		return err
	}
	h.ActionRecorder.Record(r, userUuid, "tag.deleted", "tag", tagUuid, nil)

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

func (h *Handler) workspaceIDFromContext(r *http.Request) (string, error) {
	return requestctx.WorkspaceID(r)
}
