package controller

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/domain/service"
	authdto "github.com/sunibgram/webapi/dto/auth"
)

type CacheController struct {
	cacheService service.ICacheService
}

func NewCacheController(cacheService service.ICacheService) *CacheController {
	return &CacheController{cacheService: cacheService}
}

func (c *CacheController) Clear(ctx *fiber.Ctx) error {
	if err := c.cacheService.ClearAll(); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "cache cleared successfully"})
}
