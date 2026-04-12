package notes

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"note_service/internal/apperror"
	handlermodel "note_service/internal/handlers/notes"
	"note_service/internal/storage"
)

type service struct {
	storage storage.Storage
}

func NewService(storage storage.Storage) *service {
	return &service{storage: storage}
}

func (s *service) GetOne(noteUUID string) (note handlermodel.Note, err error) {
	note, err = s.storage.FindOne(noteUUID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return handlermodel.Note{}, err
		}
		return handlermodel.Note{}, fmt.Errorf("get note: %w", err)
	}

	return note, nil
}

func (s *service) GetByCategoryUUID(categoryUUID string) (notes []handlermodel.Note, err error) {
	notes, err = s.storage.FindByCategoryUUID(categoryUUID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("get notes by category: %w", err)
	}

	return notes, nil
}

func (s *service) Create(dto handlermodel.CreateNoteDTO) (noteUUID string, err error) {
	if strings.TrimSpace(dto.Header) == "" {
		return "", apperror.BadRequestError("header is required")
	}
	if strings.TrimSpace(dto.Body) == "" {
		return "", apperror.BadRequestError("body is required")
	}
	if strings.TrimSpace(dto.CategoryUuid) == "" {
		return "", apperror.BadRequestError("category_uuid is required")
	}
	if err = s.storage.CheckTagsExist(dto.Tags); err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return "", apperror.BadRequestError("one or more tags do not exist")
		}
		return "", fmt.Errorf("check tags before create note: %w", err)
	}

	noteUUID, err = s.storage.Create(handlermodel.Note{
		Header:       dto.Header,
		Body:         dto.Body,
		ShortBody:    makeShortBody(dto.Body),
		CreatedDate:  time.Now().Unix(),
		CategoryUuid: dto.CategoryUuid,
		Tags:         dto.Tags,
	})
	if err != nil {
		return "", fmt.Errorf("create note: %w", err)
	}

	return noteUUID, nil
}

func (s *service) Update(noteUUID string, dto handlermodel.UpdateNoteDTO, tagsUpdate bool) (err error) {
	if dto.Header == "" && dto.Body == "" && dto.CategoryUuid == "" && !tagsUpdate {
		return apperror.BadRequestError("empty update payload")
	}

	if dto.Body != "" {
		dto.Body = strings.TrimSpace(dto.Body)
		if dto.Body == "" {
			return apperror.BadRequestError("body is required")
		}
	}

	if dto.Header != "" {
		dto.Header = strings.TrimSpace(dto.Header)
		if dto.Header == "" {
			return apperror.BadRequestError("header is required")
		}
	}
	if tagsUpdate {
		if err = s.storage.CheckTagsExist(dto.Tags); err != nil {
			if errors.Is(err, apperror.ErrNotFound) {
				return apperror.BadRequestError("one or more tags do not exist")
			}
			return fmt.Errorf("check tags before update note: %w", err)
		}
	}

	err = s.storage.Update(noteUUID, handlermodel.Note{
		Header:       dto.Header,
		Body:         dto.Body,
		ShortBody:    makeShortBody(dto.Body),
		CategoryUuid: dto.CategoryUuid,
		Tags:         dto.Tags,
	}, tagsUpdate)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return err
		}
		return fmt.Errorf("update note: %w", err)
	}

	return nil
}

func (s *service) Delete(noteUUID string) (err error) {
	err = s.storage.Delete(noteUUID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return err
		}
		return fmt.Errorf("delete note: %w", err)
	}

	return nil
}

func makeShortBody(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}

	body = strings.Join(strings.Fields(body), " ")
	if len(body) <= 120 {
		return body
	}

	return body[:117] + "..."
}
