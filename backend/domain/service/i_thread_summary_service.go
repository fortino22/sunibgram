package service

import (
	"context"
	"io"
)

type IThreadSummaryService interface {
	StreamSummary(ctx context.Context, threadID uint, w io.Writer) error
	GetCachedSummary(ctx context.Context, threadID uint) (string, bool)
	InvalidateCache(ctx context.Context, threadID uint) error
}
