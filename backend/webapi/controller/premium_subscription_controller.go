package controller

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
)

type PremiumSubscriptionController struct {
	premiumSubscriptionService service.IPremiumSubscriptionService
}

func NewPremiumSubscriptionController(premiumSubscriptionService service.IPremiumSubscriptionService) *PremiumSubscriptionController {
	return &PremiumSubscriptionController{premiumSubscriptionService: premiumSubscriptionService}
}

func (c *PremiumSubscriptionController) ActivateTrial(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	subscription, err := c.premiumSubscriptionService.ActivateTrial(claims.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrTrialAlreadyActivated) {
			return ctx.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return HandleServiceError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(toPremiumSubscriptionResponse(subscription))
}

func toPremiumSubscriptionResponse(subscription *model.PremiumSubscription) dto.PremiumSubscriptionResponse {
	return dto.PremiumSubscriptionResponse{
		ID:                    subscription.ID,
		UserID:                subscription.UserID,
		IsTrial:               subscription.IsTrial,
		SubscriptionStartDate: subscription.SubscriptionStartDate,
		SubscriptionEndDate:   subscription.SubscriptionEndDate,
		CreatedAt:             subscription.CreatedAt,
		UpdatedAt:             subscription.UpdatedAt,
	}
}
