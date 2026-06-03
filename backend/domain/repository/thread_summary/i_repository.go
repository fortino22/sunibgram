package threadsummaryrepoiface

import (
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
}
