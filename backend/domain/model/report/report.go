package reportmodel

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	threadmodel "github.com/sunibgram/domain/model/thread"
	usermodel "github.com/sunibgram/domain/model/user"
	violationmodel "github.com/sunibgram/domain/model/violation"
)

const (
	ReportStatusPending  = "pending"
	ReportStatusReviewed = "reviewed"
	ReportStatusResolved = "resolved"
	ReportStatusRejected = "rejected"
)

type Report struct {
	ID             uuid.UUID                `gorm:"type:uuid;primaryKey"`
	ReporterID     uint                     `gorm:"not null;index"`
	Reporter       usermodel.User           `gorm:"foreignKey:ReporterID;references:UserID;constraint:OnDelete:CASCADE"`
	ReportedUserID *uint                    `gorm:"index"`
	ReportedUser   *usermodel.User          `gorm:"foreignKey:ReportedUserID;references:UserID;constraint:OnDelete:SET NULL"`
	ReportedPostID *uint                    `gorm:"column:reported_post_id;index"`
	ReportedPost   *threadmodel.Thread      `gorm:"foreignKey:ReportedPostID;references:ThreadID;constraint:OnDelete:SET NULL"`
	ViolationID    uuid.UUID                `gorm:"type:uuid;not null;index"`
	Violation      violationmodel.Violation `gorm:"foreignKey:ViolationID;references:ID;constraint:OnDelete:RESTRICT"`
	Explanation    string                   `gorm:"type:text;not null"`
	Status         string                   `gorm:"type:varchar(20);not null;default:'pending';index"`
	CreatedAt      time.Time                `gorm:"type:timestamptz"`
	UpdatedAt      time.Time                `gorm:"type:timestamptz"`
}

type ReportUserInfo struct {
	UserID     uint
	Username   string
	UserAvatar string
	IsPremium  bool
	Faculty    *ReportUserFaculty
}

type ReportUserFaculty struct {
	FacultyID   uint
	FacultyName string
}

type ReportDetail struct {
	Report       Report
	ReporterUser *ReportUserInfo
	ReportedUser *ReportUserInfo
}

func (r *Report) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	if r.Status == "" {
		r.Status = ReportStatusPending
	}
	return nil
}
