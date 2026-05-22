package controller

import (
	"github.com/gofiber/fiber/v2"

	hashtagmodel "github.com/sunibgram/domain/model/hashtag"
	"github.com/sunibgram/domain/service"
	authdto "github.com/sunibgram/webapi/dto/auth"
	hashtagdto "github.com/sunibgram/webapi/dto/hashtag"
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
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": toTrendingHashtagResponses(hashtags)})
}

func toTrendingHashtagResponses(hashtags []hashtagmodel.TrendingHashtag) []hashtagdto.TrendingHashtagResponse {
	responses := make([]hashtagdto.TrendingHashtagResponse, len(hashtags))
	for i, hashtag := range hashtags {
		responses[i] = hashtagdto.TrendingHashtagResponse{
			Name:        hashtag.Name,
			ThreadCount: hashtag.ThreadCount,
		}
	}
	return responses
}
