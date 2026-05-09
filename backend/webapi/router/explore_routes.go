package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/webapi/controller"
)

func registerExploreRoutes(api fiber.Router, ctrl *controller.ExploreController, optionalAuth fiber.Handler) {
	api.Get("/explore/search", optionalAuth, ctrl.ExploreSearch)
}
