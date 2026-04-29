package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/webapi/controller"
)

func registerThreadRoutes(
	api fiber.Router,
	threadCtrl *controller.ThreadController,
	threadMediaCtrl *controller.ThreadMediaController,
	likeCtrl *controller.LikeController,
	bookmarkCtrl *controller.BookmarkController,
	repostCtrl *controller.RepostController,
	auth fiber.Handler,
) {
	api.Get("/likes", auth, likeCtrl.GetMyLikes)
	api.Get("/bookmarks", auth, bookmarkCtrl.GetMyBookmarks)
	api.Get("/reposts", auth, repostCtrl.GetMyReposts)

	threads := api.Group("/threads", auth)
	threads.Post("/", threadCtrl.Create)
	threads.Get("/", threadCtrl.GetRoots)
	threads.Get("/me", threadCtrl.GetMyPosts)
	threads.Get("/:id", threadCtrl.GetByID)
	threads.Get("/:id/chain", threadCtrl.GetChain)
	threads.Put("/:id", threadCtrl.Update)
	threads.Delete("/:id", threadCtrl.Delete)

	threads.Post("/:id/media", threadMediaCtrl.Upload)
	threads.Put("/:id/media", threadMediaCtrl.Update)
	threads.Delete("/:id/media", threadMediaCtrl.Delete)

	threads.Post("/:id/like", likeCtrl.Like)
	threads.Delete("/:id/unlike", likeCtrl.Unlike)
	threads.Post("/:id/bookmark", bookmarkCtrl.Bookmark)
	threads.Delete("/:id/unbookmark", bookmarkCtrl.Unbookmark)
	threads.Post("/:id/repost", repostCtrl.Repost)
	threads.Delete("/:id/unrepost", repostCtrl.Unrepost)
}
