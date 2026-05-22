package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/webapi/controller"
)

func registerThreadPollRoutes(
	api fiber.Router,
	pollCtrl *controller.ThreadPollController,
	auth fiber.Handler,
) {
	threadsAuth := api.Group("/threads", auth)
	threadsAuth.Post("/:id/poll/vote", pollCtrl.Vote)
}
