package eventctrl

import (
	"context"
	"errors"
	"log"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	eventreviewmodel "github.com/sunibgram/domain/model/event_review"
	mediamodel "github.com/sunibgram/domain/model/media"
	threadmodel "github.com/sunibgram/domain/model/thread"
	bookmarksvciface "github.com/sunibgram/domain/service/bookmark"
	eventsvciface "github.com/sunibgram/domain/service/event"
	eventthreadpinsvciface "github.com/sunibgram/domain/service/event_thread_pin"
	likesvciface "github.com/sunibgram/domain/service/like"
	threadsvciface "github.com/sunibgram/domain/service/thread"
	threadmediasvciface "github.com/sunibgram/domain/service/thread_media"
	threadpinsvciface "github.com/sunibgram/domain/service/thread_pin"
	threadpollsvciface "github.com/sunibgram/domain/service/thread_poll"
	userprofilesvciface "github.com/sunibgram/domain/service/user_profile"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	threadctrl "github.com/sunibgram/webapi/controller/thread"
	utils "github.com/sunibgram/webapi/dto/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	eventdto "github.com/sunibgram/webapi/dto/event"
	eventreviewdto "github.com/sunibgram/webapi/dto/event_review"
	threaddto "github.com/sunibgram/webapi/dto/thread"
)

type EventController struct {
	eventService      eventsvciface.IEventService
	threadService     threadsvciface.IThreadService
	mediaService      threadmediasvciface.IThreadMediaService
	likeService       likesvciface.ILikeService
	bookmarkService   bookmarksvciface.IBookmarkService
	threadPinService  threadpinsvciface.IThreadPinService
	profileService    userprofilesvciface.IUserProfileService
	pollService       threadpollsvciface.IThreadPollService
	eventThreadPinSvc eventthreadpinsvciface.IEventThreadPinService
}

func NewEventController(eventService eventsvciface.IEventService, threadService threadsvciface.IThreadService, mediaService threadmediasvciface.IThreadMediaService, likeService likesvciface.ILikeService, bookmarkService bookmarksvciface.IBookmarkService, threadPinService threadpinsvciface.IThreadPinService, profileService userprofilesvciface.IUserProfileService, pollService threadpollsvciface.IThreadPollService, eventThreadPinSvc eventthreadpinsvciface.IEventThreadPinService) *EventController {
	return &EventController{eventService: eventService, threadService: threadService, mediaService: mediaService, likeService: likeService, bookmarkService: bookmarkService, threadPinService: threadPinService, profileService: profileService, pollService: pollService, eventThreadPinSvc: eventThreadPinSvc}
}

func (c *EventController) Create(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	req, banner, attachments, err := parseCreateEventRequest(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	_, err = c.eventService.Create(ctx.Context(), claims.UserID, req.Title, req.Description, req.Location, req.Type, req.StartTime.Time, req.EndTime.Time, banner, attachments)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "success"})
}

func (c *EventController) GetAll(ctx *fiber.Ctx) error {
	if _, ok := ctrlshared.RequireAuth(ctx); !ok {
		return nil
	}
	_, limit := ctrlshared.ParsePagination(ctx)
	trending, running, future, err := c.eventService.GetAll(limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}
	return ctx.Status(fiber.StatusOK).JSON(eventdto.EventGroupsResponse{
		Trending: c.toListResponses(trending),
		Running:  c.toListResponses(running),
		Future:   c.toListResponses(future),
	})
}

