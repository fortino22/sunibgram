package usermodel

import (
	facultymodel "github.com/sunibgram/domain/model/faculty"
)

type UserFaculty struct {
	UserID    uint                 `gorm:"primaryKey;autoIncrement:false"`
	FacultyID uint                 `gorm:"primaryKey;autoIncrement:false"`
	User      User                 `gorm:"foreignKey:UserID;references:UserID"`
	Faculty   facultymodel.Faculty `gorm:"foreignKey:FacultyID;references:FacultyID"`
}
