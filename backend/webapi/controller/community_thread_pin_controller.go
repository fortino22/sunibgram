package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
)

type CommunityThreadPinController struct {
	pinService service.ICommunityThreadPinService
}

func NewCommunityThreadPinController(pinService service.ICommunityThreadPinService) *CommunityThreadPinController {
	return &CommunityThreadPinController{pinService: pinService}
}

func (c *CommunityThreadPinController) Pin(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	communityID, ok := ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}
	threadID, ok := ParseIDParam(ctx, "threadId", "thread")
	if !ok {
		return nil
	}

	if err := c.pinService.Pin(ctx.Context(), communityID, claims.UserID, threadID); err != nil {
		return pinErrorResponse(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "thread pinned successfully"})
}

func (c *CommunityThreadPinController) Unpin(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	communityID, ok := ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}
	threadID, ok := ParseIDParam(ctx, "threadId", "thread")
	if !ok {
		return nil
	}

	if err := c.pinService.Unpin(ctx.Context(), communityID, claims.UserID, threadID); err != nil {
		return pinErrorResponse(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "thread unpinned successfully"})
}

func pinErrorResponse(ctx *fiber.Ctx, err error) error {
	switch err.Error() {
	case "forbidden":
		return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
	case "community not found", "thread not found":
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	default:
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
}
