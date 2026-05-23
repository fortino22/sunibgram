package communitythreadpinctrl

import (
	"github.com/gofiber/fiber/v2"

	communitythreadpinsvciface "github.com/sunibgram/domain/service/community_thread_pin"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
)

type CommunityThreadPinController struct {
	pinService communitythreadpinsvciface.ICommunityThreadPinService
}

func NewCommunityThreadPinController(pinService communitythreadpinsvciface.ICommunityThreadPinService) *CommunityThreadPinController {
	return &CommunityThreadPinController{pinService: pinService}
}

func (c *CommunityThreadPinController) Pin(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	communityID, ok := ctrlshared.ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}
	threadID, ok := ctrlshared.ParseIDParam(ctx, "threadId", "thread")
	if !ok {
		return nil
	}

	if err := c.pinService.Pin(ctx.Context(), communityID, claims.UserID, threadID); err != nil {
		return pinErrorResponse(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "thread pinned successfully"})
}

func (c *CommunityThreadPinController) Unpin(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	communityID, ok := ctrlshared.ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}
	threadID, ok := ctrlshared.ParseIDParam(ctx, "threadId", "thread")
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
		return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
	case "community not found", "thread not found":
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
	default:
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
}
