package tags

type Tag struct {
	Uuid     string `json:"uuid,omitempty" bson:"-"`
	UserUuid string `json:"user_uuid,omitempty" bson:"user_uuid,omitempty"`
	Name     string `json:"name,omitempty" bson:"name,omitempty"`
}

type CreateTagDTO struct {
	UserUuid string `json:"user_uuid"`
	Name     string `json:"name"`
}
