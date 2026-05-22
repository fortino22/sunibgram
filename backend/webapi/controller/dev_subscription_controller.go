package controller

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/domain/service"
)

type DevSubscriptionController struct {
	PremiumSubscriptionService service.IPremiumSubscriptionService
}

func NewDevSubscriptionController(premiumSubscriptionService service.IPremiumSubscriptionService) *DevSubscriptionController {
	return &DevSubscriptionController{PremiumSubscriptionService: premiumSubscriptionService}
}

func (c *DevSubscriptionController) ExpireForDev(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	if err := c.PremiumSubscriptionService.ExpireForDev(claims.UserID); err != nil {
		return HandleServiceError(ctx, err)
	}

	return ctx.JSON(fiber.Map{"message": "subscription expired for development"})
}
