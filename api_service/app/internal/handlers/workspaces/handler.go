package workspaces

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"myproject/internal/apperror"
	categoryclient "myproject/internal/client/category"
	fileclient "myproject/internal/client/file"
	noteclient "myproject/internal/client/note"
	userclient "myproject/internal/client/user"
	"myproject/internal/requestctx"
	"myproject/pkg/logging"
	"myproject/pkg/middleware/jwt"
	workspacemw "myproject/pkg/middleware/workspace"
)

const (
	workspaceURL        = "/api/workspaces/{uuid}"
	workspaceMembersURL = "/api/workspaces/{uuid}/members"
	workspaceMemberURL  = "/api/workspaces/{uuid}/members/{member_uuid}"
	workspaceInvitesURL = "/api/workspaces/{uuid}/invites"
)

type Handler struct {
	Logger          logging.Logger
	UserService     userclient.UserService
	CategoryService categoryclient.CategoryService
	NoteService     noteclient.NoteService
	FileService     fileclient.FileService
}

type Overview struct {
	Workspace      userclient.Workspace `json:"workspace"`
	Role           string               `json:"role"`
	Status         string               `json:"status"`
	CanInvite      bool                 `json:"can_invite"`
	MembersCount   int                  `json:"members_count"`
	Stats          OverviewStats        `json:"stats"`
	UpcomingEvents []noteclient.Note    `json:"upcoming_events"`
}

type OverviewStats struct {
	CategoriesCount int64  `json:"categories_count"`
	NotesCount      int64  `json:"notes_count"`
	TagsCount       int64  `json:"tags_count"`
	FilesCount      int64  `json:"files_count"`
	LastActivityAt  *int64 `json:"last_activity_at"`
}

type CreateInviteRequest struct {
	Email     string `json:"email"`
	Role      string `json:"role,omitempty"`
	ExpiresAt int64  `json:"expires_at,omitempty"`
}

type UpdateMemberRequest struct {
	Role   string `json:"role,omitempty"`
	Status string `json:"status,omitempty"`
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc(http.MethodGet+" "+workspaceURL, jwt.JWTMiddleware(workspacemw.Middleware(h.UserService, apperror.Middleware(h.GetOverview))))
	mux.HandleFunc(http.MethodGet+" "+workspaceMembersURL, jwt.JWTMiddleware(workspacemw.Middleware(h.UserService, apperror.Middleware(h.GetMembers))))
	mux.HandleFunc(http.MethodPatch+" "+workspaceMemberURL, jwt.JWTMiddleware(workspacemw.Middleware(h.UserService, apperror.Middleware(h.UpdateMember))))
	mux.HandleFunc(http.MethodGet+" "+workspaceInvitesURL, jwt.JWTMiddleware(workspacemw.Middleware(h.UserService, apperror.Middleware(h.GetInvites))))
	mux.HandleFunc(http.MethodPost+" "+workspaceInvitesURL, jwt.JWTMiddleware(workspacemw.Middleware(h.UserService, apperror.Middleware(h.CreateInvite))))
}

