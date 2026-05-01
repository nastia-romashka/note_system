package tags

type Tag struct {
	Uuid string `json:"uuid,omitempty" bson:"-"`
	Name string `json:"name,omitempty" bson:"name,omitempty"`
}

type CreateTagDTO struct {
	Name string `json:"name"`
}
