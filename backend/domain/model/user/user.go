package usermodel

import (
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	rolemodel "github.com/sunibgram/domain/model/role"
)

const allowedEmailDomain = "@binus.ac.id"

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// ValidateName mirrors frontend: trimmed, min 2 chars.
func ValidateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("name is required")
	}
	if utf8.RuneCountInString(name) < 2 {
		return errors.New("name must be at least 2 characters")
	}
	return nil
}

// ValidateUsername mirrors frontend: trimmed, 3-30 chars, [a-zA-Z0-9_] only.
func ValidateUsername(username string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return errors.New("username is required")
	}
	n := utf8.RuneCountInString(username)
	if n < 3 {
		return errors.New("username must be at least 3 characters")
	}
	if n > 30 {
		return errors.New("username must be 30 characters or less")
	}
	if !usernameRegex.MatchString(username) {
		return errors.New("username may only contain letters, numbers, and underscores")
	}
	return nil
}

type TokenPair struct {
	AccessToken          string
	SID                  string
	AccessTokenExpiresAt time.Time
}

type User struct {
	UserID   uint   `gorm:"primaryKey;autoIncrement"`
	Name     string `gorm:"not null"`
	UserName string `gorm:"uniqueIndex;not null"`
	// Email is included in (de)serialization so the Redis cache round-trip for
	// FindByID preserves it. The User struct is never returned raw to clients —
	// email exposure is gated by DTOs (e.g. UserProfileResponse only sets it on
	// self-view). Password stays json:"-" — it must never serialize anywhere.
	Email     string `gorm:"uniqueIndex;not null" json:"email"`
	Password  string `gorm:"not null" json:"-"`
	RoleID    uint   `gorm:"not null;default:1"`
	IsPremium bool   `gorm:"not null;default:false"`
	// OnboardingCompleted gates the first-time follow-5-people flow. Permanent
	// once true (never reset on unfollow). Serialized so the FindByID Redis cache
	// round-trip preserves it; exposed to clients only on self-view DTOs.
	OnboardingCompleted bool           `gorm:"not null;default:false" json:"onboarding_completed"`
	// TermsAccepted records the user agreed to the Terms and Privacy Policy at
	// registration. Required true to register; serialized so the FindByID Redis
	// cache round-trip preserves it.
	TermsAccepted bool           `gorm:"not null;default:false" json:"terms_accepted"`
	Role          rolemodel.Role `gorm:"foreignKey:RoleID;references:RoleID"`
	UserFaculty         *UserFaculty   `gorm:"foreignKey:UserID;references:UserID;<-:false"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (u *User) ValidateRegister() error {
	if err := ValidateName(u.Name); err != nil {
		return err
	}
	if err := ValidateUsername(u.UserName); err != nil {
		return err
	}
	if strings.TrimSpace(u.Email) == "" {
		return errors.New("email is required")
	}
	if !strings.HasSuffix(strings.ToLower(u.Email), allowedEmailDomain) {
		return errors.New("email must use @binus.ac.id domain")
	}
	if err := ValidatePassword(u.Password); err != nil {
		return err
	}
	if !u.TermsAccepted {
		return errors.New("you must accept the terms and conditions")
	}
	return nil
}

// ValidatePassword enforces minimum length and character-class diversity.
// Shared by Register, ChangePassword, ForgotPassword so the policy can only
// drift in one place. Allows pre-existing hashes in DB to remain valid for
// login (no length check on login path); the policy only fires on mutation.
func ValidatePassword(pw string) error {
	if len(pw) < 10 {
		return errors.New("password must be at least 10 characters")
	}
	var hasUpper, hasLower, hasDigit bool
	for _, r := range pw {
		switch {
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= '0' && r <= '9':
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return errors.New("password must contain uppercase, lowercase, and digit")
	}
	return nil
}
