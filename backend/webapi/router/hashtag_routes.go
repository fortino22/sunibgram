package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/webapi/controller"
)

func registerHashtagRoutes(api fiber.Router, ctrl *controller.HashtagController, auth fiber.Handler) {
	api.Get("/hashtags/trending", auth, ctrl.GetTrending)
}
