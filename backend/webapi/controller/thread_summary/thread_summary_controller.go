package threadsummaryctrl

import (
	"bufio"
	"context"
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"

	threadsummaryuc "github.com/sunibgram/application/usecase/thread_summary"
	threadsummarysvciface "github.com/sunibgram/domain/service/thread_summary"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
)

type ThreadSummaryController struct {
	summaryService threadsummarysvciface.IThreadSummaryService
}

func NewThreadSummaryController(summaryService threadsummarysvciface.IThreadSummaryService) *ThreadSummaryController {
	return &ThreadSummaryController{summaryService: summaryService}
}

func (c *ThreadSummaryController) GetSummary(ctx *fiber.Ctx) error {
	threadID, ok := ctrlshared.ParseIDParam(ctx, "id", "thread")
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

// StreamSummary streams the summary as plain text as it is generated, so the
// client can render it live (typewriter effect). Cached summaries are written
// in one shot.
func (c *ThreadSummaryController) StreamSummary(ctx *fiber.Ctx) error {
	threadID, ok := ctrlshared.ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}

	ctx.Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Set("Cache-Control", "no-cache")
	ctx.Set("X-Accel-Buffering", "no")

	// Detach from the fiber request context — the body stream writer runs after
	// the handler returns, so ctx.Context() would already be released.
	reqCtx := context.Background()

	ctx.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		if err := c.summaryService.StreamSummarize(reqCtx, threadID, w); err != nil {
			log.Printf("thread summary stream failed for thread %d: %v", threadID, err)
		}
		_ = w.Flush()
	}))

	return nil
}
