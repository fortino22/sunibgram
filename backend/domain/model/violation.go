package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Violation struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        string    `gorm:"type:varchar(100);not null;uniqueIndex"`
	Description string    `gorm:"type:text;not null"`
	CreatedAt   time.Time `gorm:"type:timestamptz"`
	UpdatedAt   time.Time `gorm:"type:timestamptz"`
}

func (v *Violation) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}
