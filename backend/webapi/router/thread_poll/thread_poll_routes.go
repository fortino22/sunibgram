package threadpollroute

import (
	"github.com/gofiber/fiber/v2"

	threadpollctrl "github.com/sunibgram/webapi/controller/thread_poll"
)

func RegisterThreadPollRoutes(
	api fiber.Router,
	pollCtrl *threadpollctrl.ThreadPollController,
	auth fiber.Handler,
) {
	threadsAuth := api.Group("/threads", auth)
	threadsAuth.Post("/:id/poll/vote", pollCtrl.Vote)
}
