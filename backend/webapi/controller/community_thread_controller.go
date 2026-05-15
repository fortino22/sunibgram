package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
)

type CommunityThreadController struct {
	threadService    service.IThreadService
	mediaService     service.IThreadMediaService
	likeService      service.ILikeService
	bookmarkService  service.IBookmarkService
	threadPinService service.IThreadPinService
	profileService   service.IUserProfileService
}

func NewCommunityThreadController(
	threadService service.IThreadService,
	mediaService service.IThreadMediaService,
	likeService service.ILikeService,
	bookmarkService service.IBookmarkService,
	threadPinService service.IThreadPinService,
	profileService service.IUserProfileService,
) *CommunityThreadController {
	return &CommunityThreadController{
		threadService:    threadService,
		mediaService:     mediaService,
		likeService:      likeService,
		bookmarkService:  bookmarkService,
		threadPinService: threadPinService,
		profileService:   profileService,
	}
}

func (c *CommunityThreadController) Create(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	communityID, ok := ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	content := ctx.FormValue("content")
	eventID, _, err := parseOptionalEventID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	files := threadMediaFiles(ctx)
	if err := model.ValidateThreadMediaCount(len(files)); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	thread, err := c.threadService.CreateCommunityThread(ctx.Context(), communityID, claims.UserID, content, eventID)
	if err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	if len(files) > 0 {
		if _, err := c.mediaService.Upload(ctx.Context(), thread.ThreadID, claims.UserID, files); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
		}
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "thread created successfully"})
}

func (c *CommunityThreadController) GetCommunityThreadTop(ctx *fiber.Ctx) error {
	return c.getCommunityThreads(ctx, c.threadService.GetCommunityThreadTop)
}

func (c *CommunityThreadController) GetCommunityThreadLatest(ctx *fiber.Ctx) error {
	return c.getCommunityThreads(ctx, c.threadService.GetCommunityThreadLatest)
}

func (c *CommunityThreadController) getCommunityThreads(
	ctx *fiber.Ctx,
	load func(communityID, userID uint, page, limit int) ([]model.Thread, int64, error),
) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	communityID, ok := ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	page, limit := ParsePagination(ctx)

	threads, total, err := load(communityID, claims.UserID, page, limit)
	if err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	tc := &ThreadController{
		threadService:    c.threadService,
		mediaService:     c.mediaService,
		likeService:      c.likeService,
		bookmarkService:  c.bookmarkService,
		threadPinService: c.threadPinService,
		profileService:   c.profileService,
	}
	data := tc.buildFeedItemsWithPin(threads, &claims.UserID)
	response := make([]dto.CommunityThreadFeedItemResponse, len(data))
	for i, item := range data {
		response[i] = dto.CommunityThreadFeedItemResponse{
			FeedItemResponse: item,
			UserAvatar:       item.Author.AvatarURL,
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedCommunityThreadResponse{
		Data: response,
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *CommunityThreadController) GetCommunityThreadMedia(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	communityID, ok := ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	threads, err := c.threadService.GetCommunityThreadMedia(communityID, claims.UserID)
	if err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	data := make([]dto.CommunityThreadMediaResponse, len(threads))
	for i, thread := range threads {
		media := make([]dto.ThreadMediaResponse, len(thread.Media))
		for j, item := range thread.Media {
			media[j] = dto.ThreadMediaResponse{
				MediaID: item.MediaID,
				URL:     c.mediaService.GetURL(item.Bucket, item.ObjectName),
			}
		}
		data[i] = dto.CommunityThreadMediaResponse{
			ThreadID: thread.ThreadID,
			Media:    media,
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": data})
}
