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
}

func NewCommunityThreadController(threadService service.IThreadService, mediaService service.IThreadMediaService, likeService service.ILikeService, bookmarkService service.IBookmarkService, threadPinService service.IThreadPinService) *CommunityThreadController {
	return &CommunityThreadController{threadService: threadService, mediaService: mediaService, likeService: likeService, bookmarkService: bookmarkService, threadPinService: threadPinService}
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

func (c *CommunityThreadController) GetAll(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	communityID, ok := ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	page, limit := ParsePagination(ctx)
	sort := ctx.Query("sort", "latest")
	if sort != "latest" && sort != "top" {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "sort must be 'latest' or 'top'"})
	}

	threads, total, err := c.threadService.GetCommunityThreads(communityID, claims.UserID, page, limit, sort)
	if err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	tc := &ThreadController{threadService: c.threadService, mediaService: c.mediaService, likeService: c.likeService, bookmarkService: c.bookmarkService, threadPinService: c.threadPinService}
	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFeedResponse{
		Data: tc.buildFeedItemsWithPin(threads, &claims.UserID),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}
