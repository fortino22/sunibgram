package usersettingsmodel

import (
	"errors"
	"strings"

	usermodel "github.com/sunibgram/domain/model/user"
)

type UserTheme string

const (
	UserThemeBlack UserTheme = "black"
	UserThemeWhite UserTheme = "white"
)

type UserSettings struct {
	UserID    uint           `gorm:"primaryKey;autoIncrement:false"`
	User      usermodel.User `gorm:"foreignKey:UserID;references:UserID;constraint:OnUpdate:NO ACTION,OnDelete:CASCADE"`
	IsPrivate bool           `gorm:"not null;default:false"`
	UserTheme UserTheme      `gorm:"type:varchar(20);not null;default:'black'"`
}

func (UserSettings) TableName() string {
	return "user_settings"
}

// ParseUserTheme normalizes raw input (lowercase, trimmed) to a UserTheme.
func ParseUserTheme(raw string) UserTheme {
	return UserTheme(strings.ToLower(strings.TrimSpace(raw)))
}

// ValidateUserTheme reports whether t is an allowed theme value.
func ValidateUserTheme(t UserTheme) error {
	switch t {
	case UserThemeBlack, UserThemeWhite:
		return nil
	default:
		return errors.New("user_theme must be one of: black, white")
	}
}
