package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/webapi/controller"
)

func registerUserRoutes(api fiber.Router, followCtrl *controller.FollowController, userCtrl *controller.UserController, auth fiber.Handler) {
	users := api.Group("/users")
	users.Get("/me/profile", auth, userCtrl.GetMyProfile)
	users.Get("/:id/profile", userCtrl.GetProfile)
	users.Post("/:id/follow", auth, followCtrl.Follow)
	users.Delete("/:id/unfollow", auth, followCtrl.Unfollow)
}
