package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/webapi/controller"
)

func registerPremiumSubscriptionRoutes(api fiber.Router, premiumSubscriptionCtrl *controller.PremiumSubscriptionController, auth fiber.Handler) {
	subscriptions := api.Group("/subscriptions")

	subscriptions.Use(auth)
	subscriptions.Post("/trial/activate", premiumSubscriptionCtrl.ActivateTrial)
}
