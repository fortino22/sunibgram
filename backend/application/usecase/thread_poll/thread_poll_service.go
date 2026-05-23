package threadpolluc

import (
	"context"
	"errors"
	"mime/multipart"
	"strings"
	"sync"
	"time"

	storageuc "github.com/sunibgram/application/usecase/storage"
	threadmodel "github.com/sunibgram/domain/model/thread"
	threadpollmodel "github.com/sunibgram/domain/model/thread_poll"
	threadrepoiface "github.com/sunibgram/domain/repository/thread"
	threadpollrepoiface "github.com/sunibgram/domain/repository/thread_poll"
	threadpollsvciface "github.com/sunibgram/domain/service/thread_poll"
)

type ThreadPollService struct {
	threadRepo threadrepoiface.IThreadRepository
	pollRepo   threadpollrepoiface.IThreadPollRepository
	storage    *storageuc.StorageUsecase
}

func NewThreadPollService(
	threadRepo threadrepoiface.IThreadRepository,
	pollRepo threadpollrepoiface.IThreadPollRepository,
	storage *storageuc.StorageUsecase,
) threadpollsvciface.IThreadPollService {
	return &ThreadPollService{threadRepo: threadRepo, pollRepo: pollRepo, storage: storage}
}

func (s *ThreadPollService) Create(ctx context.Context, threadID, userID uint, input threadpollsvciface.CreatePollInput) (*threadpollmodel.ThreadPoll, error) {
	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return nil, errors.New("thread not found")
	}
	if thread.UserID != userID {
		return nil, errors.New("forbidden")
	}
	if err := threadpollmodel.ValidatePollAllowed(thread.Type); err != nil {
		return nil, err
	}
	if existing, _ := s.pollRepo.FindByThreadID(threadID); existing != nil {
		return nil, errors.New("thread already has a poll")
	}

	question := strings.TrimSpace(input.Question)
	if err := threadpollmodel.ValidatePollQuestion(question); err != nil {
		return nil, err
	}
	if err := threadpollmodel.ValidatePollOptionsCount(len(input.Labels)); err != nil {
		return nil, err
	}
	if len(input.Images) > 0 && len(input.Images) != len(input.Labels) {
		return nil, errors.New("poll option images must align with labels")
	}
	if err := threadpollmodel.ValidatePollDuration(input.DurationSeconds); err != nil {
		return nil, err
	}

	labels := make([]string, len(input.Labels))
	for i, raw := range input.Labels {
		label := strings.TrimSpace(raw)
		if err := threadpollmodel.ValidatePollLabel(label); err != nil {
			return nil, err
		}
		labels[i] = label
	}

	bucket, err := s.storage.ResolveBucket("poll")
	if err != nil {
		return nil, err
	}

	uploaded, err := s.uploadOptionImages(ctx, bucket, input.Images)
	if err != nil {
		s.cleanupObjects(ctx, bucket, uploaded)
		return nil, err
	}

	endsAt := time.Now().Add(time.Duration(input.DurationSeconds) * time.Second)
	poll := &threadpollmodel.ThreadPoll{
		ThreadID: threadID,
		Question: question,
		EndsAt:   endsAt,
	}
	options := make([]threadpollmodel.ThreadPollOption, len(labels))
	for i, label := range labels {
		opt := threadpollmodel.ThreadPollOption{
			Label:    label,
			Position: int16(i),
		}
		if i < len(uploaded) && uploaded[i] != "" {
			opt.ObjectName = uploaded[i]
			opt.Bucket = bucket
		}
		options[i] = opt
	}

	if err := s.pollRepo.Create(poll, options); err != nil {
		s.cleanupObjects(ctx, bucket, uploaded)
		return nil, err
	}
	return poll, nil
}

func (s *ThreadPollService) GetByThreadID(threadID, requesterID uint) (*threadpollmodel.ThreadPoll, *uint, error) {
	poll, err := s.pollRepo.FindByThreadID(threadID)
	if err != nil {
		return nil, nil, err
	}
	if requesterID == 0 {
		return poll, nil, nil
	}
	votes, err := s.pollRepo.FindUserVotes(requesterID, []uint{poll.PollID})
	if err != nil {
		return poll, nil, nil
	}
	if optionID, ok := votes[poll.PollID]; ok {
		return poll, &optionID, nil
	}
	return poll, nil, nil
}

