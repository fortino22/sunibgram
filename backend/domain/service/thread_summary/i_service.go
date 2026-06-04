package threadsummarysvciface

import (
	"context"
	"io"
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
	// RecordView records that the user summarized this thread (summary session).
	RecordView(ctx context.Context, userID, threadID uint) error
}
