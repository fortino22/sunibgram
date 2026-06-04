package seeder

import (
	"log"

	"gorm.io/gorm"

	facultylocationmodel "github.com/sunibgram/domain/model/faculty_location"
)

// facultyLocations is the campus-group reference list (8 campuses + the
// "All Regions" sentinel), grouped by area/region. SortOrder drives selector
// order; the sentinel sorts first (-1) in the audience picker.
var facultyLocations = []facultylocationmodel.FacultyLocation{
	{Name: "All Regions", Address: "", Region: "All", IsAllRegions: true, SortOrder: -1},

	// Area Jakarta & Sekitarnya (Jabodetabek)
	{Name: "BINUS Kemanggisan", Address: "Kebon Jeruk, Jakarta Barat", Region: "Jabodetabek", SortOrder: 1},
	{Name: "BINUS Senayan", Address: "Senayan, Jakarta Selatan", Region: "Jabodetabek", SortOrder: 2},
	{Name: "BINUS Alam Sutera", Address: "Alam Sutera, Kota Tangerang", Region: "Jabodetabek", SortOrder: 3},
	{Name: "BINUS Bekasi", Address: "Summarecon Bekasi, Bekasi Utara", Region: "Jabodetabek", SortOrder: 4},

	// Area Jawa & Luar Jabodetabek
	{Name: "BINUS Bandung", Address: "Pasirkaliki, Kota Bandung", Region: "Jawa & Luar Jabodetabek", SortOrder: 5},
	{Name: "BINUS Semarang", Address: "Semarang Barat, Kota Semarang", Region: "Jawa & Luar Jabodetabek", SortOrder: 6},
	{Name: "BINUS Malang", Address: "Araya, Kabupaten Malang", Region: "Jawa & Luar Jabodetabek", SortOrder: 7},

	// Area Sumatra (Medan)
	{Name: "BINUS Medan", Address: "Kesawan, Kota Medan", Region: "Sumatra (Medan)", SortOrder: 8},
}

// SeedFacultyLocation upserts the campus reference list by Name (idempotent).
func SeedFacultyLocation(db *gorm.DB) {
	for _, loc := range facultyLocations {
		row := loc
		if err := db.Where(facultylocationmodel.FacultyLocation{Name: loc.Name}).
			Assign(facultylocationmodel.FacultyLocation{
				Address:      loc.Address,
				Region:       loc.Region,
				IsAllRegions: loc.IsAllRegions,
				SortOrder:    loc.SortOrder,
			}).
			FirstOrCreate(&row).Error; err != nil {
			log.Printf("faculty_location seeder: failed to seed %s: %v", loc.Name, err)
			continue
		}
	}
}