func (c *EventController) GetPending(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	if claims.RoleName != "admin" {
		return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
	}
	page, limit := ctrlshared.ParsePagination(ctx)
	events, reviews, total, err := c.eventService.GetAdminEvents(ctx.Query("status"), ctx.Query("name"), page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(eventdto.PaginatedEventWithReviewHistoryResponse{
		Data: c.toAdminEventReviewHistoryResponses(ctx.Context(), events, reviews),
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *EventController) GetMyEvents(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	page, limit := ctrlshared.ParsePagination(ctx)
	events, reviews, total, err := c.eventService.GetMyEvents(claims.UserID, ctx.Query("status"), page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(eventdto.PaginatedMyEventResponse{
		Data: c.toEventReviewHistoryResponses(ctx.Context(), events, reviews),
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *EventController) GetByID(ctx *fiber.Ctx) error {
	if _, ok := ctrlshared.RequireAuth(ctx); !ok {
		return nil
	}
	event, err := c.eventService.GetByID(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: "event not found"})
	}
	return ctx.Status(fiber.StatusOK).JSON(c.toResponse(ctx.Context(), event))
}

func (c *EventController) Update(ctx *fiber.Ctx) error {
	if _, ok := ctrlshared.RequireAuth(ctx); !ok {
		return nil
	}
	req, banner, clearBanner, attachments, err := parseUpdateEventRequest(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}
	_, err = c.eventService.Update(ctx.Context(), ctx.Params("id"), req.Title, req.Description, req.Location, req.Type, req.StartTime.Time, req.EndTime.Time, banner, clearBanner, attachments)
	if err != nil {
		if err.Error() == "event not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "success"})
}

func (c *EventController) Delete(ctx *fiber.Ctx) error {
	if _, ok := ctrlshared.RequireAuth(ctx); !ok {
		return nil
	}
	if err := c.eventService.Delete(ctx.Context(), ctx.Params("id")); err != nil {
		if err.Error() == "event not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "event deleted successfully"})
}

func (c *EventController) ApproveEvent(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	if claims.RoleName != "admin" {
		return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
	}
	adminID := strconv.FormatUint(uint64(claims.UserID), 10)
	err := c.eventService.ApproveEvent(ctx.Params("id"), adminID)
	if err != nil {
		if err.Error() == "event not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Event approved successfully"})
}

func (c *EventController) DeclineEvent(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	if claims.RoleName != "admin" {
		return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
	}
	var req eventreviewdto.DeclineEventRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}
	adminID := strconv.FormatUint(uint64(claims.UserID), 10)
	err := c.eventService.DeclineEvent(ctx.Params("id"), adminID, req.Reason)
	if err != nil {
		if err.Error() == "event not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Event declined successfully"})
}

func (c *EventController) RequestRevision(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	if claims.RoleName != "admin" {
		return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
	}
	var req eventreviewdto.RevisionRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}
	adminID := strconv.FormatUint(uint64(claims.UserID), 10)
	if err := c.eventService.RequestRevision(ctx.Params("id"), adminID, req.Note); err != nil {
		if err.Error() == "event not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Event revision requested successfully"})
}

func (c *EventController) GetThreads(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	page, limit := ctrlshared.ParsePagination(ctx)
	filter := ctx.Query("filter")
	sort := ctx.Query("sort", "latest")
	eventID := ctx.Params("id")
	threads, parents, total, err := c.eventService.GetThreads(ctx.Context(), &claims.UserID, eventID, filter, sort, page, limit)
	if err != nil {
		if err.Error() == "event not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	var pinnedThreadID uint
	if filter == "official" && page == 1 && c.eventThreadPinSvc != nil {
		if pin, _ := c.eventThreadPinSvc.GetPinned(eventID); pin != nil {
			pinnedThreadID = pin.ThreadID
			for i, t := range threads {
				if t.ThreadID == pin.ThreadID {
					threads = append(threads[:i], threads[i+1:]...)
					parents = append(parents[:i], parents[i+1:]...)
					break
				}
			}
			if pinned, ferr := c.threadService.GetByID(pin.ThreadID); ferr == nil && pinned != nil {
				threads = append([]threadmodel.Thread{*pinned}, threads...)
				parents = append([]*threadmodel.Thread{nil}, parents...)
			} else {
				pinnedThreadID = 0
			}
		}
	}

	tc := threadctrl.NewThreadController(c.threadService, c.mediaService, nil, c.likeService, c.bookmarkService, c.threadPinService, c.profileService, c.pollService)
	data := tc.BuildUserFeedItemsNoPin(threads, parents, &claims.UserID)
	if pinnedThreadID != 0 {
		for i := range data {
			if data[i].ID == pinnedThreadID {
				t := true
				data[i].PinnedByEventAdmin = &t
				break
			}
		}
	}
	return ctx.Status(fiber.StatusOK).JSON(threaddto.PaginatedFeedResponse{
		Data: data,
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *EventController) toResponse(ctx context.Context, event *threadmodel.Event) eventdto.EventResponse {
	bannerURL := ctrlshared.BuildMediaURL(event.BannerObjectName, c.eventService.GetURL)
	var attachmentURL string
	if event.AttachmentObjectName != "" {
		var err error
		attachmentURL, err = c.eventService.GeneratePresignedAttachmentURL(ctx, event.AttachmentObjectName)
		if err != nil {
			log.Printf("presign error object=%s: %v", event.AttachmentObjectName, err)
		}
	}
	pdfURLs := c.eventPDFURLs(ctx, event)
	return eventdto.EventResponse{
		ID:            event.ID,
		UserID:        event.UserID,
		Title:         event.Title,
		Description:   event.Description,
		Location:      event.Location,
		Type:          event.Type,
		Status:        event.Status,
		StartTime:     eventdto.EventTime{Time: event.StartTime},
		EndTime:       eventdto.EventTime{Time: event.EndTime},
		BannerURL:     bannerURL,
		AttachmentURL: attachmentURL,
		PDFURLs:       pdfURLs,
		CreatedAt:     utils.Wrap(event.CreatedAt),
		UpdatedAt:     utils.Wrap(event.UpdatedAt),
	}
}

func (c *EventController) toResponses(ctx context.Context, events []threadmodel.Event) []eventdto.EventResponse {
	out := make([]eventdto.EventResponse, len(events))
	for i := range events {
		out[i] = c.toResponse(ctx, &events[i])
	}
	return out
}

func (c *EventController) toListResponse(event *threadmodel.Event) eventdto.EventListResponse {
	return eventdto.EventListResponse{
		ID:          event.ID,
		UserID:      event.UserID,
		Title:       event.Title,
		Description: event.Description,
		Location:    event.Location,
		Type:        event.Type,
		StartTime:   eventdto.EventTime{Time: event.StartTime},
		EndTime:     eventdto.EventTime{Time: event.EndTime},
		BannerURL:   ctrlshared.BuildMediaURL(event.BannerObjectName, c.eventService.GetURL),
		CreatedAt:   utils.Wrap(event.CreatedAt),
		UpdatedAt:   utils.Wrap(event.UpdatedAt),
	}
}

func (c *EventController) toListResponses(events []threadmodel.Event) []eventdto.EventListResponse {
	out := make([]eventdto.EventListResponse, len(events))
	for i := range events {
		out[i] = c.toListResponse(&events[i])
	}
	return out
}

func (c *EventController) toEventReviewHistoryResponses(ctx context.Context, events []threadmodel.Event, reviews map[string][]eventreviewmodel.EventReview) []eventdto.EventWithReviewHistoryResponse {
	out := make([]eventdto.EventWithReviewHistoryResponse, len(events))
	for i := range events {
		out[i] = eventdto.EventWithReviewHistoryResponse{
			EventResponse: c.toResponse(ctx, &events[i]),
			Notes:         c.toEventReviewNoteResponses(reviews[events[i].ID]),
		}
	}
	return out
}

func (c *EventController) toAdminEventReviewHistoryResponses(ctx context.Context, events []threadmodel.Event, reviews map[string][]eventreviewmodel.EventReview) []eventdto.AdminEventWithReviewHistoryResponse {
	out := make([]eventdto.AdminEventWithReviewHistoryResponse, len(events))
	for i := range events {
		out[i] = eventdto.AdminEventWithReviewHistoryResponse{
			EventResponse: c.toResponse(ctx, &events[i]),
			Notes:         c.toEventAdminReviewNoteResponses(reviews[events[i].ID]),
		}
	}
	return out
}

func (c *EventController) toEventReviewNoteResponses(reviews []eventreviewmodel.EventReview) []eventreviewdto.EventReviewNoteResponse {
	out := make([]eventreviewdto.EventReviewNoteResponse, len(reviews))
	for i := range reviews {
		out[i] = eventreviewdto.EventReviewNoteResponse{
			RevisionNumber: i + 1,
			Notes:          reviews[i].Note,
			CreatedAt:      utils.Wrap(reviews[i].CreatedAt),
		}
	}
	return out
}

func (c *EventController) toEventAdminReviewNoteResponses(reviews []eventreviewmodel.EventReview) []eventreviewdto.EventAdminReviewNoteResponse {
	out := make([]eventreviewdto.EventAdminReviewNoteResponse, len(reviews))
	for i := range reviews {
		out[i] = eventreviewdto.EventAdminReviewNoteResponse{
			RevisionNumber: i + 1,
			Notes:          reviews[i].Note,
			AdminName:      reviews[i].Admin.Name,
			CreatedAt:      utils.Wrap(reviews[i].CreatedAt),
		}
	}
	return out
}

func (c *EventController) eventPDFURLs(ctx context.Context, event *threadmodel.Event) []string {
	pdfs := event.PDFs
	if len(pdfs) == 0 && event.AttachmentObjectName != "" {
		pdfs = threadmodel.EventPDFs{{ObjectName: event.AttachmentObjectName}}
	}
	if len(pdfs) == 0 {
		return nil
	}
	urls := make([]string, 0, len(pdfs))
	for _, pdf := range pdfs {
		if pdf.ObjectName == "" {
			continue
		}
		url, err := c.eventService.GeneratePresignedAttachmentURL(ctx, pdf.ObjectName)
		if err != nil {
			log.Printf("presign error object=%s: %v", pdf.ObjectName, err)
			continue
		}
		urls = append(urls, url)
	}
	return urls
}

func parseCreateEventRequest(ctx *fiber.Ctx) (eventdto.CreateEventRequest, *multipart.FileHeader, []*multipart.FileHeader, error) {
	var req eventdto.CreateEventRequest
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
	if err := threadmodel.ValidateEventPDFCount(len(files)); err != nil {
		return err
	}
	for _, file := range files {
		if err := mediamodel.ValidateUploadSize(file, mediamodel.MaxDocumentUploadSize); err != nil {
			return err
		}
		if err := threadmodel.ValidateEventAttachmentExt(file.Filename); err != nil {
			return err
		}
	}
	return nil
}

func parseUpdateEventRequest(ctx *fiber.Ctx) (eventdto.UpdateEventRequest, *multipart.FileHeader, bool, []*multipart.FileHeader, error) {
	var req eventdto.UpdateEventRequest
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

func fillEventRequestFromForm(ctx *fiber.Ctx, title, description, location, eventType *string, startTime, endTime *eventdto.EventTime) error {
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

func parseEventFormTime(value string, target *eventdto.EventTime) error {
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
