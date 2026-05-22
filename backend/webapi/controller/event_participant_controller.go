package controller

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/domain/service"
	authdto "github.com/sunibgram/webapi/dto/auth"
	eventdto "github.com/sunibgram/webapi/dto/event"
	threaddto "github.com/sunibgram/webapi/dto/thread"
)

type EventParticipantController struct {
	partService  service.IEventParticipantService
	eventService service.IEventService
}

func NewEventParticipantController(partService service.IEventParticipantService, eventService service.IEventService) *EventParticipantController {
	return &EventParticipantController{partService: partService, eventService: eventService}
}

func (c *EventParticipantController) Join(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	if err := c.partService.Join(claims.UserID, ctx.Params("id")); err != nil {
		return eventParticipantErrorResponse(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "joined event successfully"})
}

func (c *EventParticipantController) Leave(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	if err := c.partService.Leave(claims.UserID, ctx.Params("id")); err != nil {
		return eventParticipantErrorResponse(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "left event successfully"})
}

func (c *EventParticipantController) GetMyJoined(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	page, limit := ParsePagination(ctx)
	events, total, err := c.partService.GetMyJoined(claims.UserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}
	data := make([]eventdto.EventMiniResponse, len(events))
	for i := range events {
		data[i] = eventdto.EventMiniResponse{
			ID:        events[i].ID,
			Title:     events[i].Title,
			BannerURL: BuildMediaURL(events[i].BannerObjectName, events[i].BannerBucket, c.eventService.GetURL),
		}
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": data,
		"meta": threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func eventParticipantErrorResponse(ctx *fiber.Ctx, err error) error {
	switch err.Error() {
	case "event not found":
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
	default:
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
}
