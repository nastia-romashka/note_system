package storage

import handlermodel "user_service/internal/handlers/users"

type Storage interface {
	Create(user handlermodel.User) (string, error)
	FindOne(userUUID string) (handlermodel.User, error)
	FindProfile(userUUID string) (handlermodel.UserProfile, error)
	FindByUsername(username string) (handlermodel.User, error)
	UpdateProfile(userUUID, username, email, passwordHash string) error
	UpdateLastLogin(userUUID string) error
	CreateAction(userUUID string, dto handlermodel.CreateUserActionDTO) error
	FindActions(userUUID string, limit, offset int) ([]handlermodel.UserAction, error)
	CreateSession(dto handlermodel.CreateUserSessionDTO) error
	RotateSession(dto handlermodel.RotateUserSessionDTO) (handlermodel.UserSession, error)
	RevokeSession(refreshTokenHash string) error
	FindWorkspacesByUser(userUUID string) ([]handlermodel.Workspace, error)
	FindWorkspaceByID(workspaceUUID string) (handlermodel.Workspace, error)
	FindPersonalWorkspace(userUUID string) (handlermodel.Workspace, error)
	CreateWorkspace(dto handlermodel.CreateWorkspaceDTO) (handlermodel.Workspace, error)
	LeaveWorkspace(workspaceUUID, userUUID string) error
	DeleteWorkspace(workspaceUUID, actorUserUUID string) error
	FindWorkspaceMembers(workspaceUUID string) ([]handlermodel.WorkspaceMember, error)
	UpdateWorkspaceMember(workspaceUUID, memberUserUUID string, dto handlermodel.UpdateWorkspaceMemberDTO) (handlermodel.WorkspaceMember, error)
	FindWorkspaceInvitesByUser(userUUID string) ([]handlermodel.WorkspaceInvite, error)
	FindWorkspaceInvitesByWorkspace(workspaceUUID, actorUserUUID string) ([]handlermodel.WorkspaceInvite, error)
	CreateWorkspaceInvite(workspaceUUID string, dto handlermodel.CreateWorkspaceInviteDTO) (handlermodel.WorkspaceInvite, error)
	AcceptWorkspaceInvite(inviteUUID, userUUID string) (handlermodel.Workspace, error)
	DeclineWorkspaceInvite(inviteUUID, userUUID string) error
	FindWorkspaceAccess(workspaceUUID, userUUID string) (handlermodel.WorkspaceAccess, error)
	Ping() error
	Close()
}
