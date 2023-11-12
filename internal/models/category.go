package models

type Category struct {
	GUID string `bson:"guid,omitempty" json:"id,omitempty"`
	Name string `bson:"name,omitempty" json:"name,omitempty"`
}
