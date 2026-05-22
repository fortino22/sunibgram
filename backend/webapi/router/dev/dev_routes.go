package devroute

import (
	"github.com/gofiber/fiber/v2"

	premiumsubscriptionctrl "github.com/sunibgram/webapi/controller/premium_subscription"
)

func RegisterDevRoutes(api fiber.Router, devSubscriptionCtrl *premiumsubscriptionctrl.DevSubscriptionController, auth fiber.Handler) {
	dev := api.Group("/dev")

	dev.Use(auth)
	dev.Post("/subscriptions/expire", devSubscriptionCtrl.ExpireForDev)
}
