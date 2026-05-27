package userroute

import (
	"github.com/gofiber/fiber/v2"

	explorectrl "github.com/sunibgram/webapi/controller/explore"
	followctrl "github.com/sunibgram/webapi/controller/follow"
)

func RegisterUserRoutes(api fiber.Router, followCtrl *followctrl.FollowController, exploreCtrl *explorectrl.ExploreController, auth fiber.Handler) {
	users := api.Group("/users")
	users.Get("/suggestions/faculty", auth, exploreCtrl.GetFacultySuggestions)
	users.Get("/suggestions/mutual", auth, exploreCtrl.GetMutualSuggestions)
	users.Get("/suggested/preview", auth, exploreCtrl.GetSuggestedPeoplePreview)
	users.Post("/:id/follow", auth, followCtrl.Follow)
	users.Delete("/:id/unfollow", auth, followCtrl.Unfollow)

	follows := api.Group("/follows")
	follows.Get("/me/following", auth, followCtrl.GetMyFollowing)
	follows.Get("/me/followers", auth, followCtrl.GetMyFollowers)
	follows.Get("/:user_id/following", auth, followCtrl.GetFollowing)
	follows.Get("/:user_id/followers", auth, followCtrl.GetFollowers)
}
