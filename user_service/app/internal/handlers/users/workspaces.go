package users

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"user_service/internal/apperror"
	"user_service/internal/events"
)

func (h *Handler) GetUserWorkspaces(w http.ResponseWriter, r *http.Request) error {
	userUUID := userUUIDFromParams(r)
	if userUUID == "" {
		return apperror.BadRequestError("empty user uuid")
	}

	workspaces, err := h.UserService.GetWorkspaces(userUUID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(workspaces)
	if err != nil {
		h.Logger.Error("failed to marshal workspaces response", "user_uuid", userUUID, "error", err)
		return fmt.Errorf("marshal workspaces: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) GetUserWorkspaceInvites(w http.ResponseWriter, r *http.Request) error {
	userUUID := userUUIDFromParams(r)
	if userUUID == "" {
		return apperror.BadRequestError("empty user uuid")
	}

	invites, err := h.UserService.GetWorkspaceInvites(userUUID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(invites)
	if err != nil {
		h.Logger.Error("failed to marshal workspace invites response", "user_uuid", userUUID, "error", err)
		return fmt.Errorf("marshal workspace invites: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) GetWorkspace(w http.ResponseWriter, r *http.Request) error {
	workspaceUUID := workspaceUUIDFromParams(r)
	if workspaceUUID == "" {
		return apperror.BadRequestError("empty workspace uuid")
	}

	workspace, err := h.UserService.GetWorkspace(workspaceUUID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(workspace)
	if err != nil {
		h.Logger.Error("failed to marshal workspace response", "workspace_uuid", workspaceUUID, "error", err)
		return fmt.Errorf("marshal workspace: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) GetPersonalWorkspace(w http.ResponseWriter, r *http.Request) error {
	userUUID := userUUIDFromParams(r)
	if userUUID == "" {
		return apperror.BadRequestError("empty user uuid")
	}

	workspace, err := h.UserService.GetPersonalWorkspace(userUUID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(workspace)
	if err != nil {
		h.Logger.Error("failed to marshal personal workspace response", "user_uuid", userUUID, "error", err)
		return fmt.Errorf("marshal personal workspace: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) CreateWorkspace(w http.ResponseWriter, r *http.Request) error {
	var dto CreateWorkspaceDTO
	defer r.Body.Close()

	if err := decodeJSONBody(r, &dto); err != nil {
		h.Logger.Warn("failed to decode create workspace payload", "error", err)
		return err
	}

	workspace, err := h.UserService.CreateWorkspace(dto)
	if err != nil {
		return err
	}

	w.Header().Set("Location", fmt.Sprintf("%s/%s", workspacesURL, workspace.Uuid))
	w.WriteHeader(http.StatusCreated)
	return nil
}

func (h *Handler) LeaveWorkspace(w http.ResponseWriter, r *http.Request) error {
	workspaceUUID := workspaceUUIDFromParams(r)
	if workspaceUUID == "" {
		return apperror.BadRequestError("empty workspace uuid")
	}

	var dto WorkspaceActorDTO
	defer r.Body.Close()

	if err := decodeJSONBody(r, &dto); err != nil {
		h.Logger.Warn("failed to decode leave workspace payload", "error", err)
		return err
	}

	if err := h.UserService.LeaveWorkspace(workspaceUUID, dto.UserUUID); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) DeleteWorkspace(w http.ResponseWriter, r *http.Request) error {
	workspaceUUID := workspaceUUIDFromParams(r)
	if workspaceUUID == "" {
		return apperror.BadRequestError("empty workspace uuid")
	}

	var dto WorkspaceActorDTO
	defer r.Body.Close()

	if err := decodeJSONBody(r, &dto); err != nil {
		h.Logger.Warn("failed to decode delete workspace payload", "error", err)
		return err
	}

	if err := h.UserService.DeleteWorkspace(workspaceUUID, dto.UserUUID); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) GetWorkspaceMembers(w http.ResponseWriter, r *http.Request) error {
	workspaceUUID := workspaceUUIDFromParams(r)
	if workspaceUUID == "" {
		return apperror.BadRequestError("empty workspace uuid")
	}

	members, err := h.UserService.GetWorkspaceMembers(workspaceUUID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(members)
	if err != nil {
		h.Logger.Error("failed to marshal workspace members response", "workspace_uuid", workspaceUUID, "error", err)
		return fmt.Errorf("marshal workspace members: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) UpdateWorkspaceMember(w http.ResponseWriter, r *http.Request) error {
	workspaceUUID := workspaceUUIDFromParams(r)
	if workspaceUUID == "" {
		return apperror.BadRequestError("empty workspace uuid")
	}

	memberUserUUID := r.PathValue("member_uuid")
	if memberUserUUID == "" {
		return apperror.BadRequestError("empty member user uuid")
	}

	var dto UpdateWorkspaceMemberDTO
	defer r.Body.Close()

	if err := decodeJSONBody(r, &dto); err != nil {
		h.Logger.Warn("failed to decode update workspace member payload", "error", err)
		return err
	}

	member, err := h.UserService.UpdateWorkspaceMember(workspaceUUID, memberUserUUID, dto)
	if err != nil {
		return err
	}

	data, err := json.Marshal(member)
	if err != nil {
		h.Logger.Error(
			"failed to marshal updated workspace member response",
			"workspace_uuid", workspaceUUID,
			"member_user_uuid", memberUserUUID,
			"error", err,
		)
		return fmt.Errorf("marshal updated workspace member: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) GetWorkspaceInvites(w http.ResponseWriter, r *http.Request) error {
	workspaceUUID := workspaceUUIDFromParams(r)
	if workspaceUUID == "" {
		return apperror.BadRequestError("empty workspace uuid")
	}

	userUUID := r.URL.Query().Get("user_id")
	if userUUID == "" {
		userUUID = r.URL.Query().Get("user_uuid")
	}
	if userUUID == "" {
		return apperror.BadRequestError("user_id is required")
	}

	invites, err := h.UserService.GetWorkspaceSentInvites(workspaceUUID, userUUID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(invites)
	if err != nil {
		h.Logger.Error("failed to marshal workspace invites response", "workspace_uuid", workspaceUUID, "user_uuid", userUUID, "error", err)
		return fmt.Errorf("marshal workspace invites: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) CreateWorkspaceInvite(w http.ResponseWriter, r *http.Request) error {
	workspaceUUID := workspaceUUIDFromParams(r)
	if workspaceUUID == "" {
		return apperror.BadRequestError("empty workspace uuid")
	}

	var dto CreateWorkspaceInviteDTO
	defer r.Body.Close()

	if err := decodeJSONBody(r, &dto); err != nil {
		h.Logger.Warn("failed to decode create workspace invite payload", "error", err)
		return err
	}

	invite, err := h.UserService.CreateWorkspaceInvite(workspaceUUID, dto)
	if err != nil {
		return err
	}

	data, err := json.Marshal(invite)
	if err != nil {
		h.Logger.Error("failed to marshal workspace invite response", "workspace_uuid", workspaceUUID, "error", err)
		return fmt.Errorf("marshal workspace invite: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) AcceptWorkspaceInvite(w http.ResponseWriter, r *http.Request) error {
	inviteUUID := workspaceUUIDFromParams(r)
	if inviteUUID == "" {
		return apperror.BadRequestError("empty invite uuid")
	}

	var dto ResolveWorkspaceInviteDTO
	defer r.Body.Close()

	if err := decodeJSONBody(r, &dto); err != nil {
		h.Logger.Warn("failed to decode accept workspace invite payload", "error", err)
		return err
	}

	workspace, err := h.UserService.AcceptWorkspaceInvite(inviteUUID, dto.UserUUID)
	if err != nil {
		return err
	}

	if publishErr := h.EventPublisher.PublishWorkspaceInviteAccepted(context.Background(), events.WorkspaceInviteAcceptedPayload{
		WorkspaceID:   workspace.Uuid,
		WorkspaceName: workspace.Name,
		WorkspaceType: workspace.Visibility,
		UserUUID:      dto.UserUUID,
		InviteUUID:    inviteUUID,
	}); publishErr != nil {
		h.Logger.Warn("failed to publish workspace.invite.accepted event", "invite_uuid", inviteUUID, "workspace_uuid", workspace.Uuid, "error", publishErr)
	}

	data, err := json.Marshal(workspace)
	if err != nil {
		h.Logger.Error("failed to marshal accepted workspace response", "invite_uuid", inviteUUID, "error", err)
		return fmt.Errorf("marshal accepted workspace: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func (h *Handler) DeclineWorkspaceInvite(w http.ResponseWriter, r *http.Request) error {
	inviteUUID := workspaceUUIDFromParams(r)
	if inviteUUID == "" {
		return apperror.BadRequestError("empty invite uuid")
	}

	var dto ResolveWorkspaceInviteDTO
	defer r.Body.Close()

	if err := decodeJSONBody(r, &dto); err != nil {
		h.Logger.Warn("failed to decode decline workspace invite payload", "error", err)
		return err
	}

	if err := h.UserService.DeclineWorkspaceInvite(inviteUUID, dto.UserUUID); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) GetWorkspaceAccess(w http.ResponseWriter, r *http.Request) error {
	workspaceUUID := workspaceUUIDFromParams(r)
	if workspaceUUID == "" {
		return apperror.BadRequestError("empty workspace uuid")
	}

	userUUID := r.URL.Query().Get("user_id")
	if userUUID == "" {
		userUUID = r.URL.Query().Get("user_uuid")
	}
	if userUUID == "" {
		return apperror.BadRequestError("user_id is required")
	}

	access, err := h.UserService.GetWorkspaceAccess(workspaceUUID, userUUID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(access)
	if err != nil {
		h.Logger.Error("failed to marshal workspace access response", "workspace_uuid", workspaceUUID, "user_uuid", userUUID, "error", err)
		return fmt.Errorf("marshal workspace access: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil
}

func workspaceUUIDFromParams(r *http.Request) string {
	return r.PathValue("uuid")
}
