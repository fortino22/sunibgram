package cachectrl

import (
	"github.com/gofiber/fiber/v2"

	cachesvciface "github.com/sunibgram/domain/service/cache"
	authdto "github.com/sunibgram/webapi/dto/auth"
)

type CacheController struct {
	cacheService cachesvciface.ICacheService
}

func NewCacheController(cacheService cachesvciface.ICacheService) *CacheController {
	return &CacheController{cacheService: cacheService}
}

func (c *CacheController) Clear(ctx *fiber.Ctx) error {
	if err := c.cacheService.ClearAll(); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "cache cleared successfully"})
}
