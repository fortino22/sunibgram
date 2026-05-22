package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	dtoutils "github.com/sunibgram/dto/utils"
	"github.com/sunibgram/infrastructure/cache"
)

const avatarCacheTTL = time.Hour

type avatarCacheProvider interface {
	AvatarCacheRepository() repository.ICacheRepository
}

type FeedBuilder struct {
	Media          service.IThreadMediaService
	LikeSvc        service.ILikeService
	BookmarkSvc    service.IBookmarkService
	ThreadSvc      service.IThreadService
	PinSvc         service.IThreadPinService
	PollSvc        service.IThreadPollService
	ProfileSvc     service.IUserProfileService
	Cache          *cache.Cache
	ViewerID       *uint
	avatarCache    map[uint]string
	isPremiumCache map[uint]bool
	facultyCache   map[uint]*dto.FeedAuthorFaculty
	flagsLoaded    bool
	liked          map[uint]bool
	bookmarked     map[uint]bool
	reposted       map[uint]bool
	pinned         map[uint]bool
	pollsLoaded    bool
	pollsByThread  map[uint]*model.ThreadPoll
	pollUserVotes  map[uint]uint
}

func avatarCacheKey(userID uint) string {
	return fmt.Sprintf("user:%d:avatar", userID)
}

func (fb FeedBuilder) ToMediaResponses(media []model.ThreadMedia) []dto.ThreadMediaResponse {
	result := make([]dto.ThreadMediaResponse, len(media))
	for i, m := range media {
		result[i] = dto.ThreadMediaResponse{MediaID: m.MediaID, URL: fb.Media.GetURL(m.Bucket, m.ObjectName)}
	}
	return result
}

func CommunityField(t *model.Thread) *dto.FeedCommunity {
	if t.CommunityID != nil && t.Community != nil {
		return &dto.FeedCommunity{ID: t.Community.CommunityID, Name: t.Community.Name}
	}
	return nil
}

func EventFields(t *model.Thread) (*string, *string) {
	if len(t.ThreadEvents) == 0 {
		return nil, nil
	}
	event := t.ThreadEvents[0].Event
	if event.ID == "" {
		return nil, nil
	}
	return &event.ID, &event.Title
}

func (fb *FeedBuilder) avatarCacheRepo() repository.ICacheRepository {
	provider, ok := fb.ProfileSvc.(avatarCacheProvider)
	if !ok {
		return nil
	}
	return provider.AvatarCacheRepository()
}

// LoadAvatarsForUserIDs batch-fetches avatar URLs, is_premium, and faculty for the given user IDs.
// Avatar and profile_meta caches are independent (different keys, different TTLs) and resolved
// in separate passes so a hot avatar cache does not mask a stale or missing profile_meta entry.
func (fb *FeedBuilder) LoadAvatarsForUserIDs(userIDs []uint) {
	if fb.ProfileSvc == nil || len(userIDs) == 0 {
		return
	}
	if fb.avatarCache == nil {
		fb.avatarCache = make(map[uint]string)
	}
	if fb.isPremiumCache == nil {
		fb.isPremiumCache = make(map[uint]bool)
	}
	if fb.facultyCache == nil {
		fb.facultyCache = make(map[uint]*dto.FeedAuthorFaculty)
	}
	seen := make(map[uint]struct{}, len(userIDs))
	needed := make([]uint, 0, len(userIDs))
	for _, id := range userIDs {
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		needed = append(needed, id)
	}
	if len(needed) == 0 {
		return
	}
	ctx := context.Background()
	cacheRepo := fb.avatarCacheRepo()

	fb.resolveAvatars(ctx, cacheRepo, needed)
	fb.resolveProfileMeta(needed)
}

