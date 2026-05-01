package user

type CreateUserDTO struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthUserDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type User struct {
	Uuid        string `json:"uuid,omitempty"`
	Username    string `json:"username,omitempty"`
	Email       string `json:"email,omitempty"`
	CreatedDate int64  `json:"created_date,omitempty"`
}
