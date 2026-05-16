package controller

import (
	"context"
	"log"
	"mime/multipart"
	"net/url"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	dtoutils "github.com/sunibgram/dto/utils"
)

type ChatController struct {
	chatService         service.IChatService
	chatMediaService    service.IChatMediaService
	chatReactionService service.IChatReactionService
	hub                 service.IHub
}

func NewChatController(chatService service.IChatService, chatMediaService service.IChatMediaService, chatReactionService service.IChatReactionService, hubs ...service.IHub) *ChatController {
	var hub service.IHub
	if len(hubs) > 0 {
		hub = hubs[0]
	}
	return &ChatController{chatService: chatService, chatMediaService: chatMediaService, chatReactionService: chatReactionService, hub: hub}
}

func (c *ChatController) Send(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	roomID, ok := ParseIDParam(ctx, "room_id", "chat room")
	if !ok {
		return nil
	}

	var message *string
	if raw := strings.TrimSpace(ctx.FormValue("message")); raw != "" {
		message = &raw
	}

	var gifURL *string
	if raw := strings.TrimSpace(ctx.FormValue("gif_url")); raw != "" {
		if err := model.ValidateGifURL(raw); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		gifURL = &raw
	}

	var replyToID *uint
	if raw := strings.TrimSpace(ctx.FormValue("reply_to_id")); raw != "" {
		parsed, err := strconv.ParseUint(raw, 10, 64)
		if err == nil && parsed > 0 {
			rid := uint(parsed)
			replyToID = &rid
		}
	}

	var files []*multipart.FileHeader
	form, _ := ctx.MultipartForm()
	if form != nil {
		files = form.File["files"]
	}
	if message == nil && gifURL == nil && len(files) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "message or media is required"})
	}

	chat, err := c.chatService.Send(ctx.Context(), roomID, claims.UserID, message, gifURL, replyToID)
	if err != nil {
		return HandleServiceError(ctx, err)
	}

	var media []model.ChatMedia
	if len(files) > 0 {
		media, err = c.chatMediaService.Upload(ctx.Context(), chat.ChatID, claims.UserID, files)
		if err != nil {
			_ = c.chatService.Delete(ctx.Context(), chat.ChatID)
			return HandleServiceError(ctx, err)
		}
	}

	c.broadcastRoom(chat.RoomID, "message:new", c.toChatResponse(ctx.Context(), chat, media, nil))

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "chat sent successfully"})
}

func (c *ChatController) GetMessages(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	roomID, ok := ParseIDParam(ctx, "room_id", "chat room")
	if !ok {
		return nil
	}
	page, limit := ParsePagination(ctx)

	markedIDs, err := c.chatService.MarkRead(roomID, claims.UserID)
	if err != nil {
		return HandleServiceError(ctx, err)
	}
	chats, total, err := c.chatService.GetMessages(roomID, claims.UserID, page, limit)
	if err != nil {
		return HandleServiceError(ctx, err)
	}

	mediaByChat := c.loadMediaByChat(chats)
	reactionsByChat := c.loadReactionsByChat(chats)
	data := make([]dto.ChatResponse, len(chats))
	for i := range chats {
		data[i] = c.toChatResponse(ctx.Context(), &chats[i], mediaByChat[chats[i].ChatID], reactionsByChat[chats[i].ChatID])
	}

	if len(markedIDs) > 0 {
		marked := make(map[uint]bool, len(markedIDs))
		for _, id := range markedIDs {
			marked[id] = true
		}
		var results []service.ChatMarkResult
		for i := range chats {
			if marked[chats[i].ChatID] {
				results = append(results, service.ChatMarkResult{
					ChatID:   chats[i].ChatID,
					SenderID: chats[i].SenderID,
					Status:   "marked",
				})
			}
		}
		c.broadcastReadReceipts(roomID, claims.UserID, results)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": data,
		"meta": dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *ChatController) Unsend(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "chat")
	if !ok {
		return nil
	}

	chat, err := c.chatService.GetByID(ctx.Context(), id)
	if err != nil {
		return HandleServiceError(ctx, err)
	}

	if err := c.chatService.Unsend(ctx.Context(), id, claims.UserID); err != nil {
		return HandleServiceError(ctx, err)
	}

	c.broadcastRoom(chat.RoomID, "message:deleted", fiber.Map{
		"chat_id":    chat.ChatID,
		"room_id":    chat.RoomID,
		"sender_id":  chat.SenderID,
		"deleted_by": claims.UserID,
	})

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "chat unsent successfully"})
}

