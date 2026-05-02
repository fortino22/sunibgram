package controller

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/dto/utils"
	"github.com/sunibgram/webapi/ctxutil"
)

type ThreadController struct {
	threadService service.IThreadService
	mediaService  service.IThreadMediaService
}

func NewThreadController(threadService service.IThreadService, mediaService service.IThreadMediaService) *ThreadController {
	return &ThreadController{threadService: threadService, mediaService: mediaService}
}

func (c *ThreadController) Create(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	content := ctx.FormValue("content")
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

	if form, err := ctx.MultipartForm(); err == nil {
		if files := form.File["files"]; len(files) > 0 {
			if _, err := c.mediaService.Upload(ctx.Context(), thread.ThreadID, claims.UserID, files); err != nil {
				return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
			}
		}
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "thread created successfully"})
}

func (c *ThreadController) GetRoots(ctx *fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))

	threads, total, err := c.threadService.GetRoots(page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	data := c.buildFeedItems(threads)

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFeedResponse{
		Data: data,
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *ThreadController) GetByID(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid thread id"})
	}

	thread, err := c.threadService.GetByID(uint(id))
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: "thread not found"})
	}

	media, _ := c.mediaService.GetByThreadID(uint(id))
	return ctx.Status(fiber.StatusOK).JSON(c.fb().toFeedItem(thread, media))
}

func (c *ThreadController) GetChain(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid thread id"})
	}

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))

	threads, total, err := c.threadService.GetChain(uint(id), page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFeedResponse{
		Data: c.buildFeedItems(threads),
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

func (c *ThreadController) GetMyPosts(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))

	threads, repostTimestamps, total, err := c.threadService.GetMyPosts(claims.UserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFeedResponse{
		Data: c.buildMyPostFeedItems(threads, repostTimestamps),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *ThreadController) GetMyReplies(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))

	threads, repostTimestamps, parents, total, err := c.threadService.GetMyReplies(claims.UserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}

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
		item := dto.FeedItemResponse{
			ID:        t.ThreadID,
			ParentID:  t.ParentID,
			RootID:    t.RootID,
			Content:   t.Content,
			CreatedAt: utils.Wrap(t.CreatedAt),
			UpdatedAt: utils.Wrap(t.UpdatedAt),
			Author:    dto.FeedAuthor{ID: t.UserID, Username: t.User.UserName},
			Stats:     dto.FeedStats{Likes: t.LikeCount, Reposts: t.RepostCount, Bookmarks: t.BookmarkCount},
			Media:     c.fb().toMediaResponses(mediaByThread[t.ThreadID]),
		}
		switch {
		case repostTimestamps[i] != nil:
			item.Type = "REPOST"
			wrapped := utils.Wrap(*repostTimestamps[i])
			item.RepostedAt = &wrapped
		case parents[i] != nil:
			item.Type = "REPLY"
			p := parents[i]
			item.ParentContext = &dto.FeedParentContext{
				ID:        p.ThreadID,
				ParentID:  p.ParentID,
				RootID:    p.RootID,
				Author:    dto.FeedAuthor{ID: p.UserID, Username: p.User.UserName},
				Content:   p.Content,
				Media:     c.fb().toMediaResponses(mediaByThread[p.ThreadID]),
				Stats:     dto.FeedStats{Likes: p.LikeCount, Reposts: p.RepostCount, Bookmarks: p.BookmarkCount},
				CreatedAt: utils.Wrap(p.CreatedAt),
				UpdatedAt: utils.Wrap(p.UpdatedAt),
			}
		default:
			item.Type = "POST"
		}
		data[i] = item
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFeedResponse{
		Data: data,
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

func (c *ThreadController) fb() feedBuilder { return feedBuilder{c.mediaService} }

func (c *ThreadController) buildFeedItems(threads []model.Thread) []dto.FeedItemResponse {
	mediaByThread := c.loadMediaByThread(threads)
	data := make([]dto.FeedItemResponse, len(threads))
	for i, t := range threads {
		data[i] = c.fb().toFeedItem(&t, mediaByThread[t.ThreadID])
	}
	return data
}

func (c *ThreadController) buildMyPostFeedItems(threads []model.Thread, repostTimestamps []*time.Time) []dto.FeedItemResponse {
	mediaByThread := c.loadMediaByThread(threads)
	data := make([]dto.FeedItemResponse, len(threads))
	for i, t := range threads {
		item := c.fb().toFeedItem(&t, mediaByThread[t.ThreadID])
		if repostTimestamps[i] != nil {
			item.Type = "REPOST"
			wrapped := utils.Wrap(*repostTimestamps[i])
			item.RepostedAt = &wrapped
		}
		data[i] = item
	}
	return data
}
