package repository

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/sunibgram/domain/model"
)

const forYouFeedTTL = 120 * time.Second

type cachedForYouFeed struct {
	Threads []model.Thread
	Total   int64
}

type forYouCandidate struct {
	thread model.Thread
	score  float64
}

func (r *threadCached) FindForYouFeed(ctx context.Context, userID uint, facultyID uint, page, limit int) ([]model.Thread, int64, error) {
	key := fmt.Sprintf("feed:foryou:%d:%d:%d", userID, page, limit)
	var cached cachedForYouFeed
	if err := r.cache.Get(ctx, key, &cached); err == nil {
		return cached.Threads, cached.Total, nil
	}

	merged := make(map[uint]forYouCandidate)
	addCandidates := func(threads []model.Thread, sourceWeight float64, includeReposts bool) {
		for _, thread := range threads {
			if thread.Type == model.ThreadTypeRepost && !includeReposts {
				continue
			}
			score := thread.FeedScore * sourceWeight
			current, ok := merged[thread.ThreadID]
			if !ok || score > current.score {
				merged[thread.ThreadID] = forYouCandidate{thread: thread, score: score}
			}
		}
	}

	following, err := r.findForYouFollowing(userID)
	if err != nil {
		return nil, 0, err
	}
	addCandidates(following, 1.0, true)

	community, err := r.findForYouCommunity(userID)
	if err != nil {
		return nil, 0, err
	}
	addCandidates(community, 0.9, false)

	faculty, err := r.findForYouFaculty(userID, facultyID)
	if err != nil {
		return nil, 0, err
	}
	addCandidates(faculty, 0.7, false)

	global, err := r.findForYouGlobal(ctx)
	if err != nil {
		return nil, 0, err
	}
	addCandidates(global, 0.5, false)

	candidates := make([]forYouCandidate, 0, len(merged))
	for _, candidate := range merged {
		candidates = append(candidates, candidate)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score == candidates[j].score {
			return candidates[i].thread.CreatedAt.After(candidates[j].thread.CreatedAt)
		}
		return candidates[i].score > candidates[j].score
	})

	capped := make([]model.Thread, 0, len(candidates))
	authorCounts := make(map[uint]int)
	for _, candidate := range candidates {
		authorID := candidate.thread.UserID
		if candidate.thread.Type == model.ThreadTypeRepost && candidate.thread.OriginalThread != nil {
			authorID = candidate.thread.OriginalThread.UserID
		}
		if authorCounts[authorID] >= 2 {
			continue
		}
		authorCounts[authorID]++
		capped = append(capped, candidate.thread)
	}

	total := int64(len(capped))
	offset := (page - 1) * limit
	if offset >= len(capped) {
		result := cachedForYouFeed{Threads: []model.Thread{}, Total: total}
		_ = r.cache.Set(ctx, key, result, forYouFeedTTL)
		return result.Threads, result.Total, nil
	}
	end := offset + limit
	if end > len(capped) {
		end = len(capped)
	}

	result := cachedForYouFeed{Threads: capped[offset:end], Total: total}
	if err := r.cache.Set(ctx, key, result, forYouFeedTTL); err == nil {
		for _, thread := range result.Threads {
			r.trackFeedCacheKey(ctx, thread.ThreadID, key)
		}
	}
	return result.Threads, result.Total, nil
}

func (r *threadCached) findForYouFollowing(userID uint) ([]model.Thread, error) {
	var threads []model.Thread
	err := preloadThreadDetails(r.db.readDB).
		Joins("JOIN user_relationships ON user_relationships.following_id = threads.user_id").
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN community_members ON community_members.community_id = threads.community_id AND community_members.user_id = ?", userID).
		Where("user_relationships.follower_id = ?", userID).
		Where("threads.created_at > NOW() - INTERVAL '7 days'").
		Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost)}).
		Where("threads.community_id IS NULL OR communities.visibility = ? OR community_members.user_id = ?", model.CommunityVisibilityPublic, userID).
		Order("threads.feed_score DESC").
		Limit(100).
		Find(&threads).Error
	return threads, err
}

func (r *threadCached) findForYouCommunity(userID uint) ([]model.Thread, error) {
	var threads []model.Thread
	err := preloadThreadDetails(r.db.readDB).
		Joins("JOIN community_members ON community_members.community_id = threads.community_id").
		Where("community_members.user_id = ?", userID).
		Where("threads.created_at > NOW() - INTERVAL '7 days'").
		Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost)}).
		Order("threads.feed_score DESC").
		Limit(100).
		Find(&threads).Error
	return threads, err
}

func (r *threadCached) findForYouFaculty(userID uint, facultyID uint) ([]model.Thread, error) {
	var threads []model.Thread
	err := preloadThreadDetails(r.db.readDB).
		Joins("JOIN user_faculties ON user_faculties.user_id = threads.user_id").
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN user_settings ON user_settings.user_id = threads.user_id").
		Where("user_faculties.faculty_id = ?", facultyID).
		Where("threads.user_id != ?", userID).
		Where("threads.created_at > NOW() - INTERVAL '7 days'").
		Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost)}).
		Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
		Where("COALESCE(user_settings.is_private, false) = ?", false).
		Order("threads.feed_score DESC").
		Limit(100).
		Find(&threads).Error
	return threads, err
}

func (r *threadCached) findForYouGlobal(ctx context.Context) ([]model.Thread, error) {
	var threads []model.Thread
	if err := r.cache.Get(ctx, "feed:trending:global", &threads); err == nil {
		return threads, nil
	}
	err := preloadThreadDetails(r.db.readDB).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN user_settings ON user_settings.user_id = threads.user_id").
		Where("threads.created_at > NOW() - INTERVAL '7 days'").
		Where("threads.type = ?", string(model.ThreadTypePost)).
		Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
		Where("COALESCE(user_settings.is_private, false) = ?", false).
		Order("threads.feed_score DESC").
		Limit(100).
		Find(&threads).Error
	if err != nil {
		return nil, err
	}
	_ = r.cache.Set(ctx, "feed:trending:global", threads, 5*time.Minute)
	return threads, nil
}
