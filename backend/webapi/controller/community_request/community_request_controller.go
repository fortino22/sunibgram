package communityrequestctrl

import (
	"github.com/gofiber/fiber/v2"

	communityrequestmodel "github.com/sunibgram/domain/model/community_request"
	communityrequestsvciface "github.com/sunibgram/domain/service/community_request"
	userprofilesvciface "github.com/sunibgram/domain/service/user_profile"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	utils "github.com/sunibgram/webapi/dto/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	communityrequestdto "github.com/sunibgram/webapi/dto/community_request"
	threaddto "github.com/sunibgram/webapi/dto/thread"
)

type CommunityRequestController struct {
	requestService communityrequestsvciface.ICommunityRequestService
	profileService userprofilesvciface.IUserProfileService
}

func NewCommunityRequestController(requestService communityrequestsvciface.ICommunityRequestService, profileService userprofilesvciface.IUserProfileService) *CommunityRequestController {
	return &CommunityRequestController{requestService: requestService, profileService: profileService}
}

func (c *CommunityRequestController) Send(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ctrlshared.ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	if err := c.requestService.Send(ctx.Context(), id, claims.UserID); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "join request sent"})
}

func (c *CommunityRequestController) Cancel(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ctrlshared.ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	if err := c.requestService.Cancel(ctx.Context(), id, claims.UserID); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "join request cancelled"})
}

func (c *CommunityRequestController) Approve(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ctrlshared.ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	targetID, ok := ctrlshared.ParseIDParam(ctx, "userId", "user")
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
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ctrlshared.ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	targetID, ok := ctrlshared.ParseIDParam(ctx, "userId", "user")
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
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ctrlshared.ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	page, limit := ctrlshared.ParsePagination(ctx)

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

func (c *CommunityRequestController) toCommunityRequestResponse(r *communityrequestmodel.CommunityRequest, avatarMap map[uint]string) communityrequestdto.CommunityRequestResponse {
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

func (c *CommunityRequestController) toCommunityRequestResponses(requests []communityrequestmodel.CommunityRequest, avatarMap map[uint]string) []communityrequestdto.CommunityRequestResponse {
	out := make([]communityrequestdto.CommunityRequestResponse, len(requests))
	for i := range requests {
		out[i] = c.toCommunityRequestResponse(&requests[i], avatarMap)
	}
	return out
}

func (c *CommunityRequestController) requestAvatarMap(requests []communityrequestmodel.CommunityRequest) (map[uint]string, error) {
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
		avatarMap[p.UserID] = ctrlshared.BuildMediaURL(p.AvatarObjectName, c.profileService.GetURL)
	}
	return avatarMap, nil
}
