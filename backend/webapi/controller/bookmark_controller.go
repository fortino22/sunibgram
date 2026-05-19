package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	dtoutils "github.com/sunibgram/dto/utils"
	webapiutils "github.com/sunibgram/webapi/utils"
)

type BookmarkController struct {
	bookmarkService service.IBookmarkService
	likeService     service.ILikeService
	threadService   service.IThreadService
	mediaService    service.IThreadMediaService
}

func NewBookmarkController(
	bookmarkService service.IBookmarkService,
	likeService service.ILikeService,
	threadService service.IThreadService,
	mediaService service.IThreadMediaService,
) *BookmarkController {
	return &BookmarkController{
		bookmarkService: bookmarkService,
		likeService:     likeService,
		threadService:   threadService,
		mediaService:    mediaService,
	}
}

func (c *BookmarkController) Bookmark(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	threadID, ok := ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}

	if err := c.bookmarkService.Bookmark(claims.UserID, threadID); err != nil {
		if err.Error() == "thread not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "bookmarked successfully"})
}

func (c *BookmarkController) Unbookmark(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	threadID, ok := ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}

	if err := c.bookmarkService.Unbookmark(claims.UserID, threadID); err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "unbookmarked successfully"})
}

func (c *BookmarkController) GetMyBookmarks(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	page, limit := ParsePagination(ctx)

	threads, timestamps, total, err := c.bookmarkService.GetMyBookmarks(claims.UserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	ids := make([]uint, len(threads))
	for i, t := range threads {
		ids[i] = t.ThreadID
	}
	allMedia, _ := c.mediaService.GetByThreadIDs(ids)
	mediaByThread := make(map[uint][]model.ThreadMedia, len(ids))
	for _, m := range allMedia {
		mediaByThread[m.ThreadID] = append(mediaByThread[m.ThreadID], m)
	}

	data := make([]dto.FeedItemResponse, len(threads))
	fb := webapiutils.FeedBuilder{
		Media:       c.mediaService,
		LikeSvc:     c.likeService,
		BookmarkSvc: c.bookmarkService,
		ThreadSvc:   c.threadService,
		ViewerID:    &claims.UserID,
	}
	fb.LoadViewerFlagsForThreads(threads)
	for i, t := range threads {
		item := fb.ToFeedItem(&t, mediaByThread[t.ThreadID])
		wrapped := dtoutils.Wrap(timestamps[i])
		item.ActionAt = &wrapped
		data[i] = item
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFeedResponse{
		Data: data,
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}
