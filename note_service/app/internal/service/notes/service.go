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

func (s *service) GetOne(noteUUID, userUUID string) (note handlermodel.Note, err error) {
	if strings.TrimSpace(userUUID) == "" {
		return handlermodel.Note{}, apperror.BadRequestError("user_uuid is required")
	}

	note, err = s.storage.FindOne(noteUUID, userUUID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return handlermodel.Note{}, err
		}
		return handlermodel.Note{}, fmt.Errorf("get note: %w", err)
	}

	return note, nil
}

func (s *service) GetByCategoryUUID(categoryUUID, userUUID string) (notes []handlermodel.Note, err error) {
	if strings.TrimSpace(userUUID) == "" {
		return nil, apperror.BadRequestError("user_uuid is required")
	}

	notes, err = s.storage.FindByCategoryUUID(categoryUUID, userUUID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("get notes by category: %w", err)
	}

	return notes, nil
}

func (s *service) GetByEventRange(from, to int64, userUUID string) (notes []handlermodel.Note, err error) {
	if strings.TrimSpace(userUUID) == "" {
		return nil, apperror.BadRequestError("user_uuid is required")
	}
	if from <= 0 || to <= 0 {
		return nil, apperror.BadRequestError("from and to are required")
	}
	if from > to {
		return nil, apperror.BadRequestError("from must be less than or equal to to")
	}

	notes, err = s.storage.FindByEventRange(from, to, userUUID)
	if err != nil {
		return nil, fmt.Errorf("get notes by event range: %w", err)
	}

	return notes, nil
}

func (s *service) GetStats(userUUID string) (stats handlermodel.NoteStats, err error) {
	if strings.TrimSpace(userUUID) == "" {
		return handlermodel.NoteStats{}, apperror.BadRequestError("user_uuid is required")
	}

	stats, err = s.storage.CountStats(userUUID)
	if err != nil {
		return handlermodel.NoteStats{}, fmt.Errorf("get note stats: %w", err)
	}

	return stats, nil
}

func (s *service) Create(dto handlermodel.CreateNoteDTO) (noteUUID string, err error) {
	if strings.TrimSpace(dto.UserUuid) == "" {
		return "", apperror.BadRequestError("user_uuid is required")
	}
	if strings.TrimSpace(dto.Header) == "" {
		return "", apperror.BadRequestError("header is required")
	}
	if strings.TrimSpace(dto.CategoryUuid) == "" {
		return "", apperror.BadRequestError("category_uuid is required")
	}
	dto.Body = strings.TrimSpace(dto.Body)
	if err = s.storage.CheckTagsExist(dto.Tags, dto.UserUuid); err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return "", apperror.BadRequestError("one or more tags do not exist")
		}
		return "", fmt.Errorf("check tags before create note: %w", err)
	}

	if err = validateEvent(dto.Event); err != nil {
		return "", err
	}
	now := time.Now().Unix()

	noteUUID, err = s.storage.Create(handlermodel.Note{
		UserUuid:     dto.UserUuid,
		Header:       dto.Header,
		Body:         dto.Body,
		ShortBody:    makeShortBody(dto.Body),
		CreatedDate:  now,
		UpdatedAt:    now,
		CategoryUuid: dto.CategoryUuid,
		Tags:         dto.Tags,
		Event:        dto.Event,
	})
	if err != nil {
		return "", fmt.Errorf("create note: %w", err)
	}

	return noteUUID, nil
}

func (s *service) Update(
	noteUUID, userUUID string,
	dto handlermodel.UpdateNoteDTO,
	headerUpdate, bodyUpdate, categoryUpdate, tagsUpdate, eventUpdate bool,
) (err error) {
	if strings.TrimSpace(userUUID) == "" {
		return apperror.BadRequestError("user_uuid is required")
	}
	if !headerUpdate && !bodyUpdate && !categoryUpdate && !tagsUpdate && !eventUpdate {
		return apperror.BadRequestError("empty update payload")
	}

	if bodyUpdate {
		dto.Body = strings.TrimSpace(dto.Body)
	}

	if headerUpdate {
		dto.Header = strings.TrimSpace(dto.Header)
		if dto.Header == "" {
			return apperror.BadRequestError("header is required")
		}
	}
	if tagsUpdate {
		if err = s.storage.CheckTagsExist(dto.Tags, userUUID); err != nil {
			if errors.Is(err, apperror.ErrNotFound) {
				return apperror.BadRequestError("one or more tags do not exist")
			}
			return fmt.Errorf("check tags before update note: %w", err)
		}
	}
	if eventUpdate {
		if err = validateEvent(dto.Event); err != nil {
			return err
		}
	}

	err = s.storage.Update(noteUUID, userUUID, handlermodel.Note{
		Header:       dto.Header,
		Body:         dto.Body,
		ShortBody:    makeShortBody(dto.Body),
		UpdatedAt:    time.Now().Unix(),
		CategoryUuid: dto.CategoryUuid,
		Tags:         dto.Tags,
		Event:        dto.Event,
	}, storage.UpdateOptions{
		Header:   headerUpdate,
		Body:     bodyUpdate,
		Category: categoryUpdate,
		Tags:     tagsUpdate,
		Event:    eventUpdate,
	})
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return err
		}
		return fmt.Errorf("update note: %w", err)
	}

	return nil
}

func validateEvent(event *handlermodel.NoteEvent) error {
	if event == nil {
		return nil
	}
	if event.StartAt <= 0 {
		return apperror.BadRequestError("event.start_at is required")
	}
	if event.EndAt <= 0 {
		return apperror.BadRequestError("event.end_at is required")
	}
	if event.EndAt < event.StartAt {
		return apperror.BadRequestError("event.end_at must be greater than or equal to event.start_at")
	}

	return nil
}

func (s *service) Delete(noteUUID, userUUID string) (err error) {
	if strings.TrimSpace(userUUID) == "" {
		return apperror.BadRequestError("user_uuid is required")
	}

	err = s.storage.Delete(noteUUID, userUUID)
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
