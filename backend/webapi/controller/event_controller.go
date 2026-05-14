package controller

import (
	"context"
	"log"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/dto/utils"
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
	if _, ok := RequireAuth(ctx); !ok {
		return nil
	}
	req, banner, attachments, err := parseCreateEventRequest(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	_, err = c.eventService.Create(ctx.Context(), req.Title, req.Description, req.Location, req.Type, req.StartTime.Time, req.EndTime.Time, banner, attachments)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "success"})
}

func (c *EventController) GetAll(ctx *fiber.Ctx) error {
	if _, ok := RequireAuth(ctx); !ok {
		return nil
	}
	_, limit := ParsePagination(ctx)
	trending, running, future, err := c.eventService.GetAll(limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	fctx := ctx.Context()
	return ctx.Status(fiber.StatusOK).JSON(dto.EventGroupsResponse{
		Trending: c.toResponses(fctx, trending),
		Running:  c.toResponses(fctx, running),
		Future:   c.toResponses(fctx, future),
	})
}

func (c *EventController) GetPending(ctx *fiber.Ctx) error {
	if _, ok := RequireAuth(ctx); !ok {
		return nil
	}
	_, limit := ParsePagination(ctx)
	events, err := c.eventService.GetPending(limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(c.toPendingResponses(ctx.Context(), events))
}

func (c *EventController) GetByID(ctx *fiber.Ctx) error {
	if _, ok := RequireAuth(ctx); !ok {
		return nil
	}
	event, err := c.eventService.GetByID(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: "event not found"})
	}
	return ctx.Status(fiber.StatusOK).JSON(c.toResponse(ctx.Context(), event))
}

func (c *EventController) Update(ctx *fiber.Ctx) error {
	if _, ok := RequireAuth(ctx); !ok {
		return nil
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
	if _, ok := RequireAuth(ctx); !ok {
		return nil
	}
	if err := c.eventService.Delete(ctx.Context(), ctx.Params("id")); err != nil {
		if err.Error() == "event not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "event deleted successfully"})
}

func (c *EventController) ApproveEvent(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	if claims.RoleName != "admin" {
		return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
	}
	adminID := strconv.FormatUint(uint64(claims.UserID), 10)
	err := c.eventService.ApproveEvent(ctx.Params("id"), adminID)
	if err != nil {
		if err.Error() == "event not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Event approved successfully"})
}

func (c *EventController) DeclineEvent(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	if claims.RoleName != "admin" {
		return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
	}
	var req dto.DeclineEventRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}
	adminID := strconv.FormatUint(uint64(claims.UserID), 10)
	err := c.eventService.DeclineEvent(ctx.Params("id"), adminID, req.Reason)
	if err != nil {
		if err.Error() == "event not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Event declined successfully"})
}

func (c *EventController) RequestRevision(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	if claims.RoleName != "admin" {
		return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
	}
	var req dto.RevisionRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}
	adminID := strconv.FormatUint(uint64(claims.UserID), 10)
	if err := c.eventService.RequestRevision(ctx.Params("id"), adminID, req.Note); err != nil {
		if err.Error() == "event not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Event revision requested successfully"})
}

