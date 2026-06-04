package facultylocationmodel

// FacultyLocation is a BINUS campus (campus-group granularity), tagged with the
// broader area/region it sits in (Jabodetabek / Jawa & Luar Jabodetabek /
// Sumatra). A user belongs to exactly one; events target one or more.
//
// The "All Regions" row (IsAllRegions=true) is a sentinel that means "everyone":
// it is hidden from the register/profile selectors and only offered when picking
// an event audience.
type FacultyLocation struct {
	ID           uint   `gorm:"primaryKey;autoIncrement"`
	Name         string `gorm:"type:varchar(150);not null;uniqueIndex"`
	Address      string `gorm:"type:varchar(500);default:''"`
	Region       string `gorm:"type:varchar(100);not null;index"`
	IsAllRegions bool   `gorm:"not null;default:false"`
	SortOrder    int    `gorm:"not null;default:0"`
}

func (FacultyLocation) TableName() string { return "faculty_locations" }
