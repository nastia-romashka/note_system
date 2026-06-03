package service

import (
	"fmt"
	"strings"

	"search_service/internal/apperror"
	handlermodel "search_service/internal/handlers/notes"
)

type NotesRepository interface {
	Search(q, userUUID, categoryUUID string, tagUUIDs []string, page, perPage int) ([]handlermodel.SearchNote, error)
	Upsert(note handlermodel.IndexedNote) error
	UpsertMany(notes []handlermodel.IndexedNote) error
	Delete(noteUUID, userUUID string) error
	DeleteByUser(userUUID string) error
}

type NotesService struct {
	repo NotesRepository
}

func NewNotesService(repo NotesRepository) *NotesService {
	return &NotesService{repo: repo}
}

func (s *NotesService) Search(q, userUUID, categoryUUID string, tagUUIDs []string, page, perPage int) ([]handlermodel.SearchNote, error) {
	if strings.TrimSpace(userUUID) == "" {
		return nil, apperror.BadRequestError("user_uuid is required")
	}

	notes, err := s.repo.Search(strings.TrimSpace(q), userUUID, strings.TrimSpace(categoryUUID), tagUUIDs, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("search notes: %w", err)
	}

	return notes, nil
}

func (s *NotesService) Upsert(note handlermodel.IndexedNote) error {
	if err := validateIndexedNote(note); err != nil {
		return err
	}

	if err := s.repo.Upsert(note); err != nil {
		return fmt.Errorf("upsert note: %w", err)
	}

	return nil
}

func (s *NotesService) UpsertMany(notes []handlermodel.IndexedNote) error {
	if len(notes) == 0 {
		return nil
	}

	for _, note := range notes {
		if err := validateIndexedNote(note); err != nil {
			return err
		}
	}

	if err := s.repo.UpsertMany(notes); err != nil {
		return fmt.Errorf("upsert notes: %w", err)
	}

	return nil
}

func (s *NotesService) Delete(noteUUID, userUUID string) error {
	if strings.TrimSpace(noteUUID) == "" {
		return apperror.BadRequestError("note id is required")
	}
	if strings.TrimSpace(userUUID) == "" {
		return apperror.BadRequestError("user_uuid is required")
	}

	if err := s.repo.Delete(noteUUID, userUUID); err != nil {
		return fmt.Errorf("delete indexed note: %w", err)
	}

	return nil
}

func (s *NotesService) DeleteByUser(userUUID string) error {
	if strings.TrimSpace(userUUID) == "" {
		return apperror.BadRequestError("user_uuid is required")
	}

	if err := s.repo.DeleteByUser(userUUID); err != nil {
		return fmt.Errorf("delete indexed notes by user: %w", err)
	}

	return nil
}

func validateIndexedNote(note handlermodel.IndexedNote) error {
	switch {
	case strings.TrimSpace(note.ID) == "":
		return apperror.BadRequestError("id is required")
	case strings.TrimSpace(note.UserUUID) == "":
		return apperror.BadRequestError("user_uuid is required")
	case strings.TrimSpace(note.Header) == "":
		return apperror.BadRequestError("header is required")
	case strings.TrimSpace(note.Body) == "":
		return apperror.BadRequestError("body is required")
	case strings.TrimSpace(note.CategoryUUID) == "":
		return apperror.BadRequestError("category_uuid is required")
	case strings.TrimSpace(note.CategoryName) == "":
		return apperror.BadRequestError("category_name is required")
	}

	return nil
}
