package bookmarkcollectionroute

import (
	"github.com/gofiber/fiber/v2"

	bookmarkcollectionctrl "github.com/sunibgram/webapi/controller/bookmark_collection"
)

func RegisterBookmarkCollectionRoutes(
	api fiber.Router,
	ctrl *bookmarkcollectionctrl.BookmarkCollectionController,
	auth fiber.Handler,
) {
	g := api.Group("/bookmark-collections", auth)
	g.Post("/", ctrl.Create)
	g.Get("/", ctrl.List)
	g.Get("/:id", ctrl.Get)
	g.Patch("/:id", ctrl.Update)
	g.Delete("/:id", ctrl.Delete)
	g.Get("/:id/threads", ctrl.ListThreads)
	g.Post("/:id/threads", ctrl.AddThread)
	g.Delete("/:id/threads/:threadId", ctrl.RemoveThread)
}
