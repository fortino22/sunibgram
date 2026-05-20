package controller

import (
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

func (c *CommunityRuleController) toResponses(rules []model.CommunityRule) []dto.CommunityRuleResponse {
	out := make([]dto.CommunityRuleResponse, len(rules))
	for i := range rules {
		out[i] = dto.CommunityRuleResponse{
			ID:          rules[i].ID,
			CommunityID: rules[i].CommunityID,
			OrderNum:    rules[i].OrderNum,
			Text:        rules[i].Text,
			CreatedAt:   utils.Wrap(rules[i].CreatedAt),
			UpdatedAt:   utils.Wrap(rules[i].UpdatedAt),
		}
	}
	return out
}
