package tags

import (
	"errors"
	"fmt"
	"strings"

	"note_service/internal/apperror"
	handlermodel "note_service/internal/handlers/tags"
	"note_service/internal/storage"
)

type service struct {
	storage storage.Storage
}

func NewService(storage storage.Storage) *service {
	return &service{storage: storage}
}

func (s *service) Get(tagUUIDs []string, userUUID, workspaceID string) (tags []handlermodel.Tag, err error) {
	if strings.TrimSpace(userUUID) == "" {
		return nil, apperror.BadRequestError("user_uuid is required")
	}
	if strings.TrimSpace(workspaceID) == "" {
		return nil, apperror.BadRequestError("workspace_id is required")
	}

	tags, err = s.storage.FindTags(tagUUIDs, storage.Scope{
		UserUUID:    userUUID,
		WorkspaceID: strings.TrimSpace(workspaceID),
	})
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("get tags: %w", err)
	}

	return tags, nil
}

func (s *service) Create(dto handlermodel.CreateTagDTO) (tagUUID string, err error) {
	dto.UserUuid = strings.TrimSpace(dto.UserUuid)
	dto.WorkspaceID = strings.TrimSpace(dto.WorkspaceID)
	dto.Name = strings.TrimSpace(dto.Name)
	if dto.UserUuid == "" {
		return "", apperror.BadRequestError("user_uuid is required")
	}
	if dto.WorkspaceID == "" {
		return "", apperror.BadRequestError("workspace_id is required")
	}
	if dto.Name == "" {
		return "", apperror.BadRequestError("tag name is required")
	}

	tagUUID, err = s.storage.CreateTag(handlermodel.Tag{
		UserUuid:    dto.UserUuid,
		WorkspaceID: dto.WorkspaceID,
		Name:        dto.Name,
	})
	if err != nil {
		return "", fmt.Errorf("create tag: %w", err)
	}

	return tagUUID, nil
}

func (s *service) Delete(tagUUID, userUUID, workspaceID string) (err error) {
	if strings.TrimSpace(userUUID) == "" {
		return apperror.BadRequestError("user_uuid is required")
	}
	if strings.TrimSpace(workspaceID) == "" {
		return apperror.BadRequestError("workspace_id is required")
	}

	err = s.storage.DeleteTag(tagUUID, storage.Scope{
		UserUUID:    userUUID,
		WorkspaceID: strings.TrimSpace(workspaceID),
	})
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return err
		}
		return fmt.Errorf("delete tag: %w", err)
	}

	return nil
}