func (c *EventController) GetEventReviews(ctx *fiber.Ctx) error {
	if _, ok := RequireAuth(ctx); !ok {
		return nil
	}
	reviews, err := c.eventService.GetEventReviews(ctx.Params("id"))
	if err != nil {
		if err.Error() == "event not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(c.toEventReviewResponses(reviews))
}

func (c *EventController) GetThreads(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	page, limit := ParsePagination(ctx)
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

func (c *EventController) toResponse(ctx context.Context, event *model.Event) dto.EventResponse {
	bannerURL := BuildMediaURL(event.BannerObjectName, event.BannerBucket, c.eventService.GetURL)
	var attachmentURL string
	if event.AttachmentObjectName != "" {
		var err error
		attachmentURL, err = c.eventService.GeneratePresignedAttachmentURL(ctx, event.AttachmentBucket, event.AttachmentObjectName)
		if err != nil {
			log.Printf("presign error bucket=%s object=%s: %v", event.AttachmentBucket, event.AttachmentObjectName, err)
		}
	}
	pdfURLs := c.eventPDFURLs(ctx, event)
	return dto.EventResponse{
		ID:               event.ID,
		Title:            event.Title,
		Description:      event.Description,
		Location:         event.Location,
		Type:             event.Type,
		Status:           event.Status,
		StartTime:        dto.EventTime{Time: event.StartTime},
		EndTime:          dto.EventTime{Time: event.EndTime},
		BannerURL:        bannerURL,
		AttachmentURL:    attachmentURL,
		PDFURLs:          pdfURLs,
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

func (c *EventController) toResponses(ctx context.Context, events []model.Event) []dto.EventResponse {
	out := make([]dto.EventResponse, len(events))
	for i := range events {
		out[i] = c.toResponse(ctx, &events[i])
	}
	return out
}

func (c *EventController) toEventReviewResponses(reviews []model.EventReview) []dto.EventReviewResponse {
	out := make([]dto.EventReviewResponse, len(reviews))
	for i := range reviews {
		out[i] = dto.EventReviewResponse{
			ID:        reviews[i].ID,
			EventID:   reviews[i].EventID,
			AdminID:   reviews[i].AdminID,
			Action:    reviews[i].Action,
			Note:      reviews[i].Note,
			CreatedAt: utils.Wrap(reviews[i].CreatedAt),
		}
	}
	return out
}

type pendingEventResponse struct {
	ID            string              `json:"id"`
	Title         string              `json:"title"`
	Description   string              `json:"description"`
	Type          string              `json:"type"`
	Status        string              `json:"status"`
	Location      string              `json:"location"`
	StartTime     dto.EventTime       `json:"start_time"`
	EndTime       dto.EventTime       `json:"end_time"`
	BannerURL     string              `json:"banner_url,omitempty"`
	AttachmentURL string              `json:"attachment_url,omitempty"`
	PDFURLs       []string            `json:"pdf_urls,omitempty"`
	CreatedAt     utils.FormattedTime `json:"created_at"`
	UpdatedAt     utils.FormattedTime `json:"updated_at"`
}

func (c *EventController) toPendingResponses(ctx context.Context, events []model.Event) []pendingEventResponse {
	out := make([]pendingEventResponse, len(events))
	for i := range events {
		e := &events[i]
		var attachmentURL string
		if e.AttachmentObjectName != "" {
			var err error
			attachmentURL, err = c.eventService.GeneratePresignedAttachmentURL(ctx, e.AttachmentBucket, e.AttachmentObjectName)
			if err != nil {
				log.Printf("presign error bucket=%s object=%s: %v", e.AttachmentBucket, e.AttachmentObjectName, err)
			}
		}
		out[i] = pendingEventResponse{
			ID:            e.ID,
			Title:         e.Title,
			Description:   e.Description,
			Type:          e.Type,
			Status:        e.Status,
			Location:      e.Location,
			StartTime:     dto.EventTime{Time: e.StartTime},
			EndTime:       dto.EventTime{Time: e.EndTime},
			BannerURL:     BuildMediaURL(e.BannerObjectName, e.BannerBucket, c.eventService.GetURL),
			AttachmentURL: attachmentURL,
			PDFURLs:       c.eventPDFURLs(ctx, e),
			CreatedAt:     utils.Wrap(e.CreatedAt),
			UpdatedAt:     utils.Wrap(e.UpdatedAt),
		}
	}
	return out
}

func (c *EventController) eventPDFURLs(ctx context.Context, event *model.Event) []string {
	pdfs := event.PDFs
	if len(pdfs) == 0 && event.AttachmentObjectName != "" {
		pdfs = model.EventPDFs{{ObjectName: event.AttachmentObjectName, Bucket: event.AttachmentBucket}}
	}
	if len(pdfs) == 0 {
		return nil
	}
	urls := make([]string, 0, len(pdfs))
	for _, pdf := range pdfs {
		if pdf.ObjectName == "" {
			continue
		}
		url, err := c.eventService.GeneratePresignedAttachmentURL(ctx, pdf.Bucket, pdf.ObjectName)
		if err != nil {
			log.Printf("presign error bucket=%s object=%s: %v", pdf.Bucket, pdf.ObjectName, err)
			continue
		}
		urls = append(urls, url)
	}
	return urls
}

func parseCreateEventRequest(ctx *fiber.Ctx) (dto.CreateEventRequest, *multipart.FileHeader, []*multipart.FileHeader, error) {
	var req dto.CreateEventRequest
	if isMultipart(ctx) {
		if err := fillEventRequestFromForm(ctx, &req.Title, &req.Description, &req.Location, &req.Type, &req.StartTime, &req.EndTime); err != nil {
			return req, nil, nil, err
		}
		banner, _ := ctx.FormFile("banner")
		attachments := eventPDFFiles(ctx)
		if err := validateEventPDFHeaders(attachments); err != nil {
			return req, nil, nil, err
		}
		return req, banner, attachments, nil
	}
	if err := ctx.BodyParser(&req); err != nil {
		return req, nil, nil, err
	}
	return req, nil, nil, nil
}

func eventPDFFiles(ctx *fiber.Ctx) []*multipart.FileHeader {
	form, _ := ctx.MultipartForm()
	if form == nil {
		return nil
	}
	return form.File["attachment"]
}

func validateEventPDFHeaders(files []*multipart.FileHeader) error {
	if err := model.ValidateEventPDFCount(len(files)); err != nil {
		return err
	}
	for _, file := range files {
		if err := model.ValidateEventAttachmentExt(file.Filename); err != nil {
			return err
		}
		if err := model.ValidateEventAttachmentContentType(file.Header.Get("Content-Type")); err != nil {
			return err
		}
	}
	return nil
}

func parseUpdateEventRequest(ctx *fiber.Ctx) (dto.UpdateEventRequest, *multipart.FileHeader, bool, error) {
	var req dto.UpdateEventRequest
	if isMultipart(ctx) {
		if err := fillEventRequestFromForm(ctx, &req.Title, &req.Description, &req.Location, nil, &req.StartTime, &req.EndTime); err != nil {
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

func fillEventRequestFromForm(ctx *fiber.Ctx, title, description, location, eventType *string, startTime, endTime *dto.EventTime) error {
	*title = ctx.FormValue("title")
	*description = ctx.FormValue("description")
	*location = ctx.FormValue("location")
	if eventType != nil {
		*eventType = ctx.FormValue("type")
	}
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
