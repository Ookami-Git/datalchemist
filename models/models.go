// models.go
package models

import "github.com/golang-jwt/jwt/v5"

// ---- Database type
type Parameters struct {
	Name  string `gorm:"primary_key"`
	Value string
}

type Users struct {
	ID       uint   `gorm:"primary_key" json:"id"`
	Name     string `gorm:"unique;not null" json:"name"`
	Type     string `gorm:"not null" json:"type"`
	Lang     string `gorm:"default:default" json:"lang"`
	Theme    string `gorm:"default:default" json:"theme"`
	Password string `json:"password"`
}

type Groups struct {
	ID          uint   `gorm:"primary_key" json:"id"`
	Name        string `gorm:"unique;not null" json:"name"`
	Description string `json:"description"`
}

type Roles struct {
	ID   uint `gorm:"primary_key" json:"id"`
	Gid  uint `gorm:"index" json:"group"`
	User uint `gorm:"index" json:"user"`

	Group Groups `gorm:"foreignKey:Gid"`
	User_ Users  `gorm:"foreignKey:User"`
}

type Acl struct {
	ID   uint `gorm:"primary_key" json:"id"`
	View uint `gorm:"index" json:"view"`
	Gid  uint `gorm:"index" json:"gid"`

	View_  Views  `gorm:"foreignKey:View"`
	Group_ Groups `gorm:"foreignKey:Gid"`
}

type Sources struct {
	ID         uint   `gorm:"primary_key" json:"id"`
	Name       string `gorm:"unique;not null" json:"name"`
	Parameters string `json:"parameters"`
	JSON       string `json:"json"`
}

type Source_require struct {
	ID      uint `gorm:"primary_key" json:"id"`
	Source  uint `gorm:"index" json:"source_id"`
	Require uint `gorm:"index" json:"required_source_id"`

	Source_  Sources `gorm:"foreignKey:Source"`
	Require_ Sources `gorm:"foreignKey:Require"`
}

type Items struct {
	ID         uint   `gorm:"primary_key" json:"id"`
	Name       string `gorm:"unique;not null" json:"name"`
	Parameters string `json:"parameters"`
	Template   string `json:"template"`
	Javascript string `json:"javascript"`
}

type Item_sources struct {
	ID     uint `gorm:"primary_key" json:"id"`
	Item   uint `gorm:"index" json:"item_id"`
	Source uint `gorm:"index" json:"source_id"`

	Item_   Items   `gorm:"foreignKey:Item"`
	Source_ Sources `gorm:"foreignKey:Source"`
}

type Views struct {
	ID         uint   `gorm:"primary_key" json:"id"`
	Name       string `gorm:"unique;not null" json:"name"`
	Parameters string `json:"parameters"`
	Protected  bool   `gorm:"default:false" json:"protected"`
}

type View_items struct {
	ID         uint   `gorm:"primary_key" json:"id"`
	View       uint   `gorm:"index" json:"view_id"`
	Item       uint   `gorm:"index" json:"item_id"`
	Parameters string `json:"parameters"`

	View_ Views `gorm:"foreignKey:View"`
	Item_ Items `gorm:"foreignKey:Item"`
}

// --- Other type
type RolesByUsers struct {
	ID     int   `json:"uid"`
	Groups []int `json:"groups"`
}

type RolesByGroups struct {
	ID    int   `json:"gid"`
	Users []int `json:"users"`
}

type MenuItem struct {
	Name     string
	Link     string
	Newtab   bool
	Subitems []MenuItem
}

type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type Secrets struct {
	ID         uint   `gorm:"primary_key" json:"id"`
	Name       string `gorm:"unique;not null" json:"name"`
	Secret     string `gorm:"not null" json:"secret"`
	KeyHash    string `gorm:"not null" json:"keyhash"`
}