package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
)

type ThreadPinController struct {
	threadPinService service.IThreadPinService
}

func NewThreadPinController(threadPinService service.IThreadPinService) *ThreadPinController {
	return &ThreadPinController{threadPinService: threadPinService}
}

func (c *ThreadPinController) Pin(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	threadID, ok := ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}

	if err := c.threadPinService.Pin(claims.UserID, threadID); err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		if err.Error() == "thread not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "pinned successfully"})
}

func (c *ThreadPinController) Unpin(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	threadID, ok := ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}

	if err := c.threadPinService.Unpin(claims.UserID, threadID); err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "unpinned successfully"})
}
