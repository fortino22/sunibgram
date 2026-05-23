package feedbackrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	feedbackmodel "github.com/sunibgram/domain/model/feedback"
	feedbackrepoiface "github.com/sunibgram/domain/repository/feedback"
	"github.com/sunibgram/infrastructure/cache"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
)

const feedbackTTL = 5 * time.Minute

type feedbackRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewFeedbackRepository(writeDB, readDB *gorm.DB, c *cache.Cache) feedbackrepoiface.IFeedbackRepository {
	return &feedbackRepository{writeDB: writeDB, readDB: readDB, cache: c}
}

func feedbackIDKey(id uuid.UUID) string {
	return fmt.Sprintf("feedback:id:%s", id.String())
}

func feedbackListKey(f feedbackrepoiface.ListFilter) string {
	return fmt.Sprintf("feedback:list:type:%s:status:%s:page:%d:limit:%d", f.Type, f.Status, f.Page, f.Limit)
}

func feedbackVotesKey(id uuid.UUID) string {
	return fmt.Sprintf("feedback:votes:%s", id.String())
}

func feedbackUserVoteKey(id uuid.UUID, userID uint) string {
	return fmt.Sprintf("feedback:vote:%s:user:%d", id.String(), userID)
}

type listCachePayload struct {
	Items []feedbackmodel.Feedback `json:"items"`
	Total int64                    `json:"total"`
}

func (r *feedbackRepository) Save(f *feedbackmodel.Feedback) error {
	if err := r.writeDB.Create(f).Error; err != nil {
		return err
	}
	if err := r.writeDB.Preload("User").First(f, "id = ?", f.ID).Error; err != nil {
		return err
	}
	r.invalidateListsAndItem(f.ID)
	return nil
}

func (r *feedbackRepository) FindByID(id uuid.UUID) (*feedbackmodel.Feedback, error) {
	ctx := context.Background()
	var f feedbackmodel.Feedback
	if r.cache != nil {
		if err := r.cache.Get(ctx, feedbackIDKey(id), &f); err == nil {
			return &f, nil
		}
	}
	if err := r.readDB.Preload("User").First(&f, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("feedback not found")
		}
		return nil, err
	}
	if r.cache != nil {
		_ = r.cache.Set(ctx, feedbackIDKey(id), f, feedbackTTL)
	}
	return &f, nil
}

func (r *feedbackRepository) FindAll(filter feedbackrepoiface.ListFilter) ([]feedbackmodel.Feedback, int64, error) {
	ctx := context.Background()
	key := feedbackListKey(filter)
	if r.cache != nil {
		var payload listCachePayload
		if err := r.cache.Get(ctx, key, &payload); err == nil {
			return payload.Items, payload.Total, nil
		}
	}

	q := r.readDB.Model(&feedbackmodel.Feedback{})
	if filter.Type != "" {
		q = q.Where("type = ?", filter.Type)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []feedbackmodel.Feedback
	offset := (filter.Page - 1) * filter.Limit
	if err := q.Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(filter.Limit).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	if r.cache != nil {
		_ = r.cache.Set(ctx, key, listCachePayload{Items: items, Total: total}, feedbackTTL)
		_ = r.cache.RegisterCacheKey(ctx, infrarepoShared.FeedbackListTag(), key)
	}
	return items, total, nil
}

func (r *feedbackRepository) UpdateStatus(id uuid.UUID, status string) error {
	result := r.writeDB.Model(&feedbackmodel.Feedback{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("feedback not found")
	}
	r.invalidateListsAndItem(id)
	return nil
}

func (r *feedbackRepository) Delete(id uuid.UUID) error {
	err := r.writeDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("feedback_id = ?", id).Delete(&feedbackmodel.FeedbackVote{}).Error; err != nil {
			return err
		}
		result := tx.Delete(&feedbackmodel.Feedback{}, "id = ?", id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("feedback not found")
		}
		return nil
	})
	if err != nil {
		return err
	}
	r.invalidateListsAndItem(id)
	r.invalidateVotes(id)
	return nil
}

// InsertVote inserts a vote row. Caller must guarantee no existing vote for
// (feedback_id, user_id). For flipping a vote, use UpdateVote.
func (r *feedbackRepository) InsertVote(feedbackID uuid.UUID, userID uint, value int) error {
	vote := feedbackmodel.FeedbackVote{
		FeedbackID: feedbackID,
		UserID:     userID,
		Value:      value,
	}
	if err := r.writeDB.Create(&vote).Error; err != nil {
		return err
	}
	r.invalidateVotes(feedbackID)
	r.invalidateUserVote(feedbackID, userID)
	return nil
}

