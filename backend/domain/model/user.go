package model

import (
	"errors"
	"strings"
	"time"
)

const allowedEmailDomain = "@binus.ac.id"

type TokenPair struct {
	AccessToken          string
	SID                  string
	AccessTokenExpiresAt time.Time
}

type User struct {
	UserID      uint         `gorm:"primaryKey;autoIncrement"`
	Name        string       `gorm:"not null"`
	UserName    string       `gorm:"uniqueIndex;not null"`
	Email       string       `gorm:"uniqueIndex;not null"`
	Password    string       `gorm:"not null"`
	RoleID      uint         `gorm:"not null;default:1"`
	IsPremium   bool         `gorm:"not null;default:false"`
	Role        Role         `gorm:"foreignKey:RoleID;references:RoleID"`
	UserFaculty *UserFaculty `gorm:"-"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
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
	if len(u.Password) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	return nil
}
