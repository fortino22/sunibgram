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

func (fb FeedBuilder) viewerFlags(threadID uint) (isLiked, isBookmarked, isReposted, isPinned *bool) {
	if fb.ViewerID == nil {
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
		return dto.FeedItemResponse{
			Type:         string(model.ThreadTypeRepost),
			ID:           orig.ThreadID,
			ParentID:     orig.ParentID,
			RootID:       orig.RootID,
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
	return dto.FeedItemResponse{
		Type:         string(t.Type),
		ID:           t.ThreadID,
		ParentID:     t.ParentID,
		RootID:       t.RootID,
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