func (c *ChatController) EditMessage(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "chat")
	if !ok {
		return nil
	}

	var req dto.EditMessageRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	chat, err := c.chatService.GetByID(ctx.Context(), id)
	if err != nil {
		return HandleServiceError(ctx, err)
	}

	if err := c.chatService.UpdateMessage(ctx.Context(), id, claims.UserID, req.Message); err != nil {
		return HandleServiceError(ctx, err)
	}

	c.broadcastRoom(chat.RoomID, "message:edited", fiber.Map{
		"chat_id":   chat.ChatID,
		"room_id":   chat.RoomID,
		"sender_id": chat.SenderID,
		"message":   strings.TrimSpace(req.Message),
		"is_edited": true,
	})

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "chat updated successfully"})
}

func (c *ChatController) MarkMessagesRead(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	var req dto.MarkMessagesReadRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}
	if req.RoomID == 0 || len(req.ChatIDs) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	results, err := c.chatService.MarkMessagesRead(ctx.Context(), req.RoomID, req.ChatIDs, claims.UserID)
	if err != nil {
		return HandleServiceError(ctx, err)
	}
	data := make([]fiber.Map, len(results))
	for i, r := range results {
		data[i] = fiber.Map{"chat_id": r.ChatID, "status": r.Status}
	}

	c.broadcastReadReceipts(req.RoomID, claims.UserID, results)

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"results": data})
}

func (c *ChatController) AddReaction(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	chatID, ok := ParseIDParam(ctx, "id", "chat")
	if !ok {
		return nil
	}

	var req dto.AddReactionRequest
	if err := ctx.BodyParser(&req); err != nil || strings.TrimSpace(req.Emoji) == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "emoji is required"})
	}

	chat, err := c.chatReactionService.Add(ctx.Context(), chatID, claims.UserID, req.Emoji)
	if err != nil {
		return HandleServiceError(ctx, err)
	}

	c.broadcastRoom(chat.RoomID, "message:reaction:added", fiber.Map{
		"chat_id": chatID,
		"room_id": chat.RoomID,
		"user_id": claims.UserID,
		"emoji":   req.Emoji,
	})

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "reaction added successfully"})
}

func (c *ChatController) RemoveReaction(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	chatID, ok := ParseIDParam(ctx, "id", "chat")
	if !ok {
		return nil
	}

	// Fiber decodes path params, but PathUnescape handles any double-encoded edge cases
	emoji, _ := url.PathUnescape(ctx.Params("emoji"))
	if strings.TrimSpace(emoji) == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "emoji is required"})
	}

	chat, err := c.chatReactionService.Remove(ctx.Context(), chatID, claims.UserID, emoji)
	if err != nil {
		return HandleServiceError(ctx, err)
	}

	c.broadcastRoom(chat.RoomID, "message:reaction:removed", fiber.Map{
		"chat_id": chatID,
		"room_id": chat.RoomID,
		"user_id": claims.UserID,
		"emoji":   emoji,
	})

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "reaction removed successfully"})
}

func (c *ChatController) broadcastRoom(roomID uint, event string, payload any) {
	if c.hub == nil {
		return
	}
	c.hub.Broadcast(roomID, event, payload)
}

func (c *ChatController) broadcastReadReceipts(roomID, readBy uint, results []service.ChatMarkResult) {
	if c.hub == nil {
		return
	}

	chatIDsBySender := make(map[uint][]uint)
	for _, result := range results {
		if result.Status != "marked" || result.SenderID == 0 {
			continue
		}
		chatIDsBySender[result.SenderID] = append(chatIDsBySender[result.SenderID], result.ChatID)
	}

	for senderID, chatIDs := range chatIDsBySender {
		c.hub.EmitToUser(senderID, "message:read", fiber.Map{
			"room_id":  roomID,
			"chat_ids": chatIDs,
			"read_by":  readBy,
		})
	}
}

