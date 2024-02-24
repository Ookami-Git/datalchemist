package database

type parameters struct {
	Name  string `gorm:"primary_key"`
	Value string
}

type users struct {
	ID         uint   `gorm:"primary_key"`
	Name       string `gorm:"unique;not null"`
	Type       string `gorm:"not null"`
	Parameters string
}

type groups struct {
	ID          uint   `gorm:"primary_key"`
	Name        string `gorm:"unique;not null"`
	Description string
}

type roles struct {
	ID     uint `gorm:"primary_key"`
	GID    uint `gorm:"index"`
	UserID uint `gorm:"index"`

	Group groups `gorm:"foreignKey:GID"`
	User  users  `gorm:"foreignKey:UserID"`
}

type acl_users struct {
	ID     uint `gorm:"primary_key"`
	ViewID uint `gorm:"index"`
	UserID uint `gorm:"index"`

	View views `gorm:"foreignKey:ViewID"`
	User users `gorm:"foreignKey:UserID"`
}

type acl_groups struct {
	ID     uint `gorm:"primary_key"`
	ViewID uint `gorm:"index"`
	GID    uint `gorm:"index"`

	View  views  `gorm:"foreignKey:ViewID"`
	Group groups `gorm:"foreignKey:GID"`
}

type sources struct {
	ID         uint   `gorm:"primary_key"`
	Name       string `gorm:"unique;not null"`
	Parameters string
	JSON       string
}

type source_require struct {
	ID        uint `gorm:"primary_key"`
	SourceID  uint `gorm:"index"`
	RequireID uint `gorm:"index"`

	Source  sources `gorm:"foreignKey:SourceID"`
	Require sources `gorm:"foreignKey:RequireID"`
}

type items struct {
	ID         uint   `gorm:"primary_key"`
	Name       string `gorm:"unique;not null"`
	Parameters string
	Template   string
}

type item_sources struct {
	ID       uint `gorm:"primary_key"`
	ItemID   uint `gorm:"index"`
	SourceID uint `gorm:"index"`

	Item   items   `gorm:"foreignKey:ItemID"`
	Source sources `gorm:"foreignKey:SourceID"`
}

type views struct {
	ID         uint   `gorm:"primary_key"`
	Name       string `gorm:"unique;not null"`
	Parameters string
}

type view_items struct {
	ID         uint `gorm:"primary_key"`
	ViewID     uint `gorm:"index"`
	ItemID     uint `gorm:"index"`
	Parameters string

	View views `gorm:"foreignKey:ViewID"`
	Item items `gorm:"foreignKey:ItemID"`
}
