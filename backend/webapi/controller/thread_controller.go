package controller

import (
	"sort"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	dtoutils "github.com/sunibgram/dto/utils"
	webapihelper "github.com/sunibgram/webapi/helper"
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
	claims, ok := webapihelper.Claims(ctx)
	if !ok {
		return webapihelper.Unauthorized(ctx)
	}

	content := ctx.FormValue("content")
	files := webapihelper.MultipartFiles(ctx, "files")
	if err := model.ValidateThreadMediaCount(len(files)); err != nil {
		return webapihelper.Error(ctx, fiber.StatusUnprocessableEntity, err.Error())
	}

	parentID := webapihelper.OptionalFormUint(ctx, "parent_id")

	thread, err := c.threadService.Create(ctx.Context(), claims.UserID, content, parentID)
	if err != nil {
		return webapihelper.Error(ctx, fiber.StatusBadRequest, err.Error())
	}

	if len(files) > 0 {
		if _, err := c.mediaService.Upload(ctx.Context(), thread.ThreadID, claims.UserID, files); err != nil {
			return webapihelper.Error(ctx, fiber.StatusBadRequest, err.Error())
		}
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "thread created successfully"})
}

func (c *ThreadController) GetRoots(ctx *fiber.Ctx) error {
	viewerID := webapihelper.ViewerID(ctx)

	page, limit := webapihelper.PageLimit(ctx)

	threads, total, err := c.threadService.GetRoots(ctx.Context(), viewerID, page, limit)
	if err != nil {
		return webapihelper.Error(ctx, fiber.StatusInternalServerError, err.Error())
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
	claims, ok := webapihelper.Claims(ctx)
	if !ok {
		return webapihelper.Unauthorized(ctx)
	}

	id, err := webapihelper.ParamUint(ctx, "id")
	if err != nil {
		return webapihelper.Error(ctx, fiber.StatusBadRequest, "invalid thread id")
	}

	thread, err := c.threadService.GetByID(id)
	if err != nil {
		return webapihelper.Error(ctx, fiber.StatusNotFound, "thread not found")
	}

	if thread.CommunityID != nil && thread.Community != nil && thread.Community.Visibility == model.CommunityVisibilityPrivate {
		if !c.memberService.IsMember(*thread.CommunityID, claims.UserID) {
			return webapihelper.Error(ctx, fiber.StatusForbidden, "forbidden")
		}
	}
	if err := c.threadService.CanViewUserThreads(ctx.Context(), claims.UserID, thread.UserID); err != nil {
		if err.Error() == "this account is private" {
			return webapihelper.Error(ctx, fiber.StatusForbidden, err.Error())
		}
		return webapihelper.Error(ctx, fiber.StatusInternalServerError, err.Error())
	}

	media, _ := c.mediaService.GetByThreadID(id)
	return ctx.Status(fiber.StatusOK).JSON(c.fbNoPin(&claims.UserID).ToFeedItem(thread, media))
}

func (c *ThreadController) GetChain(ctx *fiber.Ctx) error {
	claims, ok := webapihelper.Claims(ctx)
	if !ok {
		return webapihelper.Unauthorized(ctx)
	}

	id, err := webapihelper.ParamUint(ctx, "id")
	if err != nil {
		return webapihelper.Error(ctx, fiber.StatusBadRequest, "invalid thread id")
	}

	root, err := c.threadService.GetByID(id)
	if err != nil {
		return webapihelper.Error(ctx, fiber.StatusNotFound, "thread not found")
	}

	if root.CommunityID != nil && root.Community != nil && root.Community.Visibility == model.CommunityVisibilityPrivate {
		if !c.memberService.IsMember(*root.CommunityID, claims.UserID) {
			return webapihelper.Error(ctx, fiber.StatusForbidden, "forbidden")
		}
	}

	page, limit := webapihelper.PageLimit(ctx)

	threads, total, err := c.threadService.GetChain(ctx.Context(), id, claims.UserID, page, limit)
	if err != nil {
		if err.Error() == "this account is private" {
			return webapihelper.Error(ctx, fiber.StatusForbidden, err.Error())
		}
		return webapihelper.Error(ctx, fiber.StatusNotFound, err.Error())
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFeedResponse{
		Data: c.buildFeedItemsNoPin(threads, &claims.UserID),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *ThreadController) Update(ctx *fiber.Ctx) error {
	claims, ok := webapihelper.Claims(ctx)
	if !ok {
		return webapihelper.Unauthorized(ctx)
	}

	id, err := webapihelper.ParamUint(ctx, "id")
	if err != nil {
		return webapihelper.Error(ctx, fiber.StatusBadRequest, "invalid thread id")
	}

	content := ctx.FormValue("content")

	if _, err := c.threadService.Update(ctx.Context(), id, claims.UserID, content); err != nil {
		if err.Error() == "forbidden" {
			return webapihelper.Error(ctx, fiber.StatusForbidden, "forbidden")
		}
		return webapihelper.Error(ctx, fiber.StatusBadRequest, err.Error())
	}

	if ctx.FormValue("clear_media") == "true" {
		_ = c.mediaService.DeleteByThread(ctx.Context(), id)
	} else if files := webapihelper.MultipartFiles(ctx, "files"); len(files) > 0 {
		if err := c.mediaService.Update(ctx.Context(), id, claims.UserID, files); err != nil {
			return webapihelper.Error(ctx, fiber.StatusBadRequest, err.Error())
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "thread updated successfully"})
}

func (c *ThreadController) Delete(ctx *fiber.Ctx) error {
	claims, ok := webapihelper.Claims(ctx)
	if !ok {
		return webapihelper.Unauthorized(ctx)
	}

	id, err := webapihelper.ParamUint(ctx, "id")
	if err != nil {
		return webapihelper.Error(ctx, fiber.StatusBadRequest, "invalid thread id")
	}

	if err := c.threadService.Delete(ctx.Context(), id, claims.UserID); err != nil {
		if err.Error() == "forbidden" {
			return webapihelper.Error(ctx, fiber.StatusForbidden, "forbidden")
		}
		return webapihelper.Error(ctx, fiber.StatusNotFound, err.Error())
	}

	_ = c.mediaService.DeleteByThread(ctx.Context(), id)

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "thread deleted successfully"})
}

func (c *ThreadController) GetFollowingFeed(ctx *fiber.Ctx) error {
	claims, ok := webapihelper.Claims(ctx)
	if !ok {
		return webapihelper.Unauthorized(ctx)
	}
	page, limit := webapihelper.PageLimit(ctx)

	threads, parents, total, err := c.threadService.GetFollowingFeed(claims.UserID, page, limit)
	if err != nil {
		return webapihelper.Error(ctx, fiber.StatusInternalServerError, err.Error())
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFeedResponse{
		Data: c.buildUserFeedItemsNoPin(threads, parents, &claims.UserID),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *ThreadController) GetMyPosts(ctx *fiber.Ctx) error {
	claims, ok := webapihelper.Claims(ctx)
	if !ok {
		return webapihelper.Unauthorized(ctx)
	}
	targetUserID, err := webapihelper.ParamUint(ctx, "userID")
	if err != nil {
		return webapihelper.Error(ctx, fiber.StatusBadRequest, "invalid user id")
	}
	if err := c.threadService.CanViewUserThreads(ctx.Context(), claims.UserID, targetUserID); err != nil {
		if err.Error() == "this account is private" {
			return webapihelper.Error(ctx, fiber.StatusForbidden, err.Error())
		}
		return webapihelper.Error(ctx, fiber.StatusInternalServerError, err.Error())
	}
	page, limit := webapihelper.PageLimit(ctx)

	threads, parents, total, err := c.threadService.GetMyPosts(ctx.Context(), claims.UserID, targetUserID, page, limit)
	if err != nil {
		return webapihelper.Error(ctx, fiber.StatusInternalServerError, err.Error())
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFeedResponse{
		Data: c.buildUserFeedItemsWithPin(threads, parents, &claims.UserID),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *ThreadController) GetMyReplies(ctx *fiber.Ctx) error {
	claims, ok := webapihelper.Claims(ctx)
	if !ok {
		return webapihelper.Unauthorized(ctx)
	}
	targetUserID, err := webapihelper.ParamUint(ctx, "userID")
	if err != nil {
		return webapihelper.Error(ctx, fiber.StatusBadRequest, "invalid user id")
	}
	if err := c.threadService.CanViewUserThreads(ctx.Context(), claims.UserID, targetUserID); err != nil {
		if err.Error() == "this account is private" {
			return webapihelper.Error(ctx, fiber.StatusForbidden, err.Error())
		}
		return webapihelper.Error(ctx, fiber.StatusInternalServerError, err.Error())
	}
	page, limit := webapihelper.PageLimit(ctx)

	threads, parents, total, err := c.threadService.GetMyReplies(ctx.Context(), claims.UserID, targetUserID, page, limit)
	if err != nil {
		return webapihelper.Error(ctx, fiber.StatusInternalServerError, err.Error())
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFeedResponse{
		Data: c.buildUserFeedItemsWithPin(threads, parents, &claims.UserID),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *ThreadController) GetMyMediaThread(ctx *fiber.Ctx) error {
	claims, ok := webapihelper.Claims(ctx)
	if !ok {
		return webapihelper.Unauthorized(ctx)
	}
	targetUserID, err := webapihelper.ParamUint(ctx, "userId")
	if err != nil {
		return webapihelper.Error(ctx, fiber.StatusBadRequest, "invalid user id")
	}
	page, limit := webapihelper.PageLimit(ctx)

	threads, parents, total, err := c.threadService.GetMyMediaThread(ctx.Context(), claims.UserID, targetUserID, page, limit)
	if err != nil {
		if err.Error() == "this account is private" {
			return webapihelper.Error(ctx, fiber.StatusForbidden, err.Error())
		}
		return webapihelper.Error(ctx, fiber.StatusInternalServerError, err.Error())
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
