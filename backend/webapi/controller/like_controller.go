package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	dtoutils "github.com/sunibgram/dto/utils"
	webapiutils "github.com/sunibgram/webapi/utils"
)

type LikeController struct {
	likeService     service.ILikeService
	bookmarkService service.IBookmarkService
	threadService   service.IThreadService
	mediaService    service.IThreadMediaService
	pollService     service.IThreadPollService
}

func NewLikeController(
	likeService service.ILikeService,
	bookmarkService service.IBookmarkService,
	threadService service.IThreadService,
	mediaService service.IThreadMediaService,
	pollService service.IThreadPollService,
) *LikeController {
	return &LikeController{
		likeService:     likeService,
		bookmarkService: bookmarkService,
		threadService:   threadService,
		mediaService:    mediaService,
		pollService:     pollService,
	}
}

func (c *LikeController) Like(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	threadID, ok := ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}

	if err := c.likeService.Like(claims.UserID, threadID); err != nil {
		if err.Error() == "thread not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "liked successfully"})
}

func (c *LikeController) Unlike(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	threadID, ok := ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}

	if err := c.likeService.Unlike(claims.UserID, threadID); err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "unliked successfully"})
}

func (c *LikeController) GetMyLikes(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	page, limit := ParsePagination(ctx)

	threads, timestamps, total, err := c.likeService.GetMyLikes(claims.UserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: "internal server error"})
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
		PollSvc:     c.pollService,
		Cache:       feedBuilderCache,
		ViewerID:    &claims.UserID,
	}
	fb.LoadViewerFlagsForThreads(threads)
	fb.LoadPollsForThreads(threads)
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
