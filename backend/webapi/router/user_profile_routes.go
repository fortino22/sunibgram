package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/webapi/controller"
)

func registerProfileRoutes(api fiber.Router, ctrl *controller.UserProfileController, auth fiber.Handler) {
	profile := api.Group("/profile")
	profile.Get("/me", auth, ctrl.GetMyProfile)
	profile.Get("/:id", ctrl.GetProfile)
}
