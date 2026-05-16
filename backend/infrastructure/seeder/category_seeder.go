package seeder

import (
	"log"

	"github.com/sunibgram/domain/model"
	"gorm.io/gorm"
)

var categoryNames = []string{
	"Academic",
	"Arts",
	"Career",
	"Gaming",
	"Social",
	"Sports",
	"Student Life",
	"Technology",
	"Volunteering",
}

func SeedCategory(db *gorm.DB) {
	for _, name := range categoryNames {
		category := model.Category{Name: name}
		if err := db.Where(model.Category{Name: name}).FirstOrCreate(&category).Error; err != nil {
			log.Printf("category seeder: failed to seed category %s: %v", name, err)
			continue
		}
	}
}
