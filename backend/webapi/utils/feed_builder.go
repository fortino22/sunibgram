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
	ViewerID    *uint
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

func AuthorField(u model.User) dto.FeedAuthor {
	a := dto.FeedAuthor{ID: u.UserID, Username: u.UserName, IsPremium: u.IsPremium}
	if u.UserFaculty != nil {
		a.Faculty = &dto.FeedAuthorFaculty{
			FacultyID:   u.UserFaculty.FacultyID,
			FacultyName: u.UserFaculty.Faculty.FacultyName,
		}
	}
	return a
}

func (fb FeedBuilder) viewerFlags(threadID uint) (isLiked, isBookmarked, isReposted, isPinned bool) {
	if fb.ViewerID == nil {
		return
	}
	uid := *fb.ViewerID
	if fb.LikeSvc != nil {
		isLiked = fb.LikeSvc.IsLiked(uid, threadID)
	}
	if fb.BookmarkSvc != nil {
		isBookmarked = fb.BookmarkSvc.IsBookmarked(uid, threadID)
	}
	if fb.ThreadSvc != nil {
		isReposted = fb.ThreadSvc.IsReposted(uid, threadID)
	}
	if fb.PinSvc != nil {
		isPinned = fb.PinSvc.IsPinned(uid, threadID)
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
			Author:       AuthorField(orig.User),
			Stats:        dto.FeedStats{Likes: orig.LikeCount, Reposts: orig.RepostCount, Bookmarks: orig.BookmarkCount},
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
		Author:       AuthorField(t.User),
		Stats:        dto.FeedStats{Likes: t.LikeCount, Reposts: t.RepostCount, Bookmarks: t.BookmarkCount},
		Media:        fb.ToMediaResponses(media),
		IsLiked:      isLiked,
		IsBookmarked: isBookmarked,
		IsReposted:   isReposted,
		IsPinned:     isPinned,
	}
}
