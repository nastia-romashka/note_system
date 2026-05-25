package notes

type Note struct {
	Uuid         string   `json:"uuid,omitempty" bson:"-"`
	UserUuid     string   `json:"user_uuid,omitempty" bson:"user_uuid,omitempty"`
	Header       string   `json:"header,omitempty" bson:"header,omitempty"`
	Body         string   `json:"body,omitempty" bson:"body,omitempty"`
	ShortBody    string   `json:"short_body,omitempty" bson:"short_body,omitempty"`
	CreatedDate  int64    `json:"created_date,omitempty" bson:"created_date,omitempty"`
	CategoryUuid string   `json:"category_uuid,omitempty" bson:"category_uuid,omitempty"`
	Tags         []string `json:"tags,omitempty" bson:"tags,omitempty"`
}

type CreateNoteDTO struct {
	UserUuid     string   `json:"user_uuid"`
	Header       string   `json:"header"`
	Body         string   `json:"body"`
	CategoryUuid string   `json:"category_uuid"`
	Tags         []string `json:"tags,omitempty"`
}

type UpdateNoteDTO struct {
	UserUuid     string   `json:"user_uuid,omitempty" bson:"user_uuid,omitempty"`
	Header       string   `json:"header,omitempty" bson:"header,omitempty"`
	Body         string   `json:"body,omitempty" bson:"body,omitempty"`
	CategoryUuid string   `json:"category_uuid,omitempty" bson:"category_uuid,omitempty"`
	Tags         []string `json:"tags,omitempty" bson:"tags,omitempty"`
}

type NoteStats struct {
	NotesCount int64 `json:"notes_count"`
	TagsCount  int64 `json:"tags_count"`
}
