package utils

import (
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	dtoutils "github.com/sunibgram/dto/utils"
)

type FeedBuilder struct {
	Media service.IThreadMediaService
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

func (fb FeedBuilder) ToFeedItem(t *model.Thread, media []model.ThreadMedia) dto.FeedItemResponse {
	if t.Type == model.ThreadTypeRepost && t.OriginalThread != nil {
		orig := t.OriginalThread
		repostedAt := dtoutils.Wrap(t.CreatedAt)
		return dto.FeedItemResponse{
			Type:       string(model.ThreadTypeRepost),
			ID:         orig.ThreadID,
			ParentID:   orig.ParentID,
			RootID:     orig.RootID,
			Community:  CommunityField(orig),
			Content:    orig.Content,
			CreatedAt:  dtoutils.Wrap(orig.CreatedAt),
			UpdatedAt:  dtoutils.Wrap(orig.UpdatedAt),
			TimeAgo:    TimeAgo(orig.CreatedAt),
			RepostedAt: &repostedAt,
			Author:     dto.FeedAuthor{ID: orig.UserID, Username: orig.User.UserName},
			Stats:      dto.FeedStats{Likes: orig.LikeCount, Reposts: orig.RepostCount, Bookmarks: orig.BookmarkCount},
			Media:      fb.ToMediaResponses(media),
		}
	}
	return dto.FeedItemResponse{
		Type:      string(t.Type),
		ID:        t.ThreadID,
		ParentID:  t.ParentID,
		RootID:    t.RootID,
		Community: CommunityField(t),
		Content:   t.Content,
		CreatedAt: dtoutils.Wrap(t.CreatedAt),
		UpdatedAt: dtoutils.Wrap(t.UpdatedAt),
		TimeAgo:   TimeAgo(t.CreatedAt),
		Author:    dto.FeedAuthor{ID: t.UserID, Username: t.User.UserName},
		Stats:     dto.FeedStats{Likes: t.LikeCount, Reposts: t.RepostCount, Bookmarks: t.BookmarkCount},
		Media:     fb.ToMediaResponses(media),
	}
}
