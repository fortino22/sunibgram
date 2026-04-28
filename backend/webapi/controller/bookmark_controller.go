package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/webapi/ctxutil"
)

type BookmarkController struct {
	bookmarkService service.IBookmarkService
	mediaService    service.IThreadMediaService
}

func NewBookmarkController(bookmarkService service.IBookmarkService, mediaService service.IThreadMediaService) *BookmarkController {
	return &BookmarkController{bookmarkService: bookmarkService, mediaService: mediaService}
}

func (c *BookmarkController) Bookmark(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	threadID, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid thread id"})
	}

	if err := c.bookmarkService.Bookmark(claims.UserID, uint(threadID)); err != nil {
		if err.Error() == "thread not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "bookmarked successfully"})
}

func (c *BookmarkController) Unbookmark(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	threadID, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid thread id"})
	}

	if err := c.bookmarkService.Unbookmark(claims.UserID, uint(threadID)); err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "unbookmarked successfully"})
}

func (c *BookmarkController) GetMyBookmarks(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))

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

	data := make([]dto.BookmarkedThreadResponse, len(threads))
	for i, t := range threads {
		data[i] = dto.BookmarkedThreadResponse{
			ThreadResponse: toThreadResponse(&t, mediaByThread[t.ThreadID]),
			BookmarkedAt:   timestamps[i],
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedBookmarkedThreadsResponse{
		Data: data, Total: total, Page: page, Limit: limit,
	})
}
