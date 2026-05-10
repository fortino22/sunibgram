package controller

import (
	"sort"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/dto"
	dtoutils "github.com/sunibgram/dto/utils"
	webapiutils "github.com/sunibgram/webapi/utils"
)

func (c *ThreadController) loadMediaByThread(threads []model.Thread) map[uint][]model.ThreadMedia {
	ids := make([]uint, len(threads))
	for i, t := range threads {
		ids[i] = t.ThreadID
	}
	allMedia, _ := c.mediaService.GetByThreadIDs(ids)
	m := make(map[uint][]model.ThreadMedia, len(ids))
	for _, media := range allMedia {
		m[media.ThreadID] = append(m[media.ThreadID], media)
	}
	return m
}

func (c *ThreadController) fb(viewerID *uint) webapiutils.FeedBuilder {
	return webapiutils.FeedBuilder{
		Media:    c.mediaService,
		ViewerID: viewerID,
	}
}

func (c *ThreadController) fbWithPin(viewerID *uint) webapiutils.FeedBuilder {
	return webapiutils.FeedBuilder{
		Media:       c.mediaService,
		LikeSvc:     c.likeService,
		BookmarkSvc: c.bookmarkService,
		ThreadSvc:   c.threadService,
		PinSvc:      c.threadPinService,
		ViewerID:    viewerID,
	}
}

func (c *ThreadController) fbNoPin(viewerID *uint) webapiutils.FeedBuilder {
	return webapiutils.FeedBuilder{
		Media:       c.mediaService,
		LikeSvc:     c.likeService,
		BookmarkSvc: c.bookmarkService,
		ThreadSvc:   c.threadService,
		ViewerID:    viewerID,
	}
}

func (c *ThreadController) buildFeedItems(threads []model.Thread, viewerID *uint) []dto.FeedItemResponse {
	mediaByThread := c.loadMediaByThread(threads)
	data := make([]dto.FeedItemResponse, len(threads))
	fb := c.fb(viewerID)
	for i, t := range threads {
		data[i] = fb.ToFeedItem(&t, mediaByThread[t.ThreadID])
	}
	return data
}

func (c *ThreadController) buildFeedItemsNoPin(threads []model.Thread, viewerID *uint) []dto.FeedItemResponse {
	mediaByThread := c.loadMediaByThread(threads)
	data := make([]dto.FeedItemResponse, len(threads))
	fb := c.fbNoPin(viewerID)
	for i, t := range threads {
		data[i] = fb.ToFeedItem(&t, mediaByThread[t.ThreadID])
	}
	return data
}

func (c *ThreadController) buildFeedItemsWithPin(threads []model.Thread, viewerID *uint) []dto.FeedItemResponse {
	mediaByThread := c.loadMediaByThread(threads)
	data := make([]dto.FeedItemResponse, len(threads))
	fb := c.fbWithPin(viewerID)
	for i, t := range threads {
		data[i] = fb.ToFeedItem(&t, mediaByThread[t.ThreadID])
	}
	return data
}

func (c *ThreadController) buildUserFeedItems(threads []model.Thread, parents []*model.Thread, viewerID *uint) []dto.FeedItemResponse {
	return c.buildUserFeedItemsWithBuilder(threads, parents, c.fb(viewerID))
}

func (c *ThreadController) buildUserFeedItemsNoPin(threads []model.Thread, parents []*model.Thread, viewerID *uint) []dto.FeedItemResponse {
	return c.buildUserFeedItemsWithBuilder(threads, parents, c.fbNoPin(viewerID))
}

func (c *ThreadController) buildUserFeedItemsWithPin(threads []model.Thread, parents []*model.Thread, viewerID *uint) []dto.FeedItemResponse {
	return c.buildUserFeedItemsWithBuilder(threads, parents, c.fbWithPin(viewerID))
}

func (c *ThreadController) buildUserFeedItemsWithBuilder(threads []model.Thread, parents []*model.Thread, fb webapiutils.FeedBuilder) []dto.FeedItemResponse {
	allIDs := make([]uint, 0, len(threads)*2)
	for _, t := range threads {
		allIDs = append(allIDs, t.ThreadID)
	}
	for _, p := range parents {
		if p != nil {
			allIDs = append(allIDs, p.ThreadID)
		}
	}
	allMedia, _ := c.mediaService.GetByThreadIDs(allIDs)
	mediaByThread := make(map[uint][]model.ThreadMedia, len(allIDs))
	for _, m := range allMedia {
		mediaByThread[m.ThreadID] = append(mediaByThread[m.ThreadID], m)
	}

	data := make([]dto.FeedItemResponse, len(threads))
	for i, t := range threads {
		item := fb.ToFeedItem(&t, mediaByThread[t.ThreadID])
		if parents[i] != nil {
			p := parents[i]
			item.ParentContext = &dto.FeedParentContext{
				ID:        p.ThreadID,
				ParentID:  p.ParentID,
				RootID:    p.RootID,
				Community: webapiutils.CommunityField(p),
				Author:    webapiutils.AuthorField(p.User),
				Content:   p.Content,
				Media:     fb.ToMediaResponses(mediaByThread[p.ThreadID]),
				Stats:     dto.FeedStats{Likes: p.LikeCount, Reposts: p.RepostCount, Bookmarks: p.BookmarkCount, Comments: p.CommentCount},
				CreatedAt: dtoutils.Wrap(p.CreatedAt),
				UpdatedAt: dtoutils.Wrap(p.UpdatedAt),
			}
		}
		data[i] = item
	}
	sort.SliceStable(data, func(i, j int) bool {
		return boolValue(data[i].IsPinned) && !boolValue(data[j].IsPinned)
	})
	return data
}

func boolValue(v *bool) bool {
	return v != nil && *v
}

func (c *ThreadController) ensureFalseViewerFlags(data []dto.FeedItemResponse) {
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
