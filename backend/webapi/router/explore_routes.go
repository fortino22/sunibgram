package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/webapi/controller"
)

func registerExploreRoutes(api fiber.Router, ctrl *controller.ExploreController, auth fiber.Handler) {
	api.Get("/explore/search", auth, ctrl.ExploreSearch)
}
