package repostctrl

import (
	"github.com/gofiber/fiber/v2"

	threadsvciface "github.com/sunibgram/domain/service/thread"
	threadmediasvciface "github.com/sunibgram/domain/service/thread_media"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
)

type RepostController struct {
	threadService threadsvciface.IThreadService
	mediaService  threadmediasvciface.IThreadMediaService
}

func NewRepostController(threadService threadsvciface.IThreadService, mediaService threadmediasvciface.IThreadMediaService) *RepostController {
	return &RepostController{threadService: threadService, mediaService: mediaService}
}

func (c *RepostController) Repost(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	threadID, ok := ctrlshared.ParseIDParam(ctx, "id", "thread")
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
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	threadID, ok := ctrlshared.ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}

	if err := c.threadService.Unrepost(ctx.Context(), claims.UserID, threadID); err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "unreposted successfully"})
}
