// models.go
package models

import "github.com/golang-jwt/jwt/v5"

// ---- Database type
// Structure de donnée pour les utilisateurs
type User struct {
	ID         int         `json:"id" gorm:"primaryKey"`
	Name       string      `json:"name"`
	Parameters interface{} `json:"parameters"`
	Type       string      `json:"type"`
	IsAdmin    bool        `json:"is_admin"`
}

type Group struct {
	ID          int         `json:"id" gorm:"primaryKey"`
	Name        string      `json:"name"`
	Description interface{} `json:"description"`
}

type Source struct {
	ID         int         `json:"id" gorm:"primaryKey"`
	Name       string      `json:"name"`
	Parameters interface{} `json:"parameters"`
	JSON       interface{} `json:"json"`
}

type Sources struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type View struct {
	ID         int         `json:"id"`
	Name       string      `json:"name"`
	Parameters interface{} `json:"parameters"`
}

type Item struct {
	ID         int         `json:"id"`
	Name       string      `json:"name"`
	Parameters interface{} `json:"parameters"`
	Template   interface{} `json:"template"`
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
