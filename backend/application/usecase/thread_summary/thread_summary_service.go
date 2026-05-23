package threadsummaryuc

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
	cacherepoiface "github.com/sunibgram/domain/repository/cache"
	threadrepoiface "github.com/sunibgram/domain/repository/thread"
	threadsummarysvciface "github.com/sunibgram/domain/service/thread_summary"
	"github.com/sunibgram/infrastructure/openai"
)

const (
	threadSummaryTTL          time.Duration = 0
	threadSummaryReplyTrigger               = 10
)

var ErrThreadSummaryNotReady = errors.New("thread summary is being generated")

type ThreadSummaryService struct {
	threadRepo   threadrepoiface.IThreadRepository
	cacheRepo    cacherepoiface.ICacheRepository
	openaiClient *openai.Client
	group        singleflight.Group
}

func NewThreadSummaryService(
	threadRepo threadrepoiface.IThreadRepository,
	cacheRepo cacherepoiface.ICacheRepository,
	openaiClient *openai.Client,
) threadsummarysvciface.IThreadSummaryService {
	return &ThreadSummaryService{
		threadRepo:   threadRepo,
		cacheRepo:    cacheRepo,
		openaiClient: openaiClient,
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
		"Summarize the following comments from a university social media post into one concise paragraph. Always respond in English, even if the comments are in another language. Do not use bullet points or headers, just plain prose:\n\n%s",
		strings.Join(comments, "\n"),
	)

	var buf strings.Builder
	if err := s.openaiClient.StreamGenerate(ctx, prompt, &buf); err != nil {
		return "", err
	}

	summary := strings.TrimSpace(buf.String())
	return summary, nil
}

func (s *ThreadSummaryService) InvalidateCache(ctx context.Context, threadID uint) error {
	return s.cacheRepo.Delete(ctx, threadSummaryKey(threadID), threadSummaryReplyCountKey(threadID))
}

func (s *ThreadSummaryService) RegisterReply(ctx context.Context, threadID uint) error {
	count, err := s.cacheRepo.Incr(ctx, threadSummaryReplyCountKey(threadID))
	if err != nil {
		return err
	}
	if count < threadSummaryReplyTrigger {
		return nil
	}
	return s.InvalidateCache(ctx, threadID)
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

func threadSummaryReplyCountKey(threadID uint) string {
	return fmt.Sprintf("thread:summary:replycount:%d", threadID)
}

// satisfy io.Writer for strings.Builder (already implemented, just alias)
var _ io.Writer = (*strings.Builder)(nil)
