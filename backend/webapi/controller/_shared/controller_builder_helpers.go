package ctrlshared

import (
	threadmodel "github.com/sunibgram/domain/model/thread"
	threadmediamodel "github.com/sunibgram/domain/model/thread_media"
	bookmarksvciface "github.com/sunibgram/domain/service/bookmark"
	likesvciface "github.com/sunibgram/domain/service/like"
	threadsvciface "github.com/sunibgram/domain/service/thread"
	threadmediasvciface "github.com/sunibgram/domain/service/thread_media"
	threadpinsvciface "github.com/sunibgram/domain/service/thread_pin"
	threadpollsvciface "github.com/sunibgram/domain/service/thread_poll"
	userprofilesvciface "github.com/sunibgram/domain/service/user_profile"
	"github.com/sunibgram/infrastructure/cache"
	threaddto "github.com/sunibgram/webapi/dto/thread"
	webapiutils "github.com/sunibgram/webapi/utils"
)

// FeedBuilderCache holds the Redis cache shared across controllers via package-level init.
// Set from cmd/main.go via SetFeedBuilderCache.
var FeedBuilderCache *cache.Cache

// SetFeedBuilderCache injects the Redis cache used by the FeedBuilder for viewer-flag sets.
func SetFeedBuilderCache(c *cache.Cache) {
	FeedBuilderCache = c
}

func NewFeedBuilder(
	media threadmediasvciface.IThreadMediaService,
	like likesvciface.ILikeService,
	bookmark bookmarksvciface.IBookmarkService,
	thread threadsvciface.IThreadService,
	pin threadpinsvciface.IThreadPinService,
	profile userprofilesvciface.IUserProfileService,
	poll threadpollsvciface.IThreadPollService,
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
		Cache:       FeedBuilderCache,
		ViewerID:    viewerID,
	}
}

func LoadThreadMediaByThread(mediaService threadmediasvciface.IThreadMediaService, threads []threadmodel.Thread) map[uint][]threadmediamodel.ThreadMedia {
	ids := make([]uint, 0, len(threads))
	for _, thread := range threads {
		ids = append(ids, thread.ThreadID)
		if thread.OriginalThread != nil {
			ids = append(ids, thread.OriginalThread.ThreadID)
		}
	}
	return LoadThreadMediaByID(mediaService, ids)
}

func LoadThreadMediaByID(mediaService threadmediasvciface.IThreadMediaService, threadIDs []uint) map[uint][]threadmediamodel.ThreadMedia {
	allMedia, _ := mediaService.GetByThreadIDs(threadIDs)
	mediaByThread := make(map[uint][]threadmediamodel.ThreadMedia, len(threadIDs))
	for _, media := range allMedia {
		mediaByThread[media.ThreadID] = append(mediaByThread[media.ThreadID], media)
	}
	return mediaByThread
}

func BuildFeedItemsFromThreads(threads []threadmodel.Thread, mediaByThread map[uint][]threadmediamodel.ThreadMedia, fb webapiutils.FeedBuilder) []threaddto.FeedItemResponse {
	fb.LoadViewerFlagsForThreads(threads)
	fb.LoadPollsForThreads(threads)
	data := make([]threaddto.FeedItemResponse, len(threads))
	for i, thread := range threads {
		mediaID := thread.ThreadID
		if thread.Type == threadmodel.ThreadTypeRepost && thread.OriginalThread != nil {
			mediaID = thread.OriginalThread.ThreadID
		}
		data[i] = fb.ToFeedItem(&thread, mediaByThread[mediaID])
	}
	return data
}

func EnsureFalseFeedViewerFlags(data []threaddto.FeedItemResponse) {
	webapiutils.EnsureFalseViewerFlags(data)
}

func BoolPtrValue(v *bool) bool {
	return v != nil && *v
}
