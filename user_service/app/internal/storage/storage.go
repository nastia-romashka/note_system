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
	Ping() error
	Close()
}
