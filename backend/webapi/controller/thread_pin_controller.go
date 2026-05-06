package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/webapi/ctxutil"
)

type ThreadPinController struct {
	threadPinService service.IThreadPinService
}

func NewThreadPinController(threadPinService service.IThreadPinService) *ThreadPinController {
	return &ThreadPinController{threadPinService: threadPinService}
}

func (c *ThreadPinController) Pin(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	threadID, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid thread id"})
	}

	if err := c.threadPinService.Pin(claims.UserID, uint(threadID)); err != nil {
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
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	threadID, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid thread id"})
	}

	if err := c.threadPinService.Unpin(claims.UserID, uint(threadID)); err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "unpinned successfully"})
}
