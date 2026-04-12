package storage

import handlermodel "note_service/internal/handlers/notes"
import tagmodel "note_service/internal/handlers/tags"

type Storage interface {
	Create(note handlermodel.Note) (string, error)
	FindOne(noteUUID string) (handlermodel.Note, error)
	FindByCategoryUUID(categoryUUID string) ([]handlermodel.Note, error)
	Update(noteUUID string, note handlermodel.Note, tagsUpdate bool) error
	Delete(noteUUID string) error
	CreateTag(tag tagmodel.Tag) (string, error)
	FindTags(tagUUIDs []string) ([]tagmodel.Tag, error)
	CheckTagsExist(tagUUIDs []string) error
	DeleteTag(tagUUID string) error
	Ping() error
}
