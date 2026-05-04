package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/dto/utils"
	"github.com/sunibgram/webapi/ctxutil"
)

type CommunityMemberController struct {
	memberService service.ICommunityMemberService
}

func NewCommunityMemberController(memberService service.ICommunityMemberService) *CommunityMemberController {
	return &CommunityMemberController{memberService: memberService}
}

func (c *CommunityMemberController) Join(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid community id"})
	}

	if err := c.memberService.Join(ctx.Context(), uint(id), claims.UserID); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "joined community successfully"})
}

func (c *CommunityMemberController) Leave(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid community id"})
	}

	if err := c.memberService.Leave(ctx.Context(), uint(id), claims.UserID); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "left community successfully"})
}

func (c *CommunityMemberController) GetMembers(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid community id"})
	}

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))

	members, total, err := c.memberService.GetMembers(uint(id), page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedCommunityMemberResponse{
		Data: toCommunityMemberResponses(members),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func toCommunityMemberResponse(m *model.CommunityMember) dto.CommunityMemberResponse {
	return dto.CommunityMemberResponse{
		CommunityID: m.CommunityID,
		UserID:      m.UserID,
		Username:    m.User.UserName,
		Role:        string(m.Role),
		JoinedAt:    utils.Wrap(m.JoinedAt),
	}
}

func toCommunityMemberResponses(members []model.CommunityMember) []dto.CommunityMemberResponse {
	out := make([]dto.CommunityMemberResponse, len(members))
	for i := range members {
		out[i] = toCommunityMemberResponse(&members[i])
	}
	return out
}
