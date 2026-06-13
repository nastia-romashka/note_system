package users

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"user_service/internal/apperror"
	handlermodel "user_service/internal/handlers/users"
)

const (
	workspaceVisibilityPersonal   = "private"
	workspaceVisibilityShared     = "invite_only"
	defaultWorkspaceInviteTTLDays = 7
)

func normalizeWorkspaceRole(role string) string {
	role = strings.TrimSpace(strings.ToLower(role))
	if role == "creator" {
		return "editor"
	}
	return role
}

func (s *service) GetWorkspaces(userUUID string) ([]handlermodel.Workspace, error) {
	userUUID = strings.TrimSpace(userUUID)
	if userUUID == "" {
		return nil, apperror.BadRequestError("user_uuid is required")
	}

	workspaces, err := s.storage.FindWorkspacesByUser(userUUID)
	if err != nil {
		return nil, fmt.Errorf("get workspaces: %w", err)
	}
	if !containsPersonalWorkspace(workspaces) {
		workspace, ensureErr := s.ensurePersonalWorkspace(userUUID)
		if ensureErr != nil {
			return nil, ensureErr
		}
		workspaces = append([]handlermodel.Workspace{workspace}, workspaces...)
	}

	return workspaces, nil
}

func (s *service) GetWorkspace(workspaceUUID string) (handlermodel.Workspace, error) {
	workspaceUUID = strings.TrimSpace(workspaceUUID)
	if workspaceUUID == "" {
		return handlermodel.Workspace{}, apperror.BadRequestError("workspace_uuid is required")
	}

	workspace, err := s.storage.FindWorkspaceByID(workspaceUUID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return handlermodel.Workspace{}, err
		}
		return handlermodel.Workspace{}, fmt.Errorf("get workspace: %w", err)
	}

	return workspace, nil
}

func (s *service) GetPersonalWorkspace(userUUID string) (handlermodel.Workspace, error) {
	userUUID = strings.TrimSpace(userUUID)
	if userUUID == "" {
		return handlermodel.Workspace{}, apperror.BadRequestError("user_uuid is required")
	}

	workspace, err := s.storage.FindPersonalWorkspace(userUUID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return s.ensurePersonalWorkspace(userUUID)
		}
		return handlermodel.Workspace{}, fmt.Errorf("get personal workspace: %w", err)
	}

	return workspace, nil
}

func (s *service) CreateWorkspace(dto handlermodel.CreateWorkspaceDTO) (handlermodel.Workspace, error) {
	dto.OwnerUserUUID = strings.TrimSpace(dto.OwnerUserUUID)
	dto.Name = strings.TrimSpace(dto.Name)
	dto.Visibility = strings.TrimSpace(strings.ToLower(dto.Visibility))

	if dto.OwnerUserUUID == "" {
		return handlermodel.Workspace{}, apperror.BadRequestError("owner_user_uuid is required")
	}
	if dto.Name == "" {
		return handlermodel.Workspace{}, apperror.BadRequestError("name is required")
	}
	if dto.Visibility == "" {
		dto.Visibility = workspaceVisibilityShared
	}
	if dto.Visibility != workspaceVisibilityShared {
		return handlermodel.Workspace{}, apperror.BadRequestError("only invite_only workspaces can be created through public api")
	}

	workspace, err := s.storage.CreateWorkspace(dto)
	if err != nil {
		return handlermodel.Workspace{}, fmt.Errorf("create workspace: %w", err)
	}

	return workspace, nil
}

func (s *service) LeaveWorkspace(workspaceUUID, userUUID string) error {
	workspaceUUID = strings.TrimSpace(workspaceUUID)
	userUUID = strings.TrimSpace(userUUID)

	if workspaceUUID == "" {
		return apperror.BadRequestError("workspace_uuid is required")
	}
	if userUUID == "" {
		return apperror.BadRequestError("user_uuid is required")
	}

	if err := s.storage.LeaveWorkspace(workspaceUUID, userUUID); err != nil {
		if errors.Is(err, apperror.ErrUnauthorized) || errors.Is(err, apperror.ErrNotFound) {
			return err
		}
		return fmt.Errorf("leave workspace: %w", err)
	}

	return nil
}

func (s *service) DeleteWorkspace(workspaceUUID, actorUserUUID string) error {
	workspaceUUID = strings.TrimSpace(workspaceUUID)
	actorUserUUID = strings.TrimSpace(actorUserUUID)

	if workspaceUUID == "" {
		return apperror.BadRequestError("workspace_uuid is required")
	}
	if actorUserUUID == "" {
		return apperror.BadRequestError("actor_user_uuid is required")
	}

	if err := s.storage.DeleteWorkspace(workspaceUUID, actorUserUUID); err != nil {
		if errors.Is(err, apperror.ErrUnauthorized) || errors.Is(err, apperror.ErrNotFound) {
			return err
		}
		return fmt.Errorf("delete workspace: %w", err)
	}

	return nil
}

