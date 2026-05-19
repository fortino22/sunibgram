package migration

import "gorm.io/gorm"

func migrateUserProfiles(db *gorm.DB) {
	db.Exec("ALTER TABLE user_profiles ADD COLUMN IF NOT EXISTS gender VARCHAR(30)")
	db.Exec(`
		UPDATE user_profiles
		SET gender = 'prefer_not_to_say'
		WHERE gender IS NULL
			OR gender = ''
			OR gender NOT IN ('male', 'female', 'other', 'prefer_not_to_say')
	`)
	db.Exec("ALTER TABLE user_profiles ALTER COLUMN gender SET NOT NULL")
	db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'chk_user_profiles_gender'
			) THEN
				ALTER TABLE user_profiles ADD CONSTRAINT chk_user_profiles_gender
				CHECK (gender IN ('male', 'female', 'other', 'prefer_not_to_say'));
			END IF;
		END $$;
	`)

	db.Exec("ALTER TABLE user_profiles ADD COLUMN IF NOT EXISTS birthday_date DATE")
	db.Exec("UPDATE user_profiles SET birthday_date = DATE '1970-01-01' WHERE birthday_date IS NULL")
	db.Exec("ALTER TABLE user_profiles ALTER COLUMN birthday_date SET NOT NULL")
}
