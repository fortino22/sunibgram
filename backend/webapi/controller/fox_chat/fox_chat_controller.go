package foxchatctrl

import (
	"bufio"
	"context"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"

	foxchatsvciface "github.com/sunibgram/domain/service/fox_chat"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	foxchatdto "github.com/sunibgram/webapi/dto/fox_chat"
)

type FoxChatController struct {
	chatService foxchatsvciface.IFoxChatService
}

func NewFoxChatController(chatService foxchatsvciface.IFoxChatService) *FoxChatController {
	return &FoxChatController{chatService: chatService}
}

// Ask streams the Fox guide answer as plain text so the client can render it
// live (typewriter effect). Quota is gated before the stream body starts —
// once SetBodyStreamWriter runs the status locks to 200.
func (c *FoxChatController) Ask(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	userID := claims.UserID

	var req foxchatdto.AskRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}
	question := strings.TrimSpace(req.Question)
	if question == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "question is required"})
	}

	// Detach from the fiber request context — the body stream writer runs after
	// the handler returns, so ctx.Context() would already be released.
	reqCtx := context.Background()

	history := make([]foxchatsvciface.QAPair, 0, len(req.History))
	for _, h := range req.History {
		history = append(history, foxchatsvciface.QAPair{Question: h.Question, Answer: h.Answer})
	}

	ctx.Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Set("Cache-Control", "no-cache")
	ctx.Set("X-Accel-Buffering", "no")

	ctx.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		if err := c.chatService.StreamAsk(reqCtx, userID, question, history, w); err != nil {
			log.Printf("fox chat stream failed for user %d: %v", userID, err)
		}
		_ = w.Flush()
	}))

	return nil
}
