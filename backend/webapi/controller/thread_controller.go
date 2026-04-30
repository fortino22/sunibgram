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

	var req dto.CreateThreadRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	if _, err := c.threadService.Create(ctx.Context(), claims.UserID, req.Content, req.ParentID); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
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

	data := c.buildThreadResponses(threads)

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedThreadsResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
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
	return ctx.Status(fiber.StatusOK).JSON(toThreadResponse(thread, media))
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

	data := c.buildThreadResponses(threads)

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedThreadsResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
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

	var req dto.UpdateThreadRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	if _, err := c.threadService.Update(ctx.Context(), uint(id), claims.UserID, req.Content); err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
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

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedThreadsResponse{
		Data:  c.buildMyPostResponses(threads, repostTimestamps),
		Total: total,
		Page:  page,
		Limit: limit,
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
			Media:     toMediaResponses(mediaByThread[t.ThreadID]),
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
				Media:     toMediaResponses(mediaByThread[p.ThreadID]),
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

func (c *ThreadController) buildMyPostResponses(threads []model.Thread, repostTimestamps []*time.Time) []dto.ThreadResponse {
	ids := make([]uint, len(threads))
	for i, t := range threads {
		ids[i] = t.ThreadID
	}

	allMedia, _ := c.mediaService.GetByThreadIDs(ids)

	mediaByThread := make(map[uint][]model.ThreadMedia, len(ids))
	for _, m := range allMedia {
		mediaByThread[m.ThreadID] = append(mediaByThread[m.ThreadID], m)
	}

	data := make([]dto.ThreadResponse, len(threads))
	for i, t := range threads {
		r := toThreadResponse(&t, mediaByThread[t.ThreadID])
		if repostTimestamps[i] != nil {
			r.IsRepost = true
			wrapped := utils.Wrap(*repostTimestamps[i])
			r.RepostedAt = &wrapped
		}
		data[i] = r
	}
	return data
}

func (c *ThreadController) buildThreadResponses(threads []model.Thread) []dto.ThreadResponse {
	ids := make([]uint, len(threads))
	for i, t := range threads {
		ids[i] = t.ThreadID
	}

	allMedia, _ := c.mediaService.GetByThreadIDs(ids)

	mediaByThread := make(map[uint][]model.ThreadMedia, len(ids))
	for _, m := range allMedia {
		mediaByThread[m.ThreadID] = append(mediaByThread[m.ThreadID], m)
	}

	data := make([]dto.ThreadResponse, len(threads))
	for i, t := range threads {
		data[i] = toThreadResponse(&t, mediaByThread[t.ThreadID])
	}
	return data
}

func toThreadResponse(t *model.Thread, media []model.ThreadMedia) dto.ThreadResponse {
	return dto.ThreadResponse{
		ThreadID:      t.ThreadID,
		UserID:        t.UserID,
		Username:      t.User.UserName,
		Content:       t.Content,
		ParentID:      t.ParentID,
		RootID:        t.RootID,
		Media:         toMediaResponses(media),
		LikeCount:     t.LikeCount,
		RepostCount:   t.RepostCount,
		BookmarkCount: t.BookmarkCount,
		CreatedAt:     utils.Wrap(t.CreatedAt),
		UpdatedAt:     utils.Wrap(t.UpdatedAt),
	}
}

func toMediaResponses(media []model.ThreadMedia) []dto.ThreadMediaResponse {
	result := make([]dto.ThreadMediaResponse, len(media))
	for i, m := range media {
		result[i] = dto.ThreadMediaResponse{MediaID: m.MediaID, URL: m.URL}
	}
	return result
}
