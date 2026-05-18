package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
)

const trendingHashtagLimit = 30

type HashtagController struct {
	hashtagService service.IHashtagService
}

func NewHashtagController(hashtagService service.IHashtagService) *HashtagController {
	return &HashtagController{
		hashtagService: hashtagService,
	}
}

func (c *HashtagController) GetTrending(ctx *fiber.Ctx) error {
	hashtags, err := c.hashtagService.GetTrending(ctx.Context(), trendingHashtagLimit)
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
		}
	}
	return responses
}
