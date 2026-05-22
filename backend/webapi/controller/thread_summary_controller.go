package controller

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	threadsummaryuc "github.com/sunibgram/application/usecase/thread_summary"
	threadsummarysvciface "github.com/sunibgram/domain/service/thread_summary"
	authdto "github.com/sunibgram/webapi/dto/auth"
)

type ThreadSummaryController struct {
	summaryService threadsummarysvciface.IThreadSummaryService
}

func NewThreadSummaryController(summaryService threadsummarysvciface.IThreadSummaryService) *ThreadSummaryController {
	return &ThreadSummaryController{summaryService: summaryService}
}

func (c *ThreadSummaryController) GetSummary(ctx *fiber.Ctx) error {
	threadID, ok := ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}

	summary, err := c.summaryService.Summarize(ctx.Context(), threadID)
	if err != nil {
		if errors.Is(err, threadsummaryuc.ErrThreadSummaryNotReady) {
			return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"status": "processing"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}

	return ctx.JSON(fiber.Map{"summary": summary})
}
