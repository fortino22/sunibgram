package seeder

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/sunibgram/domain/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedAdmin(db *gorm.DB) {
	var adminRole model.Role
	if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		log.Println("admin role not found, skipping admin seeder")
		return
	}

	email := strings.TrimSpace(os.Getenv("ADMIN_EMAIL"))
	password := strings.TrimSpace(os.Getenv("ADMIN_PASSWORD"))
	name := strings.TrimSpace(os.Getenv("ADMIN_NAME"))
	username := strings.TrimSpace(os.Getenv("ADMIN_USERNAME"))

	var existing model.User
	if err := db.Where("email = ?", email).First(&existing).Error; err == nil {
		var existingProfile model.UserProfile
		if err := db.Where("user_id = ?", existing.UserID).First(&existingProfile).Error; err != nil {
			profile := model.UserProfile{
				UserID:       existing.UserID,
				Gender:       model.GenderPreferNotToSay,
				BirthdayDate: defaultSeedBirthdayDate(),
			}
			db.Create(&profile)
		}
		var existingSettings model.UserSettings
		if err := db.Where("user_id = ?", existing.UserID).First(&existingSettings).Error; err != nil {
			settings := model.UserSettings{UserID: existing.UserID}
			db.Create(&settings)
		}
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("admin seeder: failed to hash password: %v", err)
		return
	}

	admin := model.User{
		Name:      name,
		UserName:  username,
		Email:     email,
		Password:  string(hashed),
		RoleID:    adminRole.RoleID,
		IsPremium: true,
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Printf("admin seeder: failed to create admin user: %v", err)
		return
	}

	profile := model.UserProfile{
		UserID:       admin.UserID,
		Gender:       model.GenderPreferNotToSay,
		BirthdayDate: defaultSeedBirthdayDate(),
	}
	if err := db.Create(&profile).Error; err != nil {
		log.Printf("admin seeder: failed to create admin profile: %v", err)
	}

	settings := model.UserSettings{UserID: admin.UserID}
	if err := db.Create(&settings).Error; err != nil {
		log.Printf("admin seeder: failed to create admin settings: %v", err)
	}

	log.Println("admin seeder: admin user created")
}

func defaultSeedBirthdayDate() time.Time {
	return time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
}
