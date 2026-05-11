package controller

import (
	"mime/multipart"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/webapi/ctxutil"
)

type ThreadController struct {
	threadService    service.IThreadService
	mediaService     service.IThreadMediaService
	memberService    service.ICommunityMemberService
	likeService      service.ILikeService
	bookmarkService  service.IBookmarkService
	threadPinService service.IThreadPinService
	profileService   service.IUserProfileService
}

func NewThreadController(threadService service.IThreadService, mediaService service.IThreadMediaService, memberService service.ICommunityMemberService, likeService service.ILikeService, bookmarkService service.IBookmarkService, threadPinService service.IThreadPinService, profileService service.IUserProfileService) *ThreadController {
	return &ThreadController{threadService: threadService, mediaService: mediaService, memberService: memberService, likeService: likeService, bookmarkService: bookmarkService, threadPinService: threadPinService, profileService: profileService}
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

	if thread.CommunityID != nil && thread.Community != nil && thread.Community.Visibility == model.CommunityVisibilityPrivate {
		if !c.memberService.IsMember(*thread.CommunityID, claims.UserID) {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
	}
	if err := c.threadService.CanViewUserThreads(ctx.Context(), claims.UserID, thread.UserID); err != nil {
		if err.Error() == "this account is private" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	media, _ := c.mediaService.GetByThreadID(uint(id))
	fb := c.fbNoPin(&claims.UserID)
	fb.LoadAvatarsForUserIDs(threadUserIDs([]model.Thread{*thread}))
	return ctx.Status(fiber.StatusOK).JSON(fb.ToFeedItem(thread, media))
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

	if root.CommunityID != nil && root.Community != nil && root.Community.Visibility == model.CommunityVisibilityPrivate {
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

func (c *ThreadController) GetMyMediaThread(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}
	id, err := strconv.ParseUint(ctx.Params("userId"), 10, 64)
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

	threads, parents, total, err := c.threadService.GetMyMediaThread(ctx.Context(), claims.UserID, targetUserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFeedResponse{
		Data: c.buildUserFeedItemsWithPin(threads, parents, &claims.UserID),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func threadUserIDs(threads []model.Thread) []uint {
	ids := make([]uint, 0, len(threads))
	for _, t := range threads {
		if t.Type == model.ThreadTypeRepost && t.OriginalThread != nil {
			ids = append(ids, t.OriginalThread.User.UserID)
		} else {
			ids = append(ids, t.User.UserID)
		}
	}
	return ids
}

func threadMediaFiles(ctx *fiber.Ctx) []*multipart.FileHeader {
	form, err := ctx.MultipartForm()
	if err != nil || form == nil {
		return nil
	}
	return form.File["files"]
}
