package premiumsubscriptiondto

import (
	"time"

	"github.com/google/uuid"
)

type PremiumSubscriptionResponse struct {
	ID                    uuid.UUID `json:"id"`
	UserID                uint      `json:"user_id"`
	IsTrial               bool      `json:"is_trial"`
	SubscriptionStartDate time.Time `json:"subscription_start_date"`
	SubscriptionEndDate   time.Time `json:"subscription_end_date"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}
