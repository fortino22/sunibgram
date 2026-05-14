package controller

import (
	"bufio"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/service"
	"github.com/valyala/fasthttp"
)

type ThreadSummaryController struct {
	summaryService service.IThreadSummaryService
}

func NewThreadSummaryController(summaryService service.IThreadSummaryService) *ThreadSummaryController {
	return &ThreadSummaryController{summaryService: summaryService}
}

func (c *ThreadSummaryController) Stream(ctx *fiber.Ctx) error {
	threadID, ok := ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}

	ctx.Set("Content-Type", "text/event-stream")
	ctx.Set("Cache-Control", "no-cache")
	ctx.Set("Connection", "keep-alive")
	ctx.Set("X-Accel-Buffering", "no")

	reqCtx := ctx.Context()
	cached, cacheHit := c.summaryService.GetCachedSummary(reqCtx, threadID)
	ctx.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		if cacheHit {
			writeSSEData(w, cached)
			writeSSEDone(w)
			_ = w.Flush()
			return
		}

		if err := c.summaryService.StreamSummary(reqCtx, threadID, &summarySSEWriter{w: w}); err != nil {
			writeSSEData(w, "error: "+err.Error())
		}
		writeSSEDone(w)
		_ = w.Flush()
	}))

	return nil
}

type summarySSEWriter struct {
	w *bufio.Writer
}

func (w *summarySSEWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if err := writeSSEData(w.w, string(p)); err != nil {
		return 0, err
	}
	return len(p), w.w.Flush()
}

func (w *summarySSEWriter) Flush() {
	_ = w.w.Flush()
}

func writeSSEData(w *bufio.Writer, data string) error {
	_, err := fmt.Fprintf(w, "data: %s\n\n", data)
	return err
}

func writeSSEDone(w *bufio.Writer) {
	_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
}
