package passthrough

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"myproject/internal/apperror"
	userclient "myproject/internal/client/user"
	"myproject/internal/proxy"
	"myproject/internal/requestctx"
	"myproject/pkg/logging"
	"myproject/pkg/middleware/jwt"
	workspacemw "myproject/pkg/middleware/workspace"
)

const (
	categoriesURL             = "/api/categories"
	graphURL                  = "/api/graph"
	graphLinksURL             = "/api/graph/links"
	calendarURL               = "/api/calendar"
	noteURL                   = "/api/notes/{uuid}"
	tagsURL                   = "/api/tags"
	searchNotesURL            = "/api/search/notes"
	meURL                     = "/api/me"
	meActionsURL              = "/api/me/actions"
	meWorkspacesURL           = "/api/me/workspaces"
	meWorkspaceInvitesURL     = "/api/me/workspace-invites"
	workspacesURL             = "/api/workspaces"
	workspaceInviteAcceptURL  = "/api/workspaces/invites/{uuid}/accept"
	workspaceInviteDeclineURL = "/api/workspaces/invites/{uuid}/decline"
)

type Handler struct {
	Logger           logging.Logger
	CategoryService  proxy.Service
	NoteService      proxy.Service
	UserService      proxy.Service
	SearchService    proxy.Service
	WorkspaceService userclient.UserService
}

