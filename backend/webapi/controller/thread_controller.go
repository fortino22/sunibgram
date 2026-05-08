package controller

import (
	"mime/multipart"
	"sort"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	dtoutils "github.com/sunibgram/dto/utils"
	"github.com/sunibgram/webapi/ctxutil"
	webapiutils "github.com/sunibgram/webapi/utils"
)

type ThreadController struct {
	threadService    service.IThreadService
	mediaService     service.IThreadMediaService
	memberService    service.ICommunityMemberService
	likeService      service.ILikeService
	bookmarkService  service.IBookmarkService
	threadPinService service.IThreadPinService
}

func NewThreadController(threadService service.IThreadService, mediaService service.IThreadMediaService, memberService service.ICommunityMemberService, likeService service.ILikeService, bookmarkService service.IBookmarkService, threadPinService service.IThreadPinService) *ThreadController {
	return &ThreadController{threadService: threadService, mediaService: mediaService, memberService: memberService, likeService: likeService, bookmarkService: bookmarkService, threadPinService: threadPinService}
}

func (c *ThreadController) Create(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	content := ctx.FormValue("content")
	files := threadMediaFiles(ctx)
	if err := model.ValidateThreadMediaCount(len(files)); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	var parentID *uint
	if raw := ctx.FormValue("parent_id"); raw != "" {
		if pid, err := strconv.ParseUint(raw, 10, 64); err == nil {
			v := uint(pid)
			parentID = &v
		}
	}

	thread, err := c.threadService.Create(ctx.Context(), claims.UserID, content, parentID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	if len(files) > 0 {
		if _, err := c.mediaService.Upload(ctx.Context(), thread.ThreadID, claims.UserID, files); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
		}
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "thread created successfully"})
}

func (c *ThreadController) GetRoots(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	var viewerID *uint
	if ok {
		viewerID = &claims.UserID
	}

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))

	threads, total, err := c.threadService.GetRoots(ctx.Context(), viewerID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	data := c.buildFeedItemsNoPin(threads, viewerID)
	if viewerID == nil {
		c.ensureFalseViewerFlags(data)
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFeedResponse{
		Data: data,
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *ThreadController) GetByID(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid thread id"})
	}

	thread, err := c.threadService.GetByID(uint(id))
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: "thread not found"})
	}

	if thread.CommunityID != nil {
		if !c.memberService.IsMember(*thread.CommunityID, claims.UserID) {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
	}

	media, _ := c.mediaService.GetByThreadID(uint(id))
	return ctx.Status(fiber.StatusOK).JSON(c.fbNoPin(&claims.UserID).ToFeedItem(thread, media))
}

func (c *ThreadController) GetChain(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid thread id"})
	}

	root, err := c.threadService.GetByID(uint(id))
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: "thread not found"})
	}

	if root.CommunityID != nil {
		if !c.memberService.IsMember(*root.CommunityID, claims.UserID) {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
	}

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))

	threads, total, err := c.threadService.GetChain(ctx.Context(), uint(id), claims.UserID, page, limit)
	if err != nil {
		if err.Error() == "this account is private" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFeedResponse{
		Data: c.buildFeedItemsNoPin(threads, &claims.UserID),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *ThreadController) Update(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid thread id"})
	}

	content := ctx.FormValue("content")

	if _, err := c.threadService.Update(ctx.Context(), uint(id), claims.UserID, content); err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	if ctx.FormValue("clear_media") == "true" {
		_ = c.mediaService.DeleteByThread(ctx.Context(), uint(id))
	} else if form, err := ctx.MultipartForm(); err == nil {
		if files := form.File["files"]; len(files) > 0 {
			if err := c.mediaService.Update(ctx.Context(), uint(id), claims.UserID, files); err != nil {
				return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
			}
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "thread updated successfully"})
}

