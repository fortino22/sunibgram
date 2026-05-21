package controller

import (
	"context"
	"errors"
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
	eventService        service.IEventService
	threadService       service.IThreadService
	mediaService        service.IThreadMediaService
	likeService         service.ILikeService
	bookmarkService     service.IBookmarkService
	threadPinService    service.IThreadPinService
	profileService      service.IUserProfileService
	pollService         service.IThreadPollService
	eventThreadPinSvc   service.IEventThreadPinService
}

func NewEventController(eventService service.IEventService, threadService service.IThreadService, mediaService service.IThreadMediaService, likeService service.ILikeService, bookmarkService service.IBookmarkService, threadPinService service.IThreadPinService, profileService service.IUserProfileService, pollService service.IThreadPollService, eventThreadPinSvc service.IEventThreadPinService) *EventController {
	return &EventController{eventService: eventService, threadService: threadService, mediaService: mediaService, likeService: likeService, bookmarkService: bookmarkService, threadPinService: threadPinService, profileService: profileService, pollService: pollService, eventThreadPinSvc: eventThreadPinSvc}
}

func (c *EventController) Create(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	req, banner, attachments, err := parseCreateEventRequest(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	_, err = c.eventService.Create(ctx.Context(), claims.UserID, req.Title, req.Description, req.Location, req.Type, req.StartTime.Time, req.EndTime.Time, banner, attachments)
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
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: "internal server error"})
	}
	return ctx.Status(fiber.StatusOK).JSON(dto.EventGroupsResponse{
		Trending: c.toListResponses(trending),
		Running:  c.toListResponses(running),
		Future:   c.toListResponses(future),
	})
}

