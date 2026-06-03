package profilefieldchangemodel

import "time"

// Field identifies a rate-limited profile field tracked in the change log.
type Field string

const (
	FieldName     Field = "name"
	FieldUsername Field = "username"
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

// ProfileFieldChangeCooldown is the minimum interval between changes of a single
// rate-limited profile field. Each field is tracked independently.
const ProfileFieldChangeCooldown = 7 * 24 * time.Hour

// NextChangeAt returns the earliest time a field may change again given its last
// change time. A zero last value (never changed) returns zero — changeable now.
func NextChangeAt(last time.Time) time.Time {
	if last.IsZero() {
		return time.Time{}
	}
	return last.Add(ProfileFieldChangeCooldown)
}