func (fb *FeedBuilder) resolveAvatars(ctx context.Context, cacheRepo repository.ICacheRepository, needed []uint) {
	avatarMisses := make([]uint, 0, len(needed))
	pending := make([]uint, 0, len(needed))
	pendingKeys := make([]string, 0, len(needed))
	for _, id := range needed {
		if _, cached := fb.avatarCache[id]; cached {
			continue
		}
		if cacheRepo == nil {
			avatarMisses = append(avatarMisses, id)
			continue
		}
		pending = append(pending, id)
		pendingKeys = append(pendingKeys, avatarCacheKey(id))
	}
	if len(pending) > 0 {
		hits, err := cacheRepo.MGet(ctx, pendingKeys)
		if err != nil {
			avatarMisses = append(avatarMisses, pending...)
		} else {
			for i, id := range pending {
				raw, ok := hits[pendingKeys[i]]
				if !ok {
					avatarMisses = append(avatarMisses, id)
					continue
				}
				var avatarURL string
				if err := json.Unmarshal(raw, &avatarURL); err != nil {
					avatarMisses = append(avatarMisses, id)
					continue
				}
				fb.avatarCache[id] = avatarURL
			}
		}
	}
	if len(avatarMisses) == 0 {
		return
	}
	profiles, err := fb.ProfileSvc.GetProfiles(avatarMisses)
	if err != nil {
		return
	}
	missed := make(map[uint]struct{}, len(avatarMisses))
	for _, id := range avatarMisses {
		missed[id] = struct{}{}
	}
	writes := make(map[string]any, len(avatarMisses))
	for _, p := range profiles {
		delete(missed, p.UserID)
		avatarURL := ""
		if p.AvatarBucket != "" && p.AvatarObjectName != "" {
			avatarURL = fb.ProfileSvc.GetURL(p.AvatarBucket, p.AvatarObjectName)
		}
		fb.avatarCache[p.UserID] = avatarURL
		if cacheRepo != nil {
			writes[avatarCacheKey(p.UserID)] = avatarURL
		}
	}
	for id := range missed {
		fb.avatarCache[id] = ""
		if cacheRepo != nil {
			writes[avatarCacheKey(id)] = ""
		}
	}
	if cacheRepo != nil && len(writes) > 0 {
		_ = cacheRepo.MSet(ctx, writes, avatarCacheTTL)
	}
}

func (fb *FeedBuilder) resolveProfileMeta(needed []uint) {
	metaMisses := fb.loadProfileMetaFromCache(needed)
	if len(metaMisses) == 0 {
		return
	}
	users, errUsers := fb.ProfileSvc.GetUsers(metaMisses)
	if errUsers == nil {
		for _, u := range users {
			fb.isPremiumCache[u.UserID] = u.IsPremium
		}
	}
	faculties, errFac := fb.ProfileSvc.GetUserFaculties(metaMisses)
	if errFac == nil {
		for _, id := range metaMisses {
			if _, ok := fb.facultyCache[id]; !ok {
				fb.facultyCache[id] = nil
			}
		}
		for _, uf := range faculties {
			fb.facultyCache[uf.UserID] = &dto.FeedAuthorFaculty{
				FacultyID:   uf.FacultyID,
				FacultyName: uf.Faculty.FacultyName,
			}
		}
	}
	if errUsers == nil && errFac == nil {
		fb.writeProfileMetaToCache(metaMisses)
	}
}

const profileMetaTTL = 24 * time.Hour

type profileMetaCacheEntry struct {
	IsPremium   bool   `json:"is_premium"`
	FacultyID   uint   `json:"faculty_id,omitempty"`
	FacultyName string `json:"faculty_name,omitempty"`
	HasFaculty  bool   `json:"has_faculty"`
}

// ProfileMetaCacheKey is the Redis key for cached is_premium + faculty per user.
// Exported so repository invalidation paths can target it.
func ProfileMetaCacheKey(userID uint) string {
	return fmt.Sprintf("user:meta:%d", userID)
}

