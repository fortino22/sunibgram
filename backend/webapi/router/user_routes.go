package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/webapi/controller"
)

func registerUserRoutes(api fiber.Router, followCtrl *controller.FollowController, auth fiber.Handler) {
	users := api.Group("/users")
	users.Post("/:id/follow", auth, followCtrl.Follow)
	users.Delete("/:id/unfollow", auth, followCtrl.Unfollow)
}
