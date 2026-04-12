package users

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"user_service/internal/apperror"
	handlermodel "user_service/internal/handlers/users"
	"user_service/internal/storage"
)

type service struct {
	storage storage.Storage
}

func NewService(storage storage.Storage) *service {
	return &service{storage: storage}
}

func (s *service) GetOne(userUUID string) (user handlermodel.User, err error) {
	user, err = s.storage.FindOne(userUUID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return handlermodel.User{}, err
		}
		return handlermodel.User{}, fmt.Errorf("get user: %w", err)
	}

	return user, nil
}

func (s *service) Create(dto handlermodel.CreateUserDTO) (userUUID string, err error) {
	dto.Username = strings.TrimSpace(dto.Username)
	dto.Email = strings.TrimSpace(strings.ToLower(dto.Email))

	if dto.Username == "" {
		return "", apperror.BadRequestError("username is required")
	}
	if dto.Email == "" {
		return "", apperror.BadRequestError("email is required")
	}
	if strings.TrimSpace(dto.Password) == "" {
		return "", apperror.BadRequestError("password is required")
	}

	var passwordHash []byte
	passwordHash, err = bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	userUUID, err = s.storage.Create(handlermodel.User{
		Username:     dto.Username,
		Email:        dto.Email,
		PasswordHash: string(passwordHash),
		CreatedDate:  time.Now().Unix(),
	})
	if err != nil {
		return "", fmt.Errorf("create user: %w", err)
	}

	return userUUID, nil
}

func (s *service) Authenticate(dto handlermodel.AuthUserDTO) (user handlermodel.User, err error) {
	dto.Username = strings.TrimSpace(dto.Username)
	if dto.Username == "" || strings.TrimSpace(dto.Password) == "" {
		return handlermodel.User{}, apperror.BadRequestError("username and password are required")
	}

	user, err = s.storage.FindByUsername(dto.Username)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return handlermodel.User{}, apperror.NotFoundError("user not found")
		}
		return handlermodel.User{}, fmt.Errorf("find user by username: %w", err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(dto.Password)); err != nil {
		return handlermodel.User{}, apperror.BadRequestError("invalid credentials")
	}

	return user, nil
}