func (c *ThreadController) Delete(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid thread id"})
	}

	if err := c.threadService.Delete(ctx.Context(), uint(id), claims.UserID); err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	_ = c.mediaService.DeleteByThread(ctx.Context(), uint(id))

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "thread deleted successfully"})
}

func (c *ThreadController) GetFollowingFeed(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))

	threads, parents, total, err := c.threadService.GetFollowingFeed(claims.UserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFeedResponse{
		Data: c.buildUserFeedItemsNoPin(threads, parents, &claims.UserID),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *ThreadController) GetMyPosts(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}
	id, err := strconv.ParseUint(ctx.Params("userID"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid user id"})
	}
	targetUserID := uint(id)
	if err := c.threadService.CanViewUserThreads(ctx.Context(), claims.UserID, targetUserID); err != nil {
		if err.Error() == "this account is private" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))

	threads, parents, total, err := c.threadService.GetMyPosts(ctx.Context(), claims.UserID, targetUserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFeedResponse{
		Data: c.buildUserFeedItemsWithPin(threads, parents, &claims.UserID),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *ThreadController) GetMyReplies(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}
	id, err := strconv.ParseUint(ctx.Params("userID"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid user id"})
	}
	targetUserID := uint(id)
	if err := c.threadService.CanViewUserThreads(ctx.Context(), claims.UserID, targetUserID); err != nil {
		if err.Error() == "this account is private" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))

	threads, parents, total, err := c.threadService.GetMyReplies(ctx.Context(), claims.UserID, targetUserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFeedResponse{
		Data: c.buildUserFeedItemsWithPin(threads, parents, &claims.UserID),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *ThreadController) loadMediaByThread(threads []model.Thread) map[uint][]model.ThreadMedia {
	ids := make([]uint, len(threads))
	for i, t := range threads {
		ids[i] = t.ThreadID
	}
	allMedia, _ := c.mediaService.GetByThreadIDs(ids)
	m := make(map[uint][]model.ThreadMedia, len(ids))
	for _, media := range allMedia {
		m[media.ThreadID] = append(m[media.ThreadID], media)
	}
	return m
}

func (c *ThreadController) fb(viewerID *uint) webapiutils.FeedBuilder {
	return webapiutils.FeedBuilder{
		Media:    c.mediaService,
		ViewerID: viewerID,
	}
}

func (c *ThreadController) fbWithPin(viewerID *uint) webapiutils.FeedBuilder {
	return webapiutils.FeedBuilder{
		Media:       c.mediaService,
		LikeSvc:     c.likeService,
		BookmarkSvc: c.bookmarkService,
		ThreadSvc:   c.threadService,
		PinSvc:      c.threadPinService,
		ViewerID:    viewerID,
	}
}

func (c *ThreadController) fbNoPin(viewerID *uint) webapiutils.FeedBuilder {
	return webapiutils.FeedBuilder{
		Media:       c.mediaService,
		LikeSvc:     c.likeService,
		BookmarkSvc: c.bookmarkService,
		ThreadSvc:   c.threadService,
		ViewerID:    viewerID,
	}
}

func (c *ThreadController) buildFeedItems(threads []model.Thread, viewerID *uint) []dto.FeedItemResponse {
	mediaByThread := c.loadMediaByThread(threads)
	data := make([]dto.FeedItemResponse, len(threads))
	fb := c.fb(viewerID)
	for i, t := range threads {
		data[i] = fb.ToFeedItem(&t, mediaByThread[t.ThreadID])
	}
	return data
}

func (c *ThreadController) buildFeedItemsNoPin(threads []model.Thread, viewerID *uint) []dto.FeedItemResponse {
	mediaByThread := c.loadMediaByThread(threads)
	data := make([]dto.FeedItemResponse, len(threads))
	fb := c.fbNoPin(viewerID)
	for i, t := range threads {
		data[i] = fb.ToFeedItem(&t, mediaByThread[t.ThreadID])
	}
	return data
}

func (c *ThreadController) buildFeedItemsWithPin(threads []model.Thread, viewerID *uint) []dto.FeedItemResponse {
	mediaByThread := c.loadMediaByThread(threads)
	data := make([]dto.FeedItemResponse, len(threads))
	fb := c.fbWithPin(viewerID)
	for i, t := range threads {
		data[i] = fb.ToFeedItem(&t, mediaByThread[t.ThreadID])
	}
	return data
}

func (c *ThreadController) buildUserFeedItems(threads []model.Thread, parents []*model.Thread, viewerID *uint) []dto.FeedItemResponse {
	return c.buildUserFeedItemsWithBuilder(threads, parents, c.fb(viewerID))
}

func (c *ThreadController) buildUserFeedItemsNoPin(threads []model.Thread, parents []*model.Thread, viewerID *uint) []dto.FeedItemResponse {
	return c.buildUserFeedItemsWithBuilder(threads, parents, c.fbNoPin(viewerID))
}

func (c *ThreadController) buildUserFeedItemsWithPin(threads []model.Thread, parents []*model.Thread, viewerID *uint) []dto.FeedItemResponse {
	return c.buildUserFeedItemsWithBuilder(threads, parents, c.fbWithPin(viewerID))
}

func (c *ThreadController) buildUserFeedItemsWithBuilder(threads []model.Thread, parents []*model.Thread, fb webapiutils.FeedBuilder) []dto.FeedItemResponse {
	allIDs := make([]uint, 0, len(threads)*2)
	for _, t := range threads {
		allIDs = append(allIDs, t.ThreadID)
	}
	for _, p := range parents {
		if p != nil {
			allIDs = append(allIDs, p.ThreadID)
		}
	}
	allMedia, _ := c.mediaService.GetByThreadIDs(allIDs)
	mediaByThread := make(map[uint][]model.ThreadMedia, len(allIDs))
	for _, m := range allMedia {
		mediaByThread[m.ThreadID] = append(mediaByThread[m.ThreadID], m)
	}

	data := make([]dto.FeedItemResponse, len(threads))
	for i, t := range threads {
		item := fb.ToFeedItem(&t, mediaByThread[t.ThreadID])
		if parents[i] != nil {
			p := parents[i]
			item.ParentContext = &dto.FeedParentContext{
				ID:        p.ThreadID,
				ParentID:  p.ParentID,
				RootID:    p.RootID,
				Community: webapiutils.CommunityField(p),
				Author:    webapiutils.AuthorField(p.User),
				Content:   p.Content,
				Media:     fb.ToMediaResponses(mediaByThread[p.ThreadID]),
				Stats:     dto.FeedStats{Likes: p.LikeCount, Reposts: p.RepostCount, Bookmarks: p.BookmarkCount, Comments: p.CommentCount},
				CreatedAt: dtoutils.Wrap(p.CreatedAt),
				UpdatedAt: dtoutils.Wrap(p.UpdatedAt),
			}
		}
		data[i] = item
	}
	sort.SliceStable(data, func(i, j int) bool {
		return boolValue(data[i].IsPinned) && !boolValue(data[j].IsPinned)
	})
	return data
}

func boolValue(v *bool) bool {
	return v != nil && *v
}

func threadMediaFiles(ctx *fiber.Ctx) []*multipart.FileHeader {
	form, err := ctx.MultipartForm()
	if err != nil || form == nil {
		return nil
	}
	return form.File["files"]
}

func (c *ThreadController) ensureFalseViewerFlags(data []dto.FeedItemResponse) {
	falseVal := false
	for i := range data {
		if data[i].IsLiked == nil {
			data[i].IsLiked = &falseVal
		}
		if data[i].IsBookmarked == nil {
			data[i].IsBookmarked = &falseVal
		}
		if data[i].IsReposted == nil {
			data[i].IsReposted = &falseVal
		}
	}
}
