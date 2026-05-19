package migration

import "gorm.io/gorm"

func migrateReports(db *gorm.DB) {
	db.Exec("CREATE INDEX IF NOT EXISTS idx_reports_status ON reports(status)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_reports_reporter_id ON reports(reporter_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_reports_reported_user_id ON reports(reported_user_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_reports_reported_post_id ON reports(reported_post_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_reports_violation_id ON reports(violation_id)")
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
	db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'fk_reports_reported_user'
			) THEN
				ALTER TABLE reports ADD CONSTRAINT fk_reports_reported_user FOREIGN KEY (reported_user_id) REFERENCES users(user_id) ON DELETE SET NULL;
			END IF;
		END $$;
	`)
	db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'fk_reports_reported_post'
			) THEN
				ALTER TABLE reports ADD CONSTRAINT fk_reports_reported_post FOREIGN KEY (reported_post_id) REFERENCES threads(thread_id) ON DELETE SET NULL;
			END IF;
		END $$;
	`)
	db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'fk_reports_violation'
			) THEN
				ALTER TABLE reports ADD CONSTRAINT fk_reports_violation FOREIGN KEY (violation_id) REFERENCES violations(id) ON DELETE RESTRICT;
			END IF;
		END $$;
	`)
}
