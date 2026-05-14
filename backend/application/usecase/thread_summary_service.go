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

func (s *ThreadSummaryService) Summarize(ctx context.Context, threadID uint) (string, error) {
	var cached string
	if err := s.cacheRepo.Get(ctx, threadSummaryKey(threadID), &cached); err == nil {
		return cached, nil
	}

	replies, err := s.topReplies(threadID, 30)
	if err != nil {
		return "", err
	}

	if len(replies) == 0 {
		return "No comments to summarize yet.", nil
	}

	comments := make([]string, 0, len(replies))
	for _, reply := range replies {
		comments = append(comments, reply.Content)
	}
	prompt := fmt.Sprintf(
		"Summarize the following comments from a university social media post into one concise paragraph. Do not use bullet points or headers, just plain prose:\n\n%s",
		strings.Join(comments, "\n"),
	)

	var buf strings.Builder
	if err := s.ollamaClient.StreamGenerate(ctx, prompt, &buf); err != nil {
		return "", err
	}

	summary := strings.TrimSpace(buf.String())
	_ = s.cacheRepo.Set(ctx, threadSummaryKey(threadID), summary, threadSummaryTTL)
	return summary, nil
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

// satisfy io.Writer for strings.Builder (already implemented, just alias)
var _ io.Writer = (*strings.Builder)(nil)
