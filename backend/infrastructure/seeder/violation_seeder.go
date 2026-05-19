package seeder

import (
	"log"

	"github.com/sunibgram/domain/model"
	"gorm.io/gorm"
)

var violationSeeds = []model.Violation{
	{Name: "Spam", Description: "Repetitive, misleading, or unwanted promotional content."},
	{Name: "Harassment", Description: "Bullying, intimidation, threats, or targeted abusive behavior."},
	{Name: "Hate Speech", Description: "Content attacking people based on protected characteristics."},
	{Name: "Misinformation", Description: "False or misleading information that may cause harm."},
	{Name: "Nudity/Sexual Content", Description: "Explicit sexual content, nudity, or sexually suggestive material."},
	{Name: "Violence", Description: "Graphic violence, threats, or content encouraging physical harm."},
	{Name: "Copyright Infringement", Description: "Content that appears to violate intellectual property rights."},
	{Name: "Other", Description: "Other policy violations not covered by the listed categories."},
}

func SeedViolation(db *gorm.DB) {
	for _, seed := range violationSeeds {
		violation := seed
		if err := db.Where(model.Violation{Name: seed.Name}).FirstOrCreate(&violation).Error; err != nil {
			log.Printf("violation seeder: failed to seed violation %s: %v", seed.Name, err)
			continue
		}
	}
}
