package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/dto/utils"
	"github.com/sunibgram/webapi/ctxutil"
)

type EventController struct {
	eventService     service.IEventService
	threadService    service.IThreadService
	mediaService     service.IThreadMediaService
	likeService      service.ILikeService
	bookmarkService  service.IBookmarkService
	threadPinService service.IThreadPinService
}

func NewEventController(eventService service.IEventService, threadService service.IThreadService, mediaService service.IThreadMediaService, likeService service.ILikeService, bookmarkService service.IBookmarkService, threadPinService service.IThreadPinService) *EventController {
	return &EventController{eventService: eventService, threadService: threadService, mediaService: mediaService, likeService: likeService, bookmarkService: bookmarkService, threadPinService: threadPinService}
}

func (c *EventController) Create(ctx *fiber.Ctx) error {
	if _, ok := ctxutil.Get(ctx); !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}
	var req dto.CreateEventRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}
	_, err := c.eventService.Create(ctx.Context(), req.Title, req.Description, req.Location, req.StartTime.Time, req.EndTime.Time)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "success"})
}

func (c *EventController) GetAll(ctx *fiber.Ctx) error {
	if _, ok := ctxutil.Get(ctx); !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))
	events, total, err := c.eventService.GetAll(page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedEventResponse{
		Data: c.toResponses(events),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *EventController) GetByID(ctx *fiber.Ctx) error {
	if _, ok := ctxutil.Get(ctx); !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}
	event, err := c.eventService.GetByID(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: "event not found"})
	}
	return ctx.Status(fiber.StatusOK).JSON(c.toResponse(event))
}

func (c *EventController) Update(ctx *fiber.Ctx) error {
	if _, ok := ctxutil.Get(ctx); !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}
	var req dto.UpdateEventRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}
	_, err := c.eventService.Update(ctx.Context(), ctx.Params("id"), req.Title, req.Description, req.Location, req.StartTime.Time, req.EndTime.Time)
	if err != nil {
		if err.Error() == "event not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "success"})
}

func (c *EventController) Delete(ctx *fiber.Ctx) error {
	if _, ok := ctxutil.Get(ctx); !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}
	if err := c.eventService.Delete(ctx.Context(), ctx.Params("id")); err != nil {
		if err.Error() == "event not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "event deleted successfully"})
}

func (c *EventController) GetThreads(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))
	threads, total, err := c.eventService.GetThreads(ctx.Context(), ctx.Params("id"), page, limit)
	if err != nil {
		if err.Error() == "event not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	tc := &ThreadController{threadService: c.threadService, mediaService: c.mediaService, likeService: c.likeService, bookmarkService: c.bookmarkService, threadPinService: c.threadPinService}
	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFeedResponse{
		Data: tc.buildFeedItemsNoPin(threads, &claims.UserID),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *EventController) toResponse(event *model.Event) dto.EventResponse {
	return dto.EventResponse{
		ID:          event.ID,
		Title:       event.Title,
		Description: event.Description,
		Location:    event.Location,
		StartTime:   dto.EventTime{Time: event.StartTime},
		EndTime:     dto.EventTime{Time: event.EndTime},
		CreatedAt:   utils.Wrap(event.CreatedAt),
		UpdatedAt:   utils.Wrap(event.UpdatedAt),
	}
}

func (c *EventController) toResponses(events []model.Event) []dto.EventResponse {
	out := make([]dto.EventResponse, len(events))
	for i := range events {
		out[i] = c.toResponse(&events[i])
	}
	return out
}