func NewHandler(
	logger logging.Logger,
	categoryBaseURL string,
	noteBaseURL string,
	userBaseURL string,
	searchBaseURL string,
	workspaceService userclient.UserService,
) Handler {
	return Handler{
		Logger:           logger,
		CategoryService:  proxy.NewService(categoryBaseURL, 20*time.Second),
		NoteService:      proxy.NewService(noteBaseURL, 20*time.Second),
		UserService:      proxy.NewService(userBaseURL, 20*time.Second),
		SearchService:    proxy.NewService(searchBaseURL, 20*time.Second),
		WorkspaceService: workspaceService,
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc(http.MethodGet+" "+categoriesURL, jwt.JWTMiddleware(workspacemw.Middleware(h.WorkspaceService, apperror.Middleware(h.GetCategories))))
	mux.HandleFunc(http.MethodGet+" "+graphURL, jwt.JWTMiddleware(workspacemw.Middleware(h.WorkspaceService, apperror.Middleware(h.GetGraph))))
	mux.HandleFunc(http.MethodPost+" "+graphLinksURL, jwt.JWTMiddleware(workspacemw.Middleware(h.WorkspaceService, apperror.Middleware(h.CreateGraphLink))))
	mux.HandleFunc(http.MethodDelete+" "+graphLinksURL, jwt.JWTMiddleware(workspacemw.Middleware(h.WorkspaceService, apperror.Middleware(h.DeleteGraphLink))))
	mux.HandleFunc(http.MethodGet+" "+calendarURL, jwt.JWTMiddleware(workspacemw.Middleware(h.WorkspaceService, apperror.Middleware(h.GetCalendarNotes))))
	mux.HandleFunc(http.MethodGet+" "+noteURL, jwt.JWTMiddleware(workspacemw.Middleware(h.WorkspaceService, apperror.Middleware(h.GetNote))))
	mux.HandleFunc(http.MethodGet+" "+tagsURL, jwt.JWTMiddleware(workspacemw.Middleware(h.WorkspaceService, apperror.Middleware(h.GetTags))))
	mux.HandleFunc(http.MethodGet+" "+searchNotesURL, jwt.JWTMiddleware(workspacemw.Middleware(h.WorkspaceService, apperror.Middleware(h.SearchNotes))))
	mux.HandleFunc(http.MethodGet+" "+meURL, jwt.JWTMiddleware(apperror.Middleware(h.GetProfile)))
	mux.HandleFunc(http.MethodGet+" "+meActionsURL, jwt.JWTMiddleware(apperror.Middleware(h.GetActions)))
	mux.HandleFunc(http.MethodGet+" "+meWorkspacesURL, jwt.JWTMiddleware(apperror.Middleware(h.GetWorkspaces)))
	mux.HandleFunc(http.MethodGet+" "+meWorkspaceInvitesURL, jwt.JWTMiddleware(apperror.Middleware(h.GetWorkspaceInvites)))
	mux.HandleFunc(http.MethodPost+" "+workspacesURL, jwt.JWTMiddleware(apperror.Middleware(h.CreateWorkspace)))
	mux.HandleFunc(http.MethodPost+" "+workspaceInviteAcceptURL, jwt.JWTMiddleware(apperror.Middleware(h.AcceptWorkspaceInvite)))
	mux.HandleFunc(http.MethodPost+" "+workspaceInviteDeclineURL, jwt.JWTMiddleware(apperror.Middleware(h.DeclineWorkspaceInvite)))
}

func (h *Handler) GetCategories(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := requestctx.UserUUID(r)
	if err != nil {
		return err
	}
	workspaceID, err := requestctx.WorkspaceID(r)
	if err != nil {
		return err
	}

	return h.CategoryService.Forward(
		w,
		r,
		http.MethodGet,
		"categories",
		withWorkspaceContext(r.URL.Query(), userUUID, workspaceID),
		nil,
		nil,
	)
}

func (h *Handler) GetGraph(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := requestctx.UserUUID(r)
	if err != nil {
		return err
	}
	workspaceID, err := requestctx.WorkspaceID(r)
	if err != nil {
		return err
	}

	return h.CategoryService.Forward(
		w,
		r,
		http.MethodGet,
		"graph",
		withWorkspaceContext(r.URL.Query(), userUUID, workspaceID),
		nil,
		nil,
	)
}

func (h *Handler) CreateGraphLink(w http.ResponseWriter, r *http.Request) error {
	return h.forwardGraphLink(w, r, http.MethodPost)
}

func (h *Handler) DeleteGraphLink(w http.ResponseWriter, r *http.Request) error {
	return h.forwardGraphLink(w, r, http.MethodDelete)
}

func (h *Handler) GetCalendarNotes(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := requestctx.UserUUID(r)
	if err != nil {
		return err
	}
	workspaceID, err := requestctx.WorkspaceID(r)
	if err != nil {
		return err
	}

	query := withWorkspaceContext(r.URL.Query(), userUUID, workspaceID)
	if query.Get("from") == "" {
		return apperror.BadRequestError("from is required")
	}
	if query.Get("to") == "" {
		return apperror.BadRequestError("to is required")
	}

	return h.NoteService.Forward(w, r, http.MethodGet, "calendar", query, nil, nil)
}

func (h *Handler) GetNote(w http.ResponseWriter, r *http.Request) error {
	noteUUID := r.PathValue("uuid")
	if noteUUID == "" {
		return apperror.BadRequestError("empty note uuid")
	}

	userUUID, err := requestctx.UserUUID(r)
	if err != nil {
		return err
	}
	workspaceID, err := requestctx.WorkspaceID(r)
	if err != nil {
		return err
	}

	return h.NoteService.Forward(
		w,
		r,
		http.MethodGet,
		fmt.Sprintf("notes/%s", noteUUID),
		withWorkspaceContext(r.URL.Query(), userUUID, workspaceID),
		nil,
		nil,
	)
}

func (h *Handler) GetTags(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := requestctx.UserUUID(r)
	if err != nil {
		return err
	}
	workspaceID, err := requestctx.WorkspaceID(r)
	if err != nil {
		return err
	}

	return h.NoteService.Forward(
		w,
		r,
		http.MethodGet,
		"tags",
		withWorkspaceContext(r.URL.Query(), userUUID, workspaceID),
		nil,
		nil,
	)
}

func (h *Handler) SearchNotes(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := requestctx.UserUUID(r)
	if err != nil {
		return err
	}
	workspaceID, err := requestctx.WorkspaceID(r)
	if err != nil {
		return err
	}

	return h.SearchService.Forward(
		w,
		r,
		http.MethodGet,
		"search/notes",
		withWorkspaceContext(r.URL.Query(), userUUID, workspaceID),
		nil,
		nil,
	)
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := requestctx.UserUUID(r)
	if err != nil {
		return err
	}

	return h.UserService.Forward(
		w,
		r,
		http.MethodGet,
		fmt.Sprintf("users/%s/profile", userUUID),
		nil,
		nil,
		nil,
	)
}

func (h *Handler) GetActions(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := requestctx.UserUUID(r)
	if err != nil {
		return err
	}

	limit, err := positiveIntQuery(r, "limit", 50)
	if err != nil {
		return err
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := positiveIntQuery(r, "offset", 0)
	if err != nil {
		return err
	}

	query := url.Values{}
	query.Set("limit", strconv.Itoa(limit))
	query.Set("offset", strconv.Itoa(offset))

	return h.UserService.Forward(
		w,
		r,
		http.MethodGet,
		fmt.Sprintf("users/%s/actions", userUUID),
		query,
		nil,
		nil,
	)
}

func (h *Handler) GetWorkspaces(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := requestctx.UserUUID(r)
	if err != nil {
		return err
	}

	return h.UserService.Forward(
		w,
		r,
		http.MethodGet,
		fmt.Sprintf("users/%s/workspaces", userUUID),
		nil,
		nil,
		nil,
	)
}

func (h *Handler) GetWorkspaceInvites(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := requestctx.UserUUID(r)
	if err != nil {
		return err
	}

	return h.UserService.Forward(
		w,
		r,
		http.MethodGet,
		fmt.Sprintf("users/%s/workspace-invites", userUUID),
		nil,
		nil,
		nil,
	)
}

func (h *Handler) CreateWorkspace(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := requestctx.UserUUID(r)
	if err != nil {
		return err
	}

	var payload map[string]any
	defer r.Body.Close()
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return apperror.BadRequestError("can't decode")
	}

	name := strings.TrimSpace(asString(payload["name"]))
	if name == "" {
		return apperror.BadRequestError("name is required")
	}

	visibility := strings.TrimSpace(asString(payload["visibility"]))
	if visibility == "" {
		visibility = "invite_only"
	}

	body, err := json.Marshal(map[string]any{
		"name":            name,
		"owner_user_uuid": userUUID,
		"visibility":      visibility,
	})
	if err != nil {
		return fmt.Errorf("marshal workspace payload: %w", err)
	}

	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")
	headers.Set("Accept", "application/json")

	return h.UserService.Forward(
		w,
		r,
		http.MethodPost,
		"workspaces",
		nil,
		bytes.NewReader(body),
		headers,
	)
}

func (h *Handler) AcceptWorkspaceInvite(w http.ResponseWriter, r *http.Request) error {
	return h.resolveWorkspaceInvite(w, r, "accept")
}

func (h *Handler) DeclineWorkspaceInvite(w http.ResponseWriter, r *http.Request) error {
	return h.resolveWorkspaceInvite(w, r, "decline")
}

func (h *Handler) forwardGraphLink(w http.ResponseWriter, r *http.Request, method string) error {
	userUUID, err := requestctx.UserUUID(r)
	if err != nil {
		return err
	}
	workspaceID, err := requestctx.WorkspaceID(r)
	if err != nil {
		return err
	}

	var payload map[string]any
	defer r.Body.Close()
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return apperror.BadRequestError("can't decode")
	}

	sourceID := strings.TrimSpace(asString(payload["source_id"]))
	targetID := strings.TrimSpace(asString(payload["target_id"]))
	if sourceID == "" || targetID == "" {
		return apperror.BadRequestError("source_id and target_id are required")
	}

	payload["source_id"] = sourceID
	payload["target_id"] = targetID
	payload["user_uuid"] = userUUID
	payload["workspace_id"] = workspaceID

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal graph link payload: %w", err)
	}

	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")
	headers.Set("Accept", "application/json")

	return h.CategoryService.Forward(
		w,
		r,
		method,
		"graph/links",
		nil,
		bytes.NewReader(data),
		headers,
	)
}

