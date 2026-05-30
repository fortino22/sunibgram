package threadctrl

import (
	"sort"

	threadmodel "github.com/sunibgram/domain/model/thread"
	threadmediamodel "github.com/sunibgram/domain/model/thread_media"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	dtoutils "github.com/sunibgram/webapi/dto/_shared"
	threaddto "github.com/sunibgram/webapi/dto/thread"
	webapiutils "github.com/sunibgram/webapi/utils"
)

func (c *ThreadController) loadMediaByThread(threads []threadmodel.Thread) map[uint][]threadmediamodel.ThreadMedia {
	return ctrlshared.LoadThreadMediaByThread(c.mediaService, threads)
}

func (c *ThreadController) fb(viewerID *uint) webapiutils.FeedBuilder {
	return ctrlshared.NewFeedBuilder(c.mediaService, nil, nil, nil, nil, c.profileService, c.pollService, viewerID)
}

func (c *ThreadController) fbWithPin(viewerID *uint) webapiutils.FeedBuilder {
	return ctrlshared.NewFeedBuilder(c.mediaService, c.likeService, c.bookmarkService, c.threadService, c.threadPinService, c.profileService, c.pollService, viewerID)
}

func (c *ThreadController) fbNoPin(viewerID *uint) webapiutils.FeedBuilder {
	return ctrlshared.NewFeedBuilder(c.mediaService, c.likeService, c.bookmarkService, c.threadService, nil, c.profileService, c.pollService, viewerID)
}

func (c *ThreadController) buildFeedItems(threads []threadmodel.Thread, viewerID *uint) []threaddto.FeedItemResponse {
	mediaByThread := c.loadMediaByThread(threads)
	fb := c.fb(viewerID)
	fb.LoadAvatarsForUserIDs(ThreadUserIDs(threads))
	return ctrlshared.BuildFeedItemsFromThreads(threads, mediaByThread, fb)
}

func (c *ThreadController) buildFeedItemsNoPin(threads []threadmodel.Thread, viewerID *uint) []threaddto.FeedItemResponse {
	mediaByThread := c.loadMediaByThread(threads)
	fb := c.fbNoPin(viewerID)
	fb.LoadAvatarsForUserIDs(ThreadUserIDs(threads))
	return ctrlshared.BuildFeedItemsFromThreads(threads, mediaByThread, fb)
}

func (c *ThreadController) buildFeedItemsWithPin(threads []threadmodel.Thread, viewerID *uint) []threaddto.FeedItemResponse {
	mediaByThread := c.loadMediaByThread(threads)
	fb := c.fbWithPin(viewerID)
	fb.LoadAvatarsForUserIDs(ThreadUserIDs(threads))
	return ctrlshared.BuildFeedItemsFromThreads(threads, mediaByThread, fb)
}

func (c *ThreadController) buildUserFeedItems(threads []threadmodel.Thread, parents []*threadmodel.Thread, viewerID *uint) []threaddto.FeedItemResponse {
	return c.buildUserFeedItemsWithBuilder(threads, parents, c.fb(viewerID))
}

func (c *ThreadController) BuildUserFeedItemsNoPin(threads []threadmodel.Thread, parents []*threadmodel.Thread, viewerID *uint) []threaddto.FeedItemResponse {
	return c.buildUserFeedItemsWithBuilder(threads, parents, c.fbNoPin(viewerID))
}

func (c *ThreadController) BuildUserFeedItemsWithPin(threads []threadmodel.Thread, parents []*threadmodel.Thread, viewerID *uint) []threaddto.FeedItemResponse {
	return c.buildUserFeedItemsWithBuilder(threads, parents, c.fbWithPin(viewerID))
}

func (c *ThreadController) buildUserFeedItemsWithBuilder(threads []threadmodel.Thread, parents []*threadmodel.Thread, fb webapiutils.FeedBuilder) []threaddto.FeedItemResponse {
	allIDs := make([]uint, 0, len(threads)*2+len(parents))
	authorIDs := make([]uint, 0, len(threads)*2+len(parents))
	pollIDs := make([]uint, 0, len(threads)+len(parents))
	for i := range threads {
		t := &threads[i]
		allIDs = append(allIDs, t.ThreadID)
		if t.Type == threadmodel.ThreadTypeRepost && t.OriginalThread != nil {
			allIDs = append(allIDs, t.OriginalThread.ThreadID)
			pollIDs = append(pollIDs, t.OriginalThread.ThreadID)
			authorIDs = append(authorIDs, t.OriginalThread.User.UserID, t.UserID)
		} else {
			pollIDs = append(pollIDs, t.ThreadID)
			authorIDs = append(authorIDs, t.User.UserID)
		}
	}
	for _, p := range parents {
		if p != nil {
			allIDs = append(allIDs, p.ThreadID)
			authorIDs = append(authorIDs, p.User.UserID)
			pollIDs = append(pollIDs, p.ThreadID)
		}
	}
	mediaByThread := ctrlshared.LoadThreadMediaByID(c.mediaService, allIDs)
	fb.LoadAvatarsForUserIDs(authorIDs)
	fb.LoadViewerFlagsForThreads(threads)
	fb.LoadPollsForThreadIDs(pollIDs)

	data := make([]threaddto.FeedItemResponse, len(threads))
	for i, t := range threads {
		mediaID := t.ThreadID
		if t.Type == threadmodel.ThreadTypeRepost && t.OriginalThread != nil {
			mediaID = t.OriginalThread.ThreadID
		}
		item := fb.ToFeedItem(&t, mediaByThread[mediaID])
		if parents[i] != nil {
			p := parents[i]
			eventID, eventName := webapiutils.EventFields(p)
			item.ParentContext = &threaddto.FeedParentContext{
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
				Stats:           threaddto.FeedStats{Likes: p.LikeCount, Reposts: p.RepostCount, Bookmarks: p.BookmarkCount, Comments: p.CommentCount, Shares: p.ShareCount},
				CreatedAt:       dtoutils.Wrap(p.CreatedAt),
				UpdatedAt:       dtoutils.Wrap(p.UpdatedAt),
			}
		}
		data[i] = item
	}
	sort.SliceStable(data, func(i, j int) bool {
		return ctrlshared.BoolPtrValue(data[i].IsPinned) && !ctrlshared.BoolPtrValue(data[j].IsPinned)
	})
	return data
}

func (c *ThreadController) ensureFalseViewerFlags(data []threaddto.FeedItemResponse) {
	ctrlshared.EnsureFalseFeedViewerFlags(data)
}
