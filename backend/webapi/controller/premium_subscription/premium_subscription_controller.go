package premiumsubscriptionctrl

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	premiumsubscriptionrepoiface "github.com/sunibgram/domain/repository/premium_subscription"
	premiumsubscriptionsvciface "github.com/sunibgram/domain/service/premium_subscription"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
)

type PremiumSubscriptionController struct {
	premiumSubscriptionService premiumsubscriptionsvciface.IPremiumSubscriptionService
}

func NewPremiumSubscriptionController(premiumSubscriptionService premiumsubscriptionsvciface.IPremiumSubscriptionService) *PremiumSubscriptionController {
	return &PremiumSubscriptionController{premiumSubscriptionService: premiumSubscriptionService}
}

func (c *PremiumSubscriptionController) ActivateTrial(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	subscription, err := c.premiumSubscriptionService.ActivateTrial(claims.UserID)
	if err != nil {
		if errors.Is(err, premiumsubscriptionrepoiface.ErrTrialAlreadyActivated) {
			return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{"message": err.Error()})
		}
		return ctrlshared.HandleServiceError(ctx, err)
	}

	_ = subscription
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "subscribed successfully for one month"})
}
