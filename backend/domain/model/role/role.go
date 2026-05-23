package rolemodel

import "time"

const DefaultRoleName = "user"

type Role struct {
	RoleID    uint   `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
