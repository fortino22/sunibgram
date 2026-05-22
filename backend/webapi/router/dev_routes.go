package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/webapi/controller"
)

func registerDevRoutes(api fiber.Router, devSubscriptionCtrl *controller.DevSubscriptionController, auth fiber.Handler) {
	dev := api.Group("/dev")

	dev.Use(auth)
	dev.Post("/subscriptions/expire", devSubscriptionCtrl.ExpireForDev)
}
