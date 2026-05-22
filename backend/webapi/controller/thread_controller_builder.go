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
	return newFeedBuilder(c.mediaService, nil, nil, nil, nil, c.profileService, c.pollService, viewerID)
}

func (c *ThreadController) fbWithPin(viewerID *uint) webapiutils.FeedBuilder {
	return newFeedBuilder(c.mediaService, c.likeService, c.bookmarkService, c.threadService, c.threadPinService, c.profileService, c.pollService, viewerID)
}

func (c *ThreadController) fbNoPin(viewerID *uint) webapiutils.FeedBuilder {
	return newFeedBuilder(c.mediaService, c.likeService, c.bookmarkService, c.threadService, nil, c.profileService, c.pollService, viewerID)
}

func (c *ThreadController) buildFeedItems(threads []model.Thread, viewerID *uint) []dto.FeedItemResponse {
	mediaByThread := c.loadMediaByThread(threads)
	fb := c.fb(viewerID)
	fb.LoadAvatarsForUserIDs(threadUserIDs(threads))
	return buildFeedItemsFromThreads(threads, mediaByThread, fb)
}

func (c *ThreadController) buildFeedItemsNoPin(threads []model.Thread, viewerID *uint) []dto.FeedItemResponse {
	mediaByThread := c.loadMediaByThread(threads)
	fb := c.fbNoPin(viewerID)
	fb.LoadAvatarsForUserIDs(threadUserIDs(threads))
	return buildFeedItemsFromThreads(threads, mediaByThread, fb)
}

func (c *ThreadController) buildFeedItemsWithPin(threads []model.Thread, viewerID *uint) []dto.FeedItemResponse {
	mediaByThread := c.loadMediaByThread(threads)
	fb := c.fbWithPin(viewerID)
	fb.LoadAvatarsForUserIDs(threadUserIDs(threads))
	return buildFeedItemsFromThreads(threads, mediaByThread, fb)
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
		if t.OriginalThread != nil {
			allIDs = append(allIDs, t.OriginalThread.ThreadID)
		}
	}
	for _, p := range parents {
		if p != nil {
			allIDs = append(allIDs, p.ThreadID)
		}
	}
	mediaByThread := loadThreadMediaByID(c.mediaService, allIDs)

	// Collect all author IDs including parents so avatars load in one batch
	authorIDs := threadUserIDs(threads)
	for _, p := range parents {
		if p != nil {
			authorIDs = append(authorIDs, p.User.UserID)
		}
	}
	fb.LoadAvatarsForUserIDs(authorIDs)
	fb.LoadViewerFlagsForThreads(threads)

	pollIDs := make([]uint, 0, len(threads)+len(parents))
	for i := range threads {
		t := &threads[i]
		if t.Type == model.ThreadTypeRepost && t.OriginalThread != nil {
			pollIDs = append(pollIDs, t.OriginalThread.ThreadID)
		} else {
			pollIDs = append(pollIDs, t.ThreadID)
		}
	}
	for _, p := range parents {
		if p != nil {
			pollIDs = append(pollIDs, p.ThreadID)
		}
	}
	fb.LoadPollsForThreadIDs(pollIDs)

	data := make([]dto.FeedItemResponse, len(threads))
	for i, t := range threads {
		mediaID := t.ThreadID
		if t.Type == model.ThreadTypeRepost && t.OriginalThread != nil {
			mediaID = t.OriginalThread.ThreadID
		}
		item := fb.ToFeedItem(&t, mediaByThread[mediaID])
		if parents[i] != nil {
			p := parents[i]
			eventID, eventName := webapiutils.EventFields(p)
			item.ParentContext = &dto.FeedParentContext{
				ID:              p.ThreadID,
				ParentID:        p.ParentID,
				RootID:          p.RootID,
				EventID:         eventID,
				EventName:       eventName,
				Community:       webapiutils.CommunityField(p),
				Author:          fb.AuthorField(p.User),
				Content:         p.Content,
				ReplyPermission: string(p.ReplyPermission),
				Media:           fb.ToMediaResponses(mediaByThread[p.ThreadID]),
				Poll:            fb.PollForParent(p),
				Stats:           dto.FeedStats{Likes: p.LikeCount, Reposts: p.RepostCount, Bookmarks: p.BookmarkCount, Comments: p.CommentCount},
				CreatedAt:       dtoutils.Wrap(p.CreatedAt),
				UpdatedAt:       dtoutils.Wrap(p.UpdatedAt),
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
