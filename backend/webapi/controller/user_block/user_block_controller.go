package userblockctrl

import (
	"github.com/gofiber/fiber/v2"

	userblocksvciface "github.com/sunibgram/domain/service/user_block"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
)

type UserBlockController struct {
	blockService userblocksvciface.IUserBlockService
}

func NewUserBlockController(blockService userblocksvciface.IUserBlockService) *UserBlockController {
	return &UserBlockController{blockService: blockService}
}

func (c *UserBlockController) Block(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	targetID, ok := ctrlshared.ParseIDParam(ctx, "id", "user")
	if !ok {
		return nil
	}
	if err := c.blockService.Block(claims.UserID, targetID); err != nil {
		if err.Error() == "user not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusConflict).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "blocked successfully"})
}

func (c *UserBlockController) Unblock(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	targetID, ok := ctrlshared.ParseIDParam(ctx, "id", "user")
	if !ok {
		return nil
	}
	if err := c.blockService.Unblock(claims.UserID, targetID); err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "unblocked successfully"})
}
