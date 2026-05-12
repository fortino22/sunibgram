package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/dto/utils"
)

type CommunityRequestController struct {
	requestService service.ICommunityRequestService
}

func NewCommunityRequestController(requestService service.ICommunityRequestService) *CommunityRequestController {
	return &CommunityRequestController{requestService: requestService}
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
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
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
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
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
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
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
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
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
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedCommunityRequestResponse{
		Data: toCommunityRequestResponses(requests),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func toCommunityRequestResponse(r *model.CommunityRequest) dto.CommunityRequestResponse {
	return dto.CommunityRequestResponse{
		CommunityID: r.CommunityID,
		UserID:      r.UserID,
		Username:    r.User.UserName,
		Status:      string(r.Status),
		CreatedAt:   utils.Wrap(r.CreatedAt),
	}
}

func toCommunityRequestResponses(requests []model.CommunityRequest) []dto.CommunityRequestResponse {
	out := make([]dto.CommunityRequestResponse, len(requests))
	for i := range requests {
		out[i] = toCommunityRequestResponse(&requests[i])
	}
	return out
}
