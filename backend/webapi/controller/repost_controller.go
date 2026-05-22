package controller

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/domain/service"
	authdto "github.com/sunibgram/webapi/dto/auth"
)

type RepostController struct {
	threadService service.IThreadService
	mediaService  service.IThreadMediaService
}

func NewRepostController(threadService service.IThreadService, mediaService service.IThreadMediaService) *RepostController {
	return &RepostController{threadService: threadService, mediaService: mediaService}
}

func (c *RepostController) Repost(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	threadID, ok := ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}

	if err := c.threadService.Repost(ctx.Context(), claims.UserID, threadID); err != nil {
		if err.Error() == "thread not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusConflict).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "reposted successfully"})
}

func (c *RepostController) Unrepost(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	threadID, ok := ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}

	if err := c.threadService.Unrepost(ctx.Context(), claims.UserID, threadID); err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "unreposted successfully"})
}
