package threadrepo

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"time"

	"gorm.io/gorm"

	communitymodel "github.com/sunibgram/domain/model/community"
	threadmodel "github.com/sunibgram/domain/model/thread"
	"github.com/sunibgram/infrastructure/cache"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
)

const forYouRecencyWindow = 7 * 24 * time.Hour

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
	forYouFreshWindow   = 15 * time.Minute
	forYouFreshCap      = 5
	forYouFreshFloor    = 0.1

	// forYouReplyWeight demotes reply candidates inside the ranking. Replies
	// are surfaced for engagement-context but should not dominate the feed
	// over top-level posts at the same raw feed_score. 0.5 = halve a reply's
	// effective contribution. Applied in addCandidates AFTER the per-source
	// weight, BEFORE decay/floor logic.
	forYouReplyWeight = 0.5

	// forYouFreshCapRatio caps the fresh band size as a fraction of the
	// requested page limit. With limit=20 → 4 fresh slots, limit=10 → 2.
	// The hard ceiling forYouFreshCap still applies, so the fresh band never
	// dominates a desktop limit=50 request either.
	forYouFreshCapRatio = 0.2
)

// forYouEffectiveFreshCap returns the per-request fresh-band cap, clamped
// between 1 and forYouFreshCap. Proportional to page limit so mobile clients
// (limit=10) get a smaller fresh slice than desktop (limit=20+).
func forYouEffectiveFreshCap(limit int) int {
	value := int(math.Ceil(float64(limit) * forYouFreshCapRatio))
	if value < 1 {
		value = 1
	}
	if value > forYouFreshCap {
		value = forYouFreshCap
	}
	return value
}

// authorCapTiers defines position-aware author caps. The feed is scanned in
// order; the cap at position i is the smallest cap whose threshold is >= i+1.
// Effect: top-of-feed gets variety (max 5 from one author in the first 10
// positions, 10 in the first 25, 20 across the whole result). Prevents a
// single chatty author from filling the visible top of feed.
var authorCapTiers = []struct {
	upTo int // position threshold (1-based, inclusive)
	cap  int // max threads from one author within this prefix
}{
	{10, 5},
	{25, 10},
	{forYouCandidateCap, forYouAuthorCap},
}

// forYouRecencyBounds computes the two timestamps used by every source query
// (the 7-day recency cutoff and the fresh-window cutoff). Binding these as
// parameters instead of inlining NOW() lets Postgres reuse the prepared-statement
// plan across requests and keeps pgbouncer / GORM caches effective.
func forYouRecencyBounds() (recencyCutoff, freshCutoff time.Time) {
	now := time.Now()
	return now.Add(-forYouRecencyWindow), now.Add(-forYouFreshWindow)
}

// forYouFreshFirstOrder puts threads inside the fresh window at the top of
// each source query so feed_score=0 fresh posts are not cut by the LIMIT 100.
// Used by every per-source query; visibility WHERE clauses are unchanged.
func forYouFreshFirstOrder(freshCutoff time.Time) interface{} {
	return gorm.Expr("(threads.created_at > ?) DESC, threads.feed_score DESC", freshCutoff)
}

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

// Per-source cache keys carry a `:v{N}` suffix where N is the user's current
// For You cache version (see infrarepoShared.ForYouVersionKey). Mutations INCR
// the version atomically; readers GET it once and bake it into every source
// key built for this request. Old `:v{N-1}` entries orphan and TTL-expire.
func forYouSourceOwnKey(userID uint, version int64) string {
	return fmt.Sprintf("feed:foryou:user:%d:src:own:v%d", userID, version)
}

func forYouSourceFollowingKey(userID uint, version int64) string {
	return fmt.Sprintf("feed:foryou:user:%d:src:following:v%d", userID, version)
}

func forYouSourceCommunityKey(userID uint, version int64) string {
	return fmt.Sprintf("feed:foryou:user:%d:src:community:v%d", userID, version)
}

