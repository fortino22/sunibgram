package eventpinctrl

import (
	"github.com/gofiber/fiber/v2"

	eventsvciface "github.com/sunibgram/domain/service/event"
	eventpinsvciface "github.com/sunibgram/domain/service/event_pin"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	eventdto "github.com/sunibgram/webapi/dto/event"
	threaddto "github.com/sunibgram/webapi/dto/thread"
)

type EventPinController struct {
	pinService   eventpinsvciface.IEventPinService
	eventService eventsvciface.IEventService
}

func NewEventPinController(pinService eventpinsvciface.IEventPinService, eventService eventsvciface.IEventService) *EventPinController {
	return &EventPinController{pinService: pinService, eventService: eventService}
}

func (c *EventPinController) Pin(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	if err := c.pinService.Pin(claims.UserID, ctx.Params("id")); err != nil {
		return eventPinErrorResponse(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "event pinned successfully"})
}

func (c *EventPinController) Unpin(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	if err := c.pinService.Unpin(claims.UserID, ctx.Params("id")); err != nil {
		return eventPinErrorResponse(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "event unpinned successfully"})
}

func (c *EventPinController) GetMyPins(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	page, limit := ctrlshared.ParsePagination(ctx)
	events, total, err := c.pinService.GetMyPins(claims.UserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}
	data := make([]eventdto.EventMiniResponse, len(events))
	for i := range events {
		data[i] = eventdto.EventMiniResponse{
			ID:        events[i].ID,
			Title:     events[i].Title,
			BannerURL: ctrlshared.BuildMediaURL(events[i].BannerObjectName, c.eventService.GetURL),
		}
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": data,
		"meta": threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func eventPinErrorResponse(ctx *fiber.Ctx, err error) error {
	switch err.Error() {
	case "event not found":
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
	default:
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
}
