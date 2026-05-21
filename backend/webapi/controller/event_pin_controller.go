package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
)

type EventPinController struct {
	pinService   service.IEventPinService
	eventService service.IEventService
}

func NewEventPinController(pinService service.IEventPinService, eventService service.IEventService) *EventPinController {
	return &EventPinController{pinService: pinService, eventService: eventService}
}

func (c *EventPinController) Pin(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	if err := c.pinService.Pin(claims.UserID, ctx.Params("id")); err != nil {
		return eventPinErrorResponse(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "event pinned successfully"})
}

func (c *EventPinController) Unpin(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	if err := c.pinService.Unpin(claims.UserID, ctx.Params("id")); err != nil {
		return eventPinErrorResponse(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "event unpinned successfully"})
}

func (c *EventPinController) GetMyPins(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	page, limit := ParsePagination(ctx)
	events, total, err := c.pinService.GetMyPins(claims.UserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: "internal server error"})
	}
	data := make([]dto.EventMiniResponse, len(events))
	for i := range events {
		data[i] = dto.EventMiniResponse{
			ID:        events[i].ID,
			Title:     events[i].Title,
			BannerURL: BuildMediaURL(events[i].BannerObjectName, events[i].BannerBucket, c.eventService.GetURL),
		}
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": data,
		"meta": dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func eventPinErrorResponse(ctx *fiber.Ctx, err error) error {
	switch err.Error() {
	case "event not found":
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	default:
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
}
