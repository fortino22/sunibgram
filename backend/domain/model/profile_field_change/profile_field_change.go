package profilefieldchangemodel

import "time"

// Field identifies a rate-limited profile field tracked in the change log.
type Field string

const (
	FieldName            Field = "name"
	FieldUsername        Field = "username"
	FieldFacultyLocation Field = "faculty_location"
)

// ProfileFieldChange is one append-only record of a rate-limited profile field
// being changed. The latest ChangedAt per (UserID, Field) drives the per-field
// change cooldown. Keeps these timestamps off the users table and gives an
// audit trail.
type ProfileFieldChange struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	UserID    uint      `gorm:"not null;index:idx_pfc_user_field,priority:1"`
	Field     Field     `gorm:"type:varchar(20);not null;index:idx_pfc_user_field,priority:2"`
	ChangedAt time.Time `gorm:"not null"`
	// OldValue/NewValue capture the field's value before and after the change so
	// the "about this account" history can show "@old → @new". Empty for rows
	// recorded before value tracking existed.
	OldValue string `gorm:"type:varchar(50);default:''"`
	NewValue string `gorm:"type:varchar(50);default:''"`
}

func (ProfileFieldChange) TableName() string { return "profile_field_changes" }

// ProfileFieldChangeCooldown is the default minimum interval between changes of
// a rate-limited profile field (name/username). Each field is tracked
// independently.
const ProfileFieldChangeCooldown = 7 * 24 * time.Hour

// FacultyLocationChangeCooldown is the (longer) interval for changing the home
// campus — roughly 6 months.
const FacultyLocationChangeCooldown = 180 * 24 * time.Hour

// cooldownFor returns the per-field cooldown interval.
func cooldownFor(field Field) time.Duration {
	if field == FieldFacultyLocation {
		return FacultyLocationChangeCooldown
	}
	return ProfileFieldChangeCooldown
}

// NextChangeAt returns the earliest time a field may change again given its last
// change time. A zero last value (never changed) returns zero — changeable now.
func NextChangeAt(field Field, last time.Time) time.Time {
	if last.IsZero() {
		return time.Time{}
	}
	return last.Add(cooldownFor(field))
}
