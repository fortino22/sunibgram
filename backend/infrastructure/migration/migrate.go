package migration

import (
	"github.com/sunibgram/domain/model"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(&model.User{})
}
