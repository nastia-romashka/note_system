package category

type Category struct {
	Uuid       string     `json:"uuid"`
	Name       string     `json:"name"`
	Color      string     `json:"color,omitempty"`
	UserUuid   string     `json:"user_uuid,omitempty"`
	ParentUuid string     `json:"parent_uuid,omitempty"`
	Children   []Category `json:"children,omitempty"`
}

type CreateCategoryDTO struct {
	Name       string `json:"name"`
	Color      string `json:"color,omitempty"`
	UserUuid   string `json:"user_uuid"`
	ParentUuid string `json:"parent_uuid,omitempty"`
}

type UpdateCategoryDTO struct {
	Uuid       string `json:"uuid,omitempty"`
	Name       string `json:"name,omitempty"`
	Color      string `json:"color,omitempty"`
	UserUuid   string `json:"user_uuid,omitempty"`
	ParentUuid string `json:"parent_uuid,omitempty"`
}

type DeleteCategoryDTO struct {
	Uuid     string `json:"uuid"`
	UserUuid string `json:"user_uuid,omitempty"`
}

type CategoryStats struct {
	CategoriesCount int64 `json:"categories_count"`
}

type GraphNode struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Label        string `json:"label"`
	Color        string `json:"color,omitempty"`
	CategoryUuid string `json:"category_uuid,omitempty"`
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
	Uuid         string `json:"uuid"`
	UserUuid     string `json:"user_uuid"`
	CategoryUuid string `json:"category_uuid"`
	Header       string `json:"header"`
}

type UpdateGraphNoteDTO struct {
	UserUuid     string `json:"user_uuid"`
	CategoryUuid string `json:"category_uuid,omitempty"`
	Header       string `json:"header,omitempty"`
}

type DeleteGraphNoteDTO struct {
	UserUuid string `json:"user_uuid"`
}

type LinkGraphNotesDTO struct {
	UserUuid string `json:"user_uuid"`
}
