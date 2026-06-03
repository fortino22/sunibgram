package foxsessionmodel

import "time"

// Session kinds. A session groups one user's Fox-agent activity into a single
// re-openable conversation (chat) or a single summarize action (summary).
const (
	TypeChat    = "chat"
	TypeSummary = "summary"
)

// FoxSession is one entry in a user's Fox history: either a guide-chat
// conversation ("chat", many turns) or a thread summarize ("summary", one
// entry). The turns/entry themselves live in fox_histories, linked by
// SessionID; this row is the grouping header the history list shows.
//
//   - chat sessions: Title = the first question (truncated), ThreadID nil. New
//     turns append to fox_histories and touch UpdatedAt so the session sorts to
//     the top. Re-openable to continue.
//   - summary sessions: ThreadID set, Title = root-post snapshot preview,
//     IsStale mirrors the thread_summaries staleness. One row per (user, thread)
//     enforced by a partial unique index — re-summarizing refreshes it.
type FoxSession struct {
	ID       uint   `gorm:"primaryKey;autoIncrement"`
	UserID   uint   `gorm:"not null;index:idx_fox_session_user_updated,priority:1"`
	Type     string `gorm:"type:varchar(16);not null"`
	ThreadID *uint  `gorm:""`
	Title    string `gorm:"type:text"`
	IsStale  bool   `gorm:"not null;default:false"`

	CreatedAt time.Time
	UpdatedAt time.Time `gorm:"index:idx_fox_session_user_updated,priority:2,sort:desc"`
}
