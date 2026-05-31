package threadsummarysvciface

import (
	"context"
	"io"
	"time"

	threadsummarymodel "github.com/sunibgram/domain/model/thread_summary"
)

type IThreadSummaryService interface {
	Summarize(ctx context.Context, threadID uint) (string, error)
	StreamSummarize(ctx context.Context, threadID, userID uint, w io.Writer) error
	// CachedSummary returns a cached or persisted summary (free, no quota charge) if present.
	CachedSummary(ctx context.Context, threadID uint) (string, bool)
	// QuotaExceeded reports whether the user has burned their daily token budget.
	QuotaExceeded(ctx context.Context, userID uint) (bool, error)
	InvalidateCache(ctx context.Context, threadID uint) error
	RegisterReply(ctx context.Context, threadID uint) error
	// RecordView records that the user summarized this thread (history).
	RecordView(ctx context.Context, userID, threadID uint) error
	// ListHistory returns the user's summary history (newest first), the matching
	// view timestamps, and the total count. When q is non-empty, results are
	// filtered to summaries matching the query text.
	ListHistory(ctx context.Context, userID uint, page, limit int, q string) ([]threadsummarymodel.ThreadSummary, []time.Time, int64, error)
}