func (h *Handler) GetOverview(w http.ResponseWriter, r *http.Request) error {
	userUUID, workspaceID, err := h.validateWorkspaceRequest(r)
	if err != nil {
		return err
	}

	workspace, err := h.UserService.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		return err
	}

	role, err := requestctx.WorkspaceRole(r)
	if err != nil {
		return err
	}

	members, err := h.UserService.GetWorkspaceMembers(r.Context(), workspaceID)
	if err != nil {
		return err
	}

	categoryStats, err := h.CategoryService.GetStats(r.Context(), workspaceID)
	if err != nil {
		return err
	}

	noteStats, err := h.NoteService.GetStats(r.Context(), userUUID, workspaceID)
	if err != nil {
		return err
	}

	fileStats, err := h.FileService.GetStats(r.Context(), userUUID, workspaceID)
	if err != nil {
		return err
	}

	upcomingEvents, err := h.fetchUpcomingEvents(r, userUUID, workspaceID)
	if err != nil {
		return err
	}

	overview := Overview{
		Workspace:    workspace,
		Role:         role,
		Status:       "active",
		CanInvite:    role == "owner" || role == "editor",
		MembersCount: len(members),
		Stats: OverviewStats{
			CategoriesCount: categoryStats.CategoriesCount,
			NotesCount:      noteStats.NotesCount,
			TagsCount:       noteStats.TagsCount,
			FilesCount:      fileStats.FilesCount,
			LastActivityAt:  nil,
		},
		UpcomingEvents: upcomingEvents,
	}

	data, err := json.Marshal(overview)
	if err != nil {
		h.Logger.Error("failed to marshal workspace overview", "workspace_id", workspaceID, "error", err)
		return fmt.Errorf("marshal workspace overview: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) GetMembers(w http.ResponseWriter, r *http.Request) error {
	_, workspaceID, err := h.validateWorkspaceRequest(r)
	if err != nil {
		return err
	}

	members, err := h.UserService.GetWorkspaceMembers(r.Context(), workspaceID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(members)
	if err != nil {
		h.Logger.Error("failed to marshal workspace members", "workspace_id", workspaceID, "error", err)
		return fmt.Errorf("marshal workspace members: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) GetInvites(w http.ResponseWriter, r *http.Request) error {
	userUUID, workspaceID, err := h.validateWorkspaceRequest(r)
	if err != nil {
		return err
	}

	invites, err := h.UserService.GetWorkspaceInvites(r.Context(), workspaceID, userUUID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(invites)
	if err != nil {
		h.Logger.Error("failed to marshal workspace invites", "workspace_id", workspaceID, "user_uuid", userUUID, "error", err)
		return fmt.Errorf("marshal workspace invites: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) UpdateMember(w http.ResponseWriter, r *http.Request) error {
	userUUID, workspaceID, err := h.validateWorkspaceRequest(r)
	if err != nil {
		return err
	}

	memberUserUUID := strings.TrimSpace(r.PathValue("member_uuid"))
	if memberUserUUID == "" {
		return apperror.BadRequestError("empty member user uuid")
	}

	var payload UpdateMemberRequest
	defer r.Body.Close()
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return apperror.BadRequestError("can't decode")
	}

	member, err := h.UserService.UpdateWorkspaceMember(r.Context(), workspaceID, memberUserUUID, userclient.UpdateWorkspaceMemberDTO{
		ActorUserUUID: userUUID,
		Role:          strings.TrimSpace(strings.ToLower(payload.Role)),
		Status:        strings.TrimSpace(strings.ToLower(payload.Status)),
	})
	if err != nil {
		return err
	}

	data, err := json.Marshal(member)
	if err != nil {
		h.Logger.Error("failed to marshal updated workspace member", "workspace_id", workspaceID, "member_user_uuid", memberUserUUID, "error", err)
		return fmt.Errorf("marshal updated workspace member: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) CreateInvite(w http.ResponseWriter, r *http.Request) error {
	userUUID, workspaceID, err := h.validateWorkspaceRequest(r)
	if err != nil {
		return err
	}

	var payload CreateInviteRequest
	defer r.Body.Close()
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return apperror.BadRequestError("can't decode")
	}

	invite, err := h.UserService.CreateWorkspaceInvite(r.Context(), workspaceID, userclient.CreateWorkspaceInviteDTO{
		InvitedByUserUUID: userUUID,
		Email:             strings.TrimSpace(payload.Email),
		Role:              strings.TrimSpace(strings.ToLower(payload.Role)),
		ExpiresAt:         payload.ExpiresAt,
	})
	if err != nil {
		return err
	}

	data, err := json.Marshal(invite)
	if err != nil {
		h.Logger.Error("failed to marshal workspace invite", "workspace_id", workspaceID, "user_uuid", userUUID, "error", err)
		return fmt.Errorf("marshal workspace invite: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) fetchUpcomingEvents(r *http.Request, userUUID, workspaceID string) ([]noteclient.Note, error) {
	now := time.Now().Unix()
	const upcomingWindow = int64(30 * 24 * 60 * 60)

	notesData, err := h.NoteService.GetCalendarNotes(r.Context(), now, now+upcomingWindow, userUUID, workspaceID)
	if err != nil {
		return nil, err
	}

	var notes []noteclient.Note
	if err = json.Unmarshal(notesData, &notes); err != nil {
		return nil, fmt.Errorf("decode workspace upcoming events: %w", err)
	}

	if len(notes) > 5 {
		notes = notes[:5]
	}

	return notes, nil
}

func (h *Handler) validateWorkspaceRequest(r *http.Request) (string, string, error) {
	userUUID, err := requestctx.UserUUID(r)
	if err != nil {
		return "", "", err
	}

	contextWorkspaceID, err := requestctx.WorkspaceID(r)
	if err != nil {
		return "", "", err
	}

	pathWorkspaceID := strings.TrimSpace(r.PathValue("uuid"))
	if pathWorkspaceID == "" {
		return "", "", apperror.BadRequestError("empty workspace uuid")
	}
	if pathWorkspaceID != contextWorkspaceID {
		return "", "", apperror.BadRequestError("workspace context mismatch")
	}

	return userUUID, contextWorkspaceID, nil
}
