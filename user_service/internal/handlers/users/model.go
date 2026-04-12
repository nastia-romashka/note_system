package users

type User struct {
	Uuid         string `json:"uuid,omitempty" bson:"-"`
	Username     string `json:"username,omitempty" bson:"username,omitempty"`
	Email        string `json:"email,omitempty" bson:"email,omitempty"`
	PasswordHash string `json:"-" bson:"password_hash,omitempty"`
	CreatedDate  int64  `json:"created_date,omitempty" bson:"created_date,omitempty"`
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
