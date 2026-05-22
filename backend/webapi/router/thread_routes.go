package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/webapi/controller"
)

func registerThreadRoutes(
	api fiber.Router,
	threadCtrl *controller.ThreadController,
	likeCtrl *controller.LikeController,
	bookmarkCtrl *controller.BookmarkController,
	threadPinCtrl *controller.ThreadPinController,
	repostCtrl *controller.RepostController,
	summaryCtrl *controller.ThreadSummaryController,
	auth fiber.Handler,
) {
	api.Get("/likes", auth, likeCtrl.GetMyLikes)
	api.Get("/bookmarks", auth, bookmarkCtrl.GetMyBookmarks)

	threads := api.Group("/threads")
	threads.Get("/mention", auth, threadCtrl.Mention)
	threads.Get("/", auth, threadCtrl.GetRoots)

	threadsAuth := api.Group("/threads", auth)
	threadsAuth.Post("/", threadCtrl.Create)
	threadsAuth.Post("/feed/refresh", threadCtrl.RefreshFeed)
	threadsAuth.Get("/feed/new-count", threadCtrl.GetNewFeedCount)
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