func (s *ThreadPollService) GetForThreads(threadIDs []uint, requesterID uint) (map[uint]*threadpollmodel.ThreadPoll, map[uint]uint, error) {
	polls, err := s.pollRepo.FindByThreadIDs(threadIDs)
	if err != nil {
		return nil, nil, err
	}
	if len(polls) == 0 || requesterID == 0 {
		return polls, map[uint]uint{}, nil
	}
	pollIDs := make([]uint, 0, len(polls))
	for _, p := range polls {
		pollIDs = append(pollIDs, p.PollID)
	}
	votes, err := s.pollRepo.FindUserVotes(requesterID, pollIDs)
	if err != nil {
		return polls, map[uint]uint{}, nil
	}
	return polls, votes, nil
}

func (s *ThreadPollService) Vote(ctx context.Context, threadID, optionID, userID uint) (*threadpollmodel.ThreadPoll, *uint, error) {
	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return nil, nil, errors.New("thread not found")
	}
	resolvedThreadID := threadID
	resolvedOwnerID := thread.UserID
	if thread.Type == threadmodel.ThreadTypeRepost && thread.OriginalThreadID != nil {
		resolvedThreadID = *thread.OriginalThreadID
		if original, err := s.threadRepo.FindByID(resolvedThreadID); err == nil {
			resolvedOwnerID = original.UserID
		}
	}
	if resolvedOwnerID == userID {
		return nil, nil, errors.New("forbidden: cannot vote on your own poll")
	}
	poll, err := s.pollRepo.FindByThreadID(resolvedThreadID)
	if err != nil {
		return nil, nil, errors.New("poll not found")
	}
	if poll.IsExpired() {
		return nil, nil, errors.New("poll has ended")
	}
	option, err := s.pollRepo.FindOptionByID(optionID)
	if err != nil {
		return nil, nil, errors.New("option not found")
	}
	if option.PollID != poll.PollID {
		return nil, nil, errors.New("option does not belong to poll")
	}
	if err := s.pollRepo.CastVote(poll.PollID, resolvedThreadID, optionID, userID); err != nil {
		return nil, nil, err
	}
	refreshed, err := s.pollRepo.FindByThreadID(resolvedThreadID)
	if err != nil {
		return nil, nil, err
	}
	voted := optionID
	return refreshed, &voted, nil
}

func (s *ThreadPollService) DeleteByThread(ctx context.Context, threadID uint) error {
	options, err := s.pollRepo.DeleteByThreadID(threadID)
	if err != nil {
		return err
	}
	for _, opt := range options {
		if opt.ObjectName == "" || opt.Bucket == "" {
			continue
		}
		_ = s.storage.Delete(ctx, opt.Bucket, opt.ObjectName)
	}
	return nil
}

func (s *ThreadPollService) GetImageURL(bucket, objectName string) string {
	if bucket == "" || objectName == "" {
		return ""
	}
	return s.storage.GetURL(bucket, objectName)
}

func (s *ThreadPollService) HasPoll(threadID uint) bool {
	poll, err := s.pollRepo.FindByThreadID(threadID)
	return err == nil && poll != nil
}

func (s *ThreadPollService) uploadOptionImages(ctx context.Context, bucket string, headers []*multipart.FileHeader) ([]string, error) {
	if len(headers) == 0 {
		return nil, nil
	}
	result := make([]string, len(headers))
	errCh := make(chan error, len(headers))
	var wg sync.WaitGroup
	for i, header := range headers {
		if header == nil {
			continue
		}
		i, header := i, header
		wg.Add(1)
		go func() {
			defer wg.Done()
			file, err := header.Open()
			if err != nil {
				errCh <- err
				return
			}
			defer file.Close()
			if err := threadpollmodel.ValidatePollImage(header, file); err != nil {
				errCh <- err
				return
			}
			objectName, err := s.storage.Upload(ctx, bucket, file, header)
			if err != nil {
				errCh <- err
				return
			}
			result[i] = objectName
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

func (s *ThreadPollService) cleanupObjects(ctx context.Context, bucket string, objects []string) {
	for _, name := range objects {
		if name == "" {
			continue
		}
		_ = s.storage.Delete(ctx, bucket, name)
	}
}
