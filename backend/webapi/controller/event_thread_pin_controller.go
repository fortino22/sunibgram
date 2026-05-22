package controller

import (
	"github.com/gofiber/fiber/v2"

	eventthreadpinsvciface "github.com/sunibgram/domain/service/event_thread_pin"
	authdto "github.com/sunibgram/webapi/dto/auth"
)

type EventThreadPinController struct {
	pinService eventthreadpinsvciface.IEventThreadPinService
}

func NewEventThreadPinController(pinService eventthreadpinsvciface.IEventThreadPinService) *EventThreadPinController {
	return &EventThreadPinController{pinService: pinService}
}

func (c *EventThreadPinController) Pin(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	eventID := ctx.Params("id")
	threadID, ok := ParseIDParam(ctx, "threadId", "thread")
	if !ok {
		return nil
	}
	if err := c.pinService.Pin(ctx.Context(), eventID, claims.UserID, threadID); err != nil {
		return eventThreadPinErrorResponse(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "event thread pinned successfully"})
}

func (c *EventThreadPinController) Unpin(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	eventID := ctx.Params("id")
	threadID, ok := ParseIDParam(ctx, "threadId", "thread")
	if !ok {
		return nil
	}
	if err := c.pinService.Unpin(ctx.Context(), eventID, claims.UserID, threadID); err != nil {
		return eventThreadPinErrorResponse(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "event thread unpinned successfully"})
}

func eventThreadPinErrorResponse(ctx *fiber.Ctx, err error) error {
	switch err.Error() {
	case "forbidden":
		return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
	case "event not found", "thread not found":
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
	default:
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
}
