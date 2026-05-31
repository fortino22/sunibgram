package userprofileroute

import (
	"github.com/gofiber/fiber/v2"

	userprofilectrl "github.com/sunibgram/webapi/controller/user_profile"
	"github.com/sunibgram/webapi/middleware"
)

func RegisterProfileRoutes(api fiber.Router, ctrl *userprofilectrl.UserProfileController, auth fiber.Handler) {
	profile := api.Group("/profile")
	profile.Get("/me", auth, ctrl.GetMyProfile)
	profile.Put("/me", auth, ctrl.UpdateProfile)
	profile.Post("/me/onboarding/complete", auth, ctrl.CompleteOnboarding)
	profile.Get("/:id", middleware.OptionalAuth(), ctrl.GetProfile)
}
