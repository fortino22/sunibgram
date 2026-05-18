package controller

import (
	"github.com/gofiber/fiber/v2"
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
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.JSON(fiber.Map{"summary": summary})
}
