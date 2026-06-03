package threadsummarymodel

import "time"

// ThreadSummary is the canonical, persisted per-thread summary. It mirrors the
// value cached in Redis under "thread:summary:<id>" but survives cache eviction
// and powers the per-user summary history. One row per thread.
type ThreadSummary struct {
	ThreadID uint `gorm:"primaryKey;autoIncrement:false"`

	// Summary is the generated bullet-point text.
	Summary string `gorm:"type:text"`

	// ContentSnapshot is the root post content captured at generation time, so
	// the history view can show what was summarized even if the thread is later
	// edited or deleted.
	ContentSnapshot string `gorm:"type:text"`

	// Version increments on each regeneration (after a stale invalidation).
	Version int `gorm:"not null;default:1"`

	// IsStale marks a summary that must be regenerated on next read (e.g. the
	// thread crossed the reply-count invalidation trigger). The old summary stays
	// visible in history until a fresh one replaces it.
	IsStale bool `gorm:"not null;default:false"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
