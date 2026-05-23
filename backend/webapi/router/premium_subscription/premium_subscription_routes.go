package premiumsubscriptionroute

import (
	"github.com/gofiber/fiber/v2"

	premiumsubscriptionctrl "github.com/sunibgram/webapi/controller/premium_subscription"
)

func RegisterPremiumSubscriptionRoutes(api fiber.Router, premiumSubscriptionCtrl *premiumsubscriptionctrl.PremiumSubscriptionController, auth fiber.Handler) {
	subscriptions := api.Group("/subscriptions")

	subscriptions.Use(auth)
	subscriptions.Post("/trial/activate", premiumSubscriptionCtrl.ActivateTrial)
}
