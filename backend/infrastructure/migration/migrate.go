package migration

import (
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/infrastructure/seeder"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) {
	dropBadConstraint(db)

	db.AutoMigrate(&model.Role{})
	db.AutoMigrate(&model.User{})
	db.AutoMigrate(&model.UserAvatar{})

	seedRoles(db)
	seeder.SeedAdmin(db)
}

func dropBadConstraint(db *gorm.DB) {
	// Drop the incorrectly created FK that references users from roles table.
	// This only runs if the constraint exists; errors are ignored safely.
	db.Exec(`ALTER TABLE roles DROP CONSTRAINT IF EXISTS fk_users_role`)
	db.Exec(`ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_role`)
}

func seedRoles(db *gorm.DB) {
	defaultRoles := []string{"user", "lecturer", "admin"}
	for _, name := range defaultRoles {
		db.Where(model.Role{Name: name}).FirstOrCreate(&model.Role{}, model.Role{Name: name})
	}
}
