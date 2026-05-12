package controller

import (
	"mime/multipart"
	"strconv"
	"strings"

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
	req, banner, err := parseCreateEventRequest(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}
	_, err = c.eventService.Create(ctx.Context(), req.Title, req.Description, req.Location, req.StartTime.Time, req.EndTime.Time, banner)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "success"})
}

func (c *EventController) GetAll(ctx *fiber.Ctx) error {
	if _, ok := ctxutil.Get(ctx); !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))
	trending, running, future, err := c.eventService.GetAll(limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(dto.EventGroupsResponse{
		Trending: c.toResponses(trending),
		Running:  c.toResponses(running),
		Future:   c.toResponses(future),
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
	req, banner, clearBanner, err := parseUpdateEventRequest(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}
	_, err = c.eventService.Update(ctx.Context(), ctx.Params("id"), req.Title, req.Description, req.Location, req.StartTime.Time, req.EndTime.Time, banner, clearBanner)
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
	bannerURL := ""
	if event.BannerObjectName != "" {
		bannerURL = c.eventService.GetURL(event.BannerBucket, event.BannerObjectName)
	}
	return dto.EventResponse{
		ID:               event.ID,
		Title:            event.Title,
		Description:      event.Description,
		Location:         event.Location,
		StartTime:        dto.EventTime{Time: event.StartTime},
		EndTime:          dto.EventTime{Time: event.EndTime},
		BannerURL:        bannerURL,
		ParticipantCount: event.ParticipantCount,
		ThreadCount:      event.ThreadCount,
		LikeCount:        event.LikeCount,
		RepostCount:      event.RepostCount,
		WeekThreadCount:  event.WeekThreadCount,
		WeekLikeCount:    event.WeekLikeCount,
		WeekRepostCount:  event.WeekRepostCount,
		CreatedAt:        utils.Wrap(event.CreatedAt),
		UpdatedAt:        utils.Wrap(event.UpdatedAt),
	}
}

func (c *EventController) toResponses(events []model.Event) []dto.EventResponse {
	out := make([]dto.EventResponse, len(events))
	for i := range events {
		out[i] = c.toResponse(&events[i])
	}
	return out
}

func parseCreateEventRequest(ctx *fiber.Ctx) (dto.CreateEventRequest, *multipart.FileHeader, error) {
	var req dto.CreateEventRequest
	if isMultipart(ctx) {
		if err := fillEventRequestFromForm(ctx, &req.Title, &req.Description, &req.Location, &req.StartTime, &req.EndTime); err != nil {
			return req, nil, err
		}
		banner, _ := ctx.FormFile("banner")
		return req, banner, nil
	}
	if err := ctx.BodyParser(&req); err != nil {
		return req, nil, err
	}
	return req, nil, nil
}

func parseUpdateEventRequest(ctx *fiber.Ctx) (dto.UpdateEventRequest, *multipart.FileHeader, bool, error) {
	var req dto.UpdateEventRequest
	if isMultipart(ctx) {
		if err := fillEventRequestFromForm(ctx, &req.Title, &req.Description, &req.Location, &req.StartTime, &req.EndTime); err != nil {
			return req, nil, false, err
		}
		banner, _ := ctx.FormFile("banner")
		return req, banner, ctx.FormValue("clear_banner") == "true", nil
	}
	if err := ctx.BodyParser(&req); err != nil {
		return req, nil, false, err
	}
	return req, nil, false, nil
}

func fillEventRequestFromForm(ctx *fiber.Ctx, title, description, location *string, startTime, endTime *dto.EventTime) error {
	*title = ctx.FormValue("title")
	*description = ctx.FormValue("description")
	*location = ctx.FormValue("location")
	if err := parseEventFormTime(ctx.FormValue("start_time"), startTime); err != nil {
		return err
	}
	if err := parseEventFormTime(ctx.FormValue("end_time"), endTime); err != nil {
		return err
	}
	return nil
}

func parseEventFormTime(value string, target *dto.EventTime) error {
	value = strings.TrimSpace(value)
	if value == "" {
		target.Time = target.Time.UTC()
		return nil
	}
	return target.UnmarshalJSON([]byte(strconv.Quote(value)))
}

func isMultipart(ctx *fiber.Ctx) bool {
	return strings.Contains(strings.ToLower(ctx.Get("Content-Type")), "multipart/form-data")
}
