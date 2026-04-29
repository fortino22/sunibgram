package model

import (
	"errors"
	"strings"
	"time"
)

const allowedVerificationDomain = "@binus.ac.id"

type VerificationCode struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Email     string    `gorm:"not null;index"`
	Code      string    `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Used      bool      `gorm:"not null;default:false"`
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
