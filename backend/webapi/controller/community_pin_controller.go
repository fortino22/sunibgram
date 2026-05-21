package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
)

type CommunityPinController struct {
	pinService       service.ICommunityPinService
	communityService service.ICommunityService
}

func NewCommunityPinController(pinService service.ICommunityPinService, communityService service.ICommunityService) *CommunityPinController {
	return &CommunityPinController{pinService: pinService, communityService: communityService}
}

func (c *CommunityPinController) Pin(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	communityID, ok := ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}
	if err := c.pinService.Pin(claims.UserID, communityID); err != nil {
		return communityPinErrorResponse(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "community pinned successfully"})
}

func (c *CommunityPinController) Unpin(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	communityID, ok := ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}
	if err := c.pinService.Unpin(claims.UserID, communityID); err != nil {
		return communityPinErrorResponse(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "community unpinned successfully"})
}

func (c *CommunityPinController) GetMyPins(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	page, limit := ParsePagination(ctx)
	communities, total, err := c.pinService.GetMyPins(claims.UserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: "internal server error"})
	}
	slim := make([]dto.MyCommunityResponse, len(communities))
	for i := range communities {
		slim[i] = dto.MyCommunityResponse{
			CommunityID:   communities[i].CommunityID,
			CommunityName: communities[i].Name,
			BannerURL:     BuildMediaURL(communities[i].BannerObjectName, communities[i].BannerBucket, c.communityService.GetURL),
		}
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": slim,
		"meta": dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func communityPinErrorResponse(ctx *fiber.Ctx, err error) error {
	switch err.Error() {
	case "community not found":
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	default:
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
}