func (fb *FeedBuilder) loadProfileMetaFromCache(userIDs []uint) (missing []uint) {
	if fb.Cache == nil {
		missing = make([]uint, 0, len(userIDs))
		for _, id := range userIDs {
			if _, ok := fb.isPremiumCache[id]; ok {
				continue
			}
			missing = append(missing, id)
		}
		return missing
	}
	ctx := context.Background()
	pending := make([]uint, 0, len(userIDs))
	pendingKeys := make([]string, 0, len(userIDs))
	for _, id := range userIDs {
		if _, ok := fb.isPremiumCache[id]; ok {
			continue
		}
		pending = append(pending, id)
		pendingKeys = append(pendingKeys, ProfileMetaCacheKey(id))
	}
	if len(pending) == 0 {
		return nil
	}
	hits, err := fb.Cache.MGet(ctx, pendingKeys)
	if err != nil {
		return pending
	}
	for i, id := range pending {
		raw, ok := hits[pendingKeys[i]]
		if !ok {
			missing = append(missing, id)
			continue
		}
		var entry profileMetaCacheEntry
		if err := json.Unmarshal(raw, &entry); err != nil {
			missing = append(missing, id)
			continue
		}
		fb.isPremiumCache[id] = entry.IsPremium
		if entry.HasFaculty {
			fb.facultyCache[id] = &dto.FeedAuthorFaculty{
				FacultyID:   entry.FacultyID,
				FacultyName: entry.FacultyName,
			}
		} else {
			fb.facultyCache[id] = nil
		}
	}
	return missing
}

func (fb *FeedBuilder) writeProfileMetaToCache(userIDs []uint) {
	if fb.Cache == nil {
		return
	}
	ctx := context.Background()
	entries := make(map[string]any, len(userIDs))
	for _, id := range userIDs {
		entry := profileMetaCacheEntry{IsPremium: fb.isPremiumCache[id]}
		if f, ok := fb.facultyCache[id]; ok && f != nil {
			entry.HasFaculty = true
			entry.FacultyID = f.FacultyID
			entry.FacultyName = f.FacultyName
		}
		entries[ProfileMetaCacheKey(id)] = entry
	}
	_ = fb.Cache.MSet(ctx, entries, profileMetaTTL)
}

// AuthorField builds a FeedAuthor from a User.
// is_premium and faculty come from FeedBuilder caches (loaded fresh from DB via LoadAvatarsForUserIDs)
// rather than the embedded User struct, which may be stale from a thread list cache.
func (fb FeedBuilder) AuthorField(u model.User) dto.FeedAuthor {
	isPremium := u.IsPremium
	if v, ok := fb.isPremiumCache[u.UserID]; ok {
		isPremium = v
	}
	a := dto.FeedAuthor{
		ID:        u.UserID,
		Name:      u.Name,
		Username:  u.UserName,
		IsPremium: isPremium,
		AvatarURL: fb.avatarCache[u.UserID],
	}
	if f, ok := fb.facultyCache[u.UserID]; ok {
		a.Faculty = f
	} else if u.UserFaculty != nil {
		// fallback: use embedded struct if cache not populated
		a.Faculty = &dto.FeedAuthorFaculty{
			FacultyID:   u.UserFaculty.FacultyID,
			FacultyName: u.UserFaculty.Faculty.FacultyName,
		}
	}
	return a
}

func feedFlagThreadID(t *model.Thread) uint {
	if t.Type == model.ThreadTypeRepost && t.OriginalThread != nil {
		return t.OriginalThread.ThreadID
	}
	return t.ThreadID
}

// LoadPollsForThreads batch-fetches polls and the viewer's votes for the given threads.
// Uses the effective thread id (resolves repost -> original).
func (fb *FeedBuilder) LoadPollsForThreads(threads []model.Thread) {
	if fb.PollSvc == nil {
		return
	}
	seen := make(map[uint]struct{}, len(threads))
	threadIDs := make([]uint, 0, len(threads))
	for i := range threads {
		id := feedFlagThreadID(&threads[i])
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		threadIDs = append(threadIDs, id)
	}
	fb.pollsLoaded = true
	if len(threadIDs) == 0 {
		return
	}
	var viewerID uint
	if fb.ViewerID != nil {
		viewerID = *fb.ViewerID
	}
	polls, votes, err := fb.PollSvc.GetForThreads(threadIDs, viewerID)
	if err != nil {
		return
	}
	fb.pollsByThread = polls
	fb.pollUserVotes = votes
}

