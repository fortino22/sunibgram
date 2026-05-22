package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/webapi/controller"
	"github.com/sunibgram/webapi/middleware"
)

func registerProfileRoutes(api fiber.Router, ctrl *controller.UserProfileController, auth fiber.Handler) {
	profile := api.Group("/profile")
	profile.Get("/me", auth, ctrl.GetMyProfile)
	profile.Put("/me", auth, ctrl.UpdateProfile)
	profile.Get("/:id", middleware.OptionalAuth(), ctrl.GetProfile)
}
