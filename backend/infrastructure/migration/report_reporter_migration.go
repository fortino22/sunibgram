package migration

import "gorm.io/gorm"

func migrateReportReporter(db *gorm.DB) {
	db.Exec("ALTER TABLE reports ADD COLUMN IF NOT EXISTS reporter_id INTEGER")
	db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM reports WHERE reporter_id IS NULL) THEN
				ALTER TABLE reports ALTER COLUMN reporter_id SET NOT NULL;
			END IF;
		END $$;
	`)
	db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'fk_reports_reporter'
			) THEN
				ALTER TABLE reports ADD CONSTRAINT fk_reports_reporter FOREIGN KEY (reporter_id) REFERENCES users(user_id) ON DELETE CASCADE;
			END IF;
		END $$;
	`)
}
