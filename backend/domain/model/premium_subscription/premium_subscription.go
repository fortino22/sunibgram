package premiumsubscriptionmodel

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	usermodel "github.com/sunibgram/domain/model/user"
)

type PremiumSubscription struct {
	ID                    uuid.UUID      `gorm:"type:uuid;primaryKey"`
	UserID                uint           `gorm:"not null;index"`
	User                  usermodel.User `gorm:"foreignKey:UserID;references:UserID;constraint:OnUpdate:NO ACTION,OnDelete:CASCADE"`
	IsTrial               bool           `gorm:"not null;default:false"`
	SubscriptionStartDate time.Time      `gorm:"not null"`
	SubscriptionEndDate   time.Time      `gorm:"not null"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func (p *PremiumSubscription) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
