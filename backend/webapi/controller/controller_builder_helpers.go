package controller

import (
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	webapiutils "github.com/sunibgram/webapi/utils"
)

func newFeedBuilder(
	media service.IThreadMediaService,
	like service.ILikeService,
	bookmark service.IBookmarkService,
	thread service.IThreadService,
	pin service.IThreadPinService,
	profile service.IUserProfileService,
	poll service.IThreadPollService,
	viewerID *uint,
) webapiutils.FeedBuilder {
	return webapiutils.FeedBuilder{
		Media:       media,
		LikeSvc:     like,
		BookmarkSvc: bookmark,
		ThreadSvc:   thread,
		PinSvc:      pin,
		PollSvc:     poll,
		ProfileSvc:  profile,
		ViewerID:    viewerID,
	}
}

func loadThreadMediaByThread(mediaService service.IThreadMediaService, threads []model.Thread) map[uint][]model.ThreadMedia {
	ids := make([]uint, 0, len(threads))
	for _, thread := range threads {
		ids = append(ids, thread.ThreadID)
		if thread.OriginalThread != nil {
			ids = append(ids, thread.OriginalThread.ThreadID)
		}
	}
	return loadThreadMediaByID(mediaService, ids)
}

func loadThreadMediaByID(mediaService service.IThreadMediaService, threadIDs []uint) map[uint][]model.ThreadMedia {
	allMedia, _ := mediaService.GetByThreadIDs(threadIDs)
	mediaByThread := make(map[uint][]model.ThreadMedia, len(threadIDs))
	for _, media := range allMedia {
		mediaByThread[media.ThreadID] = append(mediaByThread[media.ThreadID], media)
	}
	return mediaByThread
}

func buildFeedItemsFromThreads(threads []model.Thread, mediaByThread map[uint][]model.ThreadMedia, fb webapiutils.FeedBuilder) []dto.FeedItemResponse {
	fb.LoadViewerFlagsForThreads(threads)
	fb.LoadPollsForThreads(threads)
	data := make([]dto.FeedItemResponse, len(threads))
	for i, thread := range threads {
		mediaID := thread.ThreadID
		if thread.Type == model.ThreadTypeRepost && thread.OriginalThread != nil {
			mediaID = thread.OriginalThread.ThreadID
		}
		data[i] = fb.ToFeedItem(&thread, mediaByThread[mediaID])
	}
	return data
}

func ensureFalseFeedViewerFlags(data []dto.FeedItemResponse) {
	webapiutils.EnsureFalseViewerFlags(data)
}

func boolPtrValue(v *bool) bool {
	return v != nil && *v
}
