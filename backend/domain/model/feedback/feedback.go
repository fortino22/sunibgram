package feedbackmodel

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"gorm.io/gorm"

	usermodel "github.com/sunibgram/domain/model/user"
)

const (
	TypeBug            = "bug"
	TypeFeatureRequest = "feature_request"

	StatusOpen       = "open"
	StatusPlanned    = "planned"
	StatusInProgress = "in_progress"
	StatusResolved   = "resolved"
	StatusRejected   = "rejected"
)

type Feedback struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	UserID    uint           `gorm:"not null;index"`
	User      usermodel.User `gorm:"foreignKey:UserID;references:UserID;constraint:OnDelete:CASCADE"`
	Type      string         `gorm:"type:varchar(20);not null;index"`
	Title     string         `gorm:"type:varchar(200);not null"`
	Body      string         `gorm:"type:text;not null"`
	Status    string         `gorm:"type:varchar(20);not null;default:'open';index"`
	CreatedAt time.Time      `gorm:"type:timestamptz"`
	UpdatedAt time.Time      `gorm:"type:timestamptz"`
}

func (f *Feedback) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	if f.Status == "" {
		f.Status = StatusOpen
	}
	return nil
}

type VoteCounts struct {
	Likes    int64 `json:"likes"`
	Dislikes int64 `json:"dislikes"`
}

func ValidateType(t string) error {
	switch t {
	case TypeBug, TypeFeatureRequest:
		return nil
	}
	return errors.New("type must be bug or feature_request")
}

func ValidateStatus(s string) error {
	switch s {
	case StatusOpen, StatusPlanned, StatusInProgress, StatusResolved, StatusRejected:
		return nil
	}
	return errors.New("status must be open, planned, in_progress, resolved, or rejected")
}

func ValidateTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", errors.New("title is required")
	}
	if utf8.RuneCountInString(title) > 200 {
		return "", errors.New("title must be 200 characters or less")
	}
	return title, nil
}

func ValidateBody(body string) (string, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return "", errors.New("body is required")
	}
	if utf8.RuneCountInString(body) > 5000 {
		return "", errors.New("body must be 5000 characters or less")
	}
	return body, nil
}

func ValidateVoteValue(v int) error {
	if v != 1 && v != -1 {
		return errors.New("vote value must be 1 (like) or -1 (dislike)")
	}
	return nil
}
