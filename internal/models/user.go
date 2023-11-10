package models

import "time"

type User struct {
	GUID         string     `bson:"guid,omitempty" json:"id,omitempty"`
	Login        string     `bson:"login,omitempty" json:"login,omitempty"`
	Name         string     `bson:"name,omitempty" json:"name,omitempty"`
	LastName     string     `bson:"last_name,omitempty" json:"last_name,omitempty"`
	LastLoginAt  *time.Time `bson:"last_login_at,omitempty" json:"last_login_at,omitempty"`
	CreatedAt    *time.Time `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt    *time.Time `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
	RefreshToken string     `bson:"refresh_token,omitempty" json:"-"`
	IsLogged     bool       `bson:"is_logged,omitempty" json:"is_logged,omitempty"`
}
