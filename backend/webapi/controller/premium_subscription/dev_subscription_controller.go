package premiumsubscriptionctrl

import (
	"github.com/gofiber/fiber/v2"

	premiumsubscriptionsvciface "github.com/sunibgram/domain/service/premium_subscription"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
)

type DevSubscriptionController struct {
	PremiumSubscriptionService premiumsubscriptionsvciface.IPremiumSubscriptionService
}

func NewDevSubscriptionController(premiumSubscriptionService premiumsubscriptionsvciface.IPremiumSubscriptionService) *DevSubscriptionController {
	return &DevSubscriptionController{PremiumSubscriptionService: premiumSubscriptionService}
}

func (c *DevSubscriptionController) ExpireForDev(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	if err := c.PremiumSubscriptionService.ExpireForDev(claims.UserID); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	return ctx.JSON(fiber.Map{"message": "subscription expired for development"})
}
