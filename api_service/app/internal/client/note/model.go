package note

type Note struct {
	Uuid         string   `json:"uuid,omitempty"`
	UserUuid     string   `json:"user_uuid,omitempty"`
	Header       string   `json:"header,omitempty"`
	Body         string   `json:"body,omitempty"`
	ShortBody    string   `json:"short_body,omitempty"`
	CreatedDate  int64    `json:"created_date,omitempty"`
	CategoryUuid string   `json:"category_uuid,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}

type CreateNoteDTO struct {
	UserUuid     string   `json:"user_uuid,omitempty"`
	Header       string   `json:"header"`
	Body         string   `json:"body"`
	CategoryUuid string   `json:"category_uuid"`
	Tags         []string `json:"tags,omitempty"`
}

type UpdateNoteDTO struct {
	UserUuid     string   `json:"user_uuid,omitempty"`
	Header       string   `json:"header,omitempty"`
	Body         string   `json:"body,omitempty"`
	CategoryUuid string   `json:"category_uuid,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}

type Tag struct {
	Uuid     string `json:"uuid,omitempty"`
	UserUuid string `json:"user_uuid,omitempty"`
	Name     string `json:"name,omitempty"`
}

type CreateTagDTO struct {
	UserUuid string `json:"user_uuid,omitempty"`
	Name     string `json:"name"`
}

type NoteStats struct {
	NotesCount int64 `json:"notes_count"`
	TagsCount  int64 `json:"tags_count"`
}
