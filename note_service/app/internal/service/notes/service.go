package notes

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"note_service/internal/apperror"
	"note_service/internal/events"
	handlermodel "note_service/internal/handlers/notes"
	"note_service/internal/storage"
	"note_service/pkg/logging"
)

type service struct {
	storage   storage.Storage
	publisher events.Publisher
	logger    logging.Logger
}

func NewService(storage storage.Storage, publisher events.Publisher, logger logging.Logger) *service {
	return &service{
		storage:   storage,
		publisher: publisher,
		logger:    logger.With("layer", "notes_service"),
	}
}

func (s *service) GetOne(noteUUID, userUUID, workspaceID string) (note handlermodel.Note, err error) {
	if strings.TrimSpace(userUUID) == "" {
		return handlermodel.Note{}, apperror.BadRequestError("user_uuid is required")
	}
	if strings.TrimSpace(workspaceID) == "" {
		return handlermodel.Note{}, apperror.BadRequestError("workspace_id is required")
	}

	note, err = s.storage.FindOne(noteUUID, storage.Scope{
		UserUUID:    userUUID,
		WorkspaceID: strings.TrimSpace(workspaceID),
	})
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return handlermodel.Note{}, err
		}
		return handlermodel.Note{}, fmt.Errorf("get note: %w", err)
	}

	return note, nil
}

func (s *service) GetByCategoryUUID(categoryUUID, userUUID, workspaceID string) (notes []handlermodel.Note, err error) {
	if strings.TrimSpace(userUUID) == "" {
		return nil, apperror.BadRequestError("user_uuid is required")
	}
	if strings.TrimSpace(workspaceID) == "" {
		return nil, apperror.BadRequestError("workspace_id is required")
	}

	notes, err = s.storage.FindByCategoryUUID(categoryUUID, storage.Scope{
		UserUUID:    userUUID,
		WorkspaceID: strings.TrimSpace(workspaceID),
	})
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("get notes by category: %w", err)
	}

	return notes, nil
}

func (s *service) GetByEventRange(from, to int64, userUUID, workspaceID string) (notes []handlermodel.Note, err error) {
	if strings.TrimSpace(userUUID) == "" {
		return nil, apperror.BadRequestError("user_uuid is required")
	}
	if strings.TrimSpace(workspaceID) == "" {
		return nil, apperror.BadRequestError("workspace_id is required")
	}
	if from <= 0 || to <= 0 {
		return nil, apperror.BadRequestError("from and to are required")
	}
	if from > to {
		return nil, apperror.BadRequestError("from must be less than or equal to to")
	}

	notes, err = s.storage.FindByEventRange(from, to, storage.Scope{
		UserUUID:    userUUID,
		WorkspaceID: strings.TrimSpace(workspaceID),
	})
	if err != nil {
		return nil, fmt.Errorf("get notes by event range: %w", err)
	}

	return notes, nil
}

func (s *service) GetStats(userUUID, workspaceID string) (stats handlermodel.NoteStats, err error) {
	if strings.TrimSpace(userUUID) == "" {
		return handlermodel.NoteStats{}, apperror.BadRequestError("user_uuid is required")
	}
	if strings.TrimSpace(workspaceID) == "" {
		return handlermodel.NoteStats{}, apperror.BadRequestError("workspace_id is required")
	}

	stats, err = s.storage.CountStats(storage.Scope{
		UserUUID:    userUUID,
		WorkspaceID: strings.TrimSpace(workspaceID),
	})
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
	if strings.TrimSpace(dto.WorkspaceID) == "" {
		return "", apperror.BadRequestError("workspace_id is required")
	}
	dto.Body = strings.TrimSpace(dto.Body)
	scope := storage.Scope{
		UserUUID:    dto.UserUuid,
		WorkspaceID: strings.TrimSpace(dto.WorkspaceID),
	}
	if err = s.storage.CheckTagsExist(dto.Tags, scope); err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return "", apperror.BadRequestError("one or more tags do not exist")
		}
		return "", fmt.Errorf("check tags before create note: %w", err)
	}

	if err = validateEvent(dto.Event); err != nil {
		return "", err
	}
	now := time.Now().Unix()
	workspaceID := strings.TrimSpace(dto.WorkspaceID)
	authorUserUUID := strings.TrimSpace(dto.AuthorUserUUID)
	if authorUserUUID == "" {
		authorUserUUID = dto.UserUuid
	}

	noteUUID, err = s.storage.Create(handlermodel.Note{
		UserUuid:       dto.UserUuid,
		WorkspaceID:    workspaceID,
		AuthorUserUUID: authorUserUUID,
		Header:         dto.Header,
		Body:           dto.Body,
		ShortBody:      makeShortBody(dto.Body),
		CreatedDate:    now,
		UpdatedAt:      now,
		CategoryUuid:   dto.CategoryUuid,
		Tags:           dto.Tags,
		Event:          dto.Event,
	})
	if err != nil {
		return "", fmt.Errorf("create note: %w", err)
	}

	if publishErr := s.publishNoteCreated(context.Background(), noteUUID, dto.UserUuid, workspaceID); publishErr != nil {
		s.logger.Warn(
			"failed to publish note.created event",
			"note_uuid", noteUUID,
			"workspace_id", workspaceID,
			"error", publishErr,
		)
	}

	return noteUUID, nil
}

