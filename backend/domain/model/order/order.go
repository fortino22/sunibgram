package ordermodel

import "time"

type Order struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	OrderID     string    `gorm:"uniqueIndex;not null"`
	UserID      uint      `gorm:"not null;index;index:idx_order_status,priority:2;index:idx_order_user,priority:1"`
	Amount      int64     `gorm:"not null"`
	Purpose     string    `gorm:"not null;default:'GENERAL'"`
	Status      string    `gorm:"not null;default:'PENDING';index:idx_order_status,priority:1"`
	PaymentType string    `gorm:"default:''"`
	CreatedAt   time.Time `gorm:"index:idx_order_user,priority:2,sort:desc"`
	UpdatedAt   time.Time
}
