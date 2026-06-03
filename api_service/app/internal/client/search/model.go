package search

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
