package usecase

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/infrastructure/ollama"
)

const threadSummaryTTL = 600 * time.Second

type ThreadSummaryService struct {
	threadRepo   repository.IThreadRepository
	cacheRepo    repository.ICacheRepository
	ollamaClient *ollama.Client
}

func NewThreadSummaryService(
	threadRepo repository.IThreadRepository,
	cacheRepo repository.ICacheRepository,
	ollamaClient *ollama.Client,
) service.IThreadSummaryService {
	return &ThreadSummaryService{
		threadRepo:   threadRepo,
		cacheRepo:    cacheRepo,
		ollamaClient: ollamaClient,
	}
}

func (s *ThreadSummaryService) StreamSummary(ctx context.Context, threadID uint, w io.Writer) error {
	if cached, ok := s.GetCachedSummary(ctx, threadID); ok {
		_, err := io.WriteString(w, cached)
		return err
	}

	replies, err := s.topReplies(threadID, 30)
	if err != nil {
		return err
	}

	comments := make([]string, 0, len(replies))
	for _, reply := range replies {
		comments = append(comments, reply.Content)
	}
	prompt := fmt.Sprintf(
		"Summarize the following comments on a social media post. Be concise and highlight the main points:\n\n%s",
		strings.Join(comments, "\n"),
	)

	var captured strings.Builder
	if err := s.ollamaClient.StreamGenerate(ctx, prompt, io.MultiWriter(w, &captured)); err != nil {
		return err
	}

	return s.cacheRepo.Set(ctx, threadSummaryKey(threadID), captured.String(), threadSummaryTTL)
}

func (s *ThreadSummaryService) GetCachedSummary(ctx context.Context, threadID uint) (string, bool) {
	var summary string
	if err := s.cacheRepo.Get(ctx, threadSummaryKey(threadID), &summary); err != nil {
		return "", false
	}
	return summary, true
}

func (s *ThreadSummaryService) InvalidateCache(ctx context.Context, threadID uint) error {
	return s.cacheRepo.Delete(ctx, threadSummaryKey(threadID))
}

func (s *ThreadSummaryService) topReplies(threadID uint, limit int) ([]model.Thread, error) {
	threads, _, err := s.threadRepo.FindByRootID(threadID, 1, 10000)
	if err != nil {
		return nil, err
	}

	replies := make([]model.Thread, 0, len(threads))
	for _, thread := range threads {
		if thread.ParentID != nil && *thread.ParentID == threadID {
			replies = append(replies, thread)
		}
	}

	sort.SliceStable(replies, func(i, j int) bool {
		if replies[i].LikeCount == replies[j].LikeCount {
			return replies[i].CreatedAt.After(replies[j].CreatedAt)
		}
		return replies[i].LikeCount > replies[j].LikeCount
	})

	if len(replies) > limit {
		replies = replies[:limit]
	}
	return replies, nil
}

func threadSummaryKey(threadID uint) string {
	return fmt.Sprintf("thread:summary:%d", threadID)
}
