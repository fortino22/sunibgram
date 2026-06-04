package facultylocationrepoiface

import facultylocationmodel "github.com/sunibgram/domain/model/faculty_location"

type IFacultyLocationRepository interface {
	// FindAll returns campuses ordered by region/sort_order/name. When includeAll
	// is false the "All Regions" sentinel row is excluded (register/profile use);
	// when true it is included first (event-audience use).
	FindAll(includeAll bool) ([]facultylocationmodel.FacultyLocation, error)
	// FindByIDs returns the campuses for the given ids (audience hydration).
	FindByIDs(ids []uint) ([]facultylocationmodel.FacultyLocation, error)
	// ExistsSelectable reports whether id is a real campus a user may pick
	// (excludes the "All Regions" sentinel).
	ExistsSelectable(id uint) (bool, error)
	// Exists reports whether id is any faculty_location (incl. "All Regions").
	Exists(id uint) (bool, error)
}