func (fb FeedBuilder) pollFor(t *model.Thread) *dto.PollResponse {
	if !fb.pollsLoaded || fb.PollSvc == nil || len(fb.pollsByThread) == 0 {
		return nil
	}
	id := feedFlagThreadID(t)
	poll, ok := fb.pollsByThread[id]
	if !ok || poll == nil {
		return nil
	}
	var voted *uint
	if v, ok := fb.pollUserVotes[poll.PollID]; ok {
		voted = &v
	}
	return dto.BuildPollResponse(poll, voted, fb.PollSvc.GetImageURL)
}

// PollForParent returns the poll response for a parent thread, used in parent_context.
// Parent threads are not reposts, so direct ThreadID lookup is sufficient.
func (fb FeedBuilder) PollForParent(parent *model.Thread) *dto.PollResponse {
	if parent == nil {
		return nil
	}
	return fb.pollFor(parent)
}

// LoadPollsForThreadIDs is an explicit-id variant of LoadPollsForThreads, used when the
// caller needs to merge thread + parent ids in a single batch.
func (fb *FeedBuilder) LoadPollsForThreadIDs(threadIDs []uint) {
	if fb.PollSvc == nil {
		return
	}
	seen := make(map[uint]struct{}, len(threadIDs))
	deduped := make([]uint, 0, len(threadIDs))
	for _, id := range threadIDs {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		deduped = append(deduped, id)
	}
	fb.pollsLoaded = true
	if len(deduped) == 0 {
		return
	}
	var viewerID uint
	if fb.ViewerID != nil {
		viewerID = *fb.ViewerID
	}
	polls, votes, err := fb.PollSvc.GetForThreads(deduped, viewerID)
	if err != nil {
		return
	}
	fb.pollsByThread = polls
	fb.pollUserVotes = votes
}

// LoadViewerFlagsForThreads batch-fetches viewer flags for the feed's effective thread IDs.
// Uses Redis viewer-flag sets when available; falls back to DB and hydrates the set on miss.
func (fb *FeedBuilder) LoadViewerFlagsForThreads(threads []model.Thread) {
	if fb.ViewerID == nil {
		return
	}
	seen := make(map[uint]struct{}, len(threads))
	threadIDs := make([]uint, 0, len(threads))
	for i := range threads {
		id := feedFlagThreadID(&threads[i])
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		threadIDs = append(threadIDs, id)
	}

	fb.flagsLoaded = true
	if len(threadIDs) == 0 {
		return
	}

	uid := *fb.ViewerID
	if fb.LikeSvc != nil {
		fb.liked = fb.loadFlagSet(uid, "liked", threadIDs, fb.LikeSvc.AllLikedThreadIDs, func() (map[uint]bool, error) {
			return fb.LikeSvc.AreThreadsLiked(uid, threadIDs)
		})
	}
	if fb.BookmarkSvc != nil {
		fb.bookmarked = fb.loadFlagSet(uid, "bookmarked", threadIDs, fb.BookmarkSvc.AllBookmarkedThreadIDs, func() (map[uint]bool, error) {
			return fb.BookmarkSvc.AreThreadsBookmarked(uid, threadIDs)
		})
	}
	if fb.ThreadSvc != nil {
		// reposted uses original_thread_id as member
		var allLoader func(uint) ([]uint, error)
		if fb.ThreadSvc != nil {
			allLoader = func(u uint) ([]uint, error) {
				return fb.threadRepostedIDs(u)
			}
		}
		fb.reposted = fb.loadFlagSet(uid, "reposted", threadIDs, allLoader, func() (map[uint]bool, error) {
			return fb.ThreadSvc.AreThreadsReposted(uid, threadIDs)
		})
	}
	if fb.PinSvc != nil {
		fb.pinned = fb.loadFlagSet(uid, "pinned", threadIDs, fb.PinSvc.AllPinnedThreadIDs, func() (map[uint]bool, error) {
			return fb.PinSvc.ArePinnedFor(uid, threadIDs)
		})
	}
}

