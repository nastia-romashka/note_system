package service

import (
	"fmt"
	"strings"

	"search_service/internal/apperror"
	handlermodel "search_service/internal/handlers/notes"
)

type NotesRepository interface {
	Search(q, workspaceID, categoryUUID string, tagUUIDs []string, page, perPage int) ([]handlermodel.SearchNote, error)
	Upsert(note handlermodel.IndexedNote) error
	UpsertMany(notes []handlermodel.IndexedNote) error
	Delete(noteUUID string) error
	DeleteByWorkspace(workspaceID string) error
}

type NotesService struct {
	repo NotesRepository
}

func NewNotesService(repo NotesRepository) *NotesService {
	return &NotesService{repo: repo}
}

func (s *NotesService) Search(q, workspaceID, categoryUUID string, tagUUIDs []string, page, perPage int) ([]handlermodel.SearchNote, error) {
	if strings.TrimSpace(workspaceID) == "" {
		return nil, apperror.BadRequestError("workspace_id is required")
	}

	notes, err := s.repo.Search(strings.TrimSpace(q), workspaceID, strings.TrimSpace(categoryUUID), tagUUIDs, page, perPage)
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

func (s *NotesService) Delete(noteUUID string) error {
	if strings.TrimSpace(noteUUID) == "" {
		return apperror.BadRequestError("note id is required")
	}

	if err := s.repo.Delete(noteUUID); err != nil {
		return fmt.Errorf("delete indexed note: %w", err)
	}

	return nil
}

func (s *NotesService) DeleteByWorkspace(workspaceID string) error {
	if strings.TrimSpace(workspaceID) == "" {
		return apperror.BadRequestError("workspace_id is required")
	}

	if err := s.repo.DeleteByWorkspace(workspaceID); err != nil {
		return fmt.Errorf("delete indexed notes by workspace: %w", err)
	}

	return nil
}

func validateIndexedNote(note handlermodel.IndexedNote) error {
	switch {
	case strings.TrimSpace(note.ID) == "":
		return apperror.BadRequestError("id is required")
	case strings.TrimSpace(note.WorkspaceID) == "":
		return apperror.BadRequestError("workspace_id is required")
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
