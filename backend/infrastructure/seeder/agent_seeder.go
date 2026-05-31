package seeder

import (
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	rolemodel "github.com/sunibgram/domain/model/role"
	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
	usersettingsmodel "github.com/sunibgram/domain/model/user_settings"
)

// SeedAgent creates the Fox/Hermes automation agent as a dedicated admin user,
// so its actions carry their own identity in the audit log (separate from the
// human admin seeded by SeedAdmin). Reads AGENT_EMAIL / AGENT_PASSWORD /
// AGENT_NAME / AGENT_USERNAME; skips silently if unset.
func SeedAgent(db *gorm.DB) {
	var adminRole rolemodel.Role
	if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		log.Println("admin role not found, skipping agent seeder")
		return
	}

	email := strings.TrimSpace(os.Getenv("AGENT_EMAIL"))
	password := strings.TrimSpace(os.Getenv("AGENT_PASSWORD"))
	name := strings.TrimSpace(os.Getenv("AGENT_NAME"))
	username := strings.TrimSpace(os.Getenv("AGENT_USERNAME"))

	if email == "" || password == "" || name == "" || username == "" {
		log.Println("agent seeder: AGENT_* env not set, skipping")
		return
	}

	var existing usermodel.User
	if err := db.Where("email = ?", email).First(&existing).Error; err == nil {
		// Ensure role is admin even if the account predates this seeder.
		if existing.RoleID != adminRole.RoleID {
			db.Model(&existing).Update("role_id", adminRole.RoleID)
		}
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
		log.Printf("agent seeder: failed to hash password: %v", err)
		return
	}

	agent := usermodel.User{
		Name:      name,
		UserName:  username,
		Email:     email,
		Password:  string(hashed),
		RoleID:    adminRole.RoleID,
		IsPremium: true,
	}

	if err := db.Create(&agent).Error; err != nil {
		log.Printf("agent seeder: failed to create agent user: %v", err)
		return
	}

	profile := userprofilemodel.UserProfile{
		UserID:       agent.UserID,
		Gender:       userprofilemodel.GenderPreferNotToSay,
		BirthdayDate: defaultSeedBirthdayDate(),
	}
	if err := db.Create(&profile).Error; err != nil {
		log.Printf("agent seeder: failed to create agent profile: %v", err)
	}

	settings := usersettingsmodel.UserSettings{UserID: agent.UserID}
	if err := db.Create(&settings).Error; err != nil {
		log.Printf("agent seeder: failed to create agent settings: %v", err)
	}

	log.Println("agent seeder: agent user created")
}
