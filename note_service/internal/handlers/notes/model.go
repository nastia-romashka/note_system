package notes

type Note struct {
	Uuid         string   `json:"uuid,omitempty" bson:"-"`
	Header       string   `json:"header,omitempty" bson:"header,omitempty"`
	Body         string   `json:"body,omitempty" bson:"body,omitempty"`
	ShortBody    string   `json:"short_body,omitempty" bson:"short_body,omitempty"`
	CreatedDate  int64    `json:"created_date,omitempty" bson:"created_date,omitempty"`
	CategoryUuid string   `json:"category_uuid,omitempty" bson:"category_uuid,omitempty"`
	Tags         []string `json:"tags,omitempty" bson:"tags,omitempty"`
}

type CreateNoteDTO struct {
	Header       string   `json:"header"`
	Body         string   `json:"body"`
	CategoryUuid string   `json:"category_uuid"`
	Tags         []string `json:"tags,omitempty"`
}

type UpdateNoteDTO struct {
	Header       string   `json:"header,omitempty" bson:"header,omitempty"`
	Body         string   `json:"body,omitempty" bson:"body,omitempty"`
	CategoryUuid string   `json:"category_uuid,omitempty" bson:"category_uuid,omitempty"`
	Tags         []string `json:"tags,omitempty" bson:"tags,omitempty"`
}
