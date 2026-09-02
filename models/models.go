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

// PasswordChange is the payload of the self-service password change. The
// current password is mandatory: it proves the session belongs to the account.
type PasswordChange struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

// UserPreferences is what a user is allowed to change on their own account.
// Name, type and password are deliberately excluded.
type UserPreferences struct {
	Lang  string `json:"lang"`
	Theme string `json:"theme"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type Secrets struct {
	ID      uint   `gorm:"primary_key" json:"id"`
	Name    string `gorm:"unique;not null" json:"name"`
	Secret  string `gorm:"not null" json:"secret"`
	KeyHash string `gorm:"not null" json:"keyhash"`
}

// Connectors est la configuration d'un connecteur externe, une ligne par type
// (le connecteur Git est unique : une synchronisation automatique vers deux
// dépôts n'aurait pas de sens). Config est un JSON public ; Credentials un JSON
// chiffré avec la clé de l'instance (token, passphrase, secret de webhook), qui
// ne quitte jamais le serveur.
type Connectors struct {
	ID          uint   `gorm:"primary_key" json:"id"`
	Type        string `gorm:"unique;not null" json:"type"`
	Enabled     bool   `gorm:"default:false" json:"enabled"`
	Config      string `json:"config"`
	Credentials string `json:"-"`
	KeyHash     string `json:"-"`
}

// Sync_states mémorise, par entité, l'empreinte du contenu tel qu'il était à la
// dernière synchronisation réussie : c'est la « base » de la comparaison à
// trois (local, distant, base) qui distingue une modification d'un conflit.
type Sync_states struct {
	ID        uint   `gorm:"primary_key"`
	Connector string `gorm:"index:idx_sync_state,unique;not null"`
	Kind      string `gorm:"index:idx_sync_state,unique;not null"`
	EntityID  uint   `gorm:"index:idx_sync_state,unique;not null"`
	Hash      string `gorm:"not null"`
}
