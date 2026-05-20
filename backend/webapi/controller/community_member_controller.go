package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/dto/utils"
)

type CommunityMemberController struct {
	memberService  service.ICommunityMemberService
	profileService service.IUserProfileService
}

func NewCommunityMemberController(memberService service.ICommunityMemberService, profileService service.IUserProfileService) *CommunityMemberController {
	return &CommunityMemberController{memberService: memberService, profileService: profileService}
}

func (c *CommunityMemberController) Join(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	if err := c.memberService.Join(ctx.Context(), id, claims.UserID); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "joined community successfully"})
}

func (c *CommunityMemberController) Leave(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	if err := c.memberService.Leave(ctx.Context(), id, claims.UserID); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "left community successfully"})
}

func (c *CommunityMemberController) GetMembers(ctx *fiber.Ctx) error {
	id, ok := ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	page, limit := ParsePagination(ctx)

	members, total, err := c.memberService.GetMembers(id, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedCommunityMemberResponse{
		Data: c.toMemberResponses(members),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *CommunityMemberController) RemoveMember(ctx *fiber.Ctx) error {
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

	if err := c.memberService.RemoveMember(ctx.Context(), id, claims.UserID, targetID); err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "member removed successfully"})
}

func (c *CommunityMemberController) UpdateRole(ctx *fiber.Ctx) error {
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

	var body struct {
		Role string `json:"role"`
	}
	if err := ctx.BodyParser(&body); err != nil || body.Role == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "role is required"})
	}

	role := model.CommunityMemberRole(body.Role)
	if role != model.CommunityMemberRoleMember && role != model.CommunityMemberRoleModerator && role != model.CommunityMemberRoleAdmin {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid role"})
	}

	if err := c.memberService.UpdateRole(ctx.Context(), id, claims.UserID, targetID, role); err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "role updated successfully"})
}

func (c *CommunityMemberController) toMemberResponse(m *model.CommunityMember) dto.CommunityMemberResponse {
	avatarURL := ""
	if m.Profile != nil {
		avatarURL = BuildMediaURL(m.Profile.AvatarObjectName, m.Profile.AvatarBucket, c.profileService.GetURL)
	}
	return dto.CommunityMemberResponse{
		CommunityID: m.CommunityID,
		UserID:      m.UserID,
		Username:    m.User.UserName,
		AvatarURL:   avatarURL,
		Role:        string(m.Role),
		JoinedAt:    utils.Wrap(m.JoinedAt),
	}
}

func (c *CommunityMemberController) toMemberResponses(members []model.CommunityMember) []dto.CommunityMemberResponse {
	out := make([]dto.CommunityMemberResponse, len(members))
	for i := range members {
		out[i] = c.toMemberResponse(&members[i])
	}
	return out
}
