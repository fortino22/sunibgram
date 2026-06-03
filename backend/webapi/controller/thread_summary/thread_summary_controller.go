package threadsummaryctrl

import (
	"bufio"
	"context"
	"errors"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"

	threadsummaryuc "github.com/sunibgram/application/usecase/thread_summary"
	foxhistorymodel "github.com/sunibgram/domain/model/fox_history"
	threadmediasvciface "github.com/sunibgram/domain/service/thread_media"
	threadsummarysvciface "github.com/sunibgram/domain/service/thread_summary"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	dtoutils "github.com/sunibgram/webapi/dto/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	threaddto "github.com/sunibgram/webapi/dto/thread"
	threadmediadto "github.com/sunibgram/webapi/dto/thread_media"
	threadsummarydto "github.com/sunibgram/webapi/dto/thread_summary"
)

type ThreadSummaryController struct {
	summaryService threadsummarysvciface.IThreadSummaryService
	mediaService   threadmediasvciface.IThreadMediaService
}

func NewThreadSummaryController(
	summaryService threadsummarysvciface.IThreadSummaryService,
	mediaService threadmediasvciface.IThreadMediaService,
) *ThreadSummaryController {
	return &ThreadSummaryController{summaryService: summaryService, mediaService: mediaService}
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

// GetHistory returns the authenticated user's summary history: threads they have
// summarized, newest first, each with the summarized content and the result.
func (c *ThreadSummaryController) GetHistory(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	page, limit := ctrlshared.ParsePagination(ctx)
	q := strings.TrimSpace(ctx.Query("q"))

	rows, total, err := c.summaryService.ListHistory(ctx.Context(), claims.UserID, page, limit, q)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}

	// Group current thread media by thread (summary rows only) for the detail view.
	ids := make([]uint, 0, len(rows))
	for _, r := range rows {
		if r.ThreadID != nil {
			ids = append(ids, *r.ThreadID)
		}
	}
	mediaByThread := ctrlshared.LoadThreadMediaByID(c.mediaService, ids)

	data := make([]threadsummarydto.SummaryHistoryItem, len(rows))
	for i, r := range rows {
		item := threadsummarydto.SummaryHistoryItem{
			ID:           r.ID,
			Type:         r.Type,
			Summary:      r.Body,
			SummarizedAt: dtoutils.Wrap(r.UpdatedAt),
			Media:        []threadmediadto.ThreadMediaResponse{},
		}
		if r.Type == foxhistorymodel.TypeChat {
			item.Question = r.Question
		} else {
			var threadID uint
			if r.ThreadID != nil {
				threadID = *r.ThreadID
			}
			item.ThreadID = threadID
			item.Content = r.Snapshot
			media := mediaByThread[threadID]
			mediaResp := make([]threadmediadto.ThreadMediaResponse, len(media))
			for j, m := range media {
				mediaResp[j] = threadmediadto.ThreadMediaResponse{MediaID: m.MediaID, URL: c.mediaService.GetURL(m.ObjectName)}
			}
			item.Media = mediaResp
		}
		data[i] = item
	}

	return ctx.Status(fiber.StatusOK).JSON(threadsummarydto.PaginatedSummaryHistoryResponse{
		Data: data,
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}