func (fb FeedBuilder) threadRepostedIDs(uid uint) ([]uint, error) {
	type ar interface {
		AllRepostedOriginalIDs(userID uint) ([]uint, error)
	}
	if r, ok := fb.ThreadSvc.(ar); ok {
		return r.AllRepostedOriginalIDs(uid)
	}
	return nil, nil
}

// loadFlagSet attempts Redis SMISMEMBER first. On hydration miss (set not authoritative),
// hydrates the set from DB and falls back to DB batch result for this request.
func (fb FeedBuilder) loadFlagSet(
	uid uint,
	flag string,
	threadIDs []uint,
	allLoader func(uint) ([]uint, error),
	dbBatch func() (map[uint]bool, error),
) map[uint]bool {
	if fb.Cache != nil && allLoader != nil {
		if members, ok := viewerFlagMembership(fb.Cache, uid, flag, threadIDs); ok {
			return members
		}
		// not hydrated — hydrate then check
		_ = hydrateViewerFlag(fb.Cache, uid, flag, func() ([]uint, error) { return allLoader(uid) })
		if members, ok := viewerFlagMembership(fb.Cache, uid, flag, threadIDs); ok {
			return members
		}
	}
	// fallback: DB batch
	result, _ := dbBatch()
	if result == nil {
		result = make(map[uint]bool)
	}
	return result
}

const (
	viewerFlagSetTTL = 24 * time.Hour
)

func viewerFlagSetKey(userID uint, flag string) string {
	return fmt.Sprintf("user:%d:%s", userID, flag)
}

func viewerFlagHydratedKey(userID uint, flag string) string {
	return fmt.Sprintf("user:%d:%s:hydrated", userID, flag)
}

func viewerFlagMembership(c *cache.Cache, userID uint, flag string, threadIDs []uint) (map[uint]bool, bool) {
	if c == nil || userID == 0 || len(threadIDs) == 0 {
		return nil, false
	}
	ctx := context.Background()
	if hit, _ := c.Exists(ctx, viewerFlagHydratedKey(userID, flag)); !hit {
		return nil, false
	}
	members := make([]string, len(threadIDs))
	for i, id := range threadIDs {
		members[i] = strconv.FormatUint(uint64(id), 10)
	}
	results, err := c.SMIsMember(ctx, viewerFlagSetKey(userID, flag), members)
	if err != nil {
		return nil, false
	}
	out := make(map[uint]bool, len(threadIDs))
	for i, id := range threadIDs {
		if i < len(results) && results[i] {
			out[id] = true
		}
	}
	return out, true
}

func hydrateViewerFlag(c *cache.Cache, userID uint, flag string, loader func() ([]uint, error)) error {
	if c == nil || userID == 0 {
		return nil
	}
	ctx := context.Background()
	if hit, _ := c.Exists(ctx, viewerFlagHydratedKey(userID, flag)); hit {
		return nil
	}
	ids, err := loader()
	if err != nil {
		return err
	}
	setKey := viewerFlagSetKey(userID, flag)
	_ = c.Delete(ctx, setKey)
	members := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		members = append(members, strconv.FormatUint(uint64(id), 10))
	}
	if err := c.SAddMany(ctx, setKey, members, viewerFlagSetTTL); err != nil {
		return err
	}
	return c.Set(ctx, viewerFlagHydratedKey(userID, flag), true, viewerFlagSetTTL)
}

func (fb FeedBuilder) viewerFlags(threadID uint) (isLiked, isBookmarked, isReposted, isPinned *bool) {
	if fb.ViewerID == nil {
		return
	}
	if fb.flagsLoaded {
		if fb.LikeSvc != nil {
			v := fb.liked[threadID]
			isLiked = &v
		}
		if fb.BookmarkSvc != nil {
			v := fb.bookmarked[threadID]
			isBookmarked = &v
		}
		if fb.ThreadSvc != nil {
			v := fb.reposted[threadID]
			isReposted = &v
		}
		if fb.PinSvc != nil {
			v := fb.pinned[threadID]
			isPinned = &v
		}
		return
	}
	uid := *fb.ViewerID
	if fb.LikeSvc != nil {
		v := fb.LikeSvc.IsLiked(uid, threadID)
		isLiked = &v
	}
	if fb.BookmarkSvc != nil {
		v := fb.BookmarkSvc.IsBookmarked(uid, threadID)
		isBookmarked = &v
	}
	if fb.ThreadSvc != nil {
		v := fb.ThreadSvc.IsReposted(uid, threadID)
		isReposted = &v
	}
	if fb.PinSvc != nil {
		v := fb.PinSvc.IsPinned(uid, threadID)
		isPinned = &v
	}
	return
}