func (s *service) GetWorkspaceMembers(workspaceUUID string) ([]handlermodel.WorkspaceMember, error) {
	workspaceUUID = strings.TrimSpace(workspaceUUID)
	if workspaceUUID == "" {
		return nil, apperror.BadRequestError("workspace_uuid is required")
	}

	members, err := s.storage.FindWorkspaceMembers(workspaceUUID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("get workspace members: %w", err)
	}

	return members, nil
}

func (s *service) UpdateWorkspaceMember(workspaceUUID, memberUserUUID string, dto handlermodel.UpdateWorkspaceMemberDTO) (handlermodel.WorkspaceMember, error) {
	workspaceUUID = strings.TrimSpace(workspaceUUID)
	memberUserUUID = strings.TrimSpace(memberUserUUID)
	dto.ActorUserUUID = strings.TrimSpace(dto.ActorUserUUID)
	dto.Role = normalizeWorkspaceRole(dto.Role)
	dto.Status = strings.TrimSpace(strings.ToLower(dto.Status))

	if workspaceUUID == "" {
		return handlermodel.WorkspaceMember{}, apperror.BadRequestError("workspace_uuid is required")
	}
	if memberUserUUID == "" {
		return handlermodel.WorkspaceMember{}, apperror.BadRequestError("member_user_uuid is required")
	}
	if dto.ActorUserUUID == "" {
		return handlermodel.WorkspaceMember{}, apperror.BadRequestError("actor_user_uuid is required")
	}
	if dto.Role == "" && dto.Status == "" {
		return handlermodel.WorkspaceMember{}, apperror.BadRequestError("role or status is required")
	}
	if dto.Role != "" && dto.Role != "viewer" && dto.Role != "editor" {
		return handlermodel.WorkspaceMember{}, apperror.BadRequestError("role must be viewer or editor")
	}
	if dto.Status != "" && dto.Status != "active" && dto.Status != "removed" {
		return handlermodel.WorkspaceMember{}, apperror.BadRequestError("status must be active or removed")
	}

	member, err := s.storage.UpdateWorkspaceMember(workspaceUUID, memberUserUUID, dto)
	if err != nil {
		if errors.Is(err, apperror.ErrUnauthorized) || errors.Is(err, apperror.ErrNotFound) {
			return handlermodel.WorkspaceMember{}, err
		}
		return handlermodel.WorkspaceMember{}, fmt.Errorf("update workspace member: %w", err)
	}

	return member, nil
}

func (s *service) GetWorkspaceInvites(userUUID string) ([]handlermodel.WorkspaceInvite, error) {
	userUUID = strings.TrimSpace(userUUID)
	if userUUID == "" {
		return nil, apperror.BadRequestError("user_uuid is required")
	}

	invites, err := s.storage.FindWorkspaceInvitesByUser(userUUID)
	if err != nil {
		return nil, fmt.Errorf("get workspace invites: %w", err)
	}

	return invites, nil
}

