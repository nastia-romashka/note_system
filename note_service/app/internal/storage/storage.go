package storage

import handlermodel "note_service/internal/handlers/notes"
import tagmodel "note_service/internal/handlers/tags"

type UpdateOptions struct {
	Header   bool
	Body     bool
	Category bool
	Tags     bool
	Event    bool
}

type Storage interface {
	Create(note handlermodel.Note) (string, error)
	FindOne(noteUUID, userUUID string) (handlermodel.Note, error)
	FindByCategoryUUID(categoryUUID, userUUID string) ([]handlermodel.Note, error)
	FindByEventRange(from, to int64, userUUID string) ([]handlermodel.Note, error)
	CountStats(userUUID string) (handlermodel.NoteStats, error)
	Update(noteUUID, userUUID string, note handlermodel.Note, opts UpdateOptions) error
	Delete(noteUUID, userUUID string) error
	CreateTag(tag tagmodel.Tag) (string, error)
	FindTags(tagUUIDs []string, userUUID string) ([]tagmodel.Tag, error)
	CheckTagsExist(tagUUIDs []string, userUUID string) error
	DeleteTag(tagUUID, userUUID string) error
	Ping() error
}
