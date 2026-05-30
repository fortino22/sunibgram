package threadroute

import (
	"github.com/gofiber/fiber/v2"

	bookmarkctrl "github.com/sunibgram/webapi/controller/bookmark"
	likectrl "github.com/sunibgram/webapi/controller/like"
	repostctrl "github.com/sunibgram/webapi/controller/repost"
	threadctrl "github.com/sunibgram/webapi/controller/thread"
	threadpinctrl "github.com/sunibgram/webapi/controller/thread_pin"
	threadsummaryctrl "github.com/sunibgram/webapi/controller/thread_summary"
)

func RegisterThreadRoutes(
	api fiber.Router,
	threadCtrl *threadctrl.ThreadController,
	likeCtrl *likectrl.LikeController,
	bookmarkCtrl *bookmarkctrl.BookmarkController,
	threadPinCtrl *threadpinctrl.ThreadPinController,
	repostCtrl *repostctrl.RepostController,
	summaryCtrl *threadsummaryctrl.ThreadSummaryController,
	auth fiber.Handler,
) {
	api.Get("/likes", auth, likeCtrl.GetMyLikes)
	api.Get("/bookmarks", auth, bookmarkCtrl.GetMyBookmarks)

	threads := api.Group("/threads")
	threads.Get("/mention", auth, threadCtrl.Mention)
	threads.Get("/", auth, threadCtrl.GetRoots)

	api.Post("/threads-event", auth, threadCtrl.CreateEventThread)

	threadsAuth := api.Group("/threads", auth)
	threadsAuth.Post("/", threadCtrl.Create)
	threadsAuth.Post("/feed/refresh", threadCtrl.RefreshFeed)
	threadsAuth.Get("/feed/new-count", threadCtrl.GetNewFeedCount)
	threadsAuth.Get("/feed/since", threadCtrl.GetNewForYouSince)
	threadsAuth.Get("/following", threadCtrl.GetFollowingFeed)
	threadsAuth.Get("/faculty", threadCtrl.GetFacultyThreads)
	threadsAuth.Get("/user/:userId/media", threadCtrl.GetMyMediaThread)
	threadsAuth.Get("/user/:userID", threadCtrl.GetMyPosts)
	threadsAuth.Get("/user/:userID/replies", threadCtrl.GetMyReplies)
	threadsAuth.Get("/:id/summary", summaryCtrl.GetSummary)
	threadsAuth.Get("/:id", threadCtrl.GetByID)
	threadsAuth.Get("/:id/chain", threadCtrl.GetChain)
	threadsAuth.Put("/:id", threadCtrl.Update)
	threadsAuth.Patch("/:id/reply-permission", threadCtrl.UpdateReplyPermission)
	threadsAuth.Delete("/:id", threadCtrl.Delete)
	threadsAuth.Post("/:id/pin", threadPinCtrl.Pin)
	threadsAuth.Delete("/:id/unpin", threadPinCtrl.Unpin)

	threads.Post("/:id/like", likeCtrl.Like)
	threads.Delete("/:id/unlike", likeCtrl.Unlike)
	threads.Post("/:id/bookmark", bookmarkCtrl.Bookmark)
	threads.Delete("/:id/unbookmark", bookmarkCtrl.Unbookmark)
	threads.Post("/:id/repost", repostCtrl.Repost)
	threads.Delete("/:id/unrepost", repostCtrl.Unrepost)
}
