package hashtagctrl

import (
	"github.com/gofiber/fiber/v2"

	hashtagmodel "github.com/sunibgram/domain/model/hashtag"
	hashtagsvciface "github.com/sunibgram/domain/service/hashtag"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	hashtagdto "github.com/sunibgram/webapi/dto/hashtag"
)

const (
	trendingHashtagLimit    = 30
	topTrendingHashtagLimit = 6
)

type HashtagController struct {
	hashtagService hashtagsvciface.IHashtagService
}

func NewHashtagController(hashtagService hashtagsvciface.IHashtagService) *HashtagController {
	return &HashtagController{
		hashtagService: hashtagService,
	}
}

func (c *HashtagController) GetTrending(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	hashtags, err := c.hashtagService.GetTrending(ctx.Context(), trendingHashtagLimit, claims.UserID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": toTrendingHashtagResponses(hashtags)})
}

func (c *HashtagController) GetTopTrending(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	hashtags, err := c.hashtagService.GetTrending(ctx.Context(), topTrendingHashtagLimit, claims.UserID)
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
