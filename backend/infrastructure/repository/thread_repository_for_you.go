package repository

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"gorm.io/gorm"

	communitymodel "github.com/sunibgram/domain/model/community"
	threadmodel "github.com/sunibgram/domain/model/thread"
)

const (
	forYouTrendingTTL = 2 * time.Minute
	forYouSourceTTL   = 60 * time.Second
	forYouOwnTTL      = 30 * time.Second
	forYouSeedTTL     = 10 * time.Minute

	forYouHalflifeHours = 6.0
	forYouBandWidth     = 0.5
	forYouMinScore      = 0.01
	forYouAuthorCap     = 20
	forYouCandidateCap  = 800
	forYouFreshWindow   = 1 * time.Hour
)

const mainFeedReplyInteractionSQL = "(threads.like_count + threads.repost_count + threads.comment_count)"

type forYouCandidate struct {
	thread threadmodel.Thread
	score  float64
}

func forYouFeedTypes(includeReposts bool) []string {
	if includeReposts {
		return []string{
			string(threadmodel.ThreadTypePost),
			string(threadmodel.ThreadTypeRepost),
			string(threadmodel.ThreadTypeReply),
		}
	}
	return []string{
		string(threadmodel.ThreadTypePost),
		string(threadmodel.ThreadTypeReply),
	}
}

func applyMainFeedReplyInteractionFilter(db *gorm.DB) *gorm.DB {
	return db.Where("threads.parent_id IS NULL OR " + mainFeedReplyInteractionSQL + " > 0")
}

func forYouSeedKey(userID uint) string {
	return fmt.Sprintf("feed:foryou:user:%d:seed", userID)
}

func forYouSourceOwnKey(userID uint) string {
	return fmt.Sprintf("feed:foryou:user:%d:src:own", userID)
}

func forYouSourceFollowingKey(userID uint) string {
	return fmt.Sprintf("feed:foryou:user:%d:src:following", userID)
}

func forYouSourceCommunityKey(userID uint) string {
	return fmt.Sprintf("feed:foryou:user:%d:src:community", userID)
}

func forYouNewCountKey(userID, sinceID uint) string {
	return fmt.Sprintf("feed:foryou:user:%d:newcount:%d", userID, sinceID)
}

const forYouSourceGlobalKey = "feed:trending:global"

// resolveSeed returns a stable seed for a user's For You feed. The seed is
// cached for forYouSeedTTL with a sliding window — every read refreshes the
// TTL so an active scrolling session keeps the same ordering. The seed is
// minted fresh when missing (first request, after idle expiry, or after an
// explicit refresh call).
func (r *threadCached) resolveForYouSeed(ctx context.Context, userID uint) int64 {
	var seed int64
	if err := r.cache.Get(ctx, forYouSeedKey(userID), &seed); err == nil && seed != 0 {
		_ = r.cache.Set(ctx, forYouSeedKey(userID), seed, forYouSeedTTL)
		return seed
	}
	seed = time.Now().UnixNano() ^ int64(uint64(userID)*0x9E3779B97F4A7C15)
	if seed == 0 {
		seed = 1
	}
	_ = r.cache.Set(ctx, forYouSeedKey(userID), seed, forYouSeedTTL)
	return seed
}

func (r *threadCached) InvalidateForYouSeed(ctx context.Context, userID uint) {
	_ = r.cache.Delete(ctx, forYouSeedKey(userID))
}

