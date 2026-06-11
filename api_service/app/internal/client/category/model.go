package category

type Category struct {
	Uuid           string     `json:"uuid"`
	WorkspaceID    string     `json:"workspace_id,omitempty"`
	AuthorUserUUID string     `json:"author_user_uuid,omitempty"`
	Name           string     `json:"name"`
	Color          string     `json:"color,omitempty"`
	CreatedAt      int64      `json:"created_at,omitempty"`
	ParentUuid     string     `json:"parent_uuid,omitempty"`
	Children       []Category `json:"children,omitempty"`
}

type CreateCategoryDTO struct {
	Name           string `json:"name"`
	Color          string `json:"color,omitempty"`
	WorkspaceID    string `json:"workspace_id"`
	WorkspaceName  string `json:"workspace_name,omitempty"`
	WorkspaceType  string `json:"workspace_type,omitempty"`
	AuthorUserUUID string `json:"author_user_uuid"`
	ParentUuid     string `json:"parent_uuid,omitempty"`
}

type UpdateCategoryDTO struct {
	Uuid        string `json:"uuid,omitempty"`
	WorkspaceID string `json:"workspace_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Color       string `json:"color,omitempty"`
	ParentUuid  string `json:"parent_uuid,omitempty"`
}

type DeleteCategoryDTO struct {
	Uuid        string `json:"uuid"`
	WorkspaceID string `json:"workspace_id,omitempty"`
}

type CategoryStats struct {
	CategoriesCount int64 `json:"categories_count"`
}

type GraphNode struct {
	ID             string `json:"id"`
	Type           string `json:"type"`
	Label          string `json:"label"`
	WorkspaceID    string `json:"workspace_id,omitempty"`
	AuthorUserUUID string `json:"author_user_uuid,omitempty"`
	Color          string `json:"color,omitempty"`
	CategoryUuid   string `json:"category_uuid,omitempty"`
	CreatedAt      int64  `json:"created_at,omitempty"`
}

type GraphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
}

type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

type CreateGraphNoteDTO struct {
	Uuid           string `json:"uuid"`
	WorkspaceID    string `json:"workspace_id"`
	WorkspaceName  string `json:"workspace_name,omitempty"`
	WorkspaceType  string `json:"workspace_type,omitempty"`
	AuthorUserUUID string `json:"author_user_uuid"`
	CategoryUuid   string `json:"category_uuid"`
	Header         string `json:"header"`
	CreatedDate    int64  `json:"created_date"`
}

type UpdateGraphNoteDTO struct {
	WorkspaceID  string `json:"workspace_id"`
	CategoryUuid string `json:"category_uuid,omitempty"`
	Header       string `json:"header,omitempty"`
}

type DeleteGraphNoteDTO struct {
	WorkspaceID string `json:"workspace_id"`
}

type UserGraphLinkDTO struct {
	WorkspaceID string `json:"workspace_id"`
	UserUuid    string `json:"user_uuid"`
	SourceID    string `json:"source_id"`
	TargetID    string `json:"target_id"`
}