func (h *Handler) resolveWorkspaceInvite(w http.ResponseWriter, r *http.Request, action string) error {
	userUUID, err := requestctx.UserUUID(r)
	if err != nil {
		return err
	}

	inviteUUID := strings.TrimSpace(r.PathValue("uuid"))
	if inviteUUID == "" {
		return apperror.BadRequestError("empty invite uuid")
	}

	body, err := json.Marshal(map[string]any{
		"user_uuid": userUUID,
	})
	if err != nil {
		return fmt.Errorf("marshal workspace invite resolve payload: %w", err)
	}

	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")
	headers.Set("Accept", "application/json")

	return h.UserService.Forward(
		w,
		r,
		http.MethodPost,
		fmt.Sprintf("workspaces/invites/%s/%s", inviteUUID, action),
		nil,
		bytes.NewReader(body),
		headers,
	)
}

func withWorkspaceContext(query url.Values, userUUID, workspaceID string) url.Values {
	cloned := cloneQuery(query)
	cloned.Set("user_uuid", userUUID)
	if strings.TrimSpace(workspaceID) != "" {
		cloned.Set("workspace_id", workspaceID)
	}
	return cloned
}

func cloneQuery(query url.Values) url.Values {
	cloned := make(url.Values, len(query))
	for key, values := range query {
		cloned[key] = append([]string(nil), values...)
	}

	return cloned
}

func positiveIntQuery(r *http.Request, name string, defaultValue int) (int, error) {
	rawValue := r.URL.Query().Get(name)
	if rawValue == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(rawValue)
	if err != nil || value < 0 {
		return 0, apperror.BadRequestError(fmt.Sprintf("invalid %s", name))
	}

	return value, nil
}

func asString(value any) string {
	text, _ := value.(string)
	return text
}
