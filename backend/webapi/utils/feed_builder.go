package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	dtoutils "github.com/sunibgram/dto/utils"
)

const avatarCacheTTL = time.Hour

type avatarCacheProvider interface {
	AvatarCacheRepository() repository.ICacheRepository
}

type FeedBuilder struct {
	Media       service.IThreadMediaService
	LikeSvc     service.ILikeService
	BookmarkSvc service.IBookmarkService
	ThreadSvc   service.IThreadService
	PinSvc      service.IThreadPinService
	ProfileSvc  service.IUserProfileService
	ViewerID    *uint
	avatarCache    map[uint]string
	isPremiumCache map[uint]bool
	facultyCache   map[uint]*dto.FeedAuthorFaculty
	flagsLoaded bool
	liked       map[uint]bool
	bookmarked  map[uint]bool
	reposted    map[uint]bool
	pinned      map[uint]bool
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
		if _, cached := fb.avatarCache[id]; cached {
			continue
		}
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
	misses := make([]uint, 0, len(needed))
	for _, id := range needed {
		if cacheRepo == nil {
			misses = append(misses, id)
			continue
		}
		var avatarURL string
		if err := cacheRepo.Get(ctx, avatarCacheKey(id), &avatarURL); err == nil {
			fb.avatarCache[id] = avatarURL
			continue
		}
		misses = append(misses, id)
	}
	if len(misses) == 0 {
		return
	}
	profiles, err := fb.ProfileSvc.GetProfiles(misses)
	if err != nil {
		return
	}
	missed := make(map[uint]struct{}, len(misses))
	for _, id := range misses {
		missed[id] = struct{}{}
	}
	for _, p := range profiles {
		delete(missed, p.UserID)
		avatarURL := ""
		if p.AvatarBucket != "" && p.AvatarObjectName != "" {
			avatarURL = fb.ProfileSvc.GetURL(p.AvatarBucket, p.AvatarObjectName)
		}
		fb.avatarCache[p.UserID] = avatarURL
		if cacheRepo != nil {
			_ = cacheRepo.Set(ctx, avatarCacheKey(p.UserID), avatarURL, avatarCacheTTL)
		}
	}
	for id := range missed {
		fb.avatarCache[id] = ""
		if cacheRepo != nil {
			_ = cacheRepo.Set(ctx, avatarCacheKey(id), "", avatarCacheTTL)
		}
	}

	// Batch-load is_premium fresh from DB (not from stale thread cache).
	users, err := fb.ProfileSvc.GetUsers(misses)
	if err == nil {
		for _, u := range users {
			fb.isPremiumCache[u.UserID] = u.IsPremium
		}
	}

	// Batch-load faculty fresh from DB.
	faculties, err := fb.ProfileSvc.GetUserFaculties(misses)
	if err == nil {
		// mark all misses as looked up (nil = no faculty)
		for _, id := range misses {
			fb.facultyCache[id] = nil
		}
		for _, uf := range faculties {
			fb.facultyCache[uf.UserID] = &dto.FeedAuthorFaculty{
				FacultyID:   uf.FacultyID,
				FacultyName: uf.Faculty.FacultyName,
			}
		}
	}
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

// LoadViewerFlagsForThreads batch-fetches viewer flags for the feed's effective thread IDs.
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
		fb.liked, _ = fb.LikeSvc.AreThreadsLiked(uid, threadIDs)
		if fb.liked == nil {
			fb.liked = make(map[uint]bool)
		}
	}
	if fb.BookmarkSvc != nil {
		fb.bookmarked, _ = fb.BookmarkSvc.AreThreadsBookmarked(uid, threadIDs)
		if fb.bookmarked == nil {
			fb.bookmarked = make(map[uint]bool)
		}
	}
	if fb.ThreadSvc != nil {
		fb.reposted, _ = fb.ThreadSvc.AreThreadsReposted(uid, threadIDs)
		if fb.reposted == nil {
			fb.reposted = make(map[uint]bool)
		}
	}
	if fb.PinSvc != nil {
		fb.pinned, _ = fb.PinSvc.ArePinnedFor(uid, threadIDs)
		if fb.pinned == nil {
			fb.pinned = make(map[uint]bool)
		}
	}
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
			Type:         string(model.ThreadTypeRepost),
			ID:           orig.ThreadID,
			ParentID:     orig.ParentID,
			RootID:       orig.RootID,
			EventID:      eventID,
			EventName:    eventName,
			Community:    CommunityField(orig),
			Content:      orig.Content,
			CreatedAt:    dtoutils.Wrap(orig.CreatedAt),
			UpdatedAt:    dtoutils.Wrap(orig.UpdatedAt),
			TimeAgo:      TimeAgo(orig.CreatedAt),
			RepostedAt:   &repostedAt,
			Author:       fb.AuthorField(orig.User),
			Stats:        dto.FeedStats{Likes: orig.LikeCount, Reposts: orig.RepostCount, Bookmarks: orig.BookmarkCount, Comments: orig.CommentCount},
			Media:        fb.ToMediaResponses(media),
			IsLiked:      isLiked,
			IsBookmarked: isBookmarked,
			IsReposted:   isReposted,
			IsPinned:     isPinned,
		}
	}
	isLiked, isBookmarked, isReposted, isPinned := fb.viewerFlags(t.ThreadID)
	eventID, eventName := EventFields(t)
	return dto.FeedItemResponse{
		Type:         string(t.Type),
		ID:           t.ThreadID,
		ParentID:     t.ParentID,
		RootID:       t.RootID,
		EventID:      eventID,
		EventName:    eventName,
		Community:    CommunityField(t),
		Content:      t.Content,
		CreatedAt:    dtoutils.Wrap(t.CreatedAt),
		UpdatedAt:    dtoutils.Wrap(t.UpdatedAt),
		TimeAgo:      TimeAgo(t.CreatedAt),
		Author:       fb.AuthorField(t.User),
		Stats:        dto.FeedStats{Likes: t.LikeCount, Reposts: t.RepostCount, Bookmarks: t.BookmarkCount, Comments: t.CommentCount},
		Media:        fb.ToMediaResponses(media),
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
