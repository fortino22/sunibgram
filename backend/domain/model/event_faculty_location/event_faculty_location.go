package eventfacultylocationmodel

// EventFacultyLocation is the M:N join between an event and its target campuses
// (region audience). Plain ID columns only — no embedded relations — so it can
// reference both threadmodel.Event and facultylocationmodel.FacultyLocation
// without creating an import cycle.
type EventFacultyLocation struct {
	EventID           string `gorm:"type:varchar(36);primaryKey;column:event_id"`
	FacultyLocationID uint   `gorm:"primaryKey;column:faculty_location_id;index"`
}

func (EventFacultyLocation) TableName() string { return "event_faculty_locations" }