func (c *ChatController) loadMediaByChat(chats []model.Chat) map[uint][]model.ChatMedia {
	ids := make([]uint, len(chats))
	for i := range chats {
		ids[i] = chats[i].ChatID
	}
	allMedia, _ := c.chatMediaService.GetByChatIDs(ids)
	byChat := make(map[uint][]model.ChatMedia, len(ids))
	for _, m := range allMedia {
		byChat[m.ChatID] = append(byChat[m.ChatID], m)
	}
	return byChat
}

func (c *ChatController) loadReactionsByChat(chats []model.Chat) map[uint][]model.ChatReaction {
	ids := make([]uint, len(chats))
	for i := range chats {
		ids[i] = chats[i].ChatID
	}
	allReactions, _ := c.chatReactionService.GetByChatIDs(ids)
	byChat := make(map[uint][]model.ChatReaction, len(ids))
	for _, r := range allReactions {
		byChat[r.ChatID] = append(byChat[r.ChatID], r)
	}
	return byChat
}

func (c *ChatController) toChatResponse(ctx context.Context, chat *model.Chat, media []model.ChatMedia, reactions []model.ChatReaction) dto.ChatResponse {
	var deletedAt *dtoutils.FormattedTime
	if chat.DeletedAt != nil {
		wrapped := dtoutils.Wrap(*chat.DeletedAt)
		deletedAt = &wrapped
	}
	var editedAt *dtoutils.FormattedTime
	if chat.EditedAt != nil {
		wrapped := dtoutils.Wrap(*chat.EditedAt)
		editedAt = &wrapped
	}

	var replyTo *dto.ChatReplyResponse
	if chat.ReplyTo != nil {
		replyTo = &dto.ChatReplyResponse{
			ChatID: chat.ReplyTo.ChatID,
			Sender: dto.ChatSender{
				ID:       chat.ReplyTo.SenderID,
				Username: chat.ReplyTo.Sender.UserName,
			},
			Message: chat.ReplyTo.Message,
			GifURL:  chat.ReplyTo.GifURL,
		}
	}

	return dto.ChatResponse{
		ChatID: chat.ChatID,
		RoomID: chat.RoomID,
		Sender: dto.ChatSender{
			ID:       chat.SenderID,
			Username: chat.Sender.UserName,
		},
		ReplyTo:   replyTo,
		Message:   chat.Message,
		GifURL:    chat.GifURL,
		Media:     c.toChatMediaResponses(ctx, media),
		Reactions: groupReactions(reactions),
		IsRead:    chat.IsRead,
		IsEdited:  chat.IsEdited,
		EditedAt:  editedAt,
		DeletedAt: deletedAt,
		CreatedAt: dtoutils.Wrap(chat.CreatedAt),
	}
}

func (c *ChatController) toChatMediaResponses(ctx context.Context, media []model.ChatMedia) []dto.ChatMediaResponse {
	out := make([]dto.ChatMediaResponse, len(media))
	for i := range media {
		u, err := c.chatMediaService.GeneratePresignedURL(ctx, media[i].Bucket, media[i].ObjectName)
		if err != nil {
			log.Printf("presign error bucket=%s object=%s: %v", media[i].Bucket, media[i].ObjectName, err)
		}
		out[i] = dto.ChatMediaResponse{
			MediaID:   media[i].MediaID,
			URL:       u,
			MediaType: media[i].MediaType,
			FileSize:  media[i].FileSize,
		}
	}
	return out
}

func groupReactions(reactions []model.ChatReaction) []dto.ChatReactionSummary {
	if len(reactions) == 0 {
		return []dto.ChatReactionSummary{}
	}
	byEmoji := make(map[string][]uint)
	for _, r := range reactions {
		byEmoji[r.Emoji] = append(byEmoji[r.Emoji], r.UserID)
	}
	summaries := make([]dto.ChatReactionSummary, 0, len(byEmoji))
	for emoji, userIDs := range byEmoji {
		summaries = append(summaries, dto.ChatReactionSummary{
			Emoji:   emoji,
			Count:   len(userIDs),
			UserIDs: userIDs,
		})
	}
	return summaries
}
