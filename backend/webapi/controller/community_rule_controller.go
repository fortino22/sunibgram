package controller

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	utils "github.com/sunibgram/webapi/dto/_shared"
	communityruledto "github.com/sunibgram/webapi/dto/community_rule"
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

	return ctx.Status(fiber.StatusOK).JSON(communityruledto.CommunityRuleListResponse{Data: c.toResponses(rules)})
}

func (c *CommunityRuleController) toResponses(rules []model.CommunityRule) []communityruledto.CommunityRuleResponse {
	out := make([]communityruledto.CommunityRuleResponse, len(rules))
	for i := range rules {
		out[i] = communityruledto.CommunityRuleResponse{
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
