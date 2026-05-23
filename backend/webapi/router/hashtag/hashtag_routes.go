package hashtagroute

import (
	"github.com/gofiber/fiber/v2"

	hashtagctrl "github.com/sunibgram/webapi/controller/hashtag"
)

func RegisterHashtagRoutes(api fiber.Router, ctrl *hashtagctrl.HashtagController, auth fiber.Handler) {
	api.Get("/hashtags/trending", auth, ctrl.GetTrending)
}