func (r *threadCached) FindForYouFeed(ctx context.Context, userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	seed := r.resolveForYouSeed(ctx, userID)

	merged := make(map[uint]forYouCandidate)
	now := time.Now()
	addCandidates := func(threads []threadmodel.Thread, sourceWeight float64, includeReposts, forceKeep bool) {
		for _, thread := range threads {
			if thread.Type == threadmodel.ThreadTypeRepost && !includeReposts {
				continue
			}
			ageHours := now.Sub(thread.CreatedAt).Hours()
			if ageHours < 0 {
				ageHours = 0
			}
			decay := 1.0 / (1.0 + ageHours/forYouHalflifeHours)
			score := thread.FeedScore * sourceWeight * decay
			if forceKeep {
				// Own threads always surface to the author. Very recent posts
				// (under 1h, before the feed_score scheduler has run) get
				// boosted into the top band so a just-posted thread appears
				// near the top of the author's own feed.
				floor := 0.5
				if ageHours < 1 {
					floor = 5.0
				}
				if score < floor {
					score = floor
				}
			} else if score < forYouMinScore {
				continue
			}
			current, ok := merged[thread.ThreadID]
			if !ok || score > current.score {
				merged[thread.ThreadID] = forYouCandidate{thread: thread, score: score}
			}
		}
	}

	own, err := r.findForYouOwn(ctx, userID)
	if err != nil {
		return nil, nil, 0, err
	}
	addCandidates(own, 1.0, true, true)

	global, err := r.findForYouGlobal(ctx)
	if err != nil {
		return nil, nil, 0, err
	}
	addCandidates(global, 1.0, false, false)

	following, err := r.findForYouFollowing(ctx, userID)
	if err != nil {
		return nil, nil, 0, err
	}
	addCandidates(following, 0.8, true, false)

	community, err := r.findForYouCommunity(ctx, userID)
	if err != nil {
		return nil, nil, 0, err
	}
	addCandidates(community, 0.7, false, false)

	fresh := make([]forYouCandidate, 0)
	ranked := make([]forYouCandidate, 0, len(merged))
	freshCutoff := now.Add(-forYouFreshWindow)
	for _, candidate := range merged {
		if candidate.thread.CreatedAt.After(freshCutoff) {
			fresh = append(fresh, candidate)
		} else {
			ranked = append(ranked, candidate)
		}
	}

	sort.SliceStable(fresh, func(i, j int) bool {
		return fresh[i].thread.CreatedAt.After(fresh[j].thread.CreatedAt)
	})

	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].score == ranked[j].score {
			return ranked[i].thread.CreatedAt.After(ranked[j].thread.CreatedAt)
		}
		return ranked[i].score > ranked[j].score
	})

	rng := rand.New(rand.NewSource(seed))
	ranked = shuffleWithinBands(ranked, rng)

	candidates := make([]forYouCandidate, 0, len(fresh)+len(ranked))
	candidates = append(candidates, fresh...)
	candidates = append(candidates, ranked...)

	if len(candidates) > forYouCandidateCap {
		candidates = candidates[:forYouCandidateCap]
	}

	capped := applyAuthorCap(candidates, forYouAuthorCap)

	total := int64(len(capped))
	offset := (page - 1) * limit
	if offset >= len(capped) {
		return []threadmodel.Thread{}, []*threadmodel.Thread{}, total, nil
	}
	end := offset + limit
	if end > len(capped) {
		end = len(capped)
	}

	pagedThreads := capped[offset:end]
	pagedThreads, parents, err := r.attachVisibleParents(pagedThreads, &userID)
	if err != nil {
		return nil, nil, 0, err
	}
	return pagedThreads, parents, total, nil
}

func shuffleWithinBands(candidates []forYouCandidate, rng *rand.Rand) []forYouCandidate {
	if len(candidates) < 2 {
		return candidates
	}
	bandOf := func(s float64) int64 { return int64(math.Floor(s / forYouBandWidth)) }
	shuffled := make([]forYouCandidate, 0, len(candidates))
	i := 0
	for i < len(candidates) {
		j := i + 1
		b := bandOf(candidates[i].score)
		for j < len(candidates) && bandOf(candidates[j].score) == b {
			j++
		}
		group := make([]forYouCandidate, j-i)
		copy(group, candidates[i:j])
		rng.Shuffle(len(group), func(a, b int) { group[a], group[b] = group[b], group[a] })
		shuffled = append(shuffled, group...)
		i = j
	}
	return shuffled
}

func applyAuthorCap(candidates []forYouCandidate, maxPerAuthor int) []threadmodel.Thread {
	out := make([]threadmodel.Thread, 0, len(candidates))
	counts := make(map[uint]int, len(candidates))
	for _, candidate := range candidates {
		authorID := candidate.thread.UserID
		if candidate.thread.Type == threadmodel.ThreadTypeRepost && candidate.thread.OriginalThread != nil {
			authorID = candidate.thread.OriginalThread.UserID
		}
		if counts[authorID] >= maxPerAuthor {
			continue
		}
		counts[authorID]++
		out = append(out, candidate.thread)
	}
	return out
}

// CountNewSinceID counts visible feed-eligible threads created after sinceID,
// used by the "Show N new posts" banner. Capped at 1000 (UI shows "999+").
func (r *threadCached) CountNewSinceID(ctx context.Context, userID, sinceID uint) (int64, error) {
	key := forYouNewCountKey(userID, sinceID)
	var cached int64
	if err := r.cache.Get(ctx, key, &cached); err == nil {
		return cached, nil
	}
	var count int64
	requester := userID
	query := applyThreadVisibility(r.db.readDB.WithContext(ctx).Model(&threadmodel.Thread{}), &requester).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN community_members count_new_members ON count_new_members.community_id = threads.community_id AND count_new_members.user_id = ?", userID).
		Where("threads.thread_id > ?", sinceID).
		Where("threads.type IN ?", forYouFeedTypes(true)).
		Where("threads.community_id IS NULL OR communities.visibility = ? OR count_new_members.user_id = ?", communitymodel.CommunityVisibilityPublic, userID)
	query = applyMainFeedReplyInteractionFilter(query)
	if err := query.Limit(1001).Count(&count).Error; err != nil {
		return 0, err
	}
	if count > 1000 {
		count = 1000
	}
	if err := r.cache.Set(ctx, key, count, 30*time.Second); err == nil {
		r.trackForYouKey(ctx, userID, key)
	}
	return count, nil
}

