package model

type Faculty struct {
	FacultyID   uint   `gorm:"primaryKey;autoIncrement"`
	FacultyName string `gorm:"type:varchar(100);not null;uniqueIndex"`
	TotalUsers  int    `gorm:"not null;default:0"`
}
