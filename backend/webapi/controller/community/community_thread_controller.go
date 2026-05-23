package communityctrl

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	threadmodel "github.com/sunibgram/domain/model/thread"
	threadmediamodel "github.com/sunibgram/domain/model/thread_media"
	bookmarksvciface "github.com/sunibgram/domain/service/bookmark"
	communitythreadpinsvciface "github.com/sunibgram/domain/service/community_thread_pin"
	likesvciface "github.com/sunibgram/domain/service/like"
	threadsvciface "github.com/sunibgram/domain/service/thread"
	threadmediasvciface "github.com/sunibgram/domain/service/thread_media"
	threadpinsvciface "github.com/sunibgram/domain/service/thread_pin"
	threadpollsvciface "github.com/sunibgram/domain/service/thread_poll"
	userprofilesvciface "github.com/sunibgram/domain/service/user_profile"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	threadctrl "github.com/sunibgram/webapi/controller/thread"
	threadpollctrl "github.com/sunibgram/webapi/controller/thread_poll"
	authdto "github.com/sunibgram/webapi/dto/auth"
	threaddto "github.com/sunibgram/webapi/dto/thread"
	threadmediadto "github.com/sunibgram/webapi/dto/thread_media"
)

type CommunityThreadController struct {
	threadService    threadsvciface.IThreadService
	mediaService     threadmediasvciface.IThreadMediaService
	likeService      likesvciface.ILikeService
	bookmarkService  bookmarksvciface.IBookmarkService
	threadPinService threadpinsvciface.IThreadPinService
	profileService   userprofilesvciface.IUserProfileService
	pollService      threadpollsvciface.IThreadPollService
	communityPinSvc  communitythreadpinsvciface.ICommunityThreadPinService
}

func NewCommunityThreadController(
	threadService threadsvciface.IThreadService,
	mediaService threadmediasvciface.IThreadMediaService,
	likeService likesvciface.ILikeService,
	bookmarkService bookmarksvciface.IBookmarkService,
	threadPinService threadpinsvciface.IThreadPinService,
	profileService userprofilesvciface.IUserProfileService,
	pollService threadpollsvciface.IThreadPollService,
	communityPinSvc communitythreadpinsvciface.ICommunityThreadPinService,
) *CommunityThreadController {
	return &CommunityThreadController{
		threadService:    threadService,
		mediaService:     mediaService,
		likeService:      likeService,
		bookmarkService:  bookmarkService,
		threadPinService: threadPinService,
		profileService:   profileService,
		pollService:      pollService,
		communityPinSvc:  communityPinSvc,
	}
}

func (c *CommunityThreadController) GetCommunityThreadTop(ctx *fiber.Ctx) error {
	return c.getCommunityThreads(ctx, c.threadService.GetCommunityThreadTop)
}

func (c *CommunityThreadController) GetCommunityThreadLatest(ctx *fiber.Ctx) error {
	return c.getCommunityThreads(ctx, c.threadService.GetCommunityThreadLatest)
}

