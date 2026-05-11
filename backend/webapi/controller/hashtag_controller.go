package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
)

type HashtagController struct {
	hashtagService service.IHashtagService
}

func NewHashtagController(hashtagService service.IHashtagService) *HashtagController {
	return &HashtagController{
		hashtagService: hashtagService,
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
