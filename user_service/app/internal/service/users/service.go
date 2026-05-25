package users

import (
	"errors"
	"fmt"
	"strings"

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

func (s *service) GetProfile(userUUID string) (profile handlermodel.UserProfile, err error) {
	userUUID = strings.TrimSpace(userUUID)
	if userUUID == "" {
		return handlermodel.UserProfile{}, apperror.BadRequestError("user_uuid is required")
	}

	profile, err = s.storage.FindProfile(userUUID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return handlermodel.UserProfile{}, err
		}
		return handlermodel.UserProfile{}, fmt.Errorf("get user profile: %w", err)
	}

	return profile, nil
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

	if err = s.storage.UpdateLastLogin(user.Uuid); err != nil {
		return handlermodel.User{}, fmt.Errorf("update last login: %w", err)
	}

	err = s.storage.CreateAction(user.Uuid, handlermodel.CreateUserActionDTO{
		Action:     "auth.login",
		EntityType: "user",
		EntityId:   user.Uuid,
		Status:     "success",
	})
	if err != nil {
		return handlermodel.User{}, fmt.Errorf("create login action: %w", err)
	}

	return user, nil
}

func (s *service) CreateAction(userUUID string, dto handlermodel.CreateUserActionDTO) (err error) {
	userUUID = strings.TrimSpace(userUUID)
	dto.Action = strings.TrimSpace(dto.Action)
	dto.EntityType = strings.TrimSpace(dto.EntityType)
	dto.EntityId = strings.TrimSpace(dto.EntityId)
	dto.Status = strings.TrimSpace(dto.Status)
	dto.IPAddress = strings.TrimSpace(dto.IPAddress)
	dto.UserAgent = strings.TrimSpace(dto.UserAgent)

	if userUUID == "" {
		return apperror.BadRequestError("user_uuid is required")
	}
	if dto.Action == "" {
		return apperror.BadRequestError("action is required")
	}
	if dto.EntityType == "" {
		return apperror.BadRequestError("entity_type is required")
	}
	if dto.Status == "" {
		dto.Status = "success"
	}
	if dto.Metadata == nil {
		dto.Metadata = map[string]any{}
	}

	if err = s.storage.CreateAction(userUUID, dto); err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return err
		}
		return fmt.Errorf("create user action: %w", err)
	}

	return nil
}

func (s *service) GetActions(userUUID string, limit, offset int) (actions []handlermodel.UserAction, err error) {
	userUUID = strings.TrimSpace(userUUID)
	if userUUID == "" {
		return nil, apperror.BadRequestError("user_uuid is required")
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	actions, err = s.storage.FindActions(userUUID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get user actions: %w", err)
	}

	return actions, nil
}
