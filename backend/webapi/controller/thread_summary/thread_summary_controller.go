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

func NewThreadSummaryController(
	summaryService threadsummarysvciface.IThreadSummaryService,
) *ThreadSummaryController {
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
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	userID := claims.UserID

	// Detach from the fiber request context — the body stream writer runs after
	// the handler returns, so ctx.Context() would already be released.
	reqCtx := context.Background()

	// A cached summary costs no quota (no LLM call). Serve it directly.
	cached, isCached := c.summaryService.CachedSummary(reqCtx, threadID)

	// Quota gate must happen before SetBodyStreamWriter — once the stream body
	// starts the status code locks to 200, so we can't return 429 afterwards.
	if !isCached {
		if over, err := c.summaryService.QuotaExceeded(reqCtx, userID); err == nil && over {
			return ctx.Status(fiber.StatusTooManyRequests).JSON(authdto.ErrorResponse{
				Message: "You've reached your daily summary limit. It resets tomorrow.",
			})
		}
	}

	ctx.Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Set("Cache-Control", "no-cache")
	ctx.Set("X-Accel-Buffering", "no")

	ctx.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		if isCached {
			_, _ = w.WriteString(cached)
			_ = w.Flush()
			// Cached path never enters StreamSummarize — record the view here so
			// cached hits still appear in the user's summary history.
			_ = c.summaryService.RecordView(reqCtx, userID, threadID)
			return
		}
		if err := c.summaryService.StreamSummarize(reqCtx, threadID, userID, w); err != nil {
			log.Printf("thread summary stream failed for thread %d: %v", threadID, err)
		}
		_ = w.Flush()
		// Record after generation so the persisted summary row exists first.
		_ = c.summaryService.RecordView(reqCtx, userID, threadID)
	}))

	return nil
}
