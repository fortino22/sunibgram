package seeder

import (
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	rolemodel "github.com/sunibgram/domain/model/role"
	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
	usersettingsmodel "github.com/sunibgram/domain/model/user_settings"
)

func SeedAdmin(db *gorm.DB) {
	var adminRole rolemodel.Role
	if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		log.Println("admin role not found, skipping admin seeder")
		return
	}

	email := strings.TrimSpace(os.Getenv("ADMIN_EMAIL"))
	password := strings.TrimSpace(os.Getenv("ADMIN_PASSWORD"))
	name := strings.TrimSpace(os.Getenv("ADMIN_NAME"))
	username := strings.TrimSpace(os.Getenv("ADMIN_USERNAME"))

	var existing usermodel.User
	if err := db.Where("email = ?", email).First(&existing).Error; err == nil {
		var existingProfile userprofilemodel.UserProfile
		if err := db.Where("user_id = ?", existing.UserID).First(&existingProfile).Error; err != nil {
			profile := userprofilemodel.UserProfile{
				UserID:       existing.UserID,
				Gender:       userprofilemodel.GenderPreferNotToSay,
				BirthdayDate: defaultSeedBirthdayDate(),
			}
			db.Create(&profile)
		}
		var existingSettings usersettingsmodel.UserSettings
		if err := db.Where("user_id = ?", existing.UserID).First(&existingSettings).Error; err != nil {
			settings := usersettingsmodel.UserSettings{UserID: existing.UserID}
			db.Create(&settings)
		}
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("admin seeder: failed to hash password: %v", err)
		return
	}

	admin := usermodel.User{
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

	profile := userprofilemodel.UserProfile{
		UserID:       admin.UserID,
		Gender:       userprofilemodel.GenderPreferNotToSay,
		BirthdayDate: defaultSeedBirthdayDate(),
	}
	if err := db.Create(&profile).Error; err != nil {
		log.Printf("admin seeder: failed to create admin profile: %v", err)
	}

	settings := usersettingsmodel.UserSettings{UserID: admin.UserID}
	if err := db.Create(&settings).Error; err != nil {
		log.Printf("admin seeder: failed to create admin settings: %v", err)
	}

	log.Println("admin seeder: admin user created")
}

func defaultSeedBirthdayDate() time.Time {
	return time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
}
