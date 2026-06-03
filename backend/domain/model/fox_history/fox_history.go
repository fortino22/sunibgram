package foxhistorymodel

import "time"

// History entry types.
const (
	TypeSummary = "summary"
	TypeChat    = "chat"
)

// FoxHistory is the unified per-user history of Fox-agent activity: thread
// summaries the user generated and guide-chat asks they made. One table, two
// kinds, discriminated by Type.
//
//   - summary rows: ThreadID set, Snapshot = root-post content at generation
//     time, Body = the generated summary text, Question empty. One row per
//     (user, thread) — re-summarizing the same thread touches UpdatedAt and
//     refreshes Body/Snapshot (partial unique index enforces this).
//   - chat rows: ThreadID nil, Question = the user's question, Body = the Fox
//     answer, Snapshot empty. Always appended (no dedupe).
//
// The list is self-contained (no JOIN to thread_summaries needed) so summary
// snapshots survive thread edits/deletes, same as the old UserThreadSummary +
// ThreadSummary.ContentSnapshot pairing.
type FoxHistory struct {
	ID       uint   `gorm:"primaryKey;autoIncrement"`
	UserID   uint   `gorm:"not null;index:idx_fox_history_user_updated,priority:1"`
	Type     string `gorm:"type:varchar(16);not null"`
	ThreadID *uint  `gorm:""`

	Question string `gorm:"type:text"`
	Body     string `gorm:"type:text"`
	Snapshot string `gorm:"type:text"`

	CreatedAt time.Time
	UpdatedAt time.Time `gorm:"index:idx_fox_history_user_updated,priority:2,sort:desc"`
}
