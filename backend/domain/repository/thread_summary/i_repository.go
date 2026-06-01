package threadsummaryrepoiface

import (
	"time"

	threadsummarymodel "github.com/sunibgram/domain/model/thread_summary"
)

type IThreadSummaryRepository interface {
	// GetSummary returns the persisted summary for a thread, or nil if absent.
	GetSummary(threadID uint) (*threadsummarymodel.ThreadSummary, error)

	// UpsertSummary writes a freshly generated summary. On conflict (thread_id
	// already exists) it updates the text/snapshot, bumps Version, and clears
	// the stale flag.
	UpsertSummary(summary, contentSnapshot string, threadID uint) error

	// MarkStale flags a thread's summary for regeneration on next read.
	MarkStale(threadID uint) error

	// RecordUserView upserts the per-user history row, touching UpdatedAt so the
	// thread sorts to the top of the user's history.
	RecordUserView(userID, threadID uint) error

	// ListUserHistory returns the user's summarized threads ordered by most
	// recently summarized. timestamps[i] is when the user last summarized
	// summaries[i]. When q is non-empty, results are filtered (case-insensitive)
	// to summaries whose content snapshot or summary text contains q. Returns
	// (summaries, timestamps, total) — total reflects the filter.
	ListUserHistory(userID uint, offset, limit int, q string) ([]threadsummarymodel.ThreadSummary, []time.Time, int64, error)
}
