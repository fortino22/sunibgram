package model

type UserFaculty struct {
	UserID    uint    `gorm:"primaryKey;autoIncrement:false"`
	FacultyID uint    `gorm:"primaryKey;autoIncrement:false"`
	User      User    `gorm:"foreignKey:UserID;references:UserID"`
	Faculty   Faculty `gorm:"foreignKey:FacultyID;references:FacultyID"`
}
