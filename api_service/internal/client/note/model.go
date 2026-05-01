package note

type Note struct {
	Uuid         string   `json:"uuid,omitempty"`
	Header       string   `json:"header,omitempty"`
	Body         string   `json:"body,omitempty"`
	ShortBody    string   `json:"short_body,omitempty"`
	CreatedDate  int64    `json:"created_date,omitempty"`
	CategoryUuid string   `json:"category_uuid,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}

type CreateNoteDTO struct {
	Header       string   `json:"header"`
	Body         string   `json:"body"`
	CategoryUuid string   `json:"category_uuid"`
	Tags         []string `json:"tags,omitempty"`
}

type UpdateNoteDTO struct {
	Header       string   `json:"header,omitempty"`
	Body         string   `json:"body,omitempty"`
	CategoryUuid string   `json:"category_uuid,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}

type Tag struct {
	Uuid string `json:"uuid,omitempty"`
	Name string `json:"name,omitempty"`
}

type CreateTagDTO struct {
	Name string `json:"name"`
}
