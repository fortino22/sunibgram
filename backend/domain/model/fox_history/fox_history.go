package foxhistorymodel

import "time"

// History entry types.
const (
	TypeSummary = "summary"
	TypeChat    = "chat"
)

// FoxHistory is one entry within a Fox session (see foxsessionmodel.FoxSession,
// linked by SessionID). Two kinds, discriminated by Type:
//
//   - chat entries: one conversation turn. Question = the user's question, Body =
//     the Fox answer, ThreadID/Snapshot empty. Many per chat session, ordered by
//     ID — the session is re-openable to continue.
//   - summary entries: ThreadID set, Snapshot = root-post content at generation
//     time, Body = the generated summary text, Question empty. Exactly one per
//     summary session; re-summarizing refreshes Body/Snapshot.
//
// Entries are self-contained (no JOIN to thread_summaries needed) so summary
// snapshots survive thread edits/deletes.
type FoxHistory struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	SessionID uint   `gorm:"not null;index:idx_fox_history_session,priority:1"`
	UserID    uint   `gorm:"not null;index:idx_fox_history_user_updated,priority:1"`
	Type      string `gorm:"type:varchar(16);not null"`
	ThreadID  *uint  `gorm:""`

	Question string `gorm:"type:text"`
	Body     string `gorm:"type:text"`
	Snapshot string `gorm:"type:text"`

	CreatedAt time.Time
	UpdatedAt time.Time `gorm:"index:idx_fox_history_user_updated,priority:2,sort:desc;index:idx_fox_history_session,priority:2,sort:desc"`
}
