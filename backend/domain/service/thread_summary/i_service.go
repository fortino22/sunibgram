package threadsummarysvciface

import "context"

type IThreadSummaryService interface {
	Summarize(ctx context.Context, threadID uint) (string, error)
	InvalidateCache(ctx context.Context, threadID uint) error
}
