package threadsummarysvciface

import (
	"context"
	"io"
)

type IThreadSummaryService interface {
	Summarize(ctx context.Context, threadID uint) (string, error)
	StreamSummarize(ctx context.Context, threadID uint, w io.Writer) error
	InvalidateCache(ctx context.Context, threadID uint) error
	RegisterReply(ctx context.Context, threadID uint) error
}
