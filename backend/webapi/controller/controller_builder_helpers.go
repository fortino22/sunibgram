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
	viewerID *uint,
) webapiutils.FeedBuilder {
	return webapiutils.FeedBuilder{
		Media:       media,
		LikeSvc:     like,
		BookmarkSvc: bookmark,
		ThreadSvc:   thread,
		PinSvc:      pin,
		ProfileSvc:  profile,
		ViewerID:    viewerID,
	}
}

func loadThreadMediaByThread(mediaService service.IThreadMediaService, threads []model.Thread) map[uint][]model.ThreadMedia {
	ids := make([]uint, len(threads))
	for i, thread := range threads {
		ids[i] = thread.ThreadID
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
	data := make([]dto.FeedItemResponse, len(threads))
	for i, thread := range threads {
		data[i] = fb.ToFeedItem(&thread, mediaByThread[thread.ThreadID])
	}
	return data
}

func ensureFalseFeedViewerFlags(data []dto.FeedItemResponse) {
	webapiutils.EnsureFalseViewerFlags(data)
}

func boolPtrValue(v *bool) bool {
	return v != nil && *v
}
