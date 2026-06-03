package notes

type IndexedNote struct {
	ID           string   `json:"id"`
	UserUUID     string   `json:"user_uuid"`
	Header       string   `json:"header"`
	Body         string   `json:"body"`
	ShortBody    string   `json:"short_body"`
	CategoryUUID string   `json:"category_uuid"`
	CategoryName string   `json:"category_name"`
	TagUUIDs     []string `json:"tag_uuids"`
	TagNames     []string `json:"tag_names"`
	CreatedDate  int64    `json:"created_date"`
}

type SearchNote struct {
	Uuid         string   `json:"uuid"`
	UserUuid     string   `json:"user_uuid,omitempty"`
	Header       string   `json:"header,omitempty"`
	Body         string   `json:"body,omitempty"`
	ShortBody    string   `json:"short_body,omitempty"`
	CreatedDate  int64    `json:"created_date,omitempty"`
	CategoryUuid string   `json:"category_uuid,omitempty"`
	CategoryName string   `json:"category_name,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	TagNames     []string `json:"tag_names,omitempty"`
}
