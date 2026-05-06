package migration

import (
	"log"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/infrastructure/seeder"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) {
	// Disable FK constraint creation during migration to avoid GORM reverse-FK bugs.
	db.DisableForeignKeyConstraintWhenMigrating = true

	models := []any{
		&model.Role{},
		&model.User{},
		&model.UserProfile{},
		&model.Community{},
		&model.Thread{},
		&model.ThreadMedia{},
		&model.Like{},
		&model.Bookmark{},
		&model.ThreadPin{},
		&model.VerificationCode{},
		&model.UserRelationship{},
		&model.UserStat{},
		&model.CommunityMember{},
		&model.CommunityRequest{},
		&model.Faculty{},
		&model.UserFaculty{},
		&model.Order{},
		&model.ChatRoom{},
		&model.ChatRoomMember{},
		&model.Chat{},
		&model.ChatMedia{},
	}

	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			log.Fatalf("AutoMigrate %T: %v", m, err)
		}
	}

	seedRoles(db)
	seeder.SeedAdmin(db)
	seeder.SeedDummy(db)
}

func seedRoles(db *gorm.DB) {
	defaultRoles := []string{"user", "lecturer", "admin"}
	for _, name := range defaultRoles {
		db.Where(model.Role{Name: name}).FirstOrCreate(&model.Role{}, model.Role{Name: name})
	}
}