func (c *EventController) GetPending(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	if claims.RoleName != "admin" {
		return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
	}
	page, limit := ParsePagination(ctx)
	events, reviews, total, err := c.eventService.GetAdminEvents(ctx.Query("status"), ctx.Query("name"), page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedEventWithReviewHistoryResponse{
		Data: c.toAdminEventReviewHistoryResponses(ctx.Context(), events, reviews),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *EventController) GetMyEvents(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	page, limit := ParsePagination(ctx)
	events, reviews, total, err := c.eventService.GetMyEvents(claims.UserID, ctx.Query("status"), page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedMyEventResponse{
		Data: c.toEventReviewHistoryResponses(ctx.Context(), events, reviews),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
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
	req, banner, clearBanner, attachments, err := parseUpdateEventRequest(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}
	_, err = c.eventService.Update(ctx.Context(), ctx.Params("id"), req.Title, req.Description, req.Location, req.Type, req.StartTime.Time, req.EndTime.Time, banner, clearBanner, attachments)
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

func (c *EventController) GetThreads(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	page, limit := ParsePagination(ctx)
	filter := ctx.Query("filter")
	sort := ctx.Query("sort", "latest")
	eventID := ctx.Params("id")
	threads, total, err := c.eventService.GetThreads(ctx.Context(), &claims.UserID, eventID, filter, sort, page, limit)
	if err != nil {
		if err.Error() == "event not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	var pinnedThreadID uint
	if filter == "official" && page == 1 && c.eventThreadPinSvc != nil {
		if pin, _ := c.eventThreadPinSvc.GetPinned(eventID); pin != nil {
			pinnedThreadID = pin.ThreadID
			for i, t := range threads {
				if t.ThreadID == pin.ThreadID {
					threads = append(threads[:i], threads[i+1:]...)
					break
				}
			}
			if pinned, ferr := c.threadService.GetByID(pin.ThreadID); ferr == nil && pinned != nil {
				threads = append([]model.Thread{*pinned}, threads...)
			} else {
				pinnedThreadID = 0
			}
		}
	}

	tc := &ThreadController{threadService: c.threadService, mediaService: c.mediaService, likeService: c.likeService, bookmarkService: c.bookmarkService, threadPinService: c.threadPinService, profileService: c.profileService, pollService: c.pollService}
	data := tc.buildFeedItemsNoPin(threads, &claims.UserID)
	if pinnedThreadID != 0 {
		for i := range data {
			if data[i].ID == pinnedThreadID {
				t := true
				data[i].PinnedByEventAdmin = &t
				break
			}
		}
	}
	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFeedResponse{
		Data: data,
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
		UserID:           event.UserID,
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

func (c *EventController) toListResponse(event *model.Event) dto.EventListResponse {
	return dto.EventListResponse{
		ID:          event.ID,
		UserID:      event.UserID,
		Title:       event.Title,
		Description: event.Description,
		Location:    event.Location,
		Type:        event.Type,
		StartTime:   dto.EventTime{Time: event.StartTime},
		EndTime:     dto.EventTime{Time: event.EndTime},
		BannerURL:   BuildMediaURL(event.BannerObjectName, event.BannerBucket, c.eventService.GetURL),
		CreatedAt:   utils.Wrap(event.CreatedAt),
		UpdatedAt:   utils.Wrap(event.UpdatedAt),
	}
}

func (c *EventController) toListResponses(events []model.Event) []dto.EventListResponse {
	out := make([]dto.EventListResponse, len(events))
	for i := range events {
		out[i] = c.toListResponse(&events[i])
	}
	return out
}

func (c *EventController) toEventReviewHistoryResponses(ctx context.Context, events []model.Event, reviews map[string][]model.EventReview) []dto.EventWithReviewHistoryResponse {
	out := make([]dto.EventWithReviewHistoryResponse, len(events))
	for i := range events {
		out[i] = dto.EventWithReviewHistoryResponse{
			EventResponse: c.toResponse(ctx, &events[i]),
			Notes:         c.toEventReviewNoteResponses(reviews[events[i].ID]),
		}
	}
	return out
}

func (c *EventController) toAdminEventReviewHistoryResponses(ctx context.Context, events []model.Event, reviews map[string][]model.EventReview) []dto.AdminEventWithReviewHistoryResponse {
	out := make([]dto.AdminEventWithReviewHistoryResponse, len(events))
	for i := range events {
		out[i] = dto.AdminEventWithReviewHistoryResponse{
			EventResponse: c.toResponse(ctx, &events[i]),
			Notes:         c.toEventAdminReviewNoteResponses(reviews[events[i].ID]),
		}
	}
	return out
}

func (c *EventController) toEventReviewNoteResponses(reviews []model.EventReview) []dto.EventReviewNoteResponse {
	out := make([]dto.EventReviewNoteResponse, len(reviews))
	for i := range reviews {
		out[i] = dto.EventReviewNoteResponse{
			RevisionNumber: i + 1,
			Notes:          reviews[i].Note,
			CreatedAt:      utils.Wrap(reviews[i].CreatedAt),
		}
	}
	return out
}

func (c *EventController) toEventAdminReviewNoteResponses(reviews []model.EventReview) []dto.EventAdminReviewNoteResponse {
	out := make([]dto.EventAdminReviewNoteResponse, len(reviews))
	for i := range reviews {
		out[i] = dto.EventAdminReviewNoteResponse{
			RevisionNumber: i + 1,
			Notes:          reviews[i].Note,
			AdminName:      reviews[i].Admin.Name,
			CreatedAt:      utils.Wrap(reviews[i].CreatedAt),
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
		if err := model.ValidateUploadSize(file, model.MaxDocumentUploadSize); err != nil {
			return err
		}
		if err := model.ValidateEventAttachmentExt(file.Filename); err != nil {
			return err
		}
	}
	return nil
}

func parseUpdateEventRequest(ctx *fiber.Ctx) (dto.UpdateEventRequest, *multipart.FileHeader, bool, []*multipart.FileHeader, error) {
	var req dto.UpdateEventRequest
	if isMultipart(ctx) {
		if err := fillEventRequestFromForm(ctx, &req.Title, &req.Description, &req.Location, &req.Type, &req.StartTime, &req.EndTime); err != nil {
			return req, nil, false, nil, err
		}
		banner, _ := ctx.FormFile("banner")
		attachments := eventPDFFiles(ctx)
		if err := validateEventPDFHeaders(attachments); err != nil {
			return req, nil, false, nil, err
		}
		clearBanner := ctx.FormValue("clear_banner") == "true"
		if banner != nil && clearBanner {
			return req, nil, false, nil, errors.New("cannot upload banner and clear_banner simultaneously")
		}
		return req, banner, clearBanner, attachments, nil
	}
	if err := ctx.BodyParser(&req); err != nil {
		return req, nil, false, nil, err
	}
	return req, nil, false, nil, nil
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
