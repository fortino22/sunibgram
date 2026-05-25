package seeder

import (
	"log"

	"gorm.io/gorm"

	facultymodel "github.com/sunibgram/domain/model/faculty"
)

var facultyNames = []string{
	"BINUS ASO School of Engineering",
	"BINUS Business School",
	"BINUS GRADUATE PROGRAM",
	"BINUS ONLINE",
	"Doctor of Computer Science",
	"Faculty of Digital Communication and Hotel and Tourism",
	"Faculty of Digital Communication and Hotel and Tourism",
	"Faculty of Engineering",
	"Faculty of Humanities",
	"School of Accounting",
	"School of Computer Science",
	"School of Computing and Creative Arts",
	"School of Design",
	"School of Information Systems",
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
