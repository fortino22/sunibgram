package likectrl

import (
	"github.com/gofiber/fiber/v2"

	threadmediamodel "github.com/sunibgram/domain/model/thread_media"
	bookmarksvciface "github.com/sunibgram/domain/service/bookmark"
	likesvciface "github.com/sunibgram/domain/service/like"
	threadsvciface "github.com/sunibgram/domain/service/thread"
	threadmediasvciface "github.com/sunibgram/domain/service/thread_media"
	threadpollsvciface "github.com/sunibgram/domain/service/thread_poll"
	userprofilesvciface "github.com/sunibgram/domain/service/user_profile"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	threadctrl "github.com/sunibgram/webapi/controller/thread"
	dtoutils "github.com/sunibgram/webapi/dto/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	threaddto "github.com/sunibgram/webapi/dto/thread"
	webapiutils "github.com/sunibgram/webapi/utils"
)

type LikeController struct {
	likeService     likesvciface.ILikeService
	bookmarkService bookmarksvciface.IBookmarkService
	threadService   threadsvciface.IThreadService
	mediaService    threadmediasvciface.IThreadMediaService
	pollService     threadpollsvciface.IThreadPollService
	profileService  userprofilesvciface.IUserProfileService
}

func NewLikeController(
	likeService likesvciface.ILikeService,
	bookmarkService bookmarksvciface.IBookmarkService,
	threadService threadsvciface.IThreadService,
	mediaService threadmediasvciface.IThreadMediaService,
	pollService threadpollsvciface.IThreadPollService,
	profileService userprofilesvciface.IUserProfileService,
) *LikeController {
	return &LikeController{
		likeService:     likeService,
		bookmarkService: bookmarkService,
		threadService:   threadService,
		mediaService:    mediaService,
		pollService:     pollService,
		profileService:  profileService,
	}
}

func (c *LikeController) Like(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	threadID, ok := ctrlshared.ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}

	if err := c.likeService.Like(claims.UserID, threadID); err != nil {
		if err.Error() == "thread not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		if err.Error() == "blocked" {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusConflict).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "liked successfully"})
}

func (c *LikeController) Unlike(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	threadID, ok := ctrlshared.ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}

	if err := c.likeService.Unlike(claims.UserID, threadID); err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "unliked successfully"})
}

func (c *LikeController) GetMyLikes(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	page, limit := ctrlshared.ParsePagination(ctx)

	threads, timestamps, total, err := c.likeService.GetMyLikes(claims.UserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}

	ids := make([]uint, len(threads))
	for i, t := range threads {
		ids[i] = t.ThreadID
	}
	allMedia, _ := c.mediaService.GetByThreadIDs(ids)
	mediaByThread := make(map[uint][]threadmediamodel.ThreadMedia, len(ids))
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
		Cache:       ctrlshared.FeedBuilderCache,
		ViewerID:    &claims.UserID,
	}
	fb.LoadAvatarsForUserIDs(threadctrl.ThreadUserIDs(threads))
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
