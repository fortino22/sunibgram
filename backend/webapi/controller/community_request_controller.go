package controller

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	utils "github.com/sunibgram/webapi/dto/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	communityrequestdto "github.com/sunibgram/webapi/dto/community_request"
	threaddto "github.com/sunibgram/webapi/dto/thread"
)

type CommunityRequestController struct {
	requestService service.ICommunityRequestService
	profileService service.IUserProfileService
}

func NewCommunityRequestController(requestService service.ICommunityRequestService, profileService service.IUserProfileService) *CommunityRequestController {
	return &CommunityRequestController{requestService: requestService, profileService: profileService}
}

func (c *CommunityRequestController) Send(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	if err := c.requestService.Send(ctx.Context(), id, claims.UserID); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "join request sent"})
}

func (c *CommunityRequestController) Cancel(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	if err := c.requestService.Cancel(ctx.Context(), id, claims.UserID); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "join request cancelled"})
}

func (c *CommunityRequestController) Approve(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	targetID, ok := ParseIDParam(ctx, "userId", "user")
	if !ok {
		return nil
	}

	if err := c.requestService.Approve(ctx.Context(), id, claims.UserID, targetID); err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "request approved"})
}

func (c *CommunityRequestController) Decline(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	targetID, ok := ParseIDParam(ctx, "userId", "user")
	if !ok {
		return nil
	}

	if err := c.requestService.Decline(ctx.Context(), id, claims.UserID, targetID); err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "request declined"})
}

func (c *CommunityRequestController) GetPending(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	page, limit := ParsePagination(ctx)

	requests, total, err := c.requestService.GetPending(ctx.Context(), id, claims.UserID, page, limit)
	if err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	avatarMap, err := c.requestAvatarMap(requests)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(communityrequestdto.PaginatedCommunityRequestResponse{
		Data: c.toCommunityRequestResponses(requests, avatarMap),
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *CommunityRequestController) toCommunityRequestResponse(r *model.CommunityRequest, avatarMap map[uint]string) communityrequestdto.CommunityRequestResponse {
	return communityrequestdto.CommunityRequestResponse{
		CommunityID: r.CommunityID,
		UserID:      r.UserID,
		Name:        r.User.Name,
		Username:    r.User.UserName,
		AvatarURL:   avatarMap[r.UserID],
		IsPremium:   r.User.IsPremium,
		Status:      string(r.Status),
		CreatedAt:   utils.Wrap(r.CreatedAt),
	}
}

func (c *CommunityRequestController) toCommunityRequestResponses(requests []model.CommunityRequest, avatarMap map[uint]string) []communityrequestdto.CommunityRequestResponse {
	out := make([]communityrequestdto.CommunityRequestResponse, len(requests))
	for i := range requests {
		out[i] = c.toCommunityRequestResponse(&requests[i], avatarMap)
	}
	return out
}

func (c *CommunityRequestController) requestAvatarMap(requests []model.CommunityRequest) (map[uint]string, error) {
	if len(requests) == 0 {
		return map[uint]string{}, nil
	}
	userIDs := make([]uint, 0, len(requests))
	for _, r := range requests {
		userIDs = append(userIDs, r.UserID)
	}
	profiles, err := c.profileService.GetProfiles(userIDs)
	if err != nil {
		return nil, err
	}
	avatarMap := make(map[uint]string, len(profiles))
	for _, p := range profiles {
		avatarMap[p.UserID] = BuildMediaURL(p.AvatarObjectName, p.AvatarBucket, c.profileService.GetURL)
	}
	return avatarMap, nil
}
