package seeder

import (
	"log"

	"gorm.io/gorm"

	facultymodel "github.com/sunibgram/domain/model/faculty"
)

var facultyNames = []string{
	"School of Computer Science",
	"School of Information Systems",
	"BINUS Business School",
	"School of Business Management",
	"Faculty of Engineering",
	"School of Design",
	"Faculty of Humanities",
	"School of Accounting",
	"Faculty of Digital Communication and Hotel & Tourism",
}

func SeedFaculty(db *gorm.DB) {
	for _, name := range facultyNames {
		faculty := facultymodel.Faculty{FacultyName: name}
		if err := db.Where(facultymodel.Faculty{FacultyName: name}).FirstOrCreate(&faculty).Error; err != nil {
			log.Printf("faculty seeder: failed to seed faculty %s: %v", name, err)
			continue
		}
	}
}
