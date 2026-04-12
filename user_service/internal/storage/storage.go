package storage

import handlermodel "user_service/internal/handlers/users"

type Storage interface {
	Create(user handlermodel.User) (string, error)
	FindOne(userUUID string) (handlermodel.User, error)
	FindByUsername(username string) (handlermodel.User, error)
	Ping() error
}
