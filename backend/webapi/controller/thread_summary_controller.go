package controller

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/application/usecase"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
)

type ThreadSummaryController struct {
	summaryService service.IThreadSummaryService
}

func NewThreadSummaryController(summaryService service.IThreadSummaryService) *ThreadSummaryController {
	return &ThreadSummaryController{summaryService: summaryService}
}

func (c *ThreadSummaryController) GetSummary(ctx *fiber.Ctx) error {
	threadID, ok := ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}

	summary, err := c.summaryService.Summarize(ctx.Context(), threadID)
	if err != nil {
		if errors.Is(err, usecase.ErrThreadSummaryNotReady) {
			return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"status": "processing"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: "internal server error"})
	}

	return ctx.JSON(fiber.Map{"summary": summary})
}
