package communitythreadpinmodel

import "time"

type CommunityThreadPin struct {
	CommunityID uint `gorm:"primaryKey;autoIncrement:false"`
	ThreadID    uint `gorm:"primaryKey;autoIncrement:false"`
	CreatedAt   time.Time
}
