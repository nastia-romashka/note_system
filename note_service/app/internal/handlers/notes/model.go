package notes

type Note struct {
	Uuid           string     `json:"uuid,omitempty" bson:"-"`
	UserUuid       string     `json:"user_uuid,omitempty" bson:"user_uuid,omitempty"`
	WorkspaceID    string     `json:"workspace_id,omitempty" bson:"workspace_id,omitempty"`
	AuthorUserUUID string     `json:"author_user_uuid,omitempty" bson:"author_user_uuid,omitempty"`
	Header         string     `json:"header,omitempty" bson:"header,omitempty"`
	Body           string     `json:"body,omitempty" bson:"body,omitempty"`
	ShortBody      string     `json:"short_body,omitempty" bson:"short_body,omitempty"`
	CreatedDate    int64      `json:"created_date,omitempty" bson:"created_date,omitempty"`
	UpdatedAt      int64      `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
	CategoryUuid   string     `json:"category_uuid,omitempty" bson:"category_uuid,omitempty"`
	Tags           []string   `json:"tags,omitempty" bson:"tags,omitempty"`
	Event          *NoteEvent `json:"event,omitempty" bson:"event,omitempty"`
}

type CreateNoteDTO struct {
	UserUuid       string     `json:"user_uuid"`
	WorkspaceID    string     `json:"workspace_id,omitempty"`
	AuthorUserUUID string     `json:"author_user_uuid,omitempty"`
	Header         string     `json:"header"`
	Body           string     `json:"body"`
	CategoryUuid   string     `json:"category_uuid"`
	Tags           []string   `json:"tags,omitempty"`
	Event          *NoteEvent `json:"event,omitempty"`
}

type UpdateNoteDTO struct {
	UserUuid     string     `json:"user_uuid,omitempty" bson:"user_uuid,omitempty"`
	WorkspaceID  string     `json:"workspace_id,omitempty" bson:"workspace_id,omitempty"`
	Header       string     `json:"header,omitempty" bson:"header,omitempty"`
	Body         string     `json:"body,omitempty" bson:"body,omitempty"`
	CategoryUuid string     `json:"category_uuid,omitempty" bson:"category_uuid,omitempty"`
	Tags         []string   `json:"tags,omitempty" bson:"tags,omitempty"`
	Event        *NoteEvent `json:"event,omitempty" bson:"event,omitempty"`
}

type NoteStats struct {
	NotesCount int64 `json:"notes_count"`
	TagsCount  int64 `json:"tags_count"`
}

type NoteEvent struct {
	Enabled bool  `json:"enabled" bson:"enabled"`
	StartAt int64 `json:"start_at,omitempty" bson:"start_at,omitempty"`
	EndAt   int64 `json:"end_at,omitempty" bson:"end_at,omitempty"`
}
