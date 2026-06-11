package notes

type IndexedNote struct {
	ID            string   `json:"id"`
	WorkspaceID   string   `json:"workspace_id"`
	Header        string   `json:"header"`
	Body          string   `json:"body"`
	ShortBody     string   `json:"short_body,omitempty"`
	CategoryUUID  string   `json:"category_uuid"`
	CategoryName  string   `json:"category_name"`
	TagUUIDs      []string `json:"tag_uuids,omitempty"`
	TagNames      []string `json:"tag_names,omitempty"`
	FileNamesText string   `json:"file_names_text,omitempty"`
	CreatedAt     int64    `json:"created_at"`
	UpdatedAt     int64    `json:"updated_at"`
}

type SearchNote struct {
	Uuid          string   `json:"uuid"`
	Header        string   `json:"header,omitempty"`
	Body          string   `json:"body,omitempty"`
	ShortBody     string   `json:"short_body,omitempty"`
	CreatedDate   int64    `json:"created_date,omitempty"`
	UpdatedAt     int64    `json:"updated_at,omitempty"`
	CategoryUuid  string   `json:"category_uuid,omitempty"`
	CategoryName  string   `json:"category_name,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	TagNames      []string `json:"tag_names,omitempty"`
	FileNamesText string   `json:"file_names_text,omitempty"`
}
