package migration

import (
	"log"

	"gorm.io/gorm"

	"github.com/sunibgram/domain/model"
)

func migratePremiumSubscriptions(db *gorm.DB) {
	if err := db.AutoMigrate(&model.PremiumSubscription{}); err != nil {
		log.Fatalf("AutoMigrate %T: %v", &model.PremiumSubscription{}, err)
	}

	db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS is_premium BOOLEAN NOT NULL DEFAULT false")
	db.Exec("ALTER TABLE premium_subscriptions ADD COLUMN IF NOT EXISTS is_trial BOOLEAN NOT NULL DEFAULT false")
	db.Exec(`
		DO $$
		BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'users' AND column_name = 'is_trial'
			) THEN
				UPDATE premium_subscriptions
				SET is_trial = true
				FROM users
				WHERE premium_subscriptions.user_id = users.user_id
					AND users.is_trial = true;
			END IF;
		END $$;
	`)
	db.Exec("ALTER TABLE users DROP COLUMN IF EXISTS is_trial")
	db.Exec("DROP INDEX IF EXISTS idx_premium_subscriptions_user_id")
}
