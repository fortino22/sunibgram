package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"golang.org/x/sync/singleflight"

	threadmodel "github.com/sunibgram/domain/model/thread"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/infrastructure/ollama"
)

const threadSummaryTTL = time.Hour

var ErrThreadSummaryNotReady = errors.New("thread summary is being generated")

type ThreadSummaryService struct {
	threadRepo   repository.IThreadRepository
	cacheRepo    repository.ICacheRepository
	ollamaClient *ollama.Client
	group        singleflight.Group
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
	key := threadSummaryKey(threadID)

	var cached string
	if err := s.cacheRepo.Get(ctx, key, &cached); err == nil {
		return cached, nil
	}

	s.queueSummaryGeneration(threadID, key)
	return "", ErrThreadSummaryNotReady
}

func (s *ThreadSummaryService) queueSummaryGeneration(threadID uint, key string) {
	go func() {
		_, err, _ := s.group.Do(key, func() (any, error) {
			ctx := context.Background()

			var cached string
			if err := s.cacheRepo.Get(ctx, key, &cached); err == nil {
				return cached, nil
			}

			summary, err := s.generateSummary(ctx, threadID)
			if err != nil {
				return nil, err
			}

			if err := s.cacheRepo.Set(ctx, key, summary, threadSummaryTTL); err != nil {
				return nil, err
			}

			return summary, nil
		})
		if err != nil {
			log.Printf("thread summary generation failed for thread %d: %v", threadID, err)
		}
	}()
}

func (s *ThreadSummaryService) generateSummary(ctx context.Context, threadID uint) (string, error) {
	replies, err := s.topReplies(threadID, 50)
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
	return summary, nil
}

func (s *ThreadSummaryService) InvalidateCache(ctx context.Context, threadID uint) error {
	return s.cacheRepo.Delete(ctx, threadSummaryKey(threadID))
}

func (s *ThreadSummaryService) topReplies(threadID uint, limit int) ([]threadmodel.Thread, error) {
	threads, _, err := s.threadRepo.FindByRootID(threadID, 1, limit)
	if err != nil {
		return nil, err
	}

	replies := make([]threadmodel.Thread, 0, len(threads))
	for _, thread := range threads {
		if thread.ParentID != nil && *thread.ParentID == threadID {
			replies = append(replies, thread)
		}
	}
	return replies, nil
}

func threadSummaryKey(threadID uint) string {
	return fmt.Sprintf("thread:summary:%d", threadID)
}

// satisfy io.Writer for strings.Builder (already implemented, just alias)
var _ io.Writer = (*strings.Builder)(nil)