func (c *CommunityThreadController) GetCommunityThreadsBySort(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	communityID, ok := ctrlshared.ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	sort := ctx.Query("sort", "latest")
	switch sort {
	case "top", "latest", "likes":
	default:
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid sort; allowed: top, latest, likes"})
	}

	page, limit := ctrlshared.ParsePagination(ctx)

	threads, parents, total, err := c.threadService.GetCommunityThreadsBySort(communityID, claims.UserID, sort, page, limit)
	if err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	var pinnedThreadID uint
	if page == 1 {
		if pin, _ := c.communityPinSvc.GetPinned(communityID); pin != nil {
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

	tc := c.threadCtrl()
	data := tc.BuildUserFeedItemsWithPin(threads, parents, &claims.UserID)
	response := make([]threaddto.CommunityThreadFeedItemResponse, len(data))
	for i, item := range data {
		if pinnedThreadID != 0 && item.ID == pinnedThreadID {
			t := true
			item.PinnedByAdmin = &t
		}
		response[i] = threaddto.CommunityThreadFeedItemResponse{FeedItemResponse: item}
	}

	return ctx.Status(fiber.StatusOK).JSON(threaddto.PaginatedCommunityThreadResponse{
		Data: response,
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *CommunityThreadController) GetCommunityThreadTopByName(ctx *fiber.Ctx) error {
	return c.getCommunityThreadsByQuery(ctx, c.threadService.GetAllCommunityThreadsByQueryTop)
}

func (c *CommunityThreadController) GetCommunityThreadLatestByName(ctx *fiber.Ctx) error {
	return c.getCommunityThreadsByQuery(ctx, c.threadService.GetAllCommunityThreadsByQueryLatest)
}

func (c *CommunityThreadController) CreateCommunityThread(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	communityID, ok := ctrlshared.ParseIDParam(ctx, "communityID", "community")
	if !ok {
		return nil
	}

	files := threadctrl.ThreadMediaFiles(ctx)
	if err := threadmediamodel.ValidateThreadMediaCount(len(files)); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	content := ctx.FormValue("content")
	eventID, eventIDPresent, err := threadctrl.ParseOptionalEventID(ctx)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	pollInput, _, hasPoll, err := threadpollctrl.ParseCreatePollFromMultipart(ctx)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	if content == "" && !eventIDPresent && len(files) == 0 && !hasPoll {
		var req threaddto.CreateThreadRequest
		if err := ctx.BodyParser(&req); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
		}
		content = req.Content
		eventID = req.EventID
	}

	thread, err := c.threadService.CreateCommunityThread(ctx.Context(), communityID, claims.UserID, content, eventID)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	if len(files) > 0 {
		if _, err := c.mediaService.Upload(ctx.Context(), thread.ThreadID, claims.UserID, files); err != nil {
			_ = c.threadService.Delete(ctx.Context(), thread.ThreadID, claims.UserID)
			return ctrlshared.HandleServiceError(ctx, err)
		}
	}

	if hasPoll {
		if _, err := c.pollService.Create(ctx.Context(), thread.ThreadID, claims.UserID, pollInput); err != nil {
			_ = c.mediaService.DeleteByThread(ctx.Context(), thread.ThreadID)
			_ = c.threadService.Delete(ctx.Context(), thread.ThreadID, claims.UserID)
			return ctrlshared.HandleServiceError(ctx, err)
		}
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "community thread created successfully",
	})
}

func (c *CommunityThreadController) getCommunityThreads(
	ctx *fiber.Ctx,
	load func(communityID, userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error),
) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	communityID, ok := ctrlshared.ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	page, limit := ctrlshared.ParsePagination(ctx)

	threads, parents, total, err := load(communityID, claims.UserID, page, limit)
	if err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	var pinnedThreadID uint
	if page == 1 {
		if pin, _ := c.communityPinSvc.GetPinned(communityID); pin != nil {
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

	tc := c.threadCtrl()
	data := tc.BuildUserFeedItemsWithPin(threads, parents, &claims.UserID)
	response := make([]threaddto.CommunityThreadFeedItemResponse, len(data))
	for i, item := range data {
		if pinnedThreadID != 0 && item.ID == pinnedThreadID {
			t := true
			item.PinnedByAdmin = &t
		}
		response[i] = threaddto.CommunityThreadFeedItemResponse{
			FeedItemResponse: item,
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(threaddto.PaginatedCommunityThreadResponse{
		Data: response,
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *CommunityThreadController) threadCtrl() *threadctrl.ThreadController {
	return threadctrl.NewThreadController(
		c.threadService,
		c.mediaService,
		nil,
		c.likeService,
		c.bookmarkService,
		c.threadPinService,
		c.profileService,
		c.pollService,
	)
}

func (c *CommunityThreadController) getCommunityThreadsByQuery(
	ctx *fiber.Ctx,
	load func(context.Context, uint, string, int, int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error),
) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	query := ctx.Query("query")
	page, limit := ctrlshared.ParsePagination(ctx)

	threads, parents, total, err := load(ctx.Context(), claims.UserID, query, page, limit)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	tc := c.threadCtrl()
	data := tc.BuildUserFeedItemsWithPin(threads, parents, &claims.UserID)
	response := make([]threaddto.CommunityThreadFeedItemResponse, len(data))
	for i, item := range data {
		response[i] = threaddto.CommunityThreadFeedItemResponse{
			FeedItemResponse: item,
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(threaddto.PaginatedCommunityThreadResponse{
		Data: response,
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *CommunityThreadController) GetMyCommunityThreads(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	page, limit := ctrlshared.ParsePagination(ctx)

	threads, parents, total, err := c.threadService.GetMyCommunityThreads(ctx.Context(), claims.UserID, page, limit)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	tc := c.threadCtrl()
	data := tc.BuildUserFeedItemsWithPin(threads, parents, &claims.UserID)
	response := make([]threaddto.CommunityThreadFeedItemResponse, len(data))
	for i, item := range data {
		response[i] = threaddto.CommunityThreadFeedItemResponse{FeedItemResponse: item}
	}

	return ctx.Status(fiber.StatusOK).JSON(threaddto.PaginatedCommunityThreadResponse{
		Data: response,
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *CommunityThreadController) GetAllCommunityThreads(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	var categoryID *uuid.UUID
	if raw := ctx.Query("category"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid category id"})
		}
		categoryID = &id
	}

	page, limit := ctrlshared.ParsePagination(ctx)

	threads, parents, total, err := c.threadService.GetAllCommunityThreads(ctx.Context(), claims.UserID, categoryID, page, limit)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	tc := c.threadCtrl()
	data := tc.BuildUserFeedItemsWithPin(threads, parents, &claims.UserID)
	response := make([]threaddto.CommunityThreadFeedItemResponse, len(data))
	for i, item := range data {
		response[i] = threaddto.CommunityThreadFeedItemResponse{
			FeedItemResponse: item,
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(threaddto.PaginatedCommunityThreadResponse{
		Data: response,
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *CommunityThreadController) GetCommunityThreadMedia(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	communityID, ok := ctrlshared.ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	threads, err := c.threadService.GetCommunityThreadMedia(communityID, claims.UserID)
	if err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	data := make([]threaddto.CommunityThreadMediaResponse, len(threads))
	for i, thread := range threads {
		media := make([]threadmediadto.ThreadMediaResponse, len(thread.Media))
		for j, item := range thread.Media {
			media[j] = threadmediadto.ThreadMediaResponse{
				MediaID: item.MediaID,
				URL:     c.mediaService.GetURL(item.ObjectName),
			}
		}
		data[i] = threaddto.CommunityThreadMediaResponse{
			ThreadID: thread.ThreadID,
			Media:    media,
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": data})
}
