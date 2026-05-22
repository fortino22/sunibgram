package controller

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	utils "github.com/sunibgram/webapi/dto/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	communitymemberdto "github.com/sunibgram/webapi/dto/community_member"
	threaddto "github.com/sunibgram/webapi/dto/thread"
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
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
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
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
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
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(communitymemberdto.PaginatedCommunityMemberResponse{
		Data: c.toMemberResponses(members),
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
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
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "member removed successfully"})
}

func (c *CommunityMemberController) GetCommunityUsers(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	var rolePtr *model.CommunityMemberRole
	if raw := strings.TrimSpace(ctx.Query("role")); raw != "" {
		role := model.CommunityMemberRole(strings.ToUpper(raw))
		if role != model.CommunityMemberRoleAdmin && role != model.CommunityMemberRoleMember {
			return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid role"})
		}
		rolePtr = &role
	}
	query := strings.TrimSpace(ctx.Query("query"))
	page, limit := ParsePagination(ctx)

	items, total, err := c.memberService.GetCommunityUsers(ctx.Context(), id, claims.UserID, rolePtr, query, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	data := make([]communitymemberdto.CommunityUserResponse, len(items))
	for i, item := range items {
		data[i] = communitymemberdto.CommunityUserResponse{
			UserID:      item.UserID,
			Username:    item.Username,
			Name:        item.Name,
			UserAvatar:  BuildMediaURL(item.AvatarObjectName, item.AvatarBucket, c.profileService.GetURL),
			IsPremium:   item.IsPremium,
			Role:        string(item.Role),
			IsFollowing: item.IsFollowing,
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(communitymemberdto.PaginatedCommunityUserResponse{
		Data: data,
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *CommunityMemberController) Kick(ctx *fiber.Ctx) error {
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
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "member kicked successfully"})
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
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "role is required"})
	}

	role := model.CommunityMemberRole(body.Role)
	if role != model.CommunityMemberRoleMember && role != model.CommunityMemberRoleAdmin {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid role"})
	}

	if err := c.memberService.UpdateRole(ctx.Context(), id, claims.UserID, targetID, role); err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "role updated successfully"})
}

func (c *CommunityMemberController) toMemberResponse(m *model.CommunityMember) communitymemberdto.CommunityMemberResponse {
	avatarURL := ""
	if m.Profile != nil {
		avatarURL = BuildMediaURL(m.Profile.AvatarObjectName, m.Profile.AvatarBucket, c.profileService.GetURL)
	}
	return communitymemberdto.CommunityMemberResponse{
		CommunityID: m.CommunityID,
		UserID:      m.UserID,
		Username:    m.User.UserName,
		AvatarURL:   avatarURL,
		Role:        string(m.Role),
		JoinedAt:    utils.Wrap(m.JoinedAt),
	}
}

func (c *CommunityMemberController) toMemberResponses(members []model.CommunityMember) []communitymemberdto.CommunityMemberResponse {
	out := make([]communitymemberdto.CommunityMemberResponse, len(members))
	for i := range members {
		out[i] = c.toMemberResponse(&members[i])
	}
	return out
}
