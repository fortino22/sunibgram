package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/webapi/ctxutil"
)

type HashtagController struct {
	hashtagService  service.IHashtagService
	mediaService    service.IThreadMediaService
	likeService     service.ILikeService
	bookmarkService service.IBookmarkService
	threadService   service.IThreadService
}

func NewHashtagController(
	hashtagService service.IHashtagService,
	mediaService service.IThreadMediaService,
	likeService service.ILikeService,
	bookmarkService service.IBookmarkService,
	threadService service.IThreadService,
) *HashtagController {
	return &HashtagController{
		hashtagService:  hashtagService,
		mediaService:    mediaService,
		likeService:     likeService,
		bookmarkService: bookmarkService,
		threadService:   threadService,
	}
}

func (c *HashtagController) GetTrending(ctx *fiber.Ctx) error {
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))
	hashtags, err := c.hashtagService.GetTrending(ctx.Context(), limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": toTrendingHashtagResponses(hashtags)})
}

func (c *HashtagController) GetThreadsByHashtag(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))
	var cursor *uint
	if raw := ctx.Query("cursor"); raw != "" {
		parsed, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid cursor"})
		}
		v := uint(parsed)
		cursor = &v
	}

	threads, nextCursor, err := c.hashtagService.GetThreadsByHashtag(ctx.Context(), ctx.Params("name"), cursor, claims.UserID, limit)
	if err != nil {
		if err.Error() == "hashtag is required" {
			return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	mediaByThread := loadThreadMediaByThread(c.mediaService, threads)
	fb := newFeedBuilder(c.mediaService, c.likeService, c.bookmarkService, c.threadService, nil, &claims.UserID)
	data := buildFeedItemsFromThreads(threads, mediaByThread, fb)
	return ctx.Status(fiber.StatusOK).JSON(dto.ThreadsByHashtagResponse{
		Data:       data,
		NextCursor: nextCursor,
		Limit:      normalizedHashtagLimit(limit),
	})
}

func toTrendingHashtagResponses(hashtags []model.TrendingHashtag) []dto.TrendingHashtagResponse {
	responses := make([]dto.TrendingHashtagResponse, len(hashtags))
	for i, hashtag := range hashtags {
		responses[i] = dto.TrendingHashtagResponse{
			Name:        hashtag.Name,
			ThreadCount: hashtag.ThreadCount,
			Score:       hashtag.Score,
		}
	}
	return responses
}

func normalizedHashtagLimit(limit int) int {
	if limit < 1 || limit > 100 {
		return 20
	}
	return limit
}
