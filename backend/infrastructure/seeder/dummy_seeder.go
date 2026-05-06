package seeder

import (
	"log"
	"strings"

	"github.com/sunibgram/domain/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type dummyAccount struct {
	Name     string
	Username string
	Email    string
	Password string
	RoleName string
}

var dummyAccounts = []dummyAccount{
	{Name: "Octa", Username: "Octa.hi", Email: "octa@binus.ac.id", Password: "dummy123", RoleName: "admin"},
}

func SeedDummy(db *gorm.DB) {
	roleCache := make(map[string]model.Role)

	for _, acc := range dummyAccounts {
		acc.Username = strings.TrimSpace(acc.Username)
		acc.Email = strings.TrimSpace(acc.Email)

		var existing model.User
		if err := db.Where("email = ?", acc.Email).First(&existing).Error; err == nil {
			var existingProfile model.UserProfile
			if err := db.Where("user_id = ?", existing.UserID).First(&existingProfile).Error; err != nil {
				db.Create(&model.UserProfile{UserID: existing.UserID})
			}
			continue
		}

		role, ok := roleCache[acc.RoleName]
		if !ok {
			if err := db.Where("name = ?", acc.RoleName).First(&role).Error; err != nil {
				log.Printf("dummy seeder: role '%s' not found, skipping %s", acc.RoleName, acc.Email)
				continue
			}
			roleCache[acc.RoleName] = role
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(acc.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("dummy seeder: failed to hash password for %s: %v", acc.Email, err)
			continue
		}

		user := model.User{
			Name:     acc.Name,
			UserName: acc.Username,
			Email:    acc.Email,
			Password: string(hashed),
			RoleID:   role.RoleID,
		}

		if err := db.Create(&user).Error; err != nil {
			log.Printf("dummy seeder: failed to create user %s: %v", acc.Email, err)
			continue
		}

		if err := db.Create(&model.UserProfile{UserID: user.UserID}).Error; err != nil {
			log.Printf("dummy seeder: failed to create profile for %s: %v", acc.Email, err)
			continue
		}

		log.Printf("dummy seeder: created user %s (%s)", acc.Username, acc.Email)
	}
}
