package userprofilemodel

import (
	"errors"
	"strings"
	"time"

	facultylocationmodel "github.com/sunibgram/domain/model/faculty_location"
	usermodel "github.com/sunibgram/domain/model/user"
)

type Gender string

const (
	GenderMale           Gender = "male"
	GenderFemale         Gender = "female"
	GenderOther          Gender = "other"
	GenderPreferNotToSay Gender = "prefer_not_to_say"
)

type UserProfile struct {
	ProfileID        uint           `gorm:"primaryKey;autoIncrement"`
	UserID           uint           `gorm:"not null;uniqueIndex"`
	User             usermodel.User `gorm:"foreignKey:UserID;references:UserID"`
	Gender           Gender         `gorm:"type:varchar(30)"`
	BirthdayDate     time.Time      `gorm:"type:date"`
	AvatarObjectName string         `gorm:"default:''"`
	BannerObjectName string         `gorm:"default:''"`
	Bio              string         `gorm:"type:text;default:''"`
	// FacultyLocationID is the user's home campus. Set at registration, changeable
	// once per 6 months (cooldown enforced via the profile_field_changes log).
	FacultyLocationID *uint                                 `gorm:"index"`
	FacultyLocation   *facultylocationmodel.FacultyLocation `gorm:"foreignKey:FacultyLocationID;references:ID"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func ParseGender(raw string) Gender {
	return Gender(strings.ToLower(strings.TrimSpace(raw)))
}

func ValidateGender(g Gender) error {
	switch g {
	case GenderMale, GenderFemale, GenderOther, GenderPreferNotToSay:
		return nil
	default:
		return errors.New("gender must be one of: male, female, other, prefer_not_to_say")
	}
}

// ValidateBirthdayDate validates a birthday when one is provided. Birthday is
// optional (zero time = not provided), so a zero value passes; only a malformed
// future date is rejected.
func ValidateBirthdayDate(t time.Time) error {
	if t.IsZero() {
		return nil
	}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	birthday := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	if birthday.After(today) {
		return errors.New("birthday_date cannot be in the future")
	}
	minBirthday := today.AddDate(-17, 0, 0)
	if birthday.After(minBirthday) {
		return errors.New("you must be at least 17 years old")
	}
	return nil
}

func (p *UserProfile) ValidateRequiredFields() error {
	if err := ValidateGender(p.Gender); err != nil {
		return err
	}
	return ValidateBirthdayDate(p.BirthdayDate)
}
