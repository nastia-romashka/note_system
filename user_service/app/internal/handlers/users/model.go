package users

type User struct {
	Uuid         string `json:"uuid,omitempty"`
	Username     string `json:"username,omitempty"`
	Email        string `json:"email,omitempty"`
	PasswordHash string `json:"-"`
	CreatedDate  int64  `json:"created_date,omitempty"`
}

type UserProfile struct {
	Uuid        string `json:"uuid,omitempty"`
	Username    string `json:"username,omitempty"`
	Email       string `json:"email,omitempty"`
	CreatedAt   int64  `json:"created_at,omitempty"`
	LastLoginAt *int64 `json:"last_login_at"`
}

type UserAction struct {
	Uuid       string         `json:"uuid,omitempty"`
	Action     string         `json:"action"`
	EntityType string         `json:"entity_type"`
	EntityId   string         `json:"entity_id,omitempty"`
	Status     string         `json:"status"`
	Metadata   map[string]any `json:"metadata"`
	IPAddress  string         `json:"ip_address,omitempty"`
	UserAgent  string         `json:"user_agent,omitempty"`
	CreatedAt  int64          `json:"created_at,omitempty"`
}

type UserSession struct {
	Uuid             string `json:"uuid,omitempty"`
	UserUUID         string `json:"user_uuid,omitempty"`
	RefreshTokenHash string `json:"-"`
	UserAgent        string `json:"user_agent,omitempty"`
	IPAddress        string `json:"ip_address,omitempty"`
	CreatedAt        int64  `json:"created_at,omitempty"`
	ExpiresAt        int64  `json:"expires_at,omitempty"`
	LastUsedAt       *int64 `json:"last_used_at,omitempty"`
}

type CreateUserActionDTO struct {
	Action     string         `json:"action"`
	EntityType string         `json:"entity_type"`
	EntityId   string         `json:"entity_id,omitempty"`
	Status     string         `json:"status,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	IPAddress  string         `json:"ip_address,omitempty"`
	UserAgent  string         `json:"user_agent,omitempty"`
}

type CreateUserDTO struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthUserDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdateUserProfileDTO struct {
	Username        string `json:"username,omitempty"`
	Email           string `json:"email,omitempty"`
	CurrentPassword string `json:"current_password,omitempty"`
	NewPassword     string `json:"new_password,omitempty"`
}

type CreateUserSessionDTO struct {
	UserUUID         string `json:"user_uuid"`
	RefreshTokenHash string `json:"refresh_token_hash"`
	ExpiresAt        int64  `json:"expires_at"`
	UserAgent        string `json:"user_agent,omitempty"`
	IPAddress        string `json:"ip_address,omitempty"`
}

type RotateUserSessionDTO struct {
	RefreshTokenHash    string `json:"refresh_token_hash"`
	NewRefreshTokenHash string `json:"new_refresh_token_hash"`
	ExpiresAt           int64  `json:"expires_at"`
	UserAgent           string `json:"user_agent,omitempty"`
	IPAddress           string `json:"ip_address,omitempty"`
}

type RevokeUserSessionDTO struct {
	RefreshTokenHash string `json:"refresh_token_hash"`
}

type Workspace struct {
	Uuid          string `json:"uuid,omitempty"`
	Name          string `json:"name,omitempty"`
	OwnerUserUUID string `json:"owner_user_uuid,omitempty"`
	Visibility    string `json:"visibility,omitempty"`
	IsPersonal    bool   `json:"is_personal,omitempty"`
	CreatedAt     int64  `json:"created_at,omitempty"`
	UpdatedAt     int64  `json:"updated_at,omitempty"`
}

type WorkspaceMember struct {
	WorkspaceUUID string `json:"workspace_uuid,omitempty"`
	UserUUID      string `json:"user_uuid,omitempty"`
	Username      string `json:"username,omitempty"`
	Email         string `json:"email,omitempty"`
	Role          string `json:"role,omitempty"`
	Status        string `json:"status,omitempty"`
	JoinedAt      *int64 `json:"joined_at,omitempty"`
	InvitedBy     string `json:"invited_by,omitempty"`
}

type WorkspaceInvite struct {
	Uuid              string `json:"uuid,omitempty"`
	WorkspaceUUID     string `json:"workspace_uuid,omitempty"`
	WorkspaceName     string `json:"workspace_name,omitempty"`
	Email             string `json:"email,omitempty"`
	Role              string `json:"role,omitempty"`
	InvitedByUserUUID string `json:"invited_by_user_uuid,omitempty"`
	InvitedByUsername string `json:"invited_by_username,omitempty"`
	ExpiresAt         int64  `json:"expires_at,omitempty"`
	AcceptedAt        *int64 `json:"accepted_at,omitempty"`
	DeclinedAt        *int64 `json:"declined_at,omitempty"`
	CreatedAt         int64  `json:"created_at,omitempty"`
	Status            string `json:"status,omitempty"`
}

type WorkspaceAccess struct {
	Workspace Workspace `json:"workspace"`
	UserUUID  string    `json:"user_uuid"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	Allowed   bool      `json:"allowed"`
}

type CreateWorkspaceDTO struct {
	OwnerUserUUID string `json:"owner_user_uuid"`
	Name          string `json:"name"`
	Visibility    string `json:"visibility,omitempty"`
}

type CreateWorkspaceInviteDTO struct {
	InvitedByUserUUID string `json:"invited_by_user_uuid"`
	Email             string `json:"email"`
	Role              string `json:"role,omitempty"`
	ExpiresAt         int64  `json:"expires_at,omitempty"`
}

type UpdateWorkspaceMemberDTO struct {
	ActorUserUUID string `json:"actor_user_uuid"`
	Role          string `json:"role,omitempty"`
	Status        string `json:"status,omitempty"`
}

type ResolveWorkspaceInviteDTO struct {
	UserUUID string `json:"user_uuid"`
}