// UpdateVote flips an existing vote to the opposite value.
func (r *feedbackRepository) UpdateVote(feedbackID uuid.UUID, userID uint, value int) error {
	result := r.writeDB.Model(&feedbackmodel.FeedbackVote{}).
		Where("feedback_id = ? AND user_id = ?", feedbackID, userID).
		Update("value", value)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("vote not found")
	}
	r.invalidateVotes(feedbackID)
	r.invalidateUserVote(feedbackID, userID)
	return nil
}

func (r *feedbackRepository) GetVoteCounts(feedbackID uuid.UUID) (feedbackmodel.VoteCounts, error) {
	ctx := context.Background()
	var counts feedbackmodel.VoteCounts
	if r.cache != nil {
		if err := r.cache.Get(ctx, feedbackVotesKey(feedbackID), &counts); err == nil {
			return counts, nil
		}
	}
	rows, err := r.readDB.Model(&feedbackmodel.FeedbackVote{}).
		Select("value, COUNT(*) AS c").
		Where("feedback_id = ?", feedbackID).
		Group("value").
		Rows()
	if err != nil {
		return counts, err
	}
	defer rows.Close()
	for rows.Next() {
		var value int
		var c int64
		if err := rows.Scan(&value, &c); err != nil {
			return counts, err
		}
		switch value {
		case 1:
			counts.Likes = c
		case -1:
			counts.Dislikes = c
		}
	}
	if r.cache != nil {
		_ = r.cache.Set(ctx, feedbackVotesKey(feedbackID), counts, feedbackTTL)
	}
	return counts, nil
}

func (r *feedbackRepository) GetUserVote(feedbackID uuid.UUID, userID uint) (int, bool, error) {
	if userID == 0 {
		return 0, false, nil
	}
	ctx := context.Background()
	if r.cache != nil {
		var cached int
		if err := r.cache.Get(ctx, feedbackUserVoteKey(feedbackID, userID), &cached); err == nil {
			if cached == 0 {
				return 0, false, nil
			}
			return cached, true, nil
		}
	}
	var vote feedbackmodel.FeedbackVote
	err := r.readDB.Where("feedback_id = ? AND user_id = ?", feedbackID, userID).First(&vote).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if r.cache != nil {
			_ = r.cache.Set(ctx, feedbackUserVoteKey(feedbackID, userID), 0, feedbackTTL)
		}
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	if r.cache != nil {
		_ = r.cache.Set(ctx, feedbackUserVoteKey(feedbackID, userID), vote.Value, feedbackTTL)
	}
	return vote.Value, true, nil
}

func (r *feedbackRepository) GetUserVotes(feedbackIDs []uuid.UUID, userID uint) (map[uuid.UUID]int, error) {
	out := make(map[uuid.UUID]int, len(feedbackIDs))
	if userID == 0 || len(feedbackIDs) == 0 {
		return out, nil
	}
	var votes []feedbackmodel.FeedbackVote
	if err := r.readDB.
		Where("user_id = ? AND feedback_id IN ?", userID, feedbackIDs).
		Find(&votes).Error; err != nil {
		return nil, err
	}
	for _, v := range votes {
		out[v.FeedbackID] = v.Value
	}
	return out, nil
}

func (r *feedbackRepository) GetVoteCountsBulk(feedbackIDs []uuid.UUID) (map[uuid.UUID]feedbackmodel.VoteCounts, error) {
	out := make(map[uuid.UUID]feedbackmodel.VoteCounts, len(feedbackIDs))
	if len(feedbackIDs) == 0 {
		return out, nil
	}
	type row struct {
		FeedbackID uuid.UUID
		Value      int
		C          int64
	}
	var rows []row
	if err := r.readDB.Model(&feedbackmodel.FeedbackVote{}).
		Select("feedback_id, value, COUNT(*) AS c").
		Where("feedback_id IN ?", feedbackIDs).
		Group("feedback_id, value").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		c := out[r.FeedbackID]
		switch r.Value {
		case 1:
			c.Likes = r.C
		case -1:
			c.Dislikes = r.C
		}
		out[r.FeedbackID] = c
	}
	return out, nil
}

func (r *feedbackRepository) invalidateListsAndItem(id uuid.UUID) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.Delete(ctx, feedbackIDKey(id))
	_ = r.cache.InvalidateTag(ctx, infrarepoShared.FeedbackListTag())
}

func (r *feedbackRepository) invalidateVotes(id uuid.UUID) {
	if r.cache == nil {
		return
	}
	_ = r.cache.Delete(context.Background(), feedbackVotesKey(id))
}

func (r *feedbackRepository) invalidateUserVote(id uuid.UUID, userID uint) {
	if r.cache == nil {
		return
	}
	_ = r.cache.Delete(context.Background(), feedbackUserVoteKey(id, userID))
}