func (r *threadCached) findForYouOwn(ctx context.Context, userID uint) ([]threadmodel.Thread, error) {
	key := forYouSourceOwnKey(userID)
	var cached []threadmodel.Thread
	if err := r.cache.Get(ctx, key, &cached); err == nil {
		return cached, nil
	}
	var threads []threadmodel.Thread
	query := preloadThreadDetails(r.db.readDB.WithContext(ctx)).
		Where("threads.user_id = ?", userID).
		Where("threads.created_at > NOW() - INTERVAL '7 days'").
		Where("threads.type IN ?", forYouFeedTypes(true))
	query = applyMainFeedReplyInteractionFilter(query)
	if err := query.
		Order("threads.feed_score DESC").
		Limit(100).
		Find(&threads).Error; err != nil {
		return nil, err
	}
	if err := r.cache.Set(ctx, key, threads, forYouOwnTTL); err == nil {
		r.trackForYouKey(ctx, userID, key)
	}
	return threads, nil
}

func (r *threadCached) findForYouFollowing(ctx context.Context, userID uint) ([]threadmodel.Thread, error) {
	key := forYouSourceFollowingKey(userID)
	var cached []threadmodel.Thread
	if err := r.cache.Get(ctx, key, &cached); err == nil {
		return cached, nil
	}
	var threads []threadmodel.Thread
	query := preloadThreadDetails(r.db.readDB.WithContext(ctx)).
		Joins("JOIN user_relationships ON user_relationships.following_id = threads.user_id").
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN community_members ON community_members.community_id = threads.community_id AND community_members.user_id = ?", userID).
		Where("user_relationships.follower_id = ?", userID).
		Where("threads.created_at > NOW() - INTERVAL '7 days'").
		Where("threads.type IN ?", forYouFeedTypes(true)).
		Where("threads.community_id IS NULL OR communities.visibility = ? OR community_members.user_id = ?", communitymodel.CommunityVisibilityPublic, userID)
	query = applyMainFeedReplyInteractionFilter(query)
	if err := query.
		Order("threads.feed_score DESC").
		Limit(100).
		Find(&threads).Error; err != nil {
		return nil, err
	}
	if err := r.cache.Set(ctx, key, threads, forYouSourceTTL); err == nil {
		r.trackForYouKey(ctx, userID, key)
	}
	return threads, nil
}

func (r *threadCached) findForYouCommunity(ctx context.Context, userID uint) ([]threadmodel.Thread, error) {
	key := forYouSourceCommunityKey(userID)
	var cached []threadmodel.Thread
	if err := r.cache.Get(ctx, key, &cached); err == nil {
		return cached, nil
	}
	var threads []threadmodel.Thread
	query := preloadThreadDetails(r.db.readDB.WithContext(ctx)).
		Joins("JOIN community_members ON community_members.community_id = threads.community_id").
		Where("community_members.user_id = ?", userID).
		Where("threads.created_at > NOW() - INTERVAL '7 days'").
		Where("threads.type IN ?", forYouFeedTypes(true))
	query = applyMainFeedReplyInteractionFilter(query)
	if err := query.
		Order("threads.feed_score DESC").
		Limit(100).
		Find(&threads).Error; err != nil {
		return nil, err
	}
	if err := r.cache.Set(ctx, key, threads, forYouSourceTTL); err == nil {
		r.trackForYouKey(ctx, userID, key)
	}
	return threads, nil
}

func (r *threadCached) findForYouGlobal(ctx context.Context) ([]threadmodel.Thread, error) {
	var threads []threadmodel.Thread
	if err := r.cache.Get(ctx, forYouSourceGlobalKey, &threads); err == nil {
		return threads, nil
	}
	query := preloadThreadDetails(r.db.readDB.WithContext(ctx)).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN user_settings ON user_settings.user_id = threads.user_id").
		Where("threads.created_at > NOW() - INTERVAL '7 days'").
		Where("threads.type IN ?", forYouFeedTypes(false)).
		Where("threads.community_id IS NULL OR communities.visibility = ?", communitymodel.CommunityVisibilityPublic).
		Where("COALESCE(user_settings.is_private, false) = ?", false)
	query = applyMainFeedReplyInteractionFilter(query)
	if err := query.
		Order("threads.feed_score DESC").
		Limit(100).
		Find(&threads).Error; err != nil {
		return nil, err
	}
	_ = r.cache.Set(ctx, forYouSourceGlobalKey, threads, forYouTrendingTTL)
	return threads, nil
}
