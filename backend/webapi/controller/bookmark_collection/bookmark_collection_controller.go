package bookmarkcollectionctrl

import (
	"github.com/gofiber/fiber/v2"

	threadmediamodel "github.com/sunibgram/domain/model/thread_media"
	bookmarksvciface "github.com/sunibgram/domain/service/bookmark"
	bookmarkcollectionsvciface "github.com/sunibgram/domain/service/bookmark_collection"
	likesvciface "github.com/sunibgram/domain/service/like"
	threadsvciface "github.com/sunibgram/domain/service/thread"
	threadmediasvciface "github.com/sunibgram/domain/service/thread_media"
	threadpollsvciface "github.com/sunibgram/domain/service/thread_poll"
	userprofilesvciface "github.com/sunibgram/domain/service/user_profile"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	threadctrl "github.com/sunibgram/webapi/controller/thread"
	dtoutils "github.com/sunibgram/webapi/dto/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	bookmarkcollectiondto "github.com/sunibgram/webapi/dto/bookmark_collection"
	threaddto "github.com/sunibgram/webapi/dto/thread"
	webapiutils "github.com/sunibgram/webapi/utils"
)

type BookmarkCollectionController struct {
	svc             bookmarkcollectionsvciface.IBookmarkCollectionService
	bookmarkService bookmarksvciface.IBookmarkService
	likeService     likesvciface.ILikeService
	threadService   threadsvciface.IThreadService
	mediaService    threadmediasvciface.IThreadMediaService
	pollService     threadpollsvciface.IThreadPollService
	profileService  userprofilesvciface.IUserProfileService
}

func NewBookmarkCollectionController(
	svc bookmarkcollectionsvciface.IBookmarkCollectionService,
	bookmarkService bookmarksvciface.IBookmarkService,
	likeService likesvciface.ILikeService,
	threadService threadsvciface.IThreadService,
	mediaService threadmediasvciface.IThreadMediaService,
	pollService threadpollsvciface.IThreadPollService,
	profileService userprofilesvciface.IUserProfileService,
) *BookmarkCollectionController {
	return &BookmarkCollectionController{
		svc:             svc,
		bookmarkService: bookmarkService,
		likeService:     likeService,
		threadService:   threadService,
		mediaService:    mediaService,
		pollService:     pollService,
		profileService:  profileService,
	}
}

func (c *BookmarkCollectionController) Create(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	var req bookmarkcollectiondto.CreateCollectionRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}
	collection, err := c.svc.Create(claims.UserID, req.Name, req.Description)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(bookmarkcollectiondto.ToCollectionResponse(collection))
}

func (c *BookmarkCollectionController) List(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	collections, err := c.svc.ListMine(claims.UserID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}
	data := make([]bookmarkcollectiondto.CollectionResponse, len(collections))
	for i := range collections {
		data[i] = bookmarkcollectiondto.ToCollectionResponse(&collections[i])
	}
	return ctx.Status(fiber.StatusOK).JSON(bookmarkcollectiondto.CollectionListResponse{Data: data})
}

func (c *BookmarkCollectionController) Get(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	collectionID, ok := ctrlshared.ParseIDParam(ctx, "id", "collection")
	if !ok {
		return nil
	}
	collection, err := c.svc.GetMine(claims.UserID, collectionID)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(bookmarkcollectiondto.ToCollectionResponse(collection))
}

func (c *BookmarkCollectionController) Update(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	collectionID, ok := ctrlshared.ParseIDParam(ctx, "id", "collection")
	if !ok {
		return nil
	}
	var req bookmarkcollectiondto.UpdateCollectionRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}
	if err := c.svc.Update(claims.UserID, collectionID, req.Name, req.Description); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "collection updated successfully"})
}

func (c *BookmarkCollectionController) Delete(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	collectionID, ok := ctrlshared.ParseIDParam(ctx, "id", "collection")
	if !ok {
		return nil
	}
	if err := c.svc.Delete(claims.UserID, collectionID); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "collection deleted successfully"})
}

func (c *BookmarkCollectionController) AddThread(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	collectionID, ok := ctrlshared.ParseIDParam(ctx, "id", "collection")
	if !ok {
		return nil
	}
	var req bookmarkcollectiondto.AddThreadRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}
	if req.ThreadID == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "thread_id is required"})
	}
	if err := c.svc.AddThread(claims.UserID, collectionID, req.ThreadID); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "thread added to collection successfully"})
}

func (c *BookmarkCollectionController) RemoveThread(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	collectionID, ok := ctrlshared.ParseIDParam(ctx, "id", "collection")
	if !ok {
		return nil
	}
	threadID, ok := ctrlshared.ParseIDParam(ctx, "threadId", "thread")
	if !ok {
		return nil
	}
	if err := c.svc.RemoveThread(claims.UserID, collectionID, threadID); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "thread removed from collection successfully"})
}

func (c *BookmarkCollectionController) CollectionIDsForThread(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	threadID, ok := ctrlshared.ParseIDParam(ctx, "threadId", "thread")
	if !ok {
		return nil
	}
	ids, err := c.svc.CollectionIDsForThread(claims.UserID, threadID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}
	if ids == nil {
		ids = []uint{}
	}
	return ctx.Status(fiber.StatusOK).JSON(bookmarkcollectiondto.CollectionIDsResponse{CollectionIDs: ids})
}

func (c *BookmarkCollectionController) ListThreads(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	collectionID, ok := ctrlshared.ParseIDParam(ctx, "id", "collection")
	if !ok {
		return nil
	}
	page, limit := ctrlshared.ParsePagination(ctx)
	threads, timestamps, total, err := c.svc.ListThreads(claims.UserID, collectionID, page, limit)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
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

	return ctx.Status(fiber.StatusOK).JSON(bookmarkcollectiondto.CollectionThreadsResponse{
		Data: data,
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}
