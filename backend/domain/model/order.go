package model

import "time"

type Order struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	OrderID     string    `gorm:"uniqueIndex;not null"`
	UserID      uint      `gorm:"not null;index"`
	Amount      int64     `gorm:"not null"`
	Purpose     string    `gorm:"not null;default:'GENERAL'"`
	Status      string    `gorm:"not null;default:'PENDING'"`
	PaymentType string    `gorm:"default:''"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
