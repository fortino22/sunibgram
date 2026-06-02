package usermodel

import (
	"errors"
	"strings"
	"time"

	rolemodel "github.com/sunibgram/domain/model/role"
)

const allowedEmailDomain = "@binus.ac.id"

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
	if strings.TrimSpace(u.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(u.UserName) == "" {
		return errors.New("username is required")
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
