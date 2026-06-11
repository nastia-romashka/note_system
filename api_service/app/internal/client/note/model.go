package note

type Note struct {
	Uuid           string     `json:"uuid,omitempty"`
	UserUuid       string     `json:"user_uuid,omitempty"`
	WorkspaceID    string     `json:"workspace_id,omitempty"`
	AuthorUserUUID string     `json:"author_user_uuid,omitempty"`
	Header         string     `json:"header,omitempty"`
	Body           string     `json:"body,omitempty"`
	ShortBody      string     `json:"short_body,omitempty"`
	CreatedDate    int64      `json:"created_date,omitempty"`
	UpdatedAt      int64      `json:"updated_at,omitempty"`
	CategoryUuid   string     `json:"category_uuid,omitempty"`
	Tags           []string   `json:"tags,omitempty"`
	Event          *NoteEvent `json:"event,omitempty"`
}

type CreateNoteDTO struct {
	UserUuid       string     `json:"user_uuid,omitempty"`
	WorkspaceID    string     `json:"workspace_id,omitempty"`
	AuthorUserUUID string     `json:"author_user_uuid,omitempty"`
	Header         string     `json:"header"`
	Body           string     `json:"body"`
	CategoryUuid   string     `json:"category_uuid"`
	Tags           []string   `json:"tags,omitempty"`
	Event          *NoteEvent `json:"event,omitempty"`
}

type UpdateNoteDTO struct {
	UserUuid     string     `json:"user_uuid,omitempty"`
	WorkspaceID  string     `json:"workspace_id,omitempty"`
	Header       string     `json:"header,omitempty"`
	Body         string     `json:"body,omitempty"`
	CategoryUuid string     `json:"category_uuid,omitempty"`
	Tags         []string   `json:"tags,omitempty"`
	Event        *NoteEvent `json:"event,omitempty"`
}

type Tag struct {
	Uuid        string `json:"uuid,omitempty"`
	UserUuid    string `json:"user_uuid,omitempty"`
	WorkspaceID string `json:"workspace_id,omitempty"`
	Name        string `json:"name,omitempty"`
}

type CreateTagDTO struct {
	UserUuid    string `json:"user_uuid,omitempty"`
	WorkspaceID string `json:"workspace_id,omitempty"`
	Name        string `json:"name"`
}

type NoteStats struct {
	NotesCount int64 `json:"notes_count"`
	TagsCount  int64 `json:"tags_count"`
}

type NoteEvent struct {
	Enabled bool  `json:"enabled"`
	StartAt int64 `json:"start_at,omitempty"`
	EndAt   int64 `json:"end_at,omitempty"`
}
