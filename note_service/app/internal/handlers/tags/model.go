package tags

type Tag struct {
	Uuid        string `json:"uuid,omitempty" bson:"-"`
	UserUuid    string `json:"user_uuid,omitempty" bson:"user_uuid,omitempty"`
	WorkspaceID string `json:"workspace_id,omitempty" bson:"workspace_id,omitempty"`
	Name        string `json:"name,omitempty" bson:"name,omitempty"`
}

type CreateTagDTO struct {
	UserUuid    string `json:"user_uuid"`
	WorkspaceID string `json:"workspace_id,omitempty"`
	Name        string `json:"name"`
}
