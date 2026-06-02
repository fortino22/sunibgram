package userroute

import (
	"github.com/gofiber/fiber/v2"

	explorectrl "github.com/sunibgram/webapi/controller/explore"
	followctrl "github.com/sunibgram/webapi/controller/follow"
	notificationusermutectrl "github.com/sunibgram/webapi/controller/notification_user_mute"
	userctrl "github.com/sunibgram/webapi/controller/user"
	userblockctrl "github.com/sunibgram/webapi/controller/user_block"
)

func RegisterUserRoutes(api fiber.Router, followCtrl *followctrl.FollowController, exploreCtrl *explorectrl.ExploreController, blockCtrl *userblockctrl.UserBlockController, muteCtrl *notificationusermutectrl.NotificationUserMuteController, userCtrl *userctrl.UserController, auth fiber.Handler) {
	users := api.Group("/users")
	users.Get("/count", userCtrl.CountRegularUsers) // public — no auth
	users.Get("/suggestions/faculty", auth, exploreCtrl.GetFacultySuggestions)
	users.Get("/suggestions/mutual", auth, exploreCtrl.GetMutualSuggestions)
	users.Get("/suggested/preview", auth, exploreCtrl.GetSuggestedPeoplePreview)
	users.Get("/suggested", auth, exploreCtrl.GetSuggestedPeople)
	users.Get("/onboarding/suggestions", auth, exploreCtrl.GetOnboardingSuggestions)
	users.Post("/:id/follow", auth, followCtrl.Follow)
	users.Delete("/:id/unfollow", auth, followCtrl.Unfollow)
	users.Get("/me/blocks", auth, blockCtrl.GetMyBlockedUsers)
	users.Post("/:id/block", auth, blockCtrl.Block)
	users.Delete("/:id/unblock", auth, blockCtrl.Unblock)
	users.Get("/me/mutes", auth, muteCtrl.GetMyMutedUsers)
	users.Post("/:id/mute", auth, muteCtrl.Mute)
	users.Delete("/:id/unmute", auth, muteCtrl.Unmute)

	follows := api.Group("/follows")
	follows.Get("/me/following", auth, followCtrl.GetMyFollowing)
	follows.Get("/me/followers", auth, followCtrl.GetMyFollowers)
	follows.Get("/:user_id/following", auth, followCtrl.GetFollowing)
	follows.Get("/:user_id/followers", auth, followCtrl.GetFollowers)
}