func (s *service) GetWorkspaceSentInvites(workspaceUUID, actorUserUUID string) ([]handlermodel.WorkspaceInvite, error) {
	workspaceUUID = strings.TrimSpace(workspaceUUID)
	actorUserUUID = strings.TrimSpace(actorUserUUID)

	if workspaceUUID == "" {
		return nil, apperror.BadRequestError("workspace_uuid is required")
	}
	if actorUserUUID == "" {
		return nil, apperror.BadRequestError("actor_user_uuid is required")
	}

	invites, err := s.storage.FindWorkspaceInvitesByWorkspace(workspaceUUID, actorUserUUID)
	if err != nil {
		if errors.Is(err, apperror.ErrUnauthorized) || errors.Is(err, apperror.ErrNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("get workspace sent invites: %w", err)
	}

	return invites, nil
}

func (s *service) CreateWorkspaceInvite(workspaceUUID string, dto handlermodel.CreateWorkspaceInviteDTO) (handlermodel.WorkspaceInvite, error) {
	workspaceUUID = strings.TrimSpace(workspaceUUID)
	dto.InvitedByUserUUID = strings.TrimSpace(dto.InvitedByUserUUID)
	dto.Email = strings.TrimSpace(strings.ToLower(dto.Email))
	dto.Role = normalizeWorkspaceRole(dto.Role)

	if workspaceUUID == "" {
		return handlermodel.WorkspaceInvite{}, apperror.BadRequestError("workspace_uuid is required")
	}
	if dto.InvitedByUserUUID == "" {
		return handlermodel.WorkspaceInvite{}, apperror.BadRequestError("invited_by_user_uuid is required")
	}
	if dto.Email == "" {
		return handlermodel.WorkspaceInvite{}, apperror.BadRequestError("email is required")
	}
	if dto.Role == "" {
		dto.Role = "viewer"
	}
	if dto.Role != "viewer" && dto.Role != "editor" {
		return handlermodel.WorkspaceInvite{}, apperror.BadRequestError("role must be viewer or editor")
	}
	if dto.ExpiresAt <= 0 {
		dto.ExpiresAt = time.Now().Add(defaultWorkspaceInviteTTLDays * 24 * time.Hour).Unix()
	}

	invite, err := s.storage.CreateWorkspaceInvite(workspaceUUID, dto)
	if err != nil {
		if errors.Is(err, apperror.ErrUnauthorized) || errors.Is(err, apperror.ErrNotFound) {
			return handlermodel.WorkspaceInvite{}, err
		}
		return handlermodel.WorkspaceInvite{}, fmt.Errorf("create workspace invite: %w", err)
	}

	return invite, nil
}

func (s *service) AcceptWorkspaceInvite(inviteUUID, userUUID string) (handlermodel.Workspace, error) {
	inviteUUID = strings.TrimSpace(inviteUUID)
	userUUID = strings.TrimSpace(userUUID)

	if inviteUUID == "" {
		return handlermodel.Workspace{}, apperror.BadRequestError("invite_uuid is required")
	}
	if userUUID == "" {
		return handlermodel.Workspace{}, apperror.BadRequestError("user_uuid is required")
	}

	workspace, err := s.storage.AcceptWorkspaceInvite(inviteUUID, userUUID)
	if err != nil {
		if errors.Is(err, apperror.ErrUnauthorized) || errors.Is(err, apperror.ErrNotFound) {
			return handlermodel.Workspace{}, err
		}
		return handlermodel.Workspace{}, fmt.Errorf("accept workspace invite: %w", err)
	}

	return workspace, nil
}

func (s *service) DeclineWorkspaceInvite(inviteUUID, userUUID string) error {
	inviteUUID = strings.TrimSpace(inviteUUID)
	userUUID = strings.TrimSpace(userUUID)

	if inviteUUID == "" {
		return apperror.BadRequestError("invite_uuid is required")
	}
	if userUUID == "" {
		return apperror.BadRequestError("user_uuid is required")
	}

	if err := s.storage.DeclineWorkspaceInvite(inviteUUID, userUUID); err != nil {
		if errors.Is(err, apperror.ErrUnauthorized) || errors.Is(err, apperror.ErrNotFound) {
			return err
		}
		return fmt.Errorf("decline workspace invite: %w", err)
	}

	return nil
}

func (s *service) GetWorkspaceAccess(workspaceUUID, userUUID string) (handlermodel.WorkspaceAccess, error) {
	workspaceUUID = strings.TrimSpace(workspaceUUID)
	userUUID = strings.TrimSpace(userUUID)

	if workspaceUUID == "" {
		return handlermodel.WorkspaceAccess{}, apperror.BadRequestError("workspace_uuid is required")
	}
	if userUUID == "" {
		return handlermodel.WorkspaceAccess{}, apperror.BadRequestError("user_uuid is required")
	}

	access, err := s.storage.FindWorkspaceAccess(workspaceUUID, userUUID)
	if err != nil {
		if errors.Is(err, apperror.ErrUnauthorized) || errors.Is(err, apperror.ErrNotFound) {
			return handlermodel.WorkspaceAccess{}, err
		}
		return handlermodel.WorkspaceAccess{}, fmt.Errorf("get workspace access: %w", err)
	}

	return access, nil
}

func (s *service) ensurePersonalWorkspace(userUUID string) (handlermodel.Workspace, error) {
	workspace, err := s.storage.CreateWorkspace(handlermodel.CreateWorkspaceDTO{
		OwnerUserUUID: userUUID,
		Name:          "Личное пространство",
		Visibility:    workspaceVisibilityPersonal,
	})
	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) && appErr.Code == "US-50000" {
			return handlermodel.Workspace{}, err
		}

		if errors.Is(err, apperror.ErrNotFound) {
			return handlermodel.Workspace{}, err
		}

		workspace, fallbackErr := s.storage.FindPersonalWorkspace(userUUID)
		if fallbackErr == nil {
			return workspace, nil
		}
		return handlermodel.Workspace{}, fmt.Errorf("ensure personal workspace: %w", err)
	}

	return workspace, nil
}

func containsPersonalWorkspace(workspaces []handlermodel.Workspace) bool {
	for _, workspace := range workspaces {
		if workspace.IsPersonal {
			return true
		}
	}

	return false
}
