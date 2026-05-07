package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/webapi/controller"
)

func registerUserRoutes(api fiber.Router, followCtrl *controller.FollowController, auth fiber.Handler) {
	users := api.Group("/users")
	users.Post("/:id/follow", auth, followCtrl.Follow)
	users.Delete("/:id/unfollow", auth, followCtrl.Unfollow)

	follows := api.Group("/follows")
	follows.Get("/me/following", auth, followCtrl.GetMyFollowing)
	follows.Get("/me/followers", auth, followCtrl.GetMyFollowers)
	follows.Get("/:user_id/following", auth, followCtrl.GetFollowing)
	follows.Get("/:user_id/followers", auth, followCtrl.GetFollowers)
}
