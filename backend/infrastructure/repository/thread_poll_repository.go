package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	pollBodyTTL = 30 * time.Second
	pollVoteTTL = 5 * time.Minute
)

type ThreadPollRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewThreadPollRepository(writeDB, readDB *gorm.DB, c ...*cache.Cache) domainrepo.IThreadPollRepository {
	var redisCache *cache.Cache
	if len(c) > 0 {
		redisCache = c[0]
	}
	return &ThreadPollRepository{writeDB: writeDB, readDB: readDB, cache: redisCache}
}

func pollThreadKey(threadID uint) string {
	return fmt.Sprintf("poll:thread:%d", threadID)
}

func pollUserVoteKey(pollID, userID uint) string {
	return fmt.Sprintf("poll:%d:user:%d", pollID, userID)
}

func pollUserVotePattern(pollID uint) string {
	return fmt.Sprintf("poll:%d:user:*", pollID)
}

func (r *ThreadPollRepository) Create(poll *model.ThreadPoll, options []model.ThreadPollOption) error {
	if err := r.writeDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(poll).Error; err != nil {
			return err
		}
		for i := range options {
			options[i].PollID = poll.PollID
		}
		if err := tx.Create(&options).Error; err != nil {
			return err
		}
		poll.Options = options
		return nil
	}); err != nil {
		return err
	}
	if r.cache != nil {
		ctx := context.Background()
		_ = r.cache.Delete(ctx, pollThreadKey(poll.ThreadID), threadKey(poll.ThreadID), parentThreadKey(poll.ThreadID))
	}
	return nil
}

func (r *ThreadPollRepository) FindByThreadID(threadID uint) (*model.ThreadPoll, error) {
	ctx := context.Background()
	if r.cache != nil {
		var cached model.ThreadPoll
		if err := r.cache.Get(ctx, pollThreadKey(threadID), &cached); err == nil {
			return &cached, nil
		}
	}
	var poll model.ThreadPoll
	if err := r.readDB.
		Preload("Options", func(db *gorm.DB) *gorm.DB { return db.Order("position ASC") }).
		Where("thread_id = ?", threadID).
		First(&poll).Error; err != nil {
		return nil, err
	}
	if r.cache != nil {
		_ = r.cache.Set(ctx, pollThreadKey(threadID), poll, pollBodyTTL)
	}
	return &poll, nil
}

func (r *ThreadPollRepository) FindByThreadIDs(threadIDs []uint) (map[uint]*model.ThreadPoll, error) {
	result := make(map[uint]*model.ThreadPoll)
	if len(threadIDs) == 0 {
		return result, nil
	}
	ctx := context.Background()
	missing := make([]uint, 0, len(threadIDs))
	if r.cache != nil {
		for _, id := range threadIDs {
			var cached model.ThreadPoll
			if err := r.cache.Get(ctx, pollThreadKey(id), &cached); err == nil {
				result[id] = &cached
				continue
			}
			missing = append(missing, id)
		}
	} else {
		missing = threadIDs
	}
	if len(missing) == 0 {
		return result, nil
	}

	var polls []model.ThreadPoll
	if err := r.readDB.Where("thread_id IN ?", missing).Find(&polls).Error; err != nil {
		return nil, err
	}
	if len(polls) == 0 {
		return result, nil
	}
	pollIDs := make([]uint, 0, len(polls))
	pollByID := make(map[uint]*model.ThreadPoll, len(polls))
	for i := range polls {
		pollIDs = append(pollIDs, polls[i].PollID)
		pollByID[polls[i].PollID] = &polls[i]
	}
	var options []model.ThreadPollOption
	if err := r.readDB.
		Where("poll_id IN ?", pollIDs).
		Order("position ASC").
		Find(&options).Error; err != nil {
		return nil, err
	}
	for _, opt := range options {
		if p, ok := pollByID[opt.PollID]; ok {
			p.Options = append(p.Options, opt)
		}
	}
	for _, p := range pollByID {
		result[p.ThreadID] = p
		if r.cache != nil {
			_ = r.cache.Set(ctx, pollThreadKey(p.ThreadID), *p, pollBodyTTL)
		}
	}
	return result, nil
}

func (r *ThreadPollRepository) FindUserVotes(userID uint, pollIDs []uint) (map[uint]uint, error) {
	result := make(map[uint]uint)
	if len(pollIDs) == 0 || userID == 0 {
		return result, nil
	}
	ctx := context.Background()
	missing := make([]uint, 0, len(pollIDs))
	if r.cache != nil {
		for _, pid := range pollIDs {
			var optID uint
			if err := r.cache.Get(ctx, pollUserVoteKey(pid, userID), &optID); err == nil {
				if optID != 0 {
					result[pid] = optID
				}
				continue
			}
			missing = append(missing, pid)
		}
	} else {
		missing = pollIDs
	}
	if len(missing) == 0 {
		return result, nil
	}

	var votes []model.ThreadPollVote
	if err := r.readDB.
		Where("user_id = ? AND poll_id IN ?", userID, missing).
		Find(&votes).Error; err != nil {
		return nil, err
	}
	hit := make(map[uint]uint, len(votes))
	for _, v := range votes {
		hit[v.PollID] = v.OptionID
		result[v.PollID] = v.OptionID
	}
	if r.cache != nil {
		for _, pid := range missing {
			optID := hit[pid] // 0 if no vote — cache miss-marker
			_ = r.cache.Set(ctx, pollUserVoteKey(pid, userID), optID, pollVoteTTL)
		}
	}
	return result, nil
}

