package category

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