func (s *service) Update(
	noteUUID, userUUID, workspaceID string,
	dto handlermodel.UpdateNoteDTO,
	headerUpdate, bodyUpdate, categoryUpdate, tagsUpdate, eventUpdate bool,
) (err error) {
	if strings.TrimSpace(userUUID) == "" {
		return apperror.BadRequestError("user_uuid is required")
	}
	if strings.TrimSpace(workspaceID) == "" {
		return apperror.BadRequestError("workspace_id is required")
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
		if err = s.storage.CheckTagsExist(dto.Tags, storage.Scope{
			UserUUID:    userUUID,
			WorkspaceID: strings.TrimSpace(workspaceID),
		}); err != nil {
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

	err = s.storage.Update(noteUUID, storage.Scope{
		UserUUID:    userUUID,
		WorkspaceID: strings.TrimSpace(workspaceID),
	}, handlermodel.Note{
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

	if publishErr := s.publishNoteUpdated(context.Background(), noteUUID, userUUID, workspaceID); publishErr != nil {
		s.logger.Warn(
			"failed to publish note.updated event",
			"note_uuid", noteUUID,
			"workspace_id", workspaceID,
			"error", publishErr,
		)
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

func (s *service) Delete(noteUUID, userUUID, workspaceID string) (err error) {
	if strings.TrimSpace(userUUID) == "" {
		return apperror.BadRequestError("user_uuid is required")
	}
	if strings.TrimSpace(workspaceID) == "" {
		return apperror.BadRequestError("workspace_id is required")
	}

	normalizedWorkspaceID := strings.TrimSpace(workspaceID)
	err = s.storage.Delete(noteUUID, storage.Scope{
		UserUUID:    userUUID,
		WorkspaceID: normalizedWorkspaceID,
	})
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return err
		}
		return fmt.Errorf("delete note: %w", err)
	}

	if publishErr := s.publisher.PublishNoteDeleted(context.Background(), events.NoteDeletedPayload{
		NoteUUID:    noteUUID,
		UserUUID:    userUUID,
		WorkspaceID: normalizedWorkspaceID,
	}); publishErr != nil {
		s.logger.Warn(
			"failed to publish note.deleted event",
			"note_uuid", noteUUID,
			"workspace_id", normalizedWorkspaceID,
			"error", publishErr,
		)
	}

	return nil
}

func (s *service) DeleteWorkspace(workspaceID string) error {
	workspaceID = strings.TrimSpace(workspaceID)
	if workspaceID == "" {
		return apperror.BadRequestError("workspace_id is required")
	}

	if err := s.storage.DeleteWorkspace(workspaceID); err != nil {
		return fmt.Errorf("delete workspace notes: %w", err)
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

func (s *service) publishNoteUpdated(ctx context.Context, noteUUID, userUUID, workspaceID string) error {
	note, err := s.loadNoteForEvent(noteUUID, userUUID, workspaceID)
	if err != nil {
		return err
	}

	if err = s.publisher.PublishNoteUpdated(ctx, note); err != nil {
		return fmt.Errorf("publish note.updated: %w", err)
	}

	return nil
}

func (s *service) publishNoteCreated(ctx context.Context, noteUUID, userUUID, workspaceID string) error {
	note, err := s.loadNoteForEvent(noteUUID, userUUID, workspaceID)
	if err != nil {
		return err
	}

	if err = s.publisher.PublishNoteCreated(ctx, note); err != nil {
		return fmt.Errorf("publish note.created: %w", err)
	}

	return nil
}

func (s *service) loadNoteForEvent(noteUUID, userUUID, workspaceID string) (handlermodel.Note, error) {
	note, err := s.storage.FindOne(noteUUID, storage.Scope{
		UserUUID:    userUUID,
		WorkspaceID: strings.TrimSpace(workspaceID),
	})
	if err != nil {
		return handlermodel.Note{}, fmt.Errorf("load note for event: %w", err)
	}
	return note, nil
}
