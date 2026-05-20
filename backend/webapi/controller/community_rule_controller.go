package controller

import (
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/dto/utils"
)

type CommunityRuleController struct {
	ruleService service.ICommunityRuleService
}

func NewCommunityRuleController(ruleService service.ICommunityRuleService) *CommunityRuleController {
	return &CommunityRuleController{ruleService: ruleService}
}

func (c *CommunityRuleController) List(ctx *fiber.Ctx) error {
	id, ok := ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	rules, err := c.ruleService.ListByCommunity(id)
	if err != nil {
		return HandleServiceError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.CommunityRuleListResponse{Data: c.toResponses(rules)})
}

func (c *CommunityRuleController) Create(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	communityID, ok := ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}
	req, ok := parseCommunityRuleRequest(ctx)
	if !ok {
		return nil
	}

	rule, err := c.ruleService.Create(ctx.Context(), communityID, claims.UserID, req.OrderNum, req.Text)
	if err != nil {
		return HandleServiceError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(c.toResponse(rule))
}

func (c *CommunityRuleController) Update(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	communityID, ok := ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}
	ruleID, ok := ParseIDParam(ctx, "ruleID", "rule")
	if !ok {
		return nil
	}
	req, ok := parseCommunityRuleRequest(ctx)
	if !ok {
		return nil
	}

	rule, err := c.ruleService.Update(ctx.Context(), communityID, claims.UserID, ruleID, req.OrderNum, req.Text)
	if err != nil {
		return HandleServiceError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(c.toResponse(rule))
}

func (c *CommunityRuleController) Delete(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	communityID, ok := ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}
	ruleID, ok := ParseIDParam(ctx, "ruleID", "rule")
	if !ok {
		return nil
	}

	if err := c.ruleService.Delete(ctx.Context(), communityID, claims.UserID, ruleID); err != nil {
		return HandleServiceError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "rule deleted successfully"})
}

func parseCommunityRuleRequest(ctx *fiber.Ctx) (dto.CommunityRuleRequest, bool) {
	var req dto.CommunityRuleRequest
	if err := ctx.BodyParser(&req); err != nil {
		_ = ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
		return req, false
	}
	req.Text = strings.TrimSpace(req.Text)
	if req.Text == "" {
		_ = ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "text is required"})
		return req, false
	}
	if utf8.RuneCountInString(req.Text) > 500 {
		_ = ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "text must be 500 characters or less"})
		return req, false
	}
	if req.OrderNum < 1 {
		_ = ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "order_num must be at least 1"})
		return req, false
	}
	return req, true
}

func (c *CommunityRuleController) toResponse(rule *model.CommunityRule) dto.CommunityRuleResponse {
	return dto.CommunityRuleResponse{
		ID:          rule.ID,
		CommunityID: rule.CommunityID,
		OrderNum:    rule.OrderNum,
		Text:        rule.Text,
		CreatedAt:   utils.Wrap(rule.CreatedAt),
		UpdatedAt:   utils.Wrap(rule.UpdatedAt),
	}
}

func (c *CommunityRuleController) toResponses(rules []model.CommunityRule) []dto.CommunityRuleResponse {
	out := make([]dto.CommunityRuleResponse, len(rules))
	for i := range rules {
		out[i] = c.toResponse(&rules[i])
	}
	return out
}
