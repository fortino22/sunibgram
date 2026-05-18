package utils

import (
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	dtoutils "github.com/sunibgram/dto/utils"
)

type FeedBuilder struct {
	Media       service.IThreadMediaService
	LikeSvc     service.ILikeService
	BookmarkSvc service.IBookmarkService
	ThreadSvc   service.IThreadService
	PinSvc      service.IThreadPinService
	ProfileSvc  service.IUserProfileService
	ViewerID    *uint
	avatarCache map[uint]string
	flagsLoaded bool
	liked       map[uint]bool
	bookmarked  map[uint]bool
	reposted    map[uint]bool
	pinned      map[uint]bool
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

// LoadAvatarsForUserIDs batch-fetches and caches avatar URLs for the given user IDs.
func (fb *FeedBuilder) LoadAvatarsForUserIDs(userIDs []uint) {
	if fb.ProfileSvc == nil || len(userIDs) == 0 {
		return
	}
	if fb.avatarCache == nil {
		fb.avatarCache = make(map[uint]string)
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
	profiles, err := fb.ProfileSvc.GetProfiles(needed)
	if err != nil {
		return
	}
	for _, p := range profiles {
		if p.AvatarBucket != "" && p.AvatarObjectName != "" {
			fb.avatarCache[p.UserID] = fb.ProfileSvc.GetURL(p.AvatarBucket, p.AvatarObjectName)
		} else {
			fb.avatarCache[p.UserID] = ""
		}
	}
}

// AuthorField builds a FeedAuthor from a User, resolving the avatar from the cache.
// Reading from a nil map in Go returns the zero value, so it is safe when avatarCache is nil.
func (fb FeedBuilder) AuthorField(u model.User) dto.FeedAuthor {
	a := dto.FeedAuthor{
		ID:        u.UserID,
		Username:  u.UserName,
		IsPremium: u.IsPremium,
		AvatarURL: fb.avatarCache[u.UserID],
	}
	if u.UserFaculty != nil {
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
