package models

type Product struct {
	GUID         string `bson:"guid,omitempty" json:"id,omitempty"`
	CategoryGuid string `bson:"category_id,omitempty" json:"category_id,omitempty"`
	Name         string `bson:"name,omitempty" json:"name,omitempty"`
	Description  string `bson:"description,omitempty" json:"description,omitempty"`
	Price        int    `bson:"price,omitempty" json:"price,omitempty"`
	Quantity     int    `bson:"quantity,omitempty" json:"quantity,omitempty"`
}
