package communitypinctrl

import (
	"github.com/gofiber/fiber/v2"

	communitysvciface "github.com/sunibgram/domain/service/community"
	communitypinsvciface "github.com/sunibgram/domain/service/community_pin"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	communitydto "github.com/sunibgram/webapi/dto/community"
	threaddto "github.com/sunibgram/webapi/dto/thread"
)

type CommunityPinController struct {
	pinService       communitypinsvciface.ICommunityPinService
	communityService communitysvciface.ICommunityService
}

func NewCommunityPinController(pinService communitypinsvciface.ICommunityPinService, communityService communitysvciface.ICommunityService) *CommunityPinController {
	return &CommunityPinController{pinService: pinService, communityService: communityService}
}

func (c *CommunityPinController) Pin(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	communityID, ok := ctrlshared.ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}
	if err := c.pinService.Pin(claims.UserID, communityID); err != nil {
		return communityPinErrorResponse(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "community pinned successfully"})
}

func (c *CommunityPinController) Unpin(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	communityID, ok := ctrlshared.ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}
	if err := c.pinService.Unpin(claims.UserID, communityID); err != nil {
		return communityPinErrorResponse(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "community unpinned successfully"})
}

func (c *CommunityPinController) GetMyPins(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	page, limit := ctrlshared.ParsePagination(ctx)
	communities, total, err := c.pinService.GetMyPins(claims.UserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}
	slim := make([]communitydto.MyCommunityResponse, len(communities))
	for i := range communities {
		slim[i] = communitydto.MyCommunityResponse{
			CommunityID:   communities[i].CommunityID,
			CommunityName: communities[i].Name,
			BannerURL:     ctrlshared.BuildMediaURL(communities[i].BannerObjectName, communities[i].BannerBucket, c.communityService.GetURL),
		}
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": slim,
		"meta": threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func communityPinErrorResponse(ctx *fiber.Ctx, err error) error {
	switch err.Error() {
	case "community not found":
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
	default:
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
}
