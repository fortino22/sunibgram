package controller

import (
	webapiutils "github.com/sunibgram/webapi/utils"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
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
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
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

	eventID, _, err := parseOptionalEventID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	thread, err := c.threadService.Create(ctx.Context(), claims.UserID, content, parentID, eventID)
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
	claims, ok := webapiutils.Get(ctx)
	var viewerID *uint
	if ok {
		viewerID = &claims.UserID
	}

	page, limit := ParsePagination(ctx)

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
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}

	thread, err := c.threadService.GetByID(id)
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

	media, _ := c.mediaService.GetByThreadID(id)
	fb := c.fbNoPin(&claims.UserID)
	fb.LoadAvatarsForUserIDs(threadUserIDs([]model.Thread{*thread}))
	return ctx.Status(fiber.StatusOK).JSON(fb.ToFeedItem(thread, media))
}

func (c *ThreadController) GetChain(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}

	root, err := c.threadService.GetByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: "thread not found"})
	}

	if root.CommunityID != nil && root.Community != nil && root.Community.Visibility == model.CommunityVisibilityPrivate {
		if !c.memberService.IsMember(*root.CommunityID, claims.UserID) {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
	}

	page, limit := ParsePagination(ctx)

	threads, total, err := c.threadService.GetChain(ctx.Context(), id, claims.UserID, page, limit)
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
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}

	content := ctx.FormValue("content")

	if _, err := c.threadService.Update(ctx.Context(), id, claims.UserID, content); err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	keepIDs, keepSent, err := parseKeepMediaIDs(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	files := threadMediaFiles(ctx)
	if err := model.ValidateThreadMediaCount(len(keepIDs) + len(files)); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	if keepSent || len(files) > 0 {
		if err := c.mediaService.UpdateGranular(ctx.Context(), id, claims.UserID, keepIDs, files); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "thread updated successfully"})
}

func (c *ThreadController) Delete(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}

	if err := c.threadService.Delete(ctx.Context(), id, claims.UserID); err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	_ = c.mediaService.DeleteByThread(ctx.Context(), id)

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "thread deleted successfully"})
}

func (c *ThreadController) GetFollowingFeed(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	page, limit := ParsePagination(ctx)

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
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	id, ok := ParseIDParam(ctx, "userID", "user")
	if !ok {
		return nil
	}
	targetUserID := id
	if err := c.threadService.CanViewUserThreads(ctx.Context(), claims.UserID, targetUserID); err != nil {
		if err.Error() == "this account is private" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	page, limit := ParsePagination(ctx)

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
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	id, ok := ParseIDParam(ctx, "userID", "user")
	if !ok {
		return nil
	}
	targetUserID := id
	if err := c.threadService.CanViewUserThreads(ctx.Context(), claims.UserID, targetUserID); err != nil {
		if err.Error() == "this account is private" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	page, limit := ParsePagination(ctx)

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
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	id, ok := ParseIDParam(ctx, "userId", "user")
	if !ok {
		return nil
	}
	targetUserID := id
	if err := c.threadService.CanViewUserThreads(ctx.Context(), claims.UserID, targetUserID); err != nil {
		if err.Error() == "this account is private" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	page, limit := ParsePagination(ctx)

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

func parseKeepMediaIDs(ctx *fiber.Ctx) ([]uint, bool, error) {
	var values []string
	sent := false

	form, err := ctx.MultipartForm()
	if err == nil && form != nil {
		if raw, ok := form.Value["keep_media_ids"]; ok {
			values = append(values, raw...)
			sent = true
		}
	}

	if !sent {
		if raw := ctx.FormValue("keep_media_ids"); raw != "" {
			values = append(values, raw)
			sent = true
		}
	}

	if !sent {
		return nil, false, nil
	}

	ids := make([]uint, 0, len(values))
	seen := make(map[uint]struct{}, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			id, err := strconv.ParseUint(part, 10, 64)
			if err != nil || id == 0 {
				return nil, true, errInvalidKeepMediaID()
			}
			mediaID := uint(id)
			if _, ok := seen[mediaID]; ok {
				continue
			}
			seen[mediaID] = struct{}{}
			ids = append(ids, mediaID)
		}
	}

	return ids, true, nil
}

func errInvalidKeepMediaID() error {
	return fiber.NewError(fiber.StatusBadRequest, "invalid keep_media_ids")
}

func parseOptionalEventID(ctx *fiber.Ctx) (*string, bool, error) {
	raw, ok := formValueWithPresence(ctx, "event_id")
	if !ok {
		raw, ok = formValueWithPresence(ctx, "eventId")
	}
	if !ok {
		return nil, false, nil
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, true, nil
	}
	return &raw, true, nil
}

func formValueWithPresence(ctx *fiber.Ctx, key string) (string, bool) {
	if form, err := ctx.MultipartForm(); err == nil && form != nil {
		if raw, ok := form.Value[key]; ok {
			if len(raw) == 0 {
				return "", true
			}
			return raw[0], true
		}
	}

	body := ctx.Request().PostArgs()
	if body.Has(key) {
		return string(body.Peek(key)), true
	}
	return "", false
}