func (fb FeedBuilder) ToFeedItem(t *model.Thread, media []model.ThreadMedia) dto.FeedItemResponse {
	if t.Type == model.ThreadTypeRepost && t.OriginalThread != nil {
		orig := t.OriginalThread
		repostedAt := dtoutils.Wrap(t.CreatedAt)
		isLiked, isBookmarked, isReposted, isPinned := fb.viewerFlags(orig.ThreadID)
		eventID, eventName := EventFields(orig)
		return dto.FeedItemResponse{
			Type:               string(model.ThreadTypeRepost),
			ID:                 t.ThreadID,
			ParentID:           orig.ParentID,
			RootID:             orig.RootID,
			EventID:            eventID,
			EventName:          eventName,
			Community:          CommunityField(orig),
			Content:            orig.Content,
			ReplyPermission:    string(orig.ReplyPermission),
			CreatedAt:          dtoutils.Wrap(orig.CreatedAt),
			UpdatedAt:          dtoutils.Wrap(orig.UpdatedAt),
			TimeAgo:            TimeAgo(orig.CreatedAt),
			RepostedAt:         &repostedAt,
			RepostedByID:       &t.UserID,
			RepostedByUsername: t.User.UserName,
			RepostedByName:     t.User.Name,
			OriginalThreadID:   &orig.ThreadID,
			Author:             fb.AuthorField(orig.User),
			Stats:              dto.FeedStats{Likes: orig.LikeCount, Reposts: orig.RepostCount, Bookmarks: orig.BookmarkCount, Comments: orig.CommentCount},
			Media:              fb.ToMediaResponses(media),
			Poll:               fb.pollFor(t),
			IsLiked:            isLiked,
			IsBookmarked:       isBookmarked,
			IsReposted:         isReposted,
			IsPinned:           isPinned,
		}
	}
	isLiked, isBookmarked, isReposted, isPinned := fb.viewerFlags(t.ThreadID)
	eventID, eventName := EventFields(t)
	return dto.FeedItemResponse{
		Type:            string(t.Type),
		ID:              t.ThreadID,
		ParentID:        t.ParentID,
		RootID:          t.RootID,
		EventID:         eventID,
		EventName:       eventName,
		Community:       CommunityField(t),
		Content:         t.Content,
		ReplyPermission: string(t.ReplyPermission),
		CreatedAt:    dtoutils.Wrap(t.CreatedAt),
		UpdatedAt:    dtoutils.Wrap(t.UpdatedAt),
		TimeAgo:      TimeAgo(t.CreatedAt),
		Author:       fb.AuthorField(t.User),
		Stats:        dto.FeedStats{Likes: t.LikeCount, Reposts: t.RepostCount, Bookmarks: t.BookmarkCount, Comments: t.CommentCount},
		Media:        fb.ToMediaResponses(media),
		Poll:         fb.pollFor(t),
		IsLiked:      isLiked,
		IsBookmarked: isBookmarked,
		IsReposted:   isReposted,
		IsPinned:     isPinned,
	}
}

func EnsureFalseViewerFlags(data []dto.FeedItemResponse) {
	falseVal := false
	for i := range data {
		if data[i].IsLiked == nil {
			data[i].IsLiked = &falseVal
		}
		if data[i].IsBookmarked == nil {
			data[i].IsBookmarked = &falseVal
		}
		if data[i].IsReposted == nil {
			data[i].IsReposted = &falseVal
		}
	}
}