func (r *ThreadPollRepository) FindOptionByID(optionID uint) (*model.ThreadPollOption, error) {
	var opt model.ThreadPollOption
	if err := r.readDB.Where("option_id = ?", optionID).First(&opt).Error; err != nil {
		return nil, err
	}
	return &opt, nil
}

// CastVote inserts a new vote or switches an existing vote to a different option.
// Atomic: option vote_counts shift via gorm.Expr; poll total_votes only increments on first vote.
func (r *ThreadPollRepository) CastVote(pollID, threadID, optionID, userID uint) error {
	if err := r.writeDB.Transaction(func(tx *gorm.DB) error {
		var existing model.ThreadPollVote
		err := tx.Where("poll_id = ? AND user_id = ?", pollID, userID).First(&existing).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil {
			if existing.OptionID == optionID {
				return nil
			}
			if err := tx.Model(&model.ThreadPollOption{}).
				Where("option_id = ?", existing.OptionID).
				UpdateColumn("vote_count", gorm.Expr("GREATEST(vote_count - 1, 0)")).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.ThreadPollOption{}).
				Where("option_id = ?", optionID).
				UpdateColumn("vote_count", gorm.Expr("vote_count + 1")).Error; err != nil {
				return err
			}
			return tx.Model(&model.ThreadPollVote{}).
				Where("vote_id = ?", existing.VoteID).
				Updates(map[string]any{"option_id": optionID}).Error
		}
		vote := model.ThreadPollVote{PollID: pollID, OptionID: optionID, UserID: userID}
		res := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "poll_id"}, {Name: "user_id"}},
			DoNothing: true,
		}).Create(&vote)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return r.switchVote(tx, pollID, optionID, userID)
		}
		if err := tx.Model(&model.ThreadPollOption{}).
			Where("option_id = ?", optionID).
			UpdateColumn("vote_count", gorm.Expr("vote_count + 1")).Error; err != nil {
			return err
		}
		return tx.Model(&model.ThreadPoll{}).
			Where("poll_id = ?", pollID).
			UpdateColumn("total_votes", gorm.Expr("total_votes + 1")).Error
	}); err != nil {
		return err
	}
	if r.cache != nil {
		ctx := context.Background()
		if threadID != 0 {
			_ = r.cache.Delete(ctx, pollThreadKey(threadID))
		}
		_ = r.cache.Set(ctx, pollUserVoteKey(pollID, userID), optionID, pollVoteTTL)
	}
	return nil
}

func (r *ThreadPollRepository) switchVote(tx *gorm.DB, pollID, optionID, userID uint) error {
	var existing model.ThreadPollVote
	if err := tx.Where("poll_id = ? AND user_id = ?", pollID, userID).First(&existing).Error; err != nil {
		return err
	}
	if existing.OptionID == optionID {
		return nil
	}
	if err := tx.Model(&model.ThreadPollOption{}).
		Where("option_id = ?", existing.OptionID).
		UpdateColumn("vote_count", gorm.Expr("GREATEST(vote_count - 1, 0)")).Error; err != nil {
		return err
	}
	if err := tx.Model(&model.ThreadPollOption{}).
		Where("option_id = ?", optionID).
		UpdateColumn("vote_count", gorm.Expr("vote_count + 1")).Error; err != nil {
		return err
	}
	return tx.Model(&model.ThreadPollVote{}).
		Where("vote_id = ?", existing.VoteID).
		Updates(map[string]any{"option_id": optionID}).Error
}

func (r *ThreadPollRepository) DeleteByThreadID(threadID uint) ([]model.ThreadPollOption, error) {
	var poll model.ThreadPoll
	if err := r.readDB.Where("thread_id = ?", threadID).First(&poll).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	var options []model.ThreadPollOption
	if err := r.readDB.Where("poll_id = ?", poll.PollID).Find(&options).Error; err != nil {
		return nil, err
	}
	if err := r.writeDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("poll_id = ?", poll.PollID).Delete(&model.ThreadPollVote{}).Error; err != nil {
			return err
		}
		if err := tx.Where("poll_id = ?", poll.PollID).Delete(&model.ThreadPollOption{}).Error; err != nil {
			return err
		}
		return tx.Where("poll_id = ?", poll.PollID).Delete(&model.ThreadPoll{}).Error
	}); err != nil {
		return nil, err
	}
	if r.cache != nil {
		ctx := context.Background()
		_ = r.cache.Delete(ctx, pollThreadKey(threadID))
		_ = r.cache.DeleteByPattern(ctx, pollUserVotePattern(poll.PollID))
	}
	return options, nil
}

// parsePollIDFromKey unused helper kept for future debug; intentionally simple.
func parsePollIDFromKey(key string) (uint, error) {
	parts := strings.Split(key, ":")
	if len(parts) < 2 {
		return 0, errors.New("invalid key")
	}
	v, err := strconv.ParseUint(parts[1], 10, 64)
	return uint(v), err
}