func forYouNewCountKey(userID, sinceID uint, version int64) string {
	return fmt.Sprintf("feed:foryou:user:%d:newcount:%d:v%d", userID, sinceID, version)
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
		r.forYouMetrics.CacheHitSeed.Add(1)
		_ = r.cache.Set(ctx, forYouSeedKey(userID), seed, forYouSeedTTL)
		return seed
	}
	r.forYouMetrics.CacheMissSeed.Add(1)
	seed = time.Now().UnixNano() ^ int64(uint64(userID)*0x9E3779B97F4A7C15)
	if seed == 0 {
		// Defensive: zero seed would collapse every user to identical RNG
		// state. The XOR collision is astronomically rare but logged so we can
		// confirm it never happens in practice.
		log.Printf("metric=feed_seed_collision user_id=%d", userID)
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
	// Single GET of the user's For You cache version. Every source cache key
	// built for this request carries this version; a concurrent mutation that
	// INCRs the version mid-request just orphans the soon-to-be-written keys.
	version := infrarepoShared.ReadForYouVersion(ctx, r.cache, userID)

	merged := make(map[uint]forYouCandidate)
	now := time.Now()
	addCandidates := func(threads []threadmodel.Thread, sourceWeight float64, includeReposts, forceKeep bool) {
		for _, thread := range threads {
			if thread.Type == threadmodel.ThreadTypeRepost && !includeReposts {
				continue
			}
			age := now.Sub(thread.CreatedAt)
			if age < 0 {
				age = 0
			}
			ageHours := age.Hours()
			isFresh := age < forYouFreshWindow
			decay := 1.0 / (1.0 + ageHours/forYouHalflifeHours)
			effectiveSourceWeight := sourceWeight
			// Replies inherit their parent's engagement context but should not
			// dominate the feed over top-level posts at equivalent feed_score.
			// Demote with a fixed multiplier (#3.6).
			if thread.Type == threadmodel.ThreadTypeReply {
				effectiveSourceWeight *= forYouReplyWeight
			}
			score := thread.FeedScore * effectiveSourceWeight * decay
			if forceKeep {
				// Own threads always surface to the author. Fresh posts get a
				// higher floor so a just-posted thread appears at the top of
				// the author's own feed before the feed_score scheduler ticks.
				floor := 0.5
				if isFresh {
					floor = 5.0
				}
				if score < floor {
					score = floor
				}
			} else if isFresh {
				// Fresh non-own thread (within forYouFreshWindow) needs a floor
				// to clear the minScore filter; the post-merge fresh/ranked
				// split places it in the fresh band, sorted by created_at,
				// regardless of this floor value.
				if score < forYouFreshFloor {
					score = forYouFreshFloor
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

	var (
		ownThreads, globalThreads, followingThreads, communityThreads []threadmodel.Thread
		ownErr, globalErr, followingErr, communityErr                 error
		wg                                                            sync.WaitGroup
	)
	wg.Add(4)
	go func() {
		defer wg.Done()
		ownThreads, ownErr = r.findForYouOwn(ctx, userID, version)
	}()
	go func() {
		defer wg.Done()
		globalThreads, globalErr = r.findForYouGlobal(ctx)
	}()
	go func() {
		defer wg.Done()
		followingThreads, followingErr = r.findForYouFollowing(ctx, userID, version)
	}()
	go func() {
		defer wg.Done()
		communityThreads, communityErr = r.findForYouCommunity(ctx, userID, version)
	}()
	wg.Wait()
	if ownErr != nil {
		return nil, nil, 0, ownErr
	}
	if globalErr != nil {
		return nil, nil, 0, globalErr
	}
	if followingErr != nil {
		return nil, nil, 0, followingErr
	}
	if communityErr != nil {
		return nil, nil, 0, communityErr
	}
	addCandidates(ownThreads, 1.0, true, true)
	addCandidates(globalThreads, 1.0, false, false)
	addCandidates(followingThreads, 0.8, true, false)
	addCandidates(communityThreads, 0.7, false, false)

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

	// Cap fresh band size proportional to page limit (#3.5). Mobile (limit=10)
	// gets a smaller fresh slice than desktop (limit=20+). Overflow spills
	// into the ranked band so the feed does not become a pure "Latest" tab
	// when the network is busy.
	freshCap := forYouEffectiveFreshCap(limit)
	if len(fresh) > freshCap {
		ranked = append(ranked, fresh[freshCap:]...)
		fresh = fresh[:freshCap]
	}

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

	capped := applyAuthorCap(candidates)

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

// applyAuthorCap implements position-aware author caps (#3.3). The feed is
// scanned in already-sorted order; at each kept-output position the function
// looks up the active cap for that position from authorCapTiers and rejects
// the candidate if its author has already filled the slot count allowed in
// any prefix that includes this position. Top of feed therefore has stricter
// variety than the tail.
//
// For repost candidates the cap counts against the ORIGINAL author so a user
// cannot dominate by reposting their own thread alongside the post itself.
func applyAuthorCap(candidates []forYouCandidate) []threadmodel.Thread {
	out := make([]threadmodel.Thread, 0, len(candidates))
	counts := make(map[uint]int, len(candidates))
	for _, candidate := range candidates {
		authorID := candidate.thread.UserID
		if candidate.thread.Type == threadmodel.ThreadTypeRepost && candidate.thread.OriginalThread != nil {
			authorID = candidate.thread.OriginalThread.UserID
		}
		// position is the rank the candidate WOULD occupy if accepted (1-based).
		position := len(out) + 1
		// The active cap is the smallest cap among the tiers whose upTo bound
		// includes this position. Iterating in upTo-ascending order means the
		// first matching tier is the strictest, which is correct.
		allowed := forYouAuthorCap
		for _, tier := range authorCapTiers {
			if position <= tier.upTo {
				allowed = tier.cap
				break
			}
		}
		if counts[authorID] >= allowed {
			continue
		}
		counts[authorID]++
		out = append(out, candidate.thread)
	}
	return out
}

// CountNewSinceID counts visible feed-eligible threads created after sinceID,
// used by the "Show N new posts" banner. Capped at 200 (UI shows "200+").
//
// Cached per (userID, sinceID, forYouVersion) with a 15 s TTL. A new post by
// any author the viewer can see triggers IncrForYouVersion(viewer), which
// orphans the cached count atomically — zero staleness, no per-since_id key
// fanout on invalidation.
func (r *threadCached) CountNewSinceID(ctx context.Context, userID, sinceID uint) (int64, error) {
	version := infrarepoShared.ReadForYouVersion(ctx, r.cache, userID)
	key := forYouNewCountKey(userID, sinceID, version)
	var cached int64
	if err := r.cache.Get(ctx, key, &cached); err == nil {
		r.forYouMetrics.CacheHitNewCount.Add(1)
		return cached, nil
	}
	r.forYouMetrics.CacheMissNewCount.Add(1)
	requester := userID
	query := ApplyThreadVisibility(r.db.readDB.WithContext(ctx).Model(&threadmodel.Thread{}), &requester).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN community_members count_new_members ON count_new_members.community_id = threads.community_id AND count_new_members.user_id = ?", userID).
		Where("threads.thread_id > ?", sinceID).
		Where("threads.type IN ?", forYouFeedTypes(true)).
		Where("threads.community_id IS NULL OR communities.visibility = ? OR count_new_members.user_id = ?", communitymodel.CommunityVisibilityPublic, userID)
	query = applyMainFeedReplyInteractionFilter(query)
	// GORM's Count() strips Limit, which would force a full-table scan on every
	// banner poll. Pluck honors Limit, so the DB returns at most 201 rows and we
	// derive the count from the slice. 201-elem uint slice ≈ 1.6 KB — negligible
	// compared to the saved scan when the matching set is huge.
	var ids []uint
	if err := query.Limit(201).Pluck("threads.thread_id", &ids).Error; err != nil {
		return 0, err
	}
	count := int64(len(ids))
	if count > 200 {
		count = 200
	}
	_ = r.cache.Set(ctx, key, count, 15*time.Second)
	return count, nil
}

// FindNewSinceID returns visible feed-eligible threads created after sinceID,
// newest-first by thread_id, capped at `limit`. Powers the "Show N new posts"
// banner click — frontend prepends these onto the existing feed instead of
// triggering a full /feed/refresh reshuffle (which can surface older threads
// with lower IDs and re-trigger the banner). Not cached: the result depends on
// sinceID, the call is rare (one per banner click), and stale rows here would
// surface as missing/duplicate posts that the client cannot self-correct.
func (r *threadCached) FindNewSinceID(ctx context.Context, userID, sinceID uint, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	requester := userID
	var threads []threadmodel.Thread
	query := ApplyThreadVisibility(PreloadThreadDetails(r.db.readDB.WithContext(ctx)), &requester).
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN community_members find_new_members ON find_new_members.community_id = threads.community_id AND find_new_members.user_id = ?", userID).
		Where("threads.thread_id > ?", sinceID).
		Where("threads.type IN ?", forYouFeedTypes(true)).
		Where("threads.community_id IS NULL OR communities.visibility = ? OR find_new_members.user_id = ?", communitymodel.CommunityVisibilityPublic, userID)
	query = applyMainFeedReplyInteractionFilter(query)
	if err := query.Order("threads.thread_id DESC").Limit(limit).Find(&threads).Error; err != nil {
		return nil, nil, err
	}
	return r.AttachReplyParents(threads, &requester)
}

func (r *threadCached) findForYouOwn(ctx context.Context, userID uint, version int64) ([]threadmodel.Thread, error) {
	key := forYouSourceOwnKey(userID, version)
	var cached []threadmodel.Thread
	if err := r.cache.Get(ctx, key, &cached); err == nil {
		r.forYouMetrics.CacheHitOwn.Add(1)
		return cached, nil
	}
	r.forYouMetrics.CacheMissOwn.Add(1)
	result, err, _ := r.forYouSingleflight.Do(key, func() (interface{}, error) {
		// Re-check cache inside the singleflight gate: a sibling caller may
		// have populated it while we were queued.
		var doubleCheck []threadmodel.Thread
		if err := r.cache.Get(ctx, key, &doubleCheck); err == nil {
			return doubleCheck, nil
		}
		started := time.Now()
		recencyCutoff, freshCutoff := forYouRecencyBounds()
		var threads []threadmodel.Thread
		query := PreloadThreadDetails(r.db.readDB.WithContext(ctx)).
			Where("threads.user_id = ?", userID).
			Where("threads.created_at > ?", recencyCutoff).
			Where("threads.type IN ?", forYouFeedTypes(true))
		query = applyMainFeedReplyInteractionFilter(query)
		err := query.
			Order(forYouFreshFirstOrder(freshCutoff)).
			Limit(100).
			Find(&threads).Error
		logSourceLatency("own", started, len(threads), err)
		if err != nil {
			return nil, err
		}
		_ = r.cache.Set(ctx, key, threads, forYouOwnTTL)
		return threads, nil
	})
	if err != nil {
		return nil, err
	}
	return result.([]threadmodel.Thread), nil
}

func (r *threadCached) findForYouFollowing(ctx context.Context, userID uint, version int64) ([]threadmodel.Thread, error) {
	key := forYouSourceFollowingKey(userID, version)
	var cached []threadmodel.Thread
	if err := r.cache.Get(ctx, key, &cached); err == nil {
		r.forYouMetrics.CacheHitFollowing.Add(1)
		return cached, nil
	}
	r.forYouMetrics.CacheMissFollowing.Add(1)
	result, err, _ := r.forYouSingleflight.Do(key, func() (interface{}, error) {
		var doubleCheck []threadmodel.Thread
		if err := r.cache.Get(ctx, key, &doubleCheck); err == nil {
			return doubleCheck, nil
		}
		started := time.Now()
		recencyCutoff, freshCutoff := forYouRecencyBounds()
		var threads []threadmodel.Thread
		query := PreloadThreadDetails(r.db.readDB.WithContext(ctx)).
			Joins("JOIN user_relationships ON user_relationships.following_id = threads.user_id").
			Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
			Joins("LEFT JOIN community_members ON community_members.community_id = threads.community_id AND community_members.user_id = ?", userID).
			Where("user_relationships.follower_id = ?", userID).
			Where("threads.created_at > ?", recencyCutoff).
			Where("threads.type IN ?", forYouFeedTypes(true)).
			Where("threads.community_id IS NULL OR communities.visibility = ? OR community_members.user_id = ?", communitymodel.CommunityVisibilityPublic, userID)
		query = applyMainFeedReplyInteractionFilter(query)
		err := query.
			Order(forYouFreshFirstOrder(freshCutoff)).
			Limit(100).
			Find(&threads).Error
		logSourceLatency("following", started, len(threads), err)
		if err != nil {
			return nil, err
		}
		_ = r.cache.Set(ctx, key, threads, forYouSourceTTL)
		return threads, nil
	})
	if err != nil {
		return nil, err
	}
	return result.([]threadmodel.Thread), nil
}

func (r *threadCached) findForYouCommunity(ctx context.Context, userID uint, version int64) ([]threadmodel.Thread, error) {
	key := forYouSourceCommunityKey(userID, version)
	var cached []threadmodel.Thread
	if err := r.cache.Get(ctx, key, &cached); err == nil {
		r.forYouMetrics.CacheHitCommunity.Add(1)
		return cached, nil
	}
	r.forYouMetrics.CacheMissCommunity.Add(1)
	result, err, _ := r.forYouSingleflight.Do(key, func() (interface{}, error) {
		var doubleCheck []threadmodel.Thread
		if err := r.cache.Get(ctx, key, &doubleCheck); err == nil {
			return doubleCheck, nil
		}
		started := time.Now()
		recencyCutoff, freshCutoff := forYouRecencyBounds()
		var threads []threadmodel.Thread
		query := PreloadThreadDetails(r.db.readDB.WithContext(ctx)).
			Joins("JOIN community_members ON community_members.community_id = threads.community_id").
			Where("community_members.user_id = ?", userID).
			Where("threads.created_at > ?", recencyCutoff).
			Where("threads.type IN ?", forYouFeedTypes(true))
		query = applyMainFeedReplyInteractionFilter(query)
		err := query.
			Order(forYouFreshFirstOrder(freshCutoff)).
			Limit(100).
			Find(&threads).Error
		logSourceLatency("community", started, len(threads), err)
		if err != nil {
			return nil, err
		}
		_ = r.cache.Set(ctx, key, threads, forYouSourceTTL)
		return threads, nil
	})
	if err != nil {
		return nil, err
	}
	return result.([]threadmodel.Thread), nil
}

func (r *threadCached) findForYouGlobal(ctx context.Context) ([]threadmodel.Thread, error) {
	// Cache shape: Redis ZSET keyed by thread_id (member, string-encoded) →
	// feed_score (score). Scheduler writes this every minute. We read the top
	// 100 IDs+scores and hydrate full thread objects via the per-thread cache
	// layer (hydrateThreads → thread:id:<id>), so the global cache itself
	// stays small (~3 KB) while preserving the full DTO shape callers expect.
	members, err := r.cache.ZRevRangeWithScores(ctx, forYouSourceGlobalKey, 0, 99)
	if err == nil && len(members) > 0 {
		r.forYouMetrics.CacheHitGlobal.Add(1)
		return r.hydrateGlobalTrending(ctx, members)
	}
	r.forYouMetrics.CacheMissGlobal.Add(1)
	// The scheduler proactively warms this key every minute; the lazy path
	// here only runs during scheduler downtime or startup. Without singleflight
	// every concurrent feed request stampedes the same global query.
	result, err, _ := r.forYouSingleflight.Do(forYouSourceGlobalKey, func() (interface{}, error) {
		// Re-check inside the singleflight gate.
		if doubleCheck, err := r.cache.ZRevRangeWithScores(ctx, forYouSourceGlobalKey, 0, 99); err == nil && len(doubleCheck) > 0 {
			return r.hydrateGlobalTrending(ctx, doubleCheck)
		}
		started := time.Now()
		recencyCutoff, freshCutoff := forYouRecencyBounds()
		var threads []threadmodel.Thread
		query := PreloadThreadDetails(r.db.readDB.WithContext(ctx)).
			Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
			Joins("LEFT JOIN user_settings ON user_settings.user_id = threads.user_id").
			Where("threads.created_at > ?", recencyCutoff).
			Where("threads.type IN ?", forYouFeedTypes(false)).
			Where("threads.community_id IS NULL OR communities.visibility = ?", communitymodel.CommunityVisibilityPublic).
			Where("COALESCE(user_settings.is_private, false) = ?", false)
		query = applyMainFeedReplyInteractionFilter(query)
		err := query.
			Order(forYouFreshFirstOrder(freshCutoff)).
			Limit(100).
			Find(&threads).Error
		logSourceLatency("global", started, len(threads), err)
		if err != nil {
			return nil, err
		}
		// Populate both the ZSET (for next-time fast path) and the per-thread
		// caches that hydrateGlobalTrending depends on.
		zMembers := make([]cache.ZSetMember, 0, len(threads))
		entries := make(map[string]any, len(threads))
		for i := range threads {
			t := threads[i]
			entries[ThreadKey(t.ThreadID)] = t
			zMembers = append(zMembers, cache.ZSetMember{
				Member: strconv.FormatUint(uint64(t.ThreadID), 10),
				Score:  t.FeedScore,
			})
		}
		_ = r.cache.MSet(ctx, entries, r.ttl)
		_ = r.cache.ZAddReplace(ctx, forYouSourceGlobalKey, zMembers, forYouTrendingTTL)
		return threads, nil
	})
	if err != nil {
		return nil, err
	}
	return result.([]threadmodel.Thread), nil
}

// hydrateGlobalTrending takes the ZSET snapshot (ID+score pairs) and returns
// fully-preloaded threads in the same score-desc order. Per-thread cache
// misses are batched into one DB query via hydrateThreads. Scores from the
// ZSET overwrite each thread's feed_score so the read path sees the same
// stable ranking the scheduler captured at write time.
func (r *threadCached) hydrateGlobalTrending(ctx context.Context, members []cache.ZSetMember) ([]threadmodel.Thread, error) {
	if len(members) == 0 {
		return nil, nil
	}
	ids := make([]uint, 0, len(members))
	scores := make(map[uint]float64, len(members))
	for _, m := range members {
		id, err := strconv.ParseUint(m.Member, 10, 64)
		if err != nil {
			continue
		}
		uid := uint(id)
		ids = append(ids, uid)
		scores[uid] = m.Score
	}
	threads, err := r.hydrateThreads(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range threads {
		if s, ok := scores[threads[i].ThreadID]; ok {
			threads[i].FeedScore = s
		}
	}
	return threads, nil
}
