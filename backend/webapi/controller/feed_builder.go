package controller

import (
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/dto/utils"
)

type feedBuilder struct {
	media service.IThreadMediaService
}

func (fb feedBuilder) toMediaResponses(media []model.ThreadMedia) []dto.ThreadMediaResponse {
	result := make([]dto.ThreadMediaResponse, len(media))
	for i, m := range media {
		result[i] = dto.ThreadMediaResponse{MediaID: m.MediaID, URL: fb.media.GetURL(m.Bucket, m.ObjectName)}
	}
	return result
}

func (fb feedBuilder) toFeedItem(t *model.Thread, media []model.ThreadMedia) dto.FeedItemResponse {
	return dto.FeedItemResponse{
		Type:      "POST",
		ID:        t.ThreadID,
		ParentID:  t.ParentID,
		RootID:    t.RootID,
		Content:   t.Content,
		CreatedAt: utils.Wrap(t.CreatedAt),
		UpdatedAt: utils.Wrap(t.UpdatedAt),
		Author:    dto.FeedAuthor{ID: t.UserID, Username: t.User.UserName},
		Stats:     dto.FeedStats{Likes: t.LikeCount, Reposts: t.RepostCount, Bookmarks: t.BookmarkCount},
		Media:     fb.toMediaResponses(media),
	}
}
