package migration

import "gorm.io/gorm"

func migrateViolations(db *gorm.DB) {
	db.Exec("CREATE INDEX IF NOT EXISTS idx_violations_name ON violations(name)")
}
