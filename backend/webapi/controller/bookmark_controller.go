package controller

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	dtoutils "github.com/sunibgram/webapi/dto/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	threaddto "github.com/sunibgram/webapi/dto/thread"
	webapiutils "github.com/sunibgram/webapi/utils"
)

type BookmarkController struct {
	bookmarkService service.IBookmarkService
	likeService     service.ILikeService
	threadService   service.IThreadService
	mediaService    service.IThreadMediaService
	pollService     service.IThreadPollService
	profileService  service.IUserProfileService
}

func NewBookmarkController(
	bookmarkService service.IBookmarkService,
	likeService service.ILikeService,
	threadService service.IThreadService,
	mediaService service.IThreadMediaService,
	pollService service.IThreadPollService,
	profileService service.IUserProfileService,
) *BookmarkController {
	return &BookmarkController{
		bookmarkService: bookmarkService,
		likeService:     likeService,
		threadService:   threadService,
		mediaService:    mediaService,
		pollService:     pollService,
		profileService:  profileService,
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
			return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusConflict).JSON(authdto.ErrorResponse{Message: err.Error()})
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
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
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
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
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

	data := make([]threaddto.FeedItemResponse, len(threads))
	fb := webapiutils.FeedBuilder{
		Media:       c.mediaService,
		LikeSvc:     c.likeService,
		BookmarkSvc: c.bookmarkService,
		ThreadSvc:   c.threadService,
		PollSvc:     c.pollService,
		ProfileSvc:  c.profileService,
		Cache:       feedBuilderCache,
		ViewerID:    &claims.UserID,
	}
	fb.LoadAvatarsForUserIDs(threadUserIDs(threads))
	fb.LoadViewerFlagsForThreads(threads)
	fb.LoadPollsForThreads(threads)
	for i, t := range threads {
		item := fb.ToFeedItem(&t, mediaByThread[t.ThreadID])
		wrapped := dtoutils.Wrap(timestamps[i])
		item.ActionAt = &wrapped
		data[i] = item
	}

	return ctx.Status(fiber.StatusOK).JSON(threaddto.PaginatedFeedResponse{
		Data: data,
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}
