package facultylocationsvciface

import facultylocationmodel "github.com/sunibgram/domain/model/faculty_location"

type IFacultyLocationService interface {
	// List returns campuses for selectors. includeAll=false (register/profile)
	// drops the "All Regions" sentinel; includeAll=true (event audience) keeps it.
	List(includeAll bool) ([]facultylocationmodel.FacultyLocation, error)
}
