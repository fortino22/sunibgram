package seeder

import (
	"log"

	"gorm.io/gorm"

	categorymodel "github.com/sunibgram/domain/model/category"
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
		category := categorymodel.Category{Name: name}
		if err := db.Where(categorymodel.Category{Name: name}).FirstOrCreate(&category).Error; err != nil {
			log.Printf("category seeder: failed to seed category %s: %v", name, err)
			continue
		}
	}
}
