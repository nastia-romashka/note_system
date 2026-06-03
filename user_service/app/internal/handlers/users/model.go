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
