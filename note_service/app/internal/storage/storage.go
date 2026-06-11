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

type Scope struct {
	UserUUID    string
	WorkspaceID string
}

type Storage interface {
	Create(note handlermodel.Note) (string, error)
	FindOne(noteUUID string, scope Scope) (handlermodel.Note, error)
	FindByCategoryUUID(categoryUUID string, scope Scope) ([]handlermodel.Note, error)
	FindByEventRange(from, to int64, scope Scope) ([]handlermodel.Note, error)
	CountStats(scope Scope) (handlermodel.NoteStats, error)
	Update(noteUUID string, scope Scope, note handlermodel.Note, opts UpdateOptions) error
	Delete(noteUUID string, scope Scope) error
	CreateTag(tag tagmodel.Tag) (string, error)
	FindTags(tagUUIDs []string, scope Scope) ([]tagmodel.Tag, error)
	CheckTagsExist(tagUUIDs []string, scope Scope) error
	DeleteTag(tagUUID string, scope Scope) error
	Ping() error
}
