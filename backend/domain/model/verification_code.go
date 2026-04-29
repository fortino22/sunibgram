package model

import (
	"errors"
	"strings"
	"time"
)

const allowedVerificationDomain = "@binus.ac.id"

type VerificationType string

const (
	VerificationTypeRegister       VerificationType = "REGISTER"
	VerificationTypeForgotPassword VerificationType = "FORGOT_PASSWORD"
)

type VerificationCode struct {
	ID        uint             `gorm:"primaryKey;autoIncrement"`
	Email     string           `gorm:"not null;index"`
	Code      string           `gorm:"not null"`
	Type      VerificationType `gorm:"not null;type:varchar(32)"`
	ExpiresAt time.Time        `gorm:"not null"`
	Used      bool             `gorm:"not null;default:false"`
	CreatedAt time.Time
}

func ValidateVerificationEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return errors.New("email is required")
	}
	if !strings.HasSuffix(strings.ToLower(email), allowedVerificationDomain) {
		return errors.New("email must use @binus.ac.id domain")
	}
	return nil
}

func ValidateVerificationType(t VerificationType) error {
	switch t {
	case VerificationTypeRegister, VerificationTypeForgotPassword:
		return nil
	default:
		return errors.New("type must be REGISTER or FORGOT_PASSWORD")
	}
}
