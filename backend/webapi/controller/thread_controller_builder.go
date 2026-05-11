package controller

import (
	"sort"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/dto"
	dtoutils "github.com/sunibgram/dto/utils"
	webapiutils "github.com/sunibgram/webapi/utils"
)

func (c *ThreadController) loadMediaByThread(threads []model.Thread) map[uint][]model.ThreadMedia {
	return loadThreadMediaByThread(c.mediaService, threads)
}

func (c *ThreadController) fb(viewerID *uint) webapiutils.FeedBuilder {
	return newFeedBuilder(c.mediaService, nil, nil, nil, nil, viewerID)
}

func (c *ThreadController) fbWithPin(viewerID *uint) webapiutils.FeedBuilder {
	return newFeedBuilder(c.mediaService, c.likeService, c.bookmarkService, c.threadService, c.threadPinService, viewerID)
}

func (c *ThreadController) fbNoPin(viewerID *uint) webapiutils.FeedBuilder {
	return newFeedBuilder(c.mediaService, c.likeService, c.bookmarkService, c.threadService, nil, viewerID)
}

func (c *ThreadController) buildFeedItems(threads []model.Thread, viewerID *uint) []dto.FeedItemResponse {
	mediaByThread := c.loadMediaByThread(threads)
	return buildFeedItemsFromThreads(threads, mediaByThread, c.fb(viewerID))
}

func (c *ThreadController) buildFeedItemsNoPin(threads []model.Thread, viewerID *uint) []dto.FeedItemResponse {
	mediaByThread := c.loadMediaByThread(threads)
	return buildFeedItemsFromThreads(threads, mediaByThread, c.fbNoPin(viewerID))
}

func (c *ThreadController) buildFeedItemsWithPin(threads []model.Thread, viewerID *uint) []dto.FeedItemResponse {
	mediaByThread := c.loadMediaByThread(threads)
	return buildFeedItemsFromThreads(threads, mediaByThread, c.fbWithPin(viewerID))
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
	mediaByThread := loadThreadMediaByID(c.mediaService, allIDs)

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
				Author:    fb.AuthorField(p.User),
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
		return boolPtrValue(data[i].IsPinned) && !boolPtrValue(data[j].IsPinned)
	})
	return data
}

func (c *ThreadController) ensureFalseViewerFlags(data []dto.FeedItemResponse) {
	ensureFalseFeedViewerFlags(data)
}
